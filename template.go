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
  ItemPic string
  HrModifDate string
  HrSize string
}

const (
  noInfo = "-"
  upName = "Back"
  zeroSize = 0
  zeroDate = 0
  backPic = "back"
  folderPic = "folder"
  filePic = "file"
  backUrl = "../"
  folderSize = 4096
)

const (
  parentRank = 0
  folderRank = 1
  fileRank = 2
)

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
      ItemPic: backPic,
      HrModifDate: noInfo,
      HrSize: noInfo,
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
        ItemPic: folderPic,
        HrModifDate: hrModifDate(x.ModTime()),
        HrSize: noInfo,
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
        ItemPic: filePic,
        HrModifDate: hrModifDate(x.ModTime()),
        HrSize: hrSize(x.Size()),
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

  t := template.New("mainTemplate")
  t, err = t.Parse(mainTemplate)
  p := toDirectory(dirs,name,cookie)
  err = t.Execute(w, p)
  if err != nil {
    return err
  }
  return nil
}

func generateAuthPage(w io.Writer) {
  fmt.Fprintf(w, "%s", authPageTemplate)
}

const mainTemplate = `
<!DOCTYPE html5>
<html>
  <head>
    <meta charset="utf-8">
    <title> Index of {{.Name}} </title>
        <!-- back_base64 -->
    <link rel="stylesheet" type="text/css" href="/__back_base64__">
        <!-- folder_base64 -->
    <link rel="stylesheet" type="text/css" href="/__folder_base64__">
        <!-- file_base64 -->
    <link rel="stylesheet" type="text/css" href="/__file_base64__">
        <!-- favicon_base64 -->
    <link rel='shortcut icon' type='image/png' href='data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAARdQTFRFAAAANJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjb////1Lch+gAAAFt0Uk5TAAABDRcbGhQIAx9vrLm4mUwQJZDf+/PCZg963bcwAtLXWQv6/umTPC0uKQX52bCoqaSUYvzcccdG2nnhkuKW+P3gyfHYbrU5IobR9OtfEhlYnLuzhz8KHB0TBmd1o1EAAAABYktHRFzq2ACXAAAACXBIWXMAAABIAAAASABGyWs+AAAAyklEQVQ4y2NgoBgwMjGzsIIBGzsHIyOmAk4ubh5eMODjFxBkwlQhJCwiGg0BYuISkphmSElHw4GojKwchgJ5MYSCaAVFJSYkwAgyT1kFSYGqmrqGJhxosWgDVUQjAx1dPX0DODA0MmZhQlUQrWKCAkzN2NEUoANzCwIKLK0IKLC2IaDA1m5UAf0U2DvgU+AozODkjE+BiysDtwIeeWU3dwYPTy9vHNIqPry+fgyc/gGBvEFYQXBIqCQTMG+GsYRHYAWskVHYMisaAAAtwAfZwspY/gAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAxOC0wMy0xNFQxNDowNjo0NCswMDowMBNWqewAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMTgtMDMtMTRUMTQ6MDY6NDQrMDA6MDBiCxFQAAAARnRFWHRzb2Z0d2FyZQBJbWFnZU1hZ2ljayA2LjcuOC05IDIwMTQtMDUtMTIgUTE2IGh0dHA6Ly93d3cuaW1hZ2VtYWdpY2sub3Jn3IbtAAAAABh0RVh0VGh1bWI6OkRvY3VtZW50OjpQYWdlcwAxp/+7LwAAABh0RVh0VGh1bWI6OkltYWdlOjpoZWlnaHQAMTkyDwByhQAAABd0RVh0VGh1bWI6OkltYWdlOjpXaWR0aAAxOTLTrCEIAAAAGXRFWHRUaHVtYjo6TWltZXR5cGUAaW1hZ2UvcG5nP7JWTgAAABd0RVh0VGh1bWI6Ok1UaW1lADE1MjEwMzY0MDS9fwamAAAAD3RFWHRUaHVtYjo6U2l6ZQAwQkKUoj7sAAAAVnRFWHRUaHVtYjo6VVJJAGZpbGU6Ly8vbW50bG9nL2Zhdmljb25zLzIwMTgtMDMtMTQvNmIwY2U4ZDY3MjA1MDY0MmZmYTZmOTk1YTU3YzYyNzkuaWNvLnBuZzEJB6oAAAAASUVORK5CYII='/>
        <!-- bootstrap_min_css -->
    <link rel="stylesheet" type="text/css" href="/__bootstrap_min_css__">
        <!-- bootstrap_sortable_css -->
    <link rel="stylesheet" type="text/css" href="/__bootstrap_sortable_css__">
        <!-- jquery_min_js -->
    <script type="text/javascript" src="/__jquery_min_js__"></script>
        <!-- bootstrap_min_js -->
    <script type="text/javascript" src="/__bootstrap_min_js__"></script>
        <!-- bootstrap_sortable_js -->
    <script type="text/javascript" src="/__bootstrap_sortable_js__"></script>
        <!-- moment_min_js -->
    <script type="text/javascript" src="/__moment_min_js__"></script>
        <!-- myscript_js -->
    <script type="text/javascript" src="/__myscript_js__"></script>
    <style>
      ul {
        list-style-type: none;
      }
    </style>
  </head>
  <body>
    <div class='container'>
        <div class='row'>
            <div class='col-md-1'></div>
            <div class='col-md-10'>
                <h3> Index of {{.Name}} </h3>
                <table class='table table-hover sortable'>
                    <thead>
                        <tr>
                            <th class='col-xs-1'></th>
                            <th class='col-xs-1'>Имя</th>
                            <th class='col-xs-1'>Изменено</th>
                            <th class='col-xs-1'>Размер</th>
                            <th class='col-xs-1'></th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Elements}}
                        <tr class='clickable-row' data-href='{{.Url}}'>
                            <td class='col-xs-1' data-value='{{.ItemRank}}'>
                                <div class='{{.ItemPic}}'></div>
                            </td>
                            <td class='col-xs-6' data-value='{{.Name}}'>
                                <a href='#' class="elem-href">{{.Name}}</a> 
                            </td>
                            <td class='col-xs-2' data-value='{{.ModifDate}}'>{{.HrModifDate}}</td>
                            <td class='col-xs-1' data-value='{{.Size}}'>{{.HrSize}}</td>
                            
                            <td class='col-xs-1'>
                              {{if not .IsDir}}
                              <a href="#" style="color:red" data-fn="{{.Url}}" class="delete-href" data-count="0">Удалить</a>
                              {{end}}
                            </td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
                <div class="upload">
                  <h3> Добавить </h3>
                  <form name="new_file" enctype="multipart/form-data" method="POST" action="./">
                    <input name="file" type="file" multiple></input>
                    <br>
                    <input type="submit" value="Закачать">
                  </form>
                </div>
                <address style='font-style:italic'>{{.Address}}</address>
            </div>
            <div class='col-md-1'></div>
        </div>
    </div>
  </body>
</html>
`

