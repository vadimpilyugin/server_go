package main

import (
  "crypto/tls"
  "printer"
)

func loadCert() tls.Certificate {
  cert,err := tls.X509KeyPair(
    /*

    Certificate
    

    */
    []byte(`-----BEGIN CERTIFICATE-----

-----END CERTIFICATE-----`),
    /*

    Private key
    

    */
    []byte(`-----BEGIN PRIVATE KEY-----

-----END PRIVATE KEY-----`))
  if err != nil {
    printer.Fatal(err)
  }
  return cert
}

func loadTlsConfig() *tls.Config {
  cfg := &tls.Config{
    MinVersion:               tls.VersionTLS12,
    CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
    PreferServerCipherSuites: true,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
        tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_RSA_WITH_AES_256_CBC_SHA,
    },
  }
  cfg.Certificates = append(cfg.Certificates, loadCert())
  return cfg
}