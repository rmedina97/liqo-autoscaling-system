package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	v1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
	"server_grpc/protos"

	"net/http"
)

// HERE START PERSONAL STRUCTS---------------// HERE START PERSONAL STRUCTS	---------------// HERE START PERSONAL STRUCTS
type Nodegroup struct {
	Id          string   `json:"id"`
	CurrentSize int32    `json:"currentSize"` //TODO struct only with the required field
	MaxSize     int32    `json:"maxSize"`
	MinSize     int32    `json:"minSize"`
	Nodes       []string `json:"nodes"` //TODO maybe put only ids of the nodes?
}

type Node struct {
	Id string `json:"id"`
}

type InstanceStatus struct {
	InstanceState     int32 //from zero to three
	InstanceErrorInfo string
}

type NodeMinInfo struct {
	Id             string         `json:"id"`
	InstanceStatus InstanceStatus `json:"--"`
}

type NodegroupMinInfo struct {
	Id      string `json:"id"`
	MaxSize int32  `json:"maxSize"`
	MinSize int32  `json:"minSize"`
}

type NodegroupCurrentSize struct {
	CurrentSize int32 `json:"currentSize"`
}

type GPUTypes struct {
	Name  string
	Specs string
}

var flag int = 0

// HERE END PERSONAL STRUCTS---------------// HERE END PERSONAL STRUCTS	---------------// HERE END PERSONAL STRUCTS

// protos->package with the files definition (_grpc.pb.go and pb.go) born from the protos file
type cloudProviderServer struct {
	protos.UnimplementedCloudProviderServer
}

// HERE START MY FUNCTIONS---------------// HERE START MY FUNCTIONS	---------------// HERE START MY FUNCTIONS
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

// HERE END MY FUNCTIONS---------------// HERE END MY FUNCTIONS	---------------// HERE END MY FUNCTIONS

// NodeGroups returns all node groups configured for this cloud provider.
func (s *cloudProviderServer) NodeGroups(ctx context.Context, req *protos.NodeGroupsRequest) (*protos.NodeGroupsResponse, error) {

	// Send a GET request to the nodegroup controller
	client, err := newClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create a client: %v", err)
	}

	reply, err := client.Get("https://localhost:9009/nodegroup") // TODO create a parameter
	if err != nil {
		return nil, fmt.Errorf("failed to get nodegroup: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		return &protos.NodeGroupsResponse{}, nil // No error, but no data
	} else if reply.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var nodeGroups []NodegroupMinInfo
	if err := json.NewDecoder(reply.Body).Decode(&nodeGroups); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	// Convert the response to the protos format
	protosNodeGroups := make([]*protos.NodeGroup, len(nodeGroups))
	for i, nodegroup := range nodeGroups {
		protosNodeGroups[i] = &protos.NodeGroup{
			Id:      nodegroup.Id,
			MinSize: nodegroup.MinSize,
			MaxSize: nodegroup.MaxSize,
		}
	}
	log.Printf("NodeGroups: %v di ritorno per chiamata all", protosNodeGroups)
	// Return the response
	return &protos.NodeGroupsResponse{
		NodeGroups: protosNodeGroups,
	}, nil

}

// NodeGroupForNode returns the node group for the given node.
// The node group id is an empty string if the node should not
// be processed by cluster autoscaler.
func (c *cloudProviderServer) NodeGroupForNode(ctx context.Context, req *protos.NodeGroupForNodeRequest) (*protos.NodeGroupForNodeResponse, error) {

	//here TODO the real computations
	client, err := newClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create a client: %v", err)
	}

	// Take the parameter
	//TODO use name or provider id? Liqo 0.10 non assegna Provider ID
	nodeId := req.Node.Name
	url := fmt.Sprintf("https://localhost:9009/nodegroup/ownership?id=%s", nodeId)

	// Send a GET request to the nodegroup controller
	reply, err := client.Get(url) // TODO create a better parameter, maybe using something more complex like DefaultClient
	if err != nil {
		return nil, fmt.Errorf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		return &protos.NodeGroupForNodeResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var nodeGroup NodegroupMinInfo
	if err := json.NewDecoder(reply.Body).Decode(&nodeGroup); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	// Convert the response to the protos format
	protosNodeGroup := &protos.NodeGroup{
		Id:      nodeGroup.Id,
		MinSize: nodeGroup.MinSize,
		MaxSize: nodeGroup.MaxSize,
	}
	log.Printf("NodeGroupForNode: %v di ritorno per chiamata get nodegroup of the node", protosNodeGroup)

	// Return the response
	return &protos.NodeGroupForNodeResponse{
		NodeGroup: protosNodeGroup,
	}, nil
}

