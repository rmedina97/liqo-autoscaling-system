package main

import (

	"log"
	"net/http"

	handler "discovery_server/handler"

)

func main() {


	mux := http.NewServeMux()
	//TODO use different handler for different routes
	mux.HandleFunc("/", handler.HandleConnection)
	err := http.ListenAndServeTLS(":9010", types.CertPem, types.KeyPem, mux)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}