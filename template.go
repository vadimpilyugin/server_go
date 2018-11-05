package main

import (
  "sort"
  "net/url"
  "io"
  "os"
  "fmt"
  "text/template"
  "strings"
  "github.com/jehiah/go-strftime"
  "time"
  "net/http"
  "github.com/vadimpilyugin/debug_print_go"
  "io/ioutil"
  "path"
)

type Directory struct {
  Name string
  Elements []Elem
  Address string
}

type Elem struct {
  Name string
  Url string
  IsDir bool
  ModifDate int64
  Size int64
  ItemRank int
  HrModifDate string
  HrSize string
  IsParent bool
  Icon string
}

const (
  noInfo = "-"
  upName = "Parent Directory"
  zeroSize = 0
  zeroDate = 0
  backUrl = "../"
  folderSize = 4096
)

const (
  parentRank = 0
  folderRank = 1
  fileRank = 2
)

func fnToIcon (s string) string {
  ext := path.Ext(s)
  icon, found := mapExtToIcon[ext]
  if found {
    return icon
  }
  return "unknown.svg"
}

var mapExtToIcon = map[string]string{".zip": "archive.svg", ".7z": "archive.svg", ".bz2": "archive.svg", ".cab": "archive.svg", ".gz": "archive.svg", ".tar": "archive.svg", ".rar": "archive.svg", ".aac": "audio.svg", ".aif": "audio.svg", ".aifc": "audio.svg", ".aiff": "audio.svg", ".ape": "audio.svg", ".au": "audio.svg", ".flac": "audio.svg", ".iff": "audio.svg", ".m4a": "audio.svg", ".mid": "audio.svg", ".mp3": "audio.svg", ".mpa": "audio.svg", ".ra": "audio.svg", ".wav": "audio.svg", ".wma": "audio.svg", ".f4a": "audio.svg", ".f4b": "audio.svg", ".oga": "audio.svg", ".ogg": "audio.svg", ".xm": "audio.svg", ".it": "audio.svg", ".s3m": "audio.svg", ".mod": "audio.svg", ".bin": "bin.svg", ".hex": "bin.svg", ".xml": "code.svg", ".doc": "doc.svg", ".docx": "doc.svg", ".docm": "doc.svg", ".dot": "doc.svg", ".dotx": "doc.svg", ".dotm": "doc.svg", ".log": "doc.svg", ".msg": "doc.svg", ".odt": "doc.svg", ".pages": "doc.svg", ".rtf": "doc.svg", ".tex": "doc.svg", ".wpd": "doc.svg", ".wps": "doc.svg", ".bmp": "img.svg", ".png": "img.svg", ".tiff": "img.svg", ".tif": "img.svg", ".gif": "img.svg", ".jpg": "img.svg", ".jpeg": "img.svg", ".jpe": "img.svg", ".psd": "img.svg", ".ai": "img.svg", ".ico": "img.svg", ".xlsx": "spreadsheet.svg", ".xlsm": "spreadsheet.svg", ".xltx": "spreadsheet.svg", ".xltm": "spreadsheet.svg", ".xlam": "spreadsheet.svg", ".xlr": "spreadsheet.svg", ".xls": "spreadsheet.svg", ".csv": "spreadsheet.svg", ".ppt": "presentation.svg", ".pptx": "presentation.svg", ".pot": "presentation.svg", ".potx": "presentation.svg", ".pptm": "presentation.svg", ".potm": "presentation.svg", ".xps": "presentation.svg", ".cpp": "c++.svg", ".c": "c.svg", ".css": "css3.svg", ".sass": "css3.svg", ".scss": "css3.svg", ".less": "css3.svg", ".ttf": "font.svg", ".TTF": "font.svg", ".woff": "font.svg", ".WOFF": "font.svg", ".woff2": "font.svg", ".WOFF2": "font.svg", ".otf": "font.svg", ".OTF": "font.svg", ".h": "h.svg", ".html": "html5.svg", ".xhtml": "html5.svg", ".shtml": "html5.svg", ".htm": "html5.svg", ".URL": "html5.svg", ".url": "html5.svg", ".nfo": "info.svg", ".info": "info.svg", ".iso": "iso.svg", ".img": "iso.svg", ".jar": "java.svg", ".java": "java.svg", ".js": "js.svg", ".json": "js.svg", ".md": "markdown.svg", ".pkg": "package.svg", ".dmg": "package.svg", ".rpm": "package.svg", ".deb": "package.svg", ".pdf": "pdf.svg", ".php": "php.svg", ".phtml": "php.svg", ".py": "py.svg", ".rb": "rb.svg", ".bat": "script.svg", ".BAT": "script.svg", ".cmd": "script.svg", ".sh": "script.svg", ".ps": "script.svg", ".exe": "script.svg", ".EXE": "script.svg", ".msi": "script.svg", ".MSI": "script.svg", ".sql": "sql.svg", ".txt": "text.svg", ".cnf": "text.svg", ".conf": "text.svg", ".map": "text.svg", ".yaml": "text.svg", ".svg": "vector.svg", ".svgz": "vector.svg", ".asf": "video.svg", ".asx": "video.svg", ".avi": "video.svg", ".flv": "video.svg", ".mkv": "video.svg", ".mov": "video.svg", ".mp4": "video.svg", ".mpg": "video.svg", ".rm": "video.svg", ".srt": "video.svg", ".swf": "video.svg", ".vob": "video.svg", ".wmv": "video.svg", ".m4v": "video.svg", ".f4v": "video.svg", ".f4p": "video.svg", ".ogv": "video.svg"}