// PricingNodePrice returns a theoretical minimum price of running a node for
// a given period of time on a perfectly matching machine.
// Implementation optional: if unimplemented return error code 12 (for `Unimplemented`)
func (c *cloudProviderServer) PricingNodePrice(ctx context.Context, req *protos.PricingNodePriceRequest) (*protos.PricingNodePriceResponse, error) {
	//here TODO the real computations
	return nil, status.Error(codes.Unimplemented, "function PricingNodePrice is Unimplemented")
}

// PricingPodPrice returns a theoretical minimum price of running a pod for a given
// period of time on a perfectly matching machine.
// Implementation optional: if unimplemented return error code 12 (for `Unimplemented`)
func (c *cloudProviderServer) PricingPodPrice(ctx context.Context, req *protos.PricingPodPriceRequest) (*protos.PricingPodPriceResponse, error) {
	//here TODO the real computations
	return nil, status.Error(codes.Unimplemented, "function PricingPodPrice is Unimplemented")
}

// GPULabel returns the label added to nodes with GPU resource.
func (c *cloudProviderServer) GPULabel(ctx context.Context, req *protos.GPULabelRequest) (*protos.GPULabelResponse, error) {

	//here TODO the real computations

	client, err := newClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create a client: %v", err)
	}

	// Send a GET request to the nodegroup controller
	reply, err := client.Get("https://localhost:9009/gpu/label") // TODO create a parameter
	if err != nil {
		return nil, fmt.Errorf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		return &protos.GPULabelResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var gpuLabel string
	if err := json.NewDecoder(reply.Body).Decode(&gpuLabel); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return &protos.GPULabelResponse{
		Label: gpuLabel,
	}, nil
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (c *cloudProviderServer) GetAvailableGPUTypes(ctx context.Context, req *protos.GetAvailableGPUTypesRequest) (*protos.GetAvailableGPUTypesResponse, error) {

	//here TODO the real computations

	client, err := newClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create a client: %v", err)
	}

	// Send a GET request to the nodegroup controller
	reply, err := client.Get("https://localhost:9009/gpu/types") // TODO create a parameter
	if err != nil {
		return nil, fmt.Errorf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		return &protos.GetAvailableGPUTypesResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var gpuLabels []string
	if err := json.NewDecoder(reply.Body).Decode(&gpuLabels); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	// Convert the response to the protos format
	mapGpu := map[string]*anypb.Any{}
	for _, label := range gpuLabels {
		mapGpu[label] = &anypb.Any{}
	}

	return &protos.GetAvailableGPUTypesResponse{
		GpuTypes: mapGpu,
	}, nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (c *cloudProviderServer) Cleanup(ctx context.Context, req *protos.CleanupRequest) (*protos.CleanupResponse, error) {
	//here TODO the real computations
	return &protos.CleanupResponse{}, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// TODO Make a new calls on nodegroup functions, otherwise will fail after a while
func (c *cloudProviderServer) Refresh(ctx context.Context, req *protos.RefreshRequest) (*protos.RefreshResponse, error) {
	//here TODO the real computations
	return &protos.RefreshResponse{}, nil
}

// NodeGroupTargetSize returns the current target size of the node group. It is possible
// that the number of nodes in Kubernetes is different at the moment but should be equal
// to the size of a node group once everything stabilizes (new nodes finish startup and
// registration or removed nodes are deleted completely).
func (c *cloudProviderServer) NodeGroupTargetSize(ctx context.Context, req *protos.NodeGroupTargetSizeRequest) (*protos.NodeGroupTargetSizeResponse, error) {

	//here TODO the real computations

	client, err := newClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create a client: %v", err)
	}

	// Take the parameter
	nodeId := req.Id
	url := fmt.Sprintf("https://localhost:9009/nodegroup/current-size?id=%s", nodeId)

	// Send a GET request to the nodegroup controller
	reply, err := client.Get(url) // TODO create a parameter
	if err != nil {
		return nil, fmt.Errorf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		return &protos.NodeGroupTargetSizeResponse{}, nil //TODO probably there is a specific error
	} else if reply.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var currentSize NodegroupCurrentSize
	if err := json.NewDecoder(reply.Body).Decode(&currentSize); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}
	log.Printf("NodeGroupTargetSize: %v di ritorno per chiamata get current size of the nodegroup", currentSize)

	return &protos.NodeGroupTargetSizeResponse{
		TargetSize: currentSize.CurrentSize,
	}, nil
}

// NodeGroupIncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use NodeGroupDeleteNodes. This function should wait until
// node group size is updated.
func (c *cloudProviderServer) NodeGroupIncreaseSize(ctx context.Context, req *protos.NodeGroupIncreaseSizeRequest) (*protos.NodeGroupIncreaseSizeResponse, error) {

	//here TODO the real computations

	//  TODO change the get with a post
	if flag == 0 {
		flag = 1
		client, err := newClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create a client: %v", err)
		}
		// Take the parameter
		nodegroupId := req.Id
		log.Printf("l'id è %s e %s", nodegroupId, req.Id)
		url := fmt.Sprintf("https://localhost:9009/nodegroup/scaleup?id=%s", nodegroupId)

		// Send a GET request to the nodegroup controller
		reply, err := client.Get(url) // TODO create a parameter
		if err != nil {
			return nil, fmt.Errorf("failed to execute get query: %v", err)
		}
		defer reply.Body.Close()

		// Check the response status code
		if reply.StatusCode == http.StatusNotFound {
			return &protos.NodeGroupIncreaseSizeResponse{}, nil //TODO probably there is a specific error
		} else if reply.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("server responded with status %d", reply.StatusCode)
		}
		flag = 0
	}
	return &protos.NodeGroupIncreaseSizeResponse{}, nil
}

