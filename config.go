package main

import (
	"log"
	"os"
	_ "embed"

	"github.com/go-ini/ini"
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
	AllowPost    bool   `ini:"allow_post"`
}

type Config struct {
	Internal `ini:"internal"`
	Network  `ini:"network"`
	Openssl  `ini:"openssl"`
	Auth     `ini:"auth"`
}

var config *Config

//go:embed config/config.ini
var defaultConfig []byte

func loadConfig(fn string) error {
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		log.Printf("Config file was not specified, using default config\n")
		config = &Config{}
		err := ini.MapTo(config, defaultConfig)
		if err != nil {
			log.Printf("Default config file parse error: %v\n", err)
			return err
		}
	} else {
		config = new(Config)
		err := ini.MapTo(config, fn)
		if err != nil {
			log.Printf("config file load failed: %v\n", err)
			return err
		}
	}
	return nil
}
