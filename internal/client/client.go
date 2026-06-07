package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a Cockpit API client
type Client struct {
	Host         string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
}

// NewClient creates a new API client
func NewClient(host, clientID, clientSecret string) *Client {
	return &Client{
		Host:         host,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// doRequest performs an HTTP request and unmarshals the response into v
func (c *Client) doRequest(method, path string, body []byte, v interface{}) error {
	url := fmt.Sprintf("%s%s", c.Host, path)
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.ClientID != "" && c.ClientSecret != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", c.ClientID, c.ClientSecret))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if v != nil {
		if err := json.Unmarshal(respBody, v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Host models
type AddHostRequest struct {
	Address  string `json:"address"`
	Port     int    `json:"port,omitempty"`
	APIToken string `json:"api_token"`
	ParentID string `json:"parent_id"`
}

type HostResponse struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Status   string `json:"status"`
}

func (c *Client) AddHost(req *AddHostRequest) (*HostResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var resp HostResponse
	err = c.doRequest("POST", "/hosts", data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetHost(id string) (*HostResponse, error) {
	var resp HostResponse
	err := c.doRequest("GET", fmt.Sprintf("/hosts/%s", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteHost(id string) error {
	return c.doRequest("DELETE", fmt.Sprintf("/hosts/%s", id), nil, nil)
}

// VM models
type VMOS struct {
	Type    string `json:"type,omitempty"`
	Variant string `json:"variant,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

type VMDisk struct {
	Size   int64  `json:"size"`
	Pool   string `json:"pool"`
	Format string `json:"format,omitempty"`
	Bus    string `json:"bus,omitempty"`
}

type VMStorage struct {
	Disks []VMDisk `json:"disks,omitempty"`
	ISO   string   `json:"iso,omitempty"`
}

type VMNetwork struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	Model  string `json:"model,omitempty"`
}

type CreateVMRequest struct {
	HostID       string      `json:"host_id"`
	ClusterID    string      `json:"cluster_id,omitempty"`
	DatacenterID string      `json:"datacenter_id,omitempty"`
	Name         string      `json:"name"`
	Memory       int         `json:"memory"` // MB
	VCPUs        int         `json:"vcpus"`
	OS           *VMOS       `json:"os,omitempty"`
	Storage      *VMStorage  `json:"storage,omitempty"`
	Networks     []VMNetwork `json:"networks,omitempty"`
	Autostart    bool        `json:"autostart,omitempty"`
}

type VMResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Memory    int    `json:"memory"`
	VCPUs     int    `json:"vcpus"`
	HostID    string `json:"host_id"`
	Autostart bool   `json:"autostart"`
}

func (c *Client) CreateVM(req *CreateVMRequest) (*VMResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var resp VMResponse
	err = c.doRequest("POST", "/vms", data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetVM(id string) (*VMResponse, error) {
	var resp VMResponse
	err := c.doRequest("GET", fmt.Sprintf("/vms/%s", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteVM(id string) error {
	return c.doRequest("DELETE", fmt.Sprintf("/vms/%s", id), nil, nil)
}

// Kubernetes Cluster models
type K8sControlPlaneRequest struct {
	CPU           int      `json:"cpu"`
	MemoryMB      int      `json:"memory_mb"`
	StoragePoolID string   `json:"storage_pool_id"`
	StorageGB     int      `json:"storage_gb"`
	DesiredCount  int      `json:"desired_count"`
	HostIDs       []string `json:"host_ids,omitempty"`
}

type K8sWorkersRequest struct {
	CPU           int      `json:"cpu"`
	MemoryMB      int      `json:"memory_mb"`
	StoragePoolID string   `json:"storage_pool_id"`
	StorageGB     int      `json:"storage_gb"`
	DesiredCount  int      `json:"desired_count"`
	HostIDs       []string `json:"host_ids,omitempty"`
}

type K8sCSITiersRequest struct {
	LocalPathEnabled bool   `json:"local_path_enabled"`
	NFSEnabled       bool   `json:"nfs_enabled"`
	NFSServer        string `json:"nfs_server,omitempty"`
	NFSPath          string `json:"nfs_path,omitempty"`
	LonghornEnabled  bool   `json:"longhorn_enabled"`
	LonghornReplicas int    `json:"longhorn_replicas,omitempty"`
}

type K8sClusterCreateRequest struct {
	ControlPlane *K8sControlPlaneRequest `json:"control_plane"`
	Workers      *K8sWorkersRequest      `json:"workers"`
	CSITiers     *K8sCSITiersRequest     `json:"csi_tiers"`
	Name         string                  `json:"name"`
	Version      string                  `json:"version"`
	Distribution string                  `json:"distribution"`
	NetworkID    string                  `json:"network_id"`
	CNI          string                  `json:"cni"`
	SSHPublicKey string                  `json:"ssh_public_key"`
	ImagePath    string                  `json:"image_path,omitempty"`
}

type K8sNodePoolResponse struct {
	ID            string `json:"id"`
	ClusterID     string `json:"cluster_id"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	CPU           int    `json:"cpu"`
	MemoryMB      int    `json:"memory_mb"`
	StoragePoolID string `json:"storage_pool_id"`
	StorageGB     int    `json:"storage_gb"`
	DesiredCount  int    `json:"desired_count"`
}

type K8sNodeResponse struct {
	ID        string `json:"id"`
	ClusterID string `json:"cluster_id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	IPAddress string `json:"ip_address"`
}

type K8sClusterResponse struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Version      string                `json:"version"`
	Distribution string                `json:"distribution"`
	NetworkID    string                `json:"network_id"`
	CNI          string                `json:"cni"`
	Status       string                `json:"status"`
	SSHPublicKey string                `json:"ssh_public_key"`
	ImagePath    string                `json:"image_path"`
	NodePools    []K8sNodePoolResponse `json:"node_pools"`
	Nodes        []K8sNodeResponse     `json:"nodes"`
}

type ScalePoolRequest struct {
	DesiredCount int `json:"desired_count"`
}

type UpgradeClusterRequest struct {
	Version string `json:"version"`
}

func (c *Client) CreateK8sCluster(req *K8sClusterCreateRequest) (*K8sClusterResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var resp K8sClusterResponse
	err = c.doRequest("POST", "/kubernetes/clusters", data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetK8sCluster(id string) (*K8sClusterResponse, error) {
	var resp K8sClusterResponse
	err := c.doRequest("GET", fmt.Sprintf("/kubernetes/clusters/%s", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteK8sCluster(id string) error {
	return c.doRequest("DELETE", fmt.Sprintf("/kubernetes/clusters/%s", id), nil, nil)
}

func (c *Client) ScaleNodePool(clusterID, poolID string, desiredCount int) error {
	req := ScalePoolRequest{DesiredCount: desiredCount}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.doRequest("POST", fmt.Sprintf("/kubernetes/clusters/%s/pools/%s/scale", clusterID, poolID), data, nil)
}

func (c *Client) UpgradeK8sCluster(clusterID string, version string) error {
	req := UpgradeClusterRequest{Version: version}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.doRequest("POST", fmt.Sprintf("/kubernetes/clusters/%s/upgrade", clusterID), data, nil)
}

func (c *Client) GetK8sClusterKubeconfig(clusterID string) (string, error) {
	url := fmt.Sprintf("%s/kubernetes/clusters/%s/kubeconfig", c.Host, clusterID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	if c.ClientID != "" && c.ClientSecret != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", c.ClientID, c.ClientSecret))
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch kubeconfig (status %d): %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
