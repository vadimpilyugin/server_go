package main

import (
  "gopkg.in/ini.v1"
  "printer"
)

type Internal struct {
  RootDir string    `ini:"root_dir"`
  ServerSoftware string   `ini:"server_software"`
  Hostname string   `ini:"hostname"`
}

type Network struct {
  ServerIp string     `ini:"server_ip"`
  ServerPort string   `ini:"server_port"`
}

type Openssl struct {
  UseSSL bool     `ini:"use_ssl"`
  RedirectHTTP bool  `ini:"redirect_http"`
  PortToRedirect string `ini:"port_to_redirect"`
}

type Auth struct {
  UseAuth bool     `ini:"use_auth"`
  Username string     `ini:"username"`
  Password string     `ini:"password"`
}

type Config struct {
  Internal    `ini:"internal"`
  Network     `ini:"network"`
  Openssl     `ini:"openssl"`
  Auth        `ini:"auth"`
}

var config *Config

func getConfig(fn string) *Config {
    c := new(Config)
    err := ini.MapTo(c, fn)
    if err != nil {
      printer.Fatal(err)
    }
    return c
}