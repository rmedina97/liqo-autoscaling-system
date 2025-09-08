package main

import (
	//"encoding/json"
	//"fmt"
	"log"
	"net/http"

	//"os/exec"
	handler "nodegroupController/handler"
	types "nodegroupController/types"
)

func main() {

	// TODO maybe search local cluster and adds all the nodes inside the same nodegroup, with the label to avoid scaledown
	mux := http.NewServeMux()
	//TODO use different handler for different routes
	mux.HandleFunc("/", handler.HandleConnection)
	err := http.ListenAndServeTLS(":9009", types.CertPem, types.KeyPem, mux)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}
