package main

import (
	"log"
	"net/http"

	cert "discovery_server/certificates"
	handler "discovery_server/handlers"
)

func main() {

	//test.TestDecode()
	mux := http.NewServeMux()
	//TODO use different handler for different routes
	mux.HandleFunc("/", handler.HandleConnection)
	err := http.ListenAndServeTLS(":9010", cert.CertPem, cert.KeyPem, mux)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}