// NodeGroupDeleteNodes deletes nodes from this node group (and also decreasing the size
// of the node group with that). Error is returned either on failure or if the given node
// doesn't belong to this node group. This function should wait until node group size is updated.
func (c *cloudProviderServer) NodeGroupDeleteNodes(ctx context.Context, req *protos.NodeGroupDeleteNodesRequest) (*protos.NodeGroupDeleteNodesResponse, error) {

	//here TODO the real computations

	//  TODO change the get with a post
	if flag == 0 {
		flag = 1
		client, err := newClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create a client: %v", err)
		}
		// Take the parameter
		nodeId := req.Nodes[0].ProviderID
		nodegroupId := req.Id
		log.Printf("l'id è %s e %s", nodegroupId, req.Id)
		url := fmt.Sprintf("https://localhost:9009/nodegroup/scaledown?id=%s&nodegroupid=%s", nodeId, nodegroupId)

		// Send a GET request to the nodegroup controller
		reply, err := client.Get(url) // TODO create a parameter
		if err != nil {
			return nil, fmt.Errorf("failed to execute get query: %v", err)
		}
		defer reply.Body.Close()

		// Check the response status code
		if reply.StatusCode == http.StatusNotFound {
			return &protos.NodeGroupDeleteNodesResponse{}, nil //TODO probably there is a specific error
		} else if reply.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("server responded with status %d", reply.StatusCode)
		}
		flag = 0
	}
	return &protos.NodeGroupDeleteNodesResponse{}, nil
}

// NodeGroupDecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the request
// for new nodes that have not been yet fulfilled. Delta should be negative. It is assumed
// that cloud provider will not delete the existing nodes if the size when there is an option
// to just decrease the target.
func (c *cloudProviderServer) NodeGroupDecreaseTargetSize(ctx context.Context, req *protos.NodeGroupDecreaseTargetSizeRequest) (*protos.NodeGroupDecreaseTargetSizeResponse, error) {
	//here TODO the real computations
	log.Printf("decrease SIZE target why? of %s", req.Id)
	return &protos.NodeGroupDecreaseTargetSizeResponse{}, nil
}

