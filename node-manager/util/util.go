package util

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	types "nodegroupController/types"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"go.yaml.in/yaml/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/clientcmd"
)

// key is id nodegroup, value is nodegroup
// TODO change from hardcoded data to a dynamic one
var mapNodegroup = map[string]types.Nodegroup{
	"STANDARD": {
		Id:          "STANDARD",
		MaxSize:     3,
		MinSize:     0,
		CurrentSize: 0,
		//Nodes:       []string{"rmedina"},
		Nodes: []string{},
	},
	"GPU": {
		Id:          "GPU",
		MaxSize:     1,
		MinSize:     0,
		CurrentSize: 0,
		Nodes:       []string{},
	},
}

// flag for cache
var nodegroupIdsCached = false

// Node list
var nodeMinInfoList []types.NodeMinInfo = make([]types.NodeMinInfo, 0, 20)

// Nodegroup list
var nodegroupListMinInfo []types.NodegroupMinInfo = make([]types.NodegroupMinInfo, 0, 5)

// key is id node, value is node
// TODO change from hardcoded data to a dynamic one
var mapNode = map[string]types.Node{
	"rmedina": {
		Id: "rmedina",
		//NodegroupId:    "STANDARD",
		NodegroupId:    "",
		InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""},
	},
}

var mapRemoteClusters = make(map[string]string)

