package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"

	printer "github.com/vadimpilyugin/debug_print_go"
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

var (
	AllowListing bool = false
	AllowGet     bool = false
)

func main() {
	home := flag.String("home", config.RootDir, "Home directory")
	addr := flag.String("addr", config.ServerIp, "Server address")
	port := flag.String("port", config.ServerPort, "Server port")
	useSSL := flag.Bool("use-ssl", config.UseSSL, "Use SSL?")
	certFile := flag.String("cert", config.CertFile, "Certificate file")
	keyFile := flag.String("key", config.KeyFile, "Private key file")
	useAuth := flag.Bool("auth", config.UseAuth, "Use authentication?")
	allowListing := flag.Bool("listing", config.AllowListing, "Allow listing?")
	allowGet := flag.Bool("allow-get", config.AllowGet, "Allow GET requests?")
	redirectHTTP := flag.Bool("redirect-http", config.RedirectHTTP, "Redirect HTTP?")
	flag.Parse()

	config.AllowListing = *allowListing
	config.AllowGet = *allowGet
	config.RootDir = *home
	config.ServerPort = *port
	config.ServerIp = *addr
	config.UseAuth = *useAuth
	config.UseSSL = *useSSL
	config.RedirectHTTP = *redirectHTTP

	if *certFile != "" {
		config.CertFile = *certFile
	}

	if *keyFile != "" {
		config.KeyFile = *keyFile
	}

	AllowListing = config.AllowListing
	AllowGet = config.AllowGet

	if AllowListing {
		AllowGet = true
		if _, err := os.Stat(config.RootDir); os.IsNotExist(err) && config.RootDir != "" {
			log.Fatalf("Directory %s does not exist!\n", config.RootDir)
		}
		if config.RootDir == "" {
			home := os.Getenv("HOME")
			if home != "" {
				log.Printf("Using $HOME=%s as a root directory\n", home)
				config.RootDir = home
			}
		}
		if config.RootDir == "" {
			log.Fatal("No root directory specified!\n")
		}
	}

	printer.Debug("", config.Internal.ServerSoftware, map[string]string{
		"Port": config.Network.ServerPort,
		"IP":   config.Network.ServerIp,
	})

	if config.Openssl.UseSSL {

		// start redirector to https
		if config.Openssl.RedirectHTTP {
			go http.ListenAndServe(":"+config.Openssl.PortToRedirect, http.HandlerFunc(redirect))
			printer.Debug("", "HTTPS Redirector v1.0", map[string]string{
				"Port": config.Openssl.PortToRedirect,
			})
		}
	}

	printer.Debug("", "----------------------")

	fileHandler := &FileHandler{Root: http.Dir(config.Internal.RootDir)}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(StaticFS)))
	mux.Handle("/", fileHandler)

	srv := &http.Server{
		Addr:    config.Network.ServerIp + ":" + config.Network.ServerPort,
		Handler: mux,
	}

	if config.Openssl.UseSSL {
		srv.TLSConfig = loadTlsConfig()
		srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
		printer.Fatal(srv.ListenAndServeTLS("", ""))
	} else {
		printer.Fatal(srv.ListenAndServe())
	}
}
