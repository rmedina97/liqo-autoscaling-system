package watcher

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type Cluster struct {
	Name       string            `json:"name"`
	Kubeconfig string            `json:"kubeconfig"`
	Resources  v1.ResourceList   `json:"resources"`
	Labels     map[string]string `json:"labels"`
	HasNat     bool              `json:"hasNat"`
}

var (
	clusters   = make([]*Cluster, 0)
	clustersMu sync.RWMutex
)

// StartClusterSecretWatcher osserva tutti i Secrets con label autoscaling=true
func StartClusterSecretWatcher(clientset *kubernetes.Clientset, namespace string, stopCh <-chan struct{}) {

	log.Printf("go routine lanciata: StartClusterSecretWatcher")
	watcher, err := clientset.CoreV1().Secrets(namespace).Watch(context.Background(), metav1.ListOptions{
		LabelSelector: "autoscaling=true",
	})
	if err != nil {
		log.Fatalf("Errore creando watcher globale: %v", err)
	}
	defer watcher.Stop()

	log.Printf("watcher globale avviato")

	for {
		select {
		case <-stopCh:
			return
		case event, ok := <-watcher.ResultChan():
			log.Printf("vento ricevuto dal watcher globale: %v", event.Type)
			//evita il panic nel caso di chiusura del canale da parte di altri
			if !ok {
				log.Println("Watcher globale chiuso, uscita")
				return
			}
			// evita il panic se non è un secret
			sec, ok := event.Object.(*v1.Secret)
			if !ok {
				continue
			}

			switch event.Type {
			case watch.Added, watch.Modified:
				addOrUpdateCluster(sec)
			case watch.Deleted:
				removeCluster(sec.Name)
			}
		}
	}
}

// upsertCluster aggiunge o aggiorna un cluster dal Secret
func addOrUpdateCluster(sec *v1.Secret) {
	cluster := parseSecretToCluster(sec)

	clustersMu.Lock()
	defer clustersMu.Unlock()

	for i, c := range clusters {
		if c.Name == sec.Name {
			clusters[i] = cluster
			log.Printf("Cluster aggiornato: %s", sec.Name)
			return
		}
	}

	clusters = append(clusters, cluster)
	log.Printf("Cluster aggiunto: %s", sec.Name)
}

// parseSecretToCluster Secret-->Cluster
func parseSecretToCluster(sec *v1.Secret) *Cluster {
	cluster := &Cluster{}

	// 1. info → JSON per Resources, Labels, HasNat
	if data, ok := sec.Data["info"]; ok {
		if err := json.Unmarshal(data, cluster); err != nil {
			log.Printf("Errore parsing info (%s): %v", sec.Name, err)
		}
	}

	// 2. kubeconfig → string
	if kc, ok := sec.Data["kubeconfig"]; ok {
		cluster.Kubeconfig = string(kc)
	}

	// 3. Name dal Secret
	cluster.Name = sec.Name

	return cluster
}

// removeCluster rimuove un cluster dalla lista
func removeCluster(name string) {
	clustersMu.Lock()
	defer clustersMu.Unlock()

	for i, c := range clusters {
		if c.Name == name {
			clusters = append(clusters[:i], clusters[i+1:]...)
			log.Printf("Cluster rimosso: %s", name)
			break
		}
	}
}

func GetClusters() []Cluster {
	clustersMu.RLock()
	defer clustersMu.RUnlock()

	// copia i dati per evitare che il chiamante modifichi direttamente la slice interna
	result := make([]Cluster, len(clusters))
	for i, c := range clusters {
		result[i] = *c
	}
	return result
}
