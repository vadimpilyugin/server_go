package main

import (
	"bytes"
	"fmt"
	"github.com/vadimpilyugin/debug_print_go"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var serverStartTime time.Time = time.Now()

const fileField = "file"

// name is '/'-separated, not filepath.Separator.
func serveFile(w http.ResponseWriter, r *http.Request, fs http.Dir, name string) {

	for x, path := range Resources {
		if x == name {
			printer.Note(name, "Static file!")
			http.ServeFile(w, r, path)
			return
		}
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
			printer.Note("Путь к файлу заканчивается на /", "Странность")
			localRedirect(w, r, "../"+path.Base(url))
			return
		}
	}

	if d.IsDir() {
		if r.Method == "POST" {
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
				path := path.Clean(string(fs) + name + "/" + fn)

				for {
					file, err := os.Open(path)
					exists := err == nil
					s := fmt.Sprintf(
						"--- File name: %s\n--- Absolute path: %s\n--- File size: %s\n--- Exists? %v\n",
						fn, path,
						hrSize(fileHeader.Size),
						exists,
					)
					printer.Debug(s, "File Upload")
					if exists {
						file.Close()
						path = path + "(1)"
					} else {
						break
					}
				}

				copyTo, err := os.Create(path)
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
			// w.WriteHeader(http.StatusOK)
			return
		}

		buf := new(bytes.Buffer)
		maxModtime, err := dirList(buf, f, name, cookie)
		if err != nil {
			printer.Error(err)
			http.Error(w, "Error reading directory", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, d.Name(), maxModtime, bytes.NewReader(buf.Bytes()))
	} else {
		if r.Method == "DELETE" {
			err := os.Remove(path.Clean(string(fs) + name))
			if err != nil {
				printer.Error(err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}
			// localRedirect(w, r, "../")
			w.WriteHeader(http.StatusOK)
		} else {
			http.ServeContent(w, r, d.Name(), d.ModTime(), f)
		}
	}
	return
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
