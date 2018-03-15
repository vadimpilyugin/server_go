package main

import (
    "crypto/tls"
    "net/http"
    "printer"
)

func redirect(w http.ResponseWriter, req *http.Request) {
  // remove/add not default ports from req.Host
  target := "https://" + req.Host + req.URL.Path 
  if len(req.URL.RawQuery) > 0 {
      target += "?" + req.URL.RawQuery
  }
  printer.Debug("Redirected http request: "+req.RequestURI, "http ~~> https")
  http.Redirect(w, req, target, http.StatusFound)
}

func main() {
    config = getConfig("config.ini")

    printer.Debug("",config.Internal.ServerSoftware,map[string]string{
      "Port":config.Network.ServerPort,
      "IP":config.Network.ServerIp,
    })

    if config.Openssl.UseSSL {

      // start redirector to https
      if config.Openssl.RedirectHTTP {
        go http.ListenAndServe(":"+config.Openssl.PortToRedirect, http.HandlerFunc(redirect))
        printer.Debug("","HTTPS Redirector v1.0", map[string]string{
          "Port":config.Openssl.PortToRedirect,
        })
      }
    }

    printer.Debug("","----------------------")

    fileHandler := &FileHandler{Root:http.Dir(config.Internal.RootDir)}

    srv := &http.Server{
      Addr: config.Network.ServerIp+":"+config.Network.ServerPort,
      Handler: fileHandler,
    }
    if config.Openssl.UseSSL {
      srv.TLSConfig = loadTlsConfig()
      srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
    }

    printer.Fatal(srv.ListenAndServeTLS("",""))
}