// NodeGroupNodes returns a list of all nodes that belong to this node group.
func (c *cloudProviderServer) NodeGroupNodes(ctx context.Context, req *protos.NodeGroupNodesRequest) (*protos.NodeGroupNodesResponse, error) {

	//here TODO the real computations

	client, err := newClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create a client: %v", err)
	}
	// Take the parameter
	nodegroupId := req.Id
	url := fmt.Sprintf("https://localhost:9009/nodegroup/nodes?id=%s", nodegroupId)

	// Send a GET request to the nodegroup controller
	reply, err := client.Get(url) // TODO create a parameter
	if err != nil {
		return nil, fmt.Errorf("failed to execute get query: %v", err)
	}
	defer reply.Body.Close()

	// Check the response status code
	if reply.StatusCode == http.StatusNotFound {
		return &protos.NodeGroupNodesResponse{}, fmt.Errorf("node with id %s doesn't exist", req.Id)
	} else if reply.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server responded with status %d", reply.StatusCode)
	}

	// Decode the JSON response
	var nodeList []NodeMinInfo
	if err := json.NewDecoder(reply.Body).Decode(&nodeList); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}
	log.Printf("NodeGroupNodes: %d lunghezza lista", len(nodeList))

	// Convert the response to the protos format
	protosNodes := make([]*protos.Instance, len(nodeList))
	for i, node := range nodeList {
		protosNodes[i] = &protos.Instance{
			Id: node.Id,
			Status: &protos.InstanceStatus{
				InstanceState: 1, //protos.InstanceStatus_InstanceState(node.InstanceStatus.InstanceState),
				//ErrorInfo: &protos.InstanceErrorInfo{
				//	ErrorMessage: "", //node.InstanceStatus.InstanceErrorInfo,
				//},
			},
		}
	}
	log.Printf("nodegroupNodes: %v di ritorno per chiamata get nodes of the nodegroup", protosNodes)

	return &protos.NodeGroupNodesResponse{
		Instances: protosNodes,
	}, nil
}

// NodeGroupTemplateNodeInfo returns a structure of an empty (as if just started) node,
// with all of the labels, capacity and allocatable information. This will be used in
// scale-up simulations to predict what would a new node look like if a node group was expanded.
// Implementation optional: if unimplemented return error code 12 (for `Unimplemented`)
func (c *cloudProviderServer) NodeGroupTemplateNodeInfo(ctx context.Context, req *protos.NodeGroupTemplateNodeInfoRequest) (*protos.NodeGroupTemplateNodeInfoResponse, error) {
	log.Printf("ASKS ABOUT TEMPLATE NODE INFO")
	//TODO write a rael functions in nodegroup manager to return the real info

	return &protos.NodeGroupTemplateNodeInfoResponse{
		NodeInfo: &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fake-node",
			},
			Status: v1.NodeStatus{
				Capacity: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("4Gi"),
					v1.ResourcePods:   resource.MustParse("110"),
				},
				Allocatable: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("4Gi"),
					v1.ResourcePods:   resource.MustParse("110"),
				},
			},
		},
	}, nil

	//return nil, status.Error(codes.Unimplemented, "function NodeGroupTemplateNodeInfo is Unimplemented")
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup.
// Implementation optional: if unimplemented return error code 12 (for `Unimplemented`)
func (c *cloudProviderServer) NodeGroupGetOptions(ctx context.Context, req *protos.NodeGroupAutoscalingOptionsRequest) (*protos.NodeGroupAutoscalingOptionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "function NodeGroupGetOptions is Unimplemented")
}

func main() {

	lis, err := net.Listen("tcp", ":9007")
	if err != nil {
		log.Fatalf("failed to start server, %v ", err)
	}

	server := grpc.NewServer()
	service := &cloudProviderServer{}

	protos.RegisterCloudProviderServer(server, service)

	server.Serve(lis)

}
