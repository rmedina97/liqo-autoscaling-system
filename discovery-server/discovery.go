package main

import (
	"log"
	"net/http"

	function "discovery_server/functions"
	handler "discovery_server/handlers"
	watcher "discovery_server/watcher"
)

func main() {

	stopCh := make(chan struct{})

	//var kubeconfigPathLocal = "/home/rmedina/.kube/config"
	clientset, erro := function.CreateKubernetesClient()
	if erro != nil {
		log.Fatalf("Errore creando il clientset: %v", erro)
	}

	go watcher.StartClusterSecretWatcher(clientset, "demo", stopCh)

	mux := http.NewServeMux()
	//TODO use different handler for different routes
	mux.HandleFunc("/", handler.HandleConnection)
	//err := http.ListenAndServeTLS(":9010", cert.CertPem, cert.KeyPem, mux)

	certPath := "/app/certificates/tls.crt"
	keyPath := "/app/certificates/tls.key"
	err := http.ListenAndServeTLS(":9010", certPath, keyPath, mux)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}
