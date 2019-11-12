package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	printer "github.com/vadimpilyugin/debug_print_go"
)

const (
	PERM_ALL   = 0644
	MODE_WRITE = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
)

type FnRepeatSt struct {
	RepeatNo int
	LastCheck int64
}

var fnRepeat = map[string]*FnRepeatSt{}

var serverStartTime time.Time = time.Now()

const fileField = "file"

func copyNo(init string) (string, int) {
	ext := filepath.Ext(init)
	init = init[:len(init)-len(ext)]
	ar := regexp.MustCompile(`(.*)\((\d+)\)$`).FindStringSubmatch(init)
	if ar == nil {
		return init, 0
	}
	num, err := strconv.Atoi(ar[2])
	if err != nil {
		fmt.Printf("Could not convert %s to num: %v\n", ar[2], err)
		return init, 0
	}
	return ar[1], num
}

func saveAs(init string, dir string) string {
	filePath := path.Join(dir,init)
	if _,found := fnRepeat[filePath]; found {
		if time.Now().Unix() - fnRepeat[filePath].LastCheck <= 600 {
			no := fnRepeat[filePath].RepeatNo
			prefix, _ := copyNo(init)
			ext := filepath.Ext(init)
			init = fmt.Sprintf("%s(%d)%s", prefix, no+1, ext)
		} else {
			delete(fnRepeat, init)
		}
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, init)); os.IsNotExist(err) {
			fmt.Printf("Checking filename '%s': ", init)
			fmt.Printf("Free!\n")
			break
		} else {
			prefix, no := copyNo(init)
			ext := filepath.Ext(init)
			init = fmt.Sprintf("%s(%d)%s", prefix, no+1, ext)
			fnRepeat[path.Join(dir,prefix+ext)] = &FnRepeatSt{no+1,time.Now().Unix()}
		}
	}
	return init
}

// name is '/'-separated, not filepath.Separator.
func serveFile(w http.ResponseWriter, r *http.Request, fs http.Dir, name string) {

	if r.Method == http.MethodGet && !AllowGet {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	for x, path := range Resources {
		if x == name {
			printer.Note(name, "Static file!")
			http.ServeFile(w, r, path)
			return
		}
	}

	if _, found := r.Header["X-Codemirror"]; found {
		w.Header().Set("Cache-Control", "no-store")
	}

	var cookie string
	var err error
	if config.Auth.UseAuth {
		cookie, err = checkCookie(w, r)
		if err != nil {
			return
		}
	}

	f, err := fs.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	// redirect to canonical path: / at end of directory url
	// r.URL.Path always begins with /
	url := r.URL.Path
	if d.IsDir() {
		if url[len(url)-1] != '/' {
			localRedirect(w, r, path.Base(url)+"/")
			return
		}
	} else {
		if url[len(url)-1] == '/' {
			printer.Note("Путь к файлу заканчивается на /")
			localRedirect(w, r, "../"+path.Base(url))
			return
		}
	}

	if d.IsDir() {
		if r.Method == "POST" {
			postStarted := time.Now()
			err := r.ParseMultipartForm(10 * mb)
			if err != nil {
				printer.Error(err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}
			for v := range r.MultipartForm.File[fileField] {
				fileHeader := r.MultipartForm.File[fileField][v]
				fn := fileHeader.Filename

				dir := filepath.Join(string(fs), name)
				duration := time.Since(postStarted)

				fn = saveAs(fn, dir)
				filePath := filepath.Join(dir, fn)

				speed := int64(float64(fileHeader.Size) / duration.Seconds())
				printer.Debug("", "File Upload", map[string]string{
					"Filename":      fn,
					"Absolute path": filePath,
					"File size":     hrSize(fileHeader.Size),
					"Duration":      fmt.Sprintf("%v", duration),
					"Speed":         hrSize(speed) + "/с",
				})

				copyTo, err := os.OpenFile(filePath, MODE_WRITE, PERM_ALL)
				if err != nil {
					printer.Fatal(err)
				}
				copyFrom, err := fileHeader.Open()
				if err != nil {
					printer.Fatal(err)
				}
				io.Copy(copyTo, copyFrom)
			}
			err = r.MultipartForm.RemoveAll()
			if err != nil {
				printer.Error(err)
			}
			localRedirect(w, r, "./")
			return
		} else if AllowListing {
			buf := new(bytes.Buffer)
			maxModtime, err := dirList(buf, f, name, cookie)
			if err != nil {
				printer.Error(err)
				http.Error(w, "Error reading directory", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(w, r, d.Name(), maxModtime, bytes.NewReader(buf.Bytes()))
			return
		}

		msg, code := "empty", http.StatusOK
		http.Error(w, msg, code)

	} else {
		if r.Method == "DELETE" {
			err := os.Remove(path.Clean(string(fs) + name))
			if err != nil {
				printer.Error(err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	}
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	if r.Method == "POST" {
		w.WriteHeader(http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusMovedPermanently)
	}
}

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value. It's important that toHTTPError does not
// actually return err.Error(), since msg and httpStatus are returned to users,
// and historically Go's ServeContent always returned just "404 Not Found" for
// all errors. We don't want to start leaking information in error messages.
func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

var headerDumpMutex sync.Mutex

func dumpRequest(req *http.Request) {
	// GET / HTTP/1.1
	req_s := fmt.Sprintf(
		"%s HTTP/%d.%d",
		req.RequestURI,
		req.ProtoMajor, req.ProtoMinor,
	)
	// [Mar 9 12:01:02]
	t := time.Now()
	date := fmt.Sprintf(
		"[%s %02d %02d:%02d:%02d]",
		t.Month().String(), t.Day(), t.Hour(), t.Minute(), t.Second(),
	)
	err := req.ParseForm()

	headerDumpMutex.Lock()
	printer.Debug("", "")
	printer.Debug("", date+" Запрос от "+req.RemoteAddr)
	printer.Debug(req_s, req.Method)
	if err != nil {
		printer.Error(err)
	} else {
		for k, vs := range req.Form {
			printer.Debug(strings.Join(vs, ","), k)
		}
	}
	if len(req.TransferEncoding) > 0 {
		printer.Debug(strings.Join(req.TransferEncoding, ","), "\tTransfer-Encoding")
	}
	if req.Close {
		printer.Debug("close", "\tConnection")
	}
	var excludeHeaders = map[string]bool{
		"Transfer-Encoding": true,
		"Trailer":           true,
	}
	for k, vs := range req.Header {
		if !excludeHeaders[k] {
			for i := range vs {
				printer.Debug(fmt.Sprintf("\t%s: %s", k, vs[i]), "")
			}
		}
	}
	headerDumpMutex.Unlock()
}

type FileHandler struct {
	Root http.Dir
}

func (f *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dumpRequest(r)

	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}

	serveFile(w, r, f.Root, path.Clean(upath))
}