const authPageTemplate = `<!DOCTYPE html5>
<html>
  <head>
    <meta charset="utf-8">
        <!-- favicon_base64 -->
    <link rel='shortcut icon' type='image/png' href='data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAARdQTFRFAAAANJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjbNJjb////1Lch+gAAAFt0Uk5TAAABDRcbGhQIAx9vrLm4mUwQJZDf+/PCZg963bcwAtLXWQv6/umTPC0uKQX52bCoqaSUYvzcccdG2nnhkuKW+P3gyfHYbrU5IobR9OtfEhlYnLuzhz8KHB0TBmd1o1EAAAABYktHRFzq2ACXAAAACXBIWXMAAABIAAAASABGyWs+AAAAyklEQVQ4y2NgoBgwMjGzsIIBGzsHIyOmAk4ubh5eMODjFxBkwlQhJCwiGg0BYuISkphmSElHw4GojKwchgJ5MYSCaAVFJSYkwAgyT1kFSYGqmrqGJhxosWgDVUQjAx1dPX0DODA0MmZhQlUQrWKCAkzN2NEUoANzCwIKLK0IKLC2IaDA1m5UAf0U2DvgU+AozODkjE+BiysDtwIeeWU3dwYPTy9vHNIqPry+fgyc/gGBvEFYQXBIqCQTMG+GsYRHYAWskVHYMisaAAAtwAfZwspY/gAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAxOC0wMy0xNFQxNDowNjo0NCswMDowMBNWqewAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMTgtMDMtMTRUMTQ6MDY6NDQrMDA6MDBiCxFQAAAARnRFWHRzb2Z0d2FyZQBJbWFnZU1hZ2ljayA2LjcuOC05IDIwMTQtMDUtMTIgUTE2IGh0dHA6Ly93d3cuaW1hZ2VtYWdpY2sub3Jn3IbtAAAAABh0RVh0VGh1bWI6OkRvY3VtZW50OjpQYWdlcwAxp/+7LwAAABh0RVh0VGh1bWI6OkltYWdlOjpoZWlnaHQAMTkyDwByhQAAABd0RVh0VGh1bWI6OkltYWdlOjpXaWR0aAAxOTLTrCEIAAAAGXRFWHRUaHVtYjo6TWltZXR5cGUAaW1hZ2UvcG5nP7JWTgAAABd0RVh0VGh1bWI6Ok1UaW1lADE1MjEwMzY0MDS9fwamAAAAD3RFWHRUaHVtYjo6U2l6ZQAwQkKUoj7sAAAAVnRFWHRUaHVtYjo6VVJJAGZpbGU6Ly8vbW50bG9nL2Zhdmljb25zLzIwMTgtMDMtMTQvNmIwY2U4ZDY3MjA1MDY0MmZmYTZmOTk1YTU3YzYyNzkuaWNvLnBuZzEJB6oAAAAASUVORK5CYII='/>
        <!-- bootstrap_min_css -->
    <link rel="stylesheet" type="text/css" href="/__bootstrap_min_css__">
  </head>
  <body>
    <div class="container">
        <div class="row">
            <div class="col-md-2">
                <form class="form-horizontal" action='/__auth__' method="POST">
                  <fieldset>
                    <div id="legend">
                      <legend class="">Login</legend>
                    </div>
                    <div class="form-group">
                      <!-- Username -->
                      <label for="username">Username</label>
                      <div class="controls">
                        <input type="text" id="username" name="username" placeholder="" class="form-control input-xlarge ">
                      </div>
                    </div>
                    <div class="form-group">
                      <!-- Password-->
                      <label for="password">Password</label>
                      <div class="controls">
                        <input type="password" id="password" name="password" placeholder="" class="form-control input-xlarge">
                      </div>
                    </div>
                    <div class="form-group">
                      <!-- Button -->
                      <div class="controls">
                        <button class="btn btn-success">Login</button>
                      </div>
                    </div>
                  </fieldset>
                </form>
            </div>
        </div>
    </div>
  </body>
</html>
`
