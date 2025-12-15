package main

import (
	"log"
	"net/http"

	cert "discovery_server/certificates"
	function "discovery_server/functions"
	handler "discovery_server/handlers"
	watcher "discovery_server/watcher"
)

func main() {

	stopCh := make(chan struct{})

	var kubeconfigPathLocal = "/home/rmedina/.kube/config"
	clientset, erro := function.CreateKubernetesClient(kubeconfigPathLocal, "standard")
	if erro != nil {
		log.Fatalf("Errore creando il clientset: %v", erro)
	}

	go watcher.StartClusterSecretWatcher(clientset, "demo", stopCh)

	//test.TestDecode()
	mux := http.NewServeMux()
	//TODO use different handler for different routes
	mux.HandleFunc("/", handler.HandleConnection)
	err := http.ListenAndServeTLS(":9010", cert.CertPem, cert.KeyPem, mux)
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}
}
