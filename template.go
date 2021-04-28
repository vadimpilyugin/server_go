package main

import (
	"bufio"
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/jehiah/go-strftime"
	printer "github.com/vadimpilyugin/debug_print_go"
)

type Directory struct {
	Name     string
	Elements []Elem
	Address  string
}

type Elem struct {
	Name        string
	Url         string
	IsDir       bool
	ModifDate   int64
	Size        int64
	ItemRank    int
	HrModifDate string
	HrSize      string
	IsParent    bool
	Icon        string
	IsViewable  bool
	IsEditable  bool
	ListNo      int
}

const (
	noInfo     = "-"
	upName     = "Parent Directory"
	zeroSize   = 0
	zeroDate   = 0
	backUrl    = "../"
	folderSize = 4096
)

const (
	parentRank = 0
	folderRank = 1
	fileRank   = 2
)

func init() {
	readMimeMap()
}

var mapMimeToIcon map[string]string
var lastModTime time.Time = time.Unix(0, 0)

//go:embed mimeToIcons.txt
var mimeData []byte

func readMimeMap() {
	mapMimeToIcon := make(map[string]string)

	scanner := bufio.NewScanner(bytes.NewBuffer(mimeData))
	for scanner.Scan() {
		line := scanner.Text()
		ar := strings.Split(line, " ")
		icon := ar[0]
		for _, v := range ar[1:] {
			mapMimeToIcon[v] = icon
		}
	}

	if err := scanner.Err(); err != nil {
		printer.Fatal(err)
	}
}

func fnToIcon(info os.FileInfo, name string) string {
	const unknown = "unknown.svg"

	s := info.Name()
	ext := path.Ext(s)
	icon, found := mapMimeToIcon[ext]
	if found {
		return icon
	}
	ext = strings.ToLower(ext)
	icon, found = mapMimeToIcon[ext]
	if found {
		return icon
	}
	// open file, read 512 bytes, determine MIME type
	f, err := os.Open(path.Join(config.Internal.RootDir, "."+name, info.Name()))
	if err != nil {
		printer.Error(err, path.Join(name, info.Name()))
		return unknown
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF || (err == io.EOF && n != 0) {
		printer.Error(err, path.Join(name, info.Name()))
		return unknown
	}
	ctype := http.DetectContentType(buf)
	mediatype, _, _ := mime.ParseMediaType(ctype)
	icon, found = mapMimeToIcon[mediatype]
	if found {
		return icon
	}
	if mediatype != "application/octet-stream" || ext != "" {
		printer.Note(mediatype, info.Name())
	} else {
		return unknown
	}
	parts := strings.Split(mediatype, "/")
	icon, found = mapMimeToIcon[parts[0]]
	if found {
		return icon
	}
	return unknown
}

func isViewable(ext string) bool {
	return ext == ".webm" || ext == ".mp4" || ext == ".mkv" || ext == ".ogg"
}

func isEditable(ext string, icon string) bool {
	return ext == ".txt" || ext == ".conf" || ext == ".json" ||
		ext == ".yaml" || ext == ".yml" || icon == "text.svg"
}

func toDirectory(dirs []os.FileInfo, name string, cookie string) Directory {
	readMimeMap()

	d := Directory{
		Name: name,
		Address: fmt.Sprintf(
			"%s at %s Port %s",
			config.Internal.ServerSoftware,
			config.Internal.Hostname,
			config.Network.ServerPort,
		),
	}

	listNo := 0

	if name != "/" {
		d.Elements = append(d.Elements, Elem{
			Name:        upName,
			Url:         backUrl,
			IsDir:       true,
			ModifDate:   zeroDate,
			Size:        folderSize,
			ItemRank:    parentRank,
			HrModifDate: "",
			HrSize:      noInfo,
			IsParent:    true,
			Icon:        "folder-home.svg",
			ListNo:      listNo,
		})
		listNo++
	}
	for _, x := range dirs {
		var elem Elem
		if x.IsDir() {
			elem = Elem{
				Name:        htmlReplacer.Replace(x.Name()),
				Url:         urlEscape(x.Name()) + "/",
				IsDir:       true,
				ModifDate:   x.ModTime().Unix(),
				Size:        folderSize,
				ItemRank:    folderRank,
				HrModifDate: hrModifDate(x.ModTime()),
				HrSize:      noInfo,
				Icon:        "folder.svg",
				ListNo:      listNo,
			}
		} else {
			elem_url := urlEscape(x.Name())
			if config.Auth.UseAuth {
				elem_url += fmt.Sprintf("?%s=%s", cookieName, cookie)
			}
			elem = Elem{
				Name:        htmlReplacer.Replace(x.Name()),
				Url:         elem_url,
				IsDir:       false,
				ModifDate:   x.ModTime().Unix(),
				Size:        x.Size(),
				ItemRank:    fileRank,
				HrModifDate: hrModifDate(x.ModTime()),
				HrSize:      hrSize(x.Size()),
				Icon:        fnToIcon(x, name),
				IsViewable:  isViewable(path.Ext(x.Name())),
				IsEditable:  isEditable(path.Ext(x.Name()), fnToIcon(x, name)),
				ListNo:      listNo,
			}
		}

		d.Elements = append(d.Elements, elem)
		listNo++
	}
	return d
}

func hrModifDate(modDate time.Time) string {
	now := time.Now()
	yesterday := now.Hour()*3600 + now.Minute()*60 + now.Second()
	const Week = 7 * 24 * 3600 // seconds
	timeElapsed := now.Sub(modDate).Seconds()

	if timeElapsed > Week {
		return strftime.Format("%a, %d %b %H:%M", modDate)
	} else if timeElapsed > float64(yesterday) {
		return strftime.Format("%a, %H:%M", modDate)
	} else { // сегодня
		return strftime.Format("%H:%M", modDate)
	}
}

const (
	_  = iota             // ignore first value by assigning to blank identifier
	kb = 1 << (10 * iota) // 1 << (10*1)
	mb                    // 1 << (10*2)
	gb                    // 1 << (10*3)
	tb                    // 1 << (10*4)
	pb                    // 1 << (10*5)
)

func hrSize(size int64) string {
	if size > 1*gb {
		return fmt.Sprintf("%.2f ГБ", float64(size)/gb)
	} else if size > 10*mb {
		return fmt.Sprintf("%d МБ", size/mb)
	} else if size > 1*mb {
		return fmt.Sprintf("%.1f МБ", float64(size)/mb)
	} else if size > 1*kb {
		return fmt.Sprintf("%d КБ", size/kb)
	} else {
		return fmt.Sprintf("%d Б", size)
	}
}

func urlEscape(name string) string {
	u := url.URL{Path: name}
	return u.String()
}

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&#34;",
	"'", "&#39;",
)

