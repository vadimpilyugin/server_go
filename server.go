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

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	configFile := flag.String("config", "", "Config file location")
	home := flag.String("home", "", "Home directory")
	addr := flag.String("addr", "", "Server address")
	port := flag.String("port", "8080", "Server port")
	useSSL := flag.Bool("use-ssl", false, "Use SSL?")
	certFile := flag.String("cert", "", "Certificate file")
	keyFile := flag.String("key", "", "Private key file")
	useAuth := flag.Bool("auth", false, "Use authentication?")
	allowListing := flag.Bool("listing", false, "Allow listing?")
	allowGet := flag.Bool("allow-get", false, "Allow GET requests?")
	allowPost := flag.Bool("allow-post", false, "Allow POST requests?")
	redirectHTTP := flag.Bool("redirect-http", false, "Redirect HTTP?")
	redirectPort := flag.String("redirect-port", "80", "Port to redirect")
	flag.Parse()

	err := loadConfig(*configFile)
	if err != nil {
		printer.Fatal(err)
	}

	if isFlagPassed("home") {
		config.RootDir = *home
	}
	if isFlagPassed("addr") {
		config.ServerIp = *addr
	}
	if isFlagPassed("port") {
		config.ServerPort = *port
	}
	if isFlagPassed("use-ssl") {
		config.UseSSL = *useSSL
	}
	if isFlagPassed("cert") {
		config.CertFile = *certFile
	}
	if isFlagPassed("key") {
		config.KeyFile = *keyFile
	}
	if isFlagPassed("auth") {
		config.UseAuth = *useAuth
	}
	if isFlagPassed("listing") {
		config.AllowListing = *allowListing
	}
	if isFlagPassed("allow-get") {
		config.AllowGet = *allowGet
	}
	if isFlagPassed("allow-post") {
		config.AllowPost = *allowPost
	}
	if isFlagPassed("redirect-http") {
		config.RedirectHTTP = *redirectHTTP
	}
	if isFlagPassed("redirect-port") {
		config.PortToRedirect = *redirectPort
	}

	if config.CertFile != "" {
		config.UseSSL = true
	}

	if config.KeyFile != "" {
		config.UseSSL = true
	}

	if config.UseSSL && (config.KeyFile == "" || config.CertFile == "") {
		printer.Debug(config.CertFile, "cert file")
		printer.Debug(config.KeyFile, "key file")
		printer.Fatal("UseSSL is true, but cert file or key file is empty")
	}

	if config.AllowListing {
		config.AllowGet = true
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
			printer.Fatal("No root directory specified!")
		}
	}

	log.Printf("Config file: %+v\n", config)

	printer.Debug("", config.Internal.ServerSoftware, map[string]string{
		"Port": config.Network.ServerPort,
		"IP":   config.Network.ServerIp,
	})

	if config.Openssl.UseSSL {

		// start redirector to https
		if config.Openssl.RedirectHTTP {
			printer.Debug("", "HTTPS Redirector v1.0", map[string]string{
				"Port": config.Openssl.PortToRedirect,
			})
			go func() {
				err := http.ListenAndServe(":"+config.Openssl.PortToRedirect, http.HandlerFunc(redirect))
				if err != nil {
					printer.Fatal(err, "HTTP Redirector failed")
				}
			}()
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
