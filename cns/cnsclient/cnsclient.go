package cnsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Azure/azure-container-networking/cns"
	"github.com/Azure/azure-container-networking/log"
)

// CNSClient specifies a client to connect to Ipam Plugin.
type CNSClient struct {
	connectionURL string
}

const (
	defaultCnsURL = "http://localhost:10090"
)

// NewCnsClient create a new cns client.
func NewCnsClient(url string) (*CNSClient, error) {
	if url == "" {
		url = defaultCnsURL
	}

	return &CNSClient{
		connectionURL: url,
	}, nil
}

// GetNetworkConfiguration Request to get network config.
func (cnsClient *CNSClient) GetNetworkConfiguration(orchestratorContext []byte) (*cns.GetNetworkContainerResponse, error) {
	var body bytes.Buffer

	httpc := &http.Client{}
	url := cnsClient.connectionURL + cns.GetNetworkContainerByOrchestratorContext
	log.Printf("GetNetworkConfiguration url %v", url)

	payload := &cns.GetNetworkContainerRequest{
		OrchestratorContext: orchestratorContext,
	}

	err := json.NewEncoder(&body).Encode(payload)
	if err != nil {
		log.Errorf("encoding json failed with %v", err)
		return nil, err
	}

	res, err := httpc.Post(url, "application/json", &body)
	if err != nil {
		log.Errorf("[Azure CNSClient] HTTP Post returned error %v", err.Error())
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("[Azure CNSClient] GetNetworkConfiguration invalid http status code: %v", res.StatusCode)
		log.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	var resp cns.GetNetworkContainerResponse

	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		log.Errorf("[Azure CNSClient] Error received while parsing GetNetworkConfiguration response resp:%v err:%v", res.Body, err.Error())
		return nil, err
	}

	if resp.Response.ReturnCode != 0 {
		log.Errorf("[Azure CNSClient] GetNetworkConfiguration received error response :%v", resp.Response.Message)
		return nil, fmt.Errorf(resp.Response.Message)
	}

	return &resp, nil
}

// CreateApipaEndpoint creates an endpoint in APIPA network for host container connectivity.
func (cnsClient *CNSClient) CreateApipaEndpoint(podName, podNamespace string /*orchestratorContext []byte*/) (*cns.CreateApipaEndpointResponse, error) {
	var body bytes.Buffer

	httpc := &http.Client{}
	url := cnsClient.connectionURL + cns.CreateApipaEndpointPath
	log.Printf("CreateApipaEndpoint url: %v", url)

	podInfo := cns.KubernetesPodInfo{PodName: podName, PodNamespace: podNamespace}
	orchestratorContext, err := json.Marshal(podInfo)
	if err != nil {
		log.Printf("Failed to marshall podInfo for orchestrator context due to error: %v", err)
		return nil, err
	}

	// What can be used here?
	payload := &cns.CreateApipaEndpointRequest{
		OrchestratorContext: orchestratorContext,
	}

	err = json.NewEncoder(&body).Encode(payload)
	if err != nil {
		log.Errorf("encoding json failed with %v", err)
		return nil, err
	}

	log.Printf("CreateApipaEndpoint posting body: %v", body)
	res, err := httpc.Post(url, "application/json", &body)
	if err != nil {
		log.Errorf("[Azure CNSClient] HTTP Post returned error %v", err.Error())
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("[Azure CNSClient] CreateEndpointForHostContainerConnectivity: Invalid http status code: %v",
			res.StatusCode)
		log.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	var resp cns.CreateApipaEndpointResponse

	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		log.Errorf("[Azure CNSClient] Error parsing CreateEndpointForHostContainerConnectivity response resp: %v err: %v",
			res.Body, err.Error())
		return nil, err
	}

	if resp.Response.ReturnCode != 0 {
		log.Errorf("[Azure CNSClient] CreateEndpointForHostContainerConnectivity received error response :%v", resp.Response.Message)
		return nil, fmt.Errorf(resp.Response.Message)
	}

	return &resp, nil
}

// DeleteApipaEndpoint deletes the endpoint in APIPA network created for host container connectivity.
func (cnsClient *CNSClient) DeleteApipaEndpoint(endpointID string) error {
	var body bytes.Buffer

	// TODO: Move this to create a reusable http client.
	httpc := &http.Client{}
	url := cnsClient.connectionURL + cns.DeleteApipaEndpointPath
	log.Printf("DeleteApipaEndpoint url: %v", url)

	payload := &cns.DeleteApipaEndpointRequest{
		EndpointID: endpointID,
	}

	err := json.NewEncoder(&body).Encode(payload)
	if err != nil {
		log.Errorf("encoding json failed with %v", err)
		return err
	}

	log.Printf("DeleteApipaEndpoint posting body: %v", body)
	res, err := httpc.Post(url, "application/json", &body)
	if err != nil {
		log.Errorf("[Azure CNSClient] HTTP Post returned error %v", err.Error())
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("[Azure CNSClient] DeleteApipaEndpoint: Invalid http status code: %v",
			res.StatusCode)
		log.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	var resp cns.DeleteApipaEndpointResponse

	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		log.Errorf("[Azure CNSClient] Error parsing DeleteApipaEndpoint response resp: %v err: %v",
			res.Body, err.Error())
		return err
	}

	if resp.Response.ReturnCode != 0 {
		log.Errorf("[Azure CNSClient] DeleteApipaEndpoint received error response :%v", resp.Response.Message)
		return fmt.Errorf(resp.Response.Message)
	}

	return nil
}
