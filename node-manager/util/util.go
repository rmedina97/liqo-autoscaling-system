package util

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	types "node-manager/types"
	"os"
	"os/exec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClusterList() ([]types.Cluster, error) {

	// Create an HTTP client
	client, err := newClient()
	if err != nil {
		log.Printf("failed to create a client: %v", err)
	}

	// Send a GET request to the discovery server
	reply, err := client.Get("https://discovery-server-svc.demo:9010/list") // TODO create a parameter
	if err != nil {
		log.Printf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		log.Printf("remote cluster not found")
		return nil, fmt.Errorf("failed to find remote cluster list")
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var clusterList []types.Cluster
	if err := json.NewDecoder(reply.Body).Decode(&clusterList); err != nil {
		log.Printf("error decoding JSON: %v", err)
	}

	return clusterList, nil
}

// Create a new client
// TODO Search if someone still uses 509 cert without san, if yes use VerifyPeerCertificate to custom accept them
func newClient() (*http.Client, error) {
	certPool := x509.NewCertPool()
	certData, err := os.ReadFile("/app/certificates/tls.crt")

	if err != nil {
		return nil, fmt.Errorf("error reading certificate: %v", err)
	}

	if !certPool.AppendCertsFromPEM(certData) {
		return nil, fmt.Errorf("failed to append certificate")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{RootCAs: certPool},
			ForceAttemptHTTP2: true,
		},
	}
	return client, nil

}

func DecodeKubeconfig(kubeconfigBase64 string) (string, string, string, error) {
	kubeconfigBytes, err := base64.StdEncoding.DecodeString(kubeconfigBase64)
	if err != nil {
		panic(err)
	}

	cfgRemote, err := clientcmd.Load(kubeconfigBytes)
	if err != nil {
		panic(err)
	}

	ctxRemote := cfgRemote.Contexts[cfgRemote.CurrentContext]
	if ctxRemote == nil {
		panic("current remote context not found")
	}

	cluster := cfgRemote.Clusters[ctxRemote.Cluster]
	if cluster == nil {
		panic("remote cluster not found")
	}

	// Parse URL del server
	urlRemote, err := url.Parse(cluster.Server)
	if err != nil {
		panic(err)
	}

	ip := urlRemote.Hostname()
	fmt.Println("IP extracted from kubeconfig:", ip)

	// Crea file temporaneo
	tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		panic(err)
	}

	// Write the file
	if _, err := tmpFile.Write(kubeconfigBytes); err != nil {
		panic(err)
	}
	tmpFile.Close()

	// Get the path of the temporary file
	kubeconfigPathRemote := tmpFile.Name()
	// TODO: Delete tmpFile.Name() because == kubeconfigPathRemote, and transform in a struct with all the info
	fmt.Println("Kubeconfig saved in:", kubeconfigPathRemote)

	return ip, kubeconfigPathRemote, tmpFile.Name(), nil
}