// Template map
var mapNodegroupTemplate = map[string]types.NodegroupTemplate{
	"GPU": {
		NodegroupId: "GPU",
		Resources: types.ResourceRange{
			Min: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("3.5"),
				v1.ResourceMemory: *resource.NewQuantity(4*1024*1024*1024, resource.BinarySI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":  *resource.NewQuantity(1, resource.DecimalSI),
			},
			Max: v1.ResourceList{
				//v1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
				v1.ResourceCPU:    resource.MustParse("3.5"),
				v1.ResourceMemory: *resource.NewQuantity(8*1024*1024*1024, resource.BinarySI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":  *resource.NewQuantity(4, resource.DecimalSI),
			},
		},
		Labels: map[string]string{
			"country":  "france",
			"provider": "liqo",
			"city":     "paris",
		},
		Cost: 25.0,
	},
	"STANDARD": {
		NodegroupId: "STANDARD",
		Resources: types.ResourceRange{
			Min: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewQuantity(2, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(2*1024*1024*1024, resource.BinarySI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":  *resource.NewQuantity(0, resource.DecimalSI),
			},
			Max: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewQuantity(2, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(3*1024*1024*1024, resource.BinarySI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				"nvidia.com/gpu":  *resource.NewQuantity(0, resource.DecimalSI),
			},
		},
		Labels: map[string]string{
			"country":  "italy",
			"provider": "liqo",
			"city":     "turin",
		},
		Cost: 1.0,
	},
}

// List of function inside the handle connection -------------------------------------------------------

// getAllNodegroups get all the nodegroups
func GetAllNodegroups() ([]types.NodegroupMinInfo, error) {

	// No existing Nodegroups
	if len(mapNodegroup) == 0 {
		return nil, nil
	} else {
		// Check if it is cached
		// TODO Design decision required to determine where these informations should be stored
		if !nodegroupIdsCached {
			for _, nodegroup := range mapNodegroup {
				nodegroupMinInfo := types.NodegroupMinInfo{Id: nodegroup.Id, MaxSize: nodegroup.MaxSize, MinSize: nodegroup.MinSize}
				nodegroupListMinInfo = append(nodegroupListMinInfo, nodegroupMinInfo)
			}
			nodegroupIdsCached = true
		}
		return nodegroupListMinInfo, nil
	}
}

// nodegroupForNode get the nodegroup for a specific node
func GetNodegroupForNode(id string) (*types.NodegroupMinInfo, error) {
	node, exist := mapNode[id]
	if !exist {
		return nil, nil //TODO change with error
	}
	if node.NodegroupId != "" { // Constraint imposed by CA; if empty, it should not be processed
		nodegroup := mapNodegroup[node.NodegroupId]
		nodegroupMinInfo := &types.NodegroupMinInfo{Id: nodegroup.Id, MaxSize: nodegroup.MaxSize, MinSize: nodegroup.MinSize}
		return nodegroupMinInfo, nil
	}
	return &types.NodegroupMinInfo{}, nil

}

// getCurrentSize get the current size of a nodegroup
func GetCurrentSize(id string) (*types.NodegroupCurrentSize, error) {
	nodegroup, exist := mapNodegroup[id]
	if !exist {
		return nil, nil
	}
	nodegroupCurrentSize := types.NodegroupCurrentSize{CurrentSize: nodegroup.CurrentSize}
	return &nodegroupCurrentSize, nil
}

// getNodegroupNodes get all the nodes of a nodegroup
func GetNodegroupNodes(id string) (*[]types.NodeMinInfo, error) {

	// Clear the list before adding the new nodes
	nodeMinInfoList = nodeMinInfoList[:0]
	nodegroup, exist := mapNodegroup[id]
	if !exist {
		return nil, nil //TODO change with error
	}
	if len(nodegroup.Nodes) == 0 {
		return nil, nil
	} else {
		for _, nodeId := range nodegroup.Nodes {
			nodeX := mapNode[nodeId]
			nodeMinInfoList = append(nodeMinInfoList, types.NodeMinInfo{Id: nodeX.Id, InstanceStatus: nodeX.InstanceStatus})
		}
		log.Printf("nodeMinInfoList: %v", nodeMinInfoList)
		return &nodeMinInfoList, nil
	}
}

// createNodegroup create a new nodegroup
func CreateNodegroup(newNodegroup types.Nodegroup) (success bool, err error) {

	if _, exists := mapNodegroup[newNodegroup.Id]; exists {
		return false, fmt.Errorf("nodegroup already exists")
	}
	mapNodegroup[newNodegroup.Id] = newNodegroup
	// TODO check if we need a lock on nodegroupiscached for the get all nodegroups
	// Update the list of nodegroups
	newNodegroupMinInfo := types.NodegroupMinInfo{Id: newNodegroup.Id, MaxSize: newNodegroup.MaxSize, MinSize: newNodegroup.MinSize}
	nodegroupListMinInfo = append(nodegroupListMinInfo, newNodegroupMinInfo)
	//nodegroupList = append(nodegroupList, newNodegroup)
	return true, nil
}

// deleteNodegroup delete the target nodegroup
func DeleteNodegroup(nodegroupId string) (success bool, err error) {

	if _, exists := mapNodegroup[nodegroupId]; !exists {
		return false, fmt.Errorf("nodegroup not found")
	}
	delete(mapNodegroup, nodegroupId)
	// Remove the nodegroup from the list
	for i, nodegroup := range nodegroupListMinInfo {
		if nodegroup.Id == nodegroupId {
			nodegroupListMinInfo = append(nodegroupListMinInfo[:i], nodegroupListMinInfo[i+1:]...)
			break
		}
	}
	// Update the cache flag if the list is empty
	if len(nodegroupListMinInfo) == 0 {
		nodegroupIdsCached = false
	}
	return true, nil
}

// scaleUpNodegroup scale up the nodegroup of a certain amount
func ScaleUpNodegroup(nodegroupId string, count int) (success bool, err error) {

	log.Printf("ScaleUpNodegroup called with query params: id %s count  %d", nodegroupId, count)

	client, err := newClient()
	if err != nil {
		log.Printf("failed to create a client: %v", err)
	}

	// Send a GET request to the discovery server
	reply, err := client.Get("https://localhost:9010/list") // TODO create a parameter
	if err != nil {
		log.Printf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		log.Printf("remote cluster not found")
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var clusterList []types.Cluster
	if err := json.NewDecoder(reply.Body).Decode(&clusterList); err != nil {
		log.Printf("error decoding JSON: %v", err)
	}

	for key, value := range mapRemoteClusters {
		log.Printf("%s: %v\n", key, value)
	}

	//--------------------------------------------------
	// Choose the first cluster that match the template (two template, one for GPU and one for standard)
	// TODO implement a better choice algorithm
	//--------------------------------------------------
	for i := 0; i < count; i++ {
		clusterchosen := types.Cluster{}
		for _, c := range clusterList {
			if c.Labels["hasGPU"] == "true" && nodegroupId == "GPU" {
				if _, exists := mapRemoteClusters[c.Name]; !exists {
					clusterchosen = c
					mapRemoteClusters[c.Name] = c.Name
					break
				}
			}
			if c.Labels["hasGPU"] == "false" && nodegroupId == "STANDARD" {
				if _, exists := mapRemoteClusters[c.Name]; !exists {
					clusterchosen = c
					mapRemoteClusters[c.Name] = c.Name
					break
				}
			}
		}

		log.Printf("Cluster chosen: %s", clusterchosen.Name)

		//--------------------------------------------------
		// Transform the kubeconfig from base64 to a file for the peering command
		//--------------------------------------------------
		kubeconfigBytes, err := base64.StdEncoding.DecodeString(clusterchosen.Kubeconfig)
		if err != nil {
			panic(err)
		}

		cfg, err := clientcmd.Load(kubeconfigBytes)
		if err != nil {
			panic(err)
		}

		ctx := cfg.Contexts[cfg.CurrentContext]
		if ctx == nil {
			panic("current context not found")
		}

		cluster := cfg.Clusters[ctx.Cluster]
		if cluster == nil {
			panic("cluster not found")
		}

		// Parse URL del server
		u, err := url.Parse(cluster.Server)
		if err != nil {
			panic(err)
		}

		ip := u.Hostname()
		fmt.Println("IP extracted from kubeconfig:", ip)

		// Crea file temporaneo
		tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
		if err != nil {
			panic(err)
		}

		defer os.Remove(tmpFile.Name())

		// Write the file
		if _, err := tmpFile.Write(kubeconfigBytes); err != nil {
			panic(err)
		}
		tmpFile.Close()

		// Get the path of the temporary file
		kubeconfigPath := tmpFile.Name()
		fmt.Println("Kubeconfig salvato in:", kubeconfigPath)

		//--------------------------------------------------
		// Peering with liqoctl two conditions: with/without nat and with/without GPU
		// TODO check if the peering is already present
		// TODO manage the error if the peering fails
		//--------------------------------------------------

		switch {
		case !clusterchosen.HasNat && nodegroupId == "STANDARD":
			// cmd := exec.Command(
			// 	"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPath, "--skip-confirm",
			// )
			// output, err := cmd.CombinedOutput()
			// if err != nil {
			// 	log.Printf("Error during SSH:%v", err)
			// }
			// log.Printf("Output: %s ", output)
			log.Printf("Cluster has no nat and request is for STANDARD")
			cmd := exec.Command(
				"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPath, "--create-resource-slice=false", "--skip-confirm",
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Error during SSH:%v", err)
			}
			log.Printf("Output: %s ", output)
			gpu := clusterchosen.Resources["nvidia.com/gpu"]
			log.Printf("CLUSTER HAS %s GPUs ", gpu.String())
			rs := types.ResourceSlice{
				APIVersion: "authentication.liqo.io/v1beta1",
				Kind:       "ResourceSlice",
				Metadata: types.Metadata{
					Name:      clusterchosen.Name,
					Namespace: "liqo-tenant-" + clusterchosen.Name,
					Labels: map[string]string{
						"liqo.io/remote-cluster-id": clusterchosen.Name,
						"liqo.io/remoteID":          clusterchosen.Name,
						"liqo.io/replication":       "true",
					},
					Annotations: map[string]string{
						"liqo.io/create-virtual-node": "true",
						"custom.annotation":           "hello-there-general-kenobi",
					},
				},
				Spec: types.ResourceSliceSpec{
					Class:             "default",
					ProviderClusterID: clusterchosen.Name,
					Resources: types.Resources{
						//CPU:    clusterchosen.Resources.Cpu().String(),
						CPU:    "1.5",
						Memory: clusterchosen.Resources.Memory().String(),
						Pods:   clusterchosen.Resources.Pods().String(),
						GPU:    gpu.String(),
					},
				},
				Status: types.Status{},
			}

			data, err := yaml.Marshal(rs)
			if err != nil {
				log.Fatal(err)
			}

			cmd1 := exec.Command("kubectl", "apply", "-f", "-")
			cmd1.Stdin = bytes.NewReader(data)
			output1, err1 := cmd1.CombinedOutput()
			if err != nil {
				log.Fatalf("kubectl apply failed: %v\n%s", err1, string(output1))
			}
			log.Println(string(output1))
			log.Printf("ResourceSlice created for cluster %s is actived?", clusterchosen.Name)

		case clusterchosen.HasNat && nodegroupId == "STANDARD":
			cmd := exec.Command(
				"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPath, "--api-server-url", ip, "--skip-confirm",
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Error during SSH:%v", err)
			}
			log.Printf("Output: %s ", output)

		case !clusterchosen.HasNat && nodegroupId == "GPU":
			log.Printf("Cluster has no nat and request is for GPU")
			cmd := exec.Command(
				"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPath, "--create-resource-slice=false", "--skip-confirm",
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Error during SSH:%v", err)
			}
			log.Printf("Output: %s ", output)
			gpu := clusterchosen.Resources["nvidia.com/gpu"]
			log.Printf("CLUSTER HAS %s GPUs ", gpu.String())
			rs := types.ResourceSlice{
				APIVersion: "authentication.liqo.io/v1beta1",
				Kind:       "ResourceSlice",
				Metadata: types.Metadata{
					Name:      clusterchosen.Name,
					Namespace: "liqo-tenant-" + clusterchosen.Name,
					Labels: map[string]string{
						"liqo.io/remote-cluster-id": clusterchosen.Name,
						"liqo.io/remoteID":          clusterchosen.Name,
						"liqo.io/replication":       "true",
					},
					Annotations: map[string]string{
						"liqo.io/create-virtual-node": "true",
						"custom.annotation":           "hello-there-general-kenobi",
					},
				},
				Spec: types.ResourceSliceSpec{
					Class:             "default",
					ProviderClusterID: clusterchosen.Name,
					Resources: types.Resources{
						//CPU:    clusterchosen.Resources.Cpu().String(),
						CPU:    "3.5",
						Memory: clusterchosen.Resources.Memory().String(),
						Pods:   clusterchosen.Resources.Pods().String(),
						GPU:    gpu.String(),
					},
				},
				Status: types.Status{},
			}

			data, err := yaml.Marshal(rs)
			if err != nil {
				log.Fatal(err)
			}

			cmd1 := exec.Command("kubectl", "apply", "-f", "-")
			cmd1.Stdin = bytes.NewReader(data)
			output1, err1 := cmd1.CombinedOutput()
			if err != nil {
				log.Fatalf("kubectl apply failed: %v\n%s", err1, string(output1))
			}
			log.Println(string(output1))
			log.Printf("ResourceSlice created for cluster %s is actived?", clusterchosen.Name)

		case clusterchosen.HasNat && nodegroupId == "GPU":
			cmd := exec.Command(
				"liqoctl", "peer", "--remote-kubeconfig", kubeconfigPath, "--create-resource-slice=false", "--api-server-url", ip, "--skip-confirm",
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Error during SSH:%v", err)
			}
			log.Printf("Output: %s ", output)
			gpu := clusterchosen.Resources["nvidia.com/gpu"]
			rs := types.ResourceSlice{
				APIVersion: "authentication.liqo.io/v1beta1",
				Kind:       "ResourceSlice",
				Metadata: types.Metadata{
					Name:      clusterchosen.Name,
					Namespace: "liqo-tenant-" + clusterchosen.Name,
					Labels: map[string]string{
						"liqo.io/remote-cluster-id": clusterchosen.Name,
						"liqo.io/remoteID":          clusterchosen.Name,
						"liqo.io/replication":       "true",
						"custom.label":              "shadow-slave",
					},
					Annotations: map[string]string{
						"liqo.io/create-virtual-node": "false",
						"custom.annotation":           "hello-there-general-kenobi",
					},
				},
				Spec: types.ResourceSliceSpec{
					Class:             "default",
					ProviderClusterID: clusterchosen.Name,
					Resources: types.Resources{
						CPU:    clusterchosen.Resources.Cpu().String(),
						Memory: clusterchosen.Resources.Memory().String(),
						Pods:   clusterchosen.Resources.Pods().String(),
						GPU:    gpu.String(),
					},
				},
				Status: types.Status{},
			}

			data, err := yaml.Marshal(rs)
			if err != nil {
				log.Fatal(err)
			}

			cmd1 := exec.Command("kubectl", "apply", "-f", "-")
			cmd1.Stdin = bytes.NewReader(data)
			output1, err1 := cmd1.CombinedOutput()
			if err != nil {
				log.Fatalf("kubectl apply failed: %v\n%s", err1, string(output1))
			}
			log.Println(string(output1))
		}

		// Wait until the node is ready
		for {
			cmdCheck := exec.Command("kubectl", "get", "node", clusterchosen.Name)
			if err := cmdCheck.Run(); err != nil {
				log.Printf("Nodo %s non ancora creato, attendo 1s...", clusterchosen.Name)
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}

		cmdWait := exec.Command("kubectl", "wait", "node/"+clusterchosen.Name, "--for=condition=Ready", "--timeout=30s")
		outputWait, err := cmdWait.CombinedOutput()
		if err != nil {
			log.Fatalf("Where is the virtual node?: %v\n%s", err, string(outputWait))
		}

		// Patch node with Provider ID
		providerId := "liqo://" + clusterchosen.Name
		patch := fmt.Sprintf(`{"spec":{"providerID":"%s"}}`, providerId)
		cmdPatch := exec.Command("kubectl", "patch", "node", clusterchosen.Name, "--type=merge", "-p", patch)
		outputPatch, err := cmdPatch.CombinedOutput()
		if err != nil {
			log.Fatalf("Patch failed: %v\n%s", err, string(outputPatch))
		}

		//--------------------------------------------------
		// Update internal structures
		//--------------------------------------------------
		mapNode[clusterchosen.Name] = types.Node{Id: clusterchosen.Name, NodegroupId: nodegroupId, InstanceStatus: types.InstanceStatus{InstanceState: 1, InstanceErrorInfo: ""}}
		nodegroup := mapNodegroup[nodegroupId]
		nodegroup.CurrentSize++
		log.Printf("Nodegroup size is %d after adding node %s", nodegroup.CurrentSize, clusterchosen.Name)
		nodegroup.Nodes = append(nodegroup.Nodes, clusterchosen.Name)
		mapNodegroup[nodegroupId] = nodegroup
	}
	return true, nil

}

// scaleDownNodegroup scale down the nodegroup killing a certain node
func ScaleDownNodegroup(nodegroupId string, nodeId string) (success bool, err error) {

	// TODO Sent request to discovery server for only the kubeconfig of that node
	client, err := newClient()
	if err != nil {
		log.Printf("failed to create a client: %v", err)
	}

	// Send a GET request to the discovery server
	reply, err := client.Get("https://localhost:9010/list") // TODO create a parameter
	if err != nil {
		log.Printf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		log.Printf("remote cluster not found")
	} else if reply.StatusCode != http.StatusOK {
		log.Printf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var clusterList []types.Cluster
	if err := json.NewDecoder(reply.Body).Decode(&clusterList); err != nil {
		log.Printf("error decoding JSON: %v", err)
	}

	//divide in nodegroup
	cleanNodeId := strings.TrimPrefix(nodeId, "liqo://")
	log.Printf("Clean node id to be deleted \n\n\n\n\n\n\n\n: %s", cleanNodeId)
	clusterchosen := types.Cluster{}
	for _, c := range clusterList {
		if c.Name == cleanNodeId {
			clusterchosen = c
			break
		}
	}

	log.Printf("Cluster chosen: %s", clusterchosen.Name)
	// -----------------------------------------
	// Decodifica
	kubeconfigBytes, err := base64.StdEncoding.DecodeString(clusterchosen.Kubeconfig)
	if err != nil {
		panic(err)
	}
	log.Printf("preso cluster %s", clusterchosen.Name)

	// Crea file temporaneo
	tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		panic(err)
	}

	defer os.Remove(tmpFile.Name())

	// Scrive il contenuto
	if _, err := tmpFile.Write(kubeconfigBytes); err != nil {
		panic(err)
	}
	tmpFile.Close()

	// Ottieni il path
	kubeconfigPath := tmpFile.Name()
	fmt.Println("Kubeconfig salvato in:", kubeconfigPath)
	// -----------------------------------------

	//log.Printf("ScaleDownNodegroup called on first: %s", nodeId)
	cmd := exec.Command(
		"liqoctl", "unpeer", "--remote-kubeconfig", kubeconfigPath, "--skip-confirm",
	)
	output, err := cmd.CombinedOutput()
	log.Printf("Fine SSH, %s %s", output, err)
	if err != nil {
		log.Printf("Error during SSH: %v", err)
	}
	nodegroup := mapNodegroup[nodegroupId]
	for i, node := range nodegroup.Nodes {
		if node == cleanNodeId { // Remove the node from the list
			nodegroup.Nodes = append(nodegroup.Nodes[:i], nodegroup.Nodes[i+1:]...)
			break
		}
	}
	nodegroup.CurrentSize--
	log.Printf("Nodegroup size is %d after removing node %s", nodegroup.CurrentSize, cleanNodeId)
	mapNodegroup[nodegroupId] = nodegroup
	delete(mapNode, cleanNodeId)
	delete(mapRemoteClusters, cleanNodeId)
	return true, nil
}

// // getTemplateNodegroup get the template of a nodegroup
func GetTemplateNodegroup(id string) (*types.NodegroupTemplate, error) {

	// No existing Nodegrouptemplate
	log.Printf("GetTemplateNodegroup called with id: %s", id)
	template, exist := mapNodegroupTemplate[id]
	if !exist {
		log.Printf("Nodegroup template not found")
		return nil, nil //TODO change with error
	}
	return &template, nil
}

// getPriceNodegroup get the price of a nodegroup
func GetPriceNodegroup(id string) (*float64, error) {

	// No existing Nodegrouptemplate
	log.Printf("GetPriceNodegroup called with id: %s", id)
	re := regexp.MustCompile(`for-([^-]+)-`)
	templateName := re.FindStringSubmatch(id)
	log.Printf("Extracted template name: %s", templateName[1])
	template, exist := mapNodegroupTemplate[templateName[1]]
	if !exist {
		log.Printf("Nodegroup template not found")
		return nil, nil //TODO change with error
	}
	log.Printf("Price of NODEGRUPPPP----------------- %s is %f", id, template.Cost)
	return &template.Cost, nil
}

// getPriceNodegroup get the price of a nodegroup
func GetPricePod() (*float64, error) {

	// Assuming same price
	var podprice float64 = 1.00
	return &podprice, nil
}

// Create a new client
// TODO Search if someone still uses 509 cert without san, if yes use VerifyPeerCertificate to custom accept them
func newClient() (*http.Client, error) {
	certPool := x509.NewCertPool()
	certData, err := os.ReadFile("cert.pem")

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