func greater(b1, b2 bool) bool {
	if b1 && !b2 {
		return true
	} else {
		return false
	}
}

func trueModTime(name string) time.Time {
	maxModTime := serverStartTime
	var files = []string{
		path.Join(config.Internal.RootDir, "."+name),
	}

	for _, k := range files {
		tmpStat, err := os.Stat(k)
		if err != nil {
			printer.Fatal(err, "true mod name")
		}
		if tmpStat.ModTime().Unix() > maxModTime.Unix() {
			maxModTime = tmpStat.ModTime()
		}
	}
	return maxModTime
}

//go:embed templates/templ.html
var dirlistTempl embed.FS

func dirList(w io.Writer, f http.File, name string, cookie string) (time.Time, error) {
	dirs, err := f.Readdir(-1)
	if err != nil {
		return serverStartTime, err
	}
	b := dirs[:0]
	for _, x := range dirs {
		if !strings.HasPrefix(x.Name(), ".") || !x.IsDir() {
			b = append(b, x)
		}
	}
	dirs = b

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[j].ModTime().UnixNano() < dirs[i].ModTime().UnixNano()
	})
	sort.SliceStable(dirs, func(i, j int) bool {
		return greater(dirs[i].IsDir(), dirs[j].IsDir())
	})

	t, err := template.ParseFS(dirlistTempl, "templates/templ.html")
	if err != nil {
		log.Fatal("Cannot parse dilist template:", err)
	}
	p := toDirectory(dirs, name, cookie)
	err = t.Execute(w, p)
	if err != nil {
		return serverStartTime, err
	}
	return trueModTime(name), nil
}

//go:embed templates/auth.html
var authTempl embed.FS

func generateAuthPage(w io.Writer) {
	data, err := authTempl.ReadFile("templates/auth.html")
	if err != nil {
		printer.Fatal(err, "auth template")
	}
	fmt.Fprintf(w, "%s", data)
}