func PeeringWithLiqoctl(clusterchosen types.Cluster,
	nodegroupId string,
	kubeconfigPathRemote string,
	kubeconfigPathLocal string,
	tmpFile string,
	ip string) error {

	// Prepare to delete the temporary kubeconfig file
	defer os.Remove(tmpFile)
	switch {
	case !clusterchosen.HasNat && nodegroupId == "STANDARD":

		log.Printf("Cluster has no nat and request is for STANDARD")
		cmd := exec.Command(
			"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPathRemote, "--create-resource-slice=false", "--skip-confirm",
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error during liqoctl peer: %v", err)
			return fmt.Errorf(" error: %w", err)
		}
		log.Printf("Output: %s ", output)
		gpu := clusterchosen.Resources["nvidia.com/gpu"]
		log.Printf("CLUSTER HAS %s GPUs ", gpu.String())

		// // GVR della CRD ResourceSlice
		// gvr := schema.GroupVersionResource{
		// 	Group:    "authentication.liqo.io",
		// 	Version:  "v1beta1",
		// 	Resource: "resourceslices",
		// }

		// // Oggetto unstructured
		// rs := &unstructured.Unstructured{
		// 	Object: map[string]interface{}{
		// 		"apiVersion": "authentication.liqo.io/v1beta1",
		// 		"kind":       "ResourceSlice",
		// 		"metadata": map[string]interface{}{
		// 			"name":      clusterchosen.Name,
		// 			"namespace": "liqo-tenant-" + clusterchosen.Name,
		// 			"labels": map[string]interface{}{
		// 				"liqo.io/remote-cluster-id": clusterchosen.Name,
		// 				"liqo.io/remoteID":          clusterchosen.Name,
		// 				"liqo.io/replication":       "true",
		// 			},
		// 			"annotations": map[string]interface{}{
		// 				"liqo.io/create-virtual-node": "true",
		// 				"custom.annotation":           "hello-there-general-kenobi",
		// 			},
		// 		},
		// 		"spec": map[string]interface{}{
		// 			"class":             "default",
		// 			"providerClusterID": clusterchosen.Name,
		// 			"resources": map[string]interface{}{
		// 				"cpu":    "1.5",
		// 				"memory": clusterchosen.Resources.Memory().String(),
		// 				"pods":   clusterchosen.Resources.Pods().String(),
		// 				"gpu":    gpu.String(),
		// 			},
		// 		},
		// 		"status": map[string]interface{}{},
		// 	},
		// }

		// // Creazione della risorsa
		// result, err := dynClient.Resource(gvr).
		// 	Namespace("liqo-tenant-"+clusterchosen.Name).
		// 	Create(context.Background(), rs, metav1.CreateOptions{})
		// if err != nil {
		// 	return fmt.Errorf("ResourceSlice creation error: %w", err)
		// }

		// log.Printf("ResourceSlice %s created", result.GetName())
		err = CreateVirtualNodeLabel(clusterchosen, kubeconfigPathLocal)
		if err != nil {
			return fmt.Errorf("ResourceSlice creation error: %w", err)
		}

	case clusterchosen.HasNat && nodegroupId == "STANDARD":
		cmd := exec.Command(
			"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPathRemote, "--create-resource-slice=false", "--api-server-url", ip, "--skip-confirm",
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("SSH error: %w", err)
		}
		log.Printf("Output: %s ", output)

		err = CreateVirtualNodeLabel(clusterchosen, kubeconfigPathLocal)
		if err != nil {
			return fmt.Errorf("ResourceSlice creation error: %w", err)
		}

	case !clusterchosen.HasNat && nodegroupId == "GPU":
		log.Printf("Cluster has no nat and request is for GPU")
		cmd := exec.Command(
			"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPathRemote, "--create-resource-slice=false", "--skip-confirm",
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("SSH error: %w", err)
		}
		log.Printf("Output: %s ", output)
		// gpu := clusterchosen.Resources["nvidia.com/gpu"]
		// log.Printf("CLUSTER HAS %s GPUs ", gpu.String())
		// rs := types.ResourceSlice{
		// 	APIVersion: "authentication.liqo.io/v1beta1",
		// 	Kind:       "ResourceSlice",
		// 	Metadata: types.Metadata{
		// 		Name:      clusterchosen.Name,
		// 		Namespace: "liqo-tenant-" + clusterchosen.Name,
		// 		Labels: map[string]string{
		// 			"liqo.io/remote-cluster-id": clusterchosen.Name,
		// 			"liqo.io/remoteID":          clusterchosen.Name,
		// 			"liqo.io/replication":       "true",
		// 		},
		// 		Annotations: map[string]string{
		// 			"liqo.io/create-virtual-node": "true",
		// 			"custom.annotation":           "hello-there-general-kenobi",
		// 		},
		// 	},
		// 	Spec: types.ResourceSliceSpec{
		// 		Class:             "default",
		// 		ProviderClusterID: clusterchosen.Name,
		// 		Resources: types.Resources{
		// 			//CPU:    clusterchosen.Resources.Cpu().String(),
		// 			CPU:    "3.5",
		// 			Memory: clusterchosen.Resources.Memory().String(),
		// 			Pods:   clusterchosen.Resources.Pods().String(),
		// 			GPU:    gpu.String(),
		// 		},
		// 	},
		// 	Status: types.Status{},
		// }

		// data, err := yaml.Marshal(rs)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// cmd1 := exec.Command("kubectl", "apply", "-f", "-")
		// cmd1.Stdin = bytes.NewReader(data)
		// output1, err1 := cmd1.CombinedOutput()
		// if err != nil {
		// 	log.Fatalf("kubectl apply failed: %v\n%s", err1, string(output1))
		// }
		// log.Println(string(output1))
		// log.Printf("ResourceSlice created for cluster %s is actived?", clusterchosen.Name)
		err = CreateVirtualNodeLabel(clusterchosen, kubeconfigPathLocal)
		if err != nil {
			return fmt.Errorf("ResourceSlice creation error: %w", err)
		}

	case clusterchosen.HasNat && nodegroupId == "GPU":
		cmd := exec.Command(
			"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPathRemote, "--create-resource-slice=false", "--api-server-url", ip, "--skip-confirm",
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("SSH error: %w", err)
		}
		log.Printf("Output: %s ", output)
		// gpu := clusterchosen.Resources["nvidia.com/gpu"]
		// rs := types.ResourceSlice{
		// 	APIVersion: "authentication.liqo.io/v1beta1",
		// 	Kind:       "ResourceSlice",
		// 	Metadata: types.Metadata{
		// 		Name:      clusterchosen.Name,
		// 		Namespace: "liqo-tenant-" + clusterchosen.Name,
		// 		Labels: map[string]string{
		// 			"liqo.io/remote-cluster-id": clusterchosen.Name,
		// 			"liqo.io/remoteID":          clusterchosen.Name,
		// 			"liqo.io/replication":       "true",
		// 			"custom.label":              "shadow-slave",
		// 		},
		// 		Annotations: map[string]string{
		// 			"liqo.io/create-virtual-node": "false",
		// 			"custom.annotation":           "hello-there-general-kenobi",
		// 		},
		// 	},
		// 	Spec: types.ResourceSliceSpec{
		// 		Class:             "default",
		// 		ProviderClusterID: clusterchosen.Name,
		// 		Resources: types.Resources{
		// 			CPU:    clusterchosen.Resources.Cpu().String(),
		// 			Memory: clusterchosen.Resources.Memory().String(),
		// 			Pods:   clusterchosen.Resources.Pods().String(),
		// 			GPU:    gpu.String(),
		// 		},
		// 	},
		// 	Status: types.Status{},
		// }

		// data, err := yaml.Marshal(rs)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// cmd1 := exec.Command("kubectl", "apply", "-f", "-")
		// cmd1.Stdin = bytes.NewReader(data)
		// output1, err1 := cmd1.CombinedOutput()
		// if err != nil {
		// 	log.Fatalf("kubectl apply failed: %v\n%s", err1, string(output1))
		// }
		// log.Println(string(output1))
		err = CreateVirtualNodeLabel(clusterchosen, kubeconfigPathLocal)
		if err != nil {
			return fmt.Errorf("ResourceSlice creation error: %w", err)
		}

	}
	return nil
}