func toDirectory(dirs []os.FileInfo, name string, cookie string) Directory {
  d := Directory{
    Name: name, 
    Address: fmt.Sprintf(
      "%s at %s Port %s",
      config.Internal.ServerSoftware,
      config.Internal.Hostname,
      config.Network.ServerPort,
    ),
  }
  if name != "/" {
    d.Elements = append(d.Elements, Elem{
      Name: upName,
      Url: backUrl,
      IsDir: true,
      ModifDate: zeroDate,
      Size: folderSize,
      ItemRank: parentRank,
      HrModifDate: "",
      HrSize: noInfo,
      IsParent: true,
      Icon: "folder-home.svg",
    })
  }
  for _,x := range dirs {
    var elem Elem
    if x.IsDir() {
      elem = Elem{
        Name: htmlReplacer.Replace(x.Name()),
        Url: urlEscape(x.Name())+"/",
        IsDir: true,
        ModifDate: x.ModTime().Unix(),
        Size: folderSize,
        ItemRank: folderRank,
        HrModifDate: hrModifDate(x.ModTime()),
        HrSize: noInfo,
        Icon: "folder.svg",
      }
    } else {
      elem_url := urlEscape(x.Name())
      if config.Auth.UseAuth {
        elem_url += fmt.Sprintf("?%s=%s",cookieName, cookie)
      }
      elem = Elem{
        Name: htmlReplacer.Replace(x.Name()),
        Url: elem_url,
        IsDir: false,
        ModifDate: x.ModTime().Unix(),
        Size: x.Size(),
        ItemRank: fileRank,
        HrModifDate: hrModifDate(x.ModTime()),
        HrSize: hrSize(x.Size()),
        Icon: fnToIcon(x.Name()),
      }
    }

    d.Elements = append(d.Elements, elem)
  }
  return d
}


func hrModifDate (modif_date time.Time) string {
  const Day = 24*3600
  const Week = 7*Day // seconds
  elapsed_time := time.Now().Sub(modif_date).Seconds()

  if (elapsed_time > Week) {
    return strftime.Format ("%a, %d %b %H:%M", modif_date);
  } else if (elapsed_time > Day) {
    return strftime.Format ("%a, %H:%M", modif_date);
  } else {  // сегодня
    return strftime.Format ("%H:%M", modif_date);
  }
}

const (
    _           = iota           // ignore first value by assigning to blank identifier
    kb = 1 << (10 * iota)        // 1 << (10*1)
    mb                           // 1 << (10*2)
    gb                           // 1 << (10*3)
    tb                           // 1 << (10*4)
    pb                           // 1 << (10*5)
)
func hrSize (size int64) string {
  if size > 1*gb {
    return fmt.Sprintf("%.2f ГБ",float64(size)/gb)
  } else if size > 10*mb {
    return fmt.Sprintf("%d МБ",size/mb)
  } else if size > 1*mb {
    return fmt.Sprintf("%.1f МБ",float64(size)/mb)
  } else if size > 1*kb {
    return fmt.Sprintf("%d КБ",size/kb)
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

func greater (b1, b2 bool) bool {
  if b1 && !b2 {
    return true
  } else {
    return false
  }
}

func dirList(w io.Writer, f http.File, name string, cookie string) error {
  dirs, err := f.Readdir(-1)
  if err != nil {
    return err
  }
  b := dirs[:0]
  for _, x := range dirs {
    if !strings.HasPrefix(x.Name(), ".") || !x.IsDir() {
      b = append(b, x)
    }
  }
  dirs = b

  sort.Slice(dirs, func(i, j int) bool { 
    return dirs[j].ModTime().Unix() < dirs[i].ModTime().Unix() 
  })
  sort.SliceStable(dirs, func(i, j int) bool { 
    return greater(dirs[i].IsDir(),dirs[j].IsDir()) 
  })

  t, err := template.ParseFiles(config.Static.DirlistTempl)
  if err != nil {
    printer.Fatal(err)
  }
  p := toDirectory(dirs,name,cookie)
  err = t.Execute(w, p)
  if err != nil {
    return err
  }
  return nil
}

func generateAuthPage(w io.Writer) {
  data, err := ioutil.ReadFile(config.Static.AuthTempl)
  if err != nil {
    printer.Fatal(err)
  }
  fmt.Fprintf(w, "%s", data)
}