package main

import (
	"log"
	"os"

	"github.com/go-ini/ini"
	printer "github.com/vadimpilyugin/debug_print_go"
)

type Internal struct {
	RootDir        string `ini:"root_dir"`
	ServerSoftware string `ini:"server_software"`
	Hostname       string `ini:"hostname"`
	DoOverwrite    bool   `ini:"overwrite"`
}

type Network struct {
	ServerIp   string `ini:"server_ip"`
	ServerPort string `ini:"server_port"`
}

type Openssl struct {
	UseSSL         bool   `ini:"use_ssl"`
	RedirectHTTP   bool   `ini:"redirect_http"`
	PortToRedirect string `ini:"port_to_redirect"`
	CertFile       string `ini:"cert_file"`
	KeyFile        string `ini:"key_file"`
}

type Auth struct {
	UseAuth      bool   `ini:"use_auth"`
	Username     string `ini:"username"`
	Password     string `ini:"password"`
	AllowListing bool   `ini:"allow_listing"`
	AllowGet     bool   `ini:"allow_get"`
}

type Static struct {
	DirlistTempl string `ini:"dirlist_template"`
	AuthTempl    string `ini:"auth_template"`
	MimeMap      string `ini:"mime_map"`
}

type Config struct {
	Internal `ini:"internal"`
	Network  `ini:"network"`
	Openssl  `ini:"openssl"`
	Auth     `ini:"auth"`
	Static   `ini:"static"`
}

var config *Config

func init() {
	fn := "config.ini"
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		log.Printf("Config file was not found, not using it\n")
		config = &Config{
			Internal: Internal{
				ServerSoftware: "Server Go",
			},
			Network: Network{
				ServerIp:   "0.0.0.0",
				ServerPort: "8080",
			},
		}
	} else {
		config = new(Config)
		err := ini.MapTo(config, fn)
		if err != nil {
			printer.Fatal(err, "config file loader")
		}
	}
}