func CreateKubernetesClient(kubeconfigPath string, client string) (dynamic.Interface, *kubernetes.Clientset) {
	// Client creation, dynamic for custom resources, clientset for core resources
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatalf("Errore caricando kubeconfig: %v", err)
	}
	if client == "dynamic" {
		dynClient, err := dynamic.NewForConfig(cfg)
		if err != nil {
			log.Fatal(err)
		}
		return dynClient, nil
	} else {
		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			log.Fatal(err)
		}
		return nil, clientset
	}
}

func CreateVirtualNodeLabel(clusterchosen types.Cluster, kubeconfigPathLocal string) error {

	// Create dynamic client
	dynClient, _ := CreateKubernetesClient(kubeconfigPathLocal, "dynamic")

	//TODO GPU can have different names, need to generalize
	// Get GPU resources
	gpu := clusterchosen.Resources["nvidia.com/gpu"]

	// GroupVersionResource della CRD ResourceSlice
	gvr := schema.GroupVersionResource{
		Group:    "authentication.liqo.io",
		Version:  "v1beta1",
		Resource: "resourceslices",
	}

	// Unstructured object for ResourceSlice
	rs := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "authentication.liqo.io/v1beta1",
			"kind":       "ResourceSlice",
			"metadata": map[string]interface{}{
				"name":      clusterchosen.Name,
				"namespace": "liqo-tenant-" + clusterchosen.Name,
				"labels": map[string]interface{}{
					"liqo.io/remote-cluster-id": clusterchosen.Name,
					"liqo.io/remoteID":          clusterchosen.Name,
					"liqo.io/replication":       "true",
				},
				"annotations": map[string]interface{}{
					"liqo.io/create-virtual-node": "true",
					"custom.annotation":           "hello-there-general-kenobi",
				},
			},
			"spec": map[string]interface{}{
				"class":             "default",
				"providerClusterID": clusterchosen.Name,
				"resources": map[string]interface{}{
					"cpu":    "1.5",
					"memory": clusterchosen.Resources.Memory().String(),
					"pods":   clusterchosen.Resources.Pods().String(),
					"gpu":    gpu.String(),
				},
			},
			"status": map[string]interface{}{},
		},
	}

	// Creazione della risorsa
	_, err := dynClient.Resource(gvr).
		Namespace("liqo-tenant-"+clusterchosen.Name).
		Create(context.Background(), rs, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("ResourceSlice creation error: %w", err)
	}
	return nil

}
