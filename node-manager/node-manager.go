package main

import (
	//"encoding/json"
	//"fmt"
	"log"
	"net/http"

	//"os/exec"
	handler "node-manager/handler"
)

func main() {

	// TODO maybe search local cluster and adds all the nodes inside the same nodegroup, with the label to avoid scaledown
	mux := http.NewServeMux()
	//TODO use different handler for different routes
	mux.HandleFunc("/", handler.HandleConnection)
	//from secret
	certPath := "/app/certificates/tls.crt"
	keyPath := "/app/certificates/tls.key"
	err := http.ListenAndServeTLS(":9009", certPath, keyPath, mux)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}
