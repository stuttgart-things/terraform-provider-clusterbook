package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	URL        string
	HTTPClient *http.Client
}

func New(url string) *Client {
	return &Client{
		URL:        url,
		HTTPClient: &http.Client{},
	}
}

// Network types

type PoolInfo struct {
	NetworkKey string  `json:"NetworkKey"`
	Total      float64 `json:"Total"`
	Available  float64 `json:"Available"`
	Assigned   float64 `json:"Assigned"`
	Pending    float64 `json:"Pending"`
}

type IPEntry struct {
	IP      string `json:"IP"`
	Digit   string `json:"Digit"`
	Status  string `json:"Status"`
	Cluster string `json:"Cluster"`
}

type ClusterSummary struct {
	Cluster string `json:"cluster"`
	IPCount int    `json:"ip_count"`
}

type ClusterInfo struct {
	Cluster string       `json:"cluster"`
	IPs     []ClusterIP  `json:"ips"`
}

type ClusterIP struct {
	Network string `json:"network"`
	IP      string `json:"ip"`
	Status  string `json:"status"`
}

// GetNetworks returns all network pools.
func (c *Client) GetNetworks() ([]PoolInfo, error) {
	var pools []PoolInfo
	if err := c.get("/api/v1/networks", &pools); err != nil {
		return nil, err
	}
	return pools, nil
}

// GetNetworkIPs returns all IPs in a network.
func (c *Client) GetNetworkIPs(networkKey string) ([]IPEntry, error) {
	var entries []IPEntry
	if err := c.get(fmt.Sprintf("/api/v1/networks/%s/ips", networkKey), &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// CreateNetwork creates a new network pool.
func (c *Client) CreateNetwork(networkKey string, ipFrom, ipTo int) error {
	body := map[string]interface{}{
		"network": networkKey,
		"ip_from": ipFrom,
		"ip_to":   ipTo,
	}
	return c.post("/api/v1/networks", body, nil)
}

// DeleteNetwork deletes a network pool.
func (c *Client) DeleteNetwork(networkKey string) error {
	return c.delete(fmt.Sprintf("/api/v1/networks/%s", networkKey))
}

// AssignIP assigns a specific IP in a network to a cluster.
func (c *Client) AssignIP(networkKey, ip, cluster, status string, createDNS bool) (*AssignResponse, error) {
	body := map[string]interface{}{
		"ip":         ip,
		"cluster":    cluster,
		"status":     status,
		"create_dns": createDNS,
	}
	var resp AssignResponse
	if err := c.post(fmt.Sprintf("/api/v1/networks/%s/assign", networkKey), body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindAndAssignIP finds an available IP and assigns it to a cluster.
func (c *Client) FindAndAssignIP(networkKey, cluster, status string, createDNS bool) (string, error) {
	entries, err := c.GetNetworkIPs(networkKey)
	if err != nil {
		return "", fmt.Errorf("listing IPs: %w", err)
	}

	var availableIP string
	for _, e := range entries {
		if e.Status == "" && e.Cluster == "" {
			availableIP = e.IP
			break
		}
	}
	if availableIP == "" {
		return "", fmt.Errorf("no available IPs in network %s", networkKey)
	}

	_, err = c.AssignIP(networkKey, availableIP, cluster, status, createDNS)
	if err != nil {
		return "", err
	}
	return availableIP, nil
}

type AssignResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	IP      string `json:"ip"`
}

// ReleaseIP releases an IP assignment.
func (c *Client) ReleaseIP(networkKey, ip string) error {
	body := map[string]interface{}{
		"ip": ip,
	}
	return c.post(fmt.Sprintf("/api/v1/networks/%s/release", networkKey), body, nil)
}

// DeleteIP deletes an IP entry entirely.
func (c *Client) DeleteIP(networkKey, ip string) error {
	return c.delete(fmt.Sprintf("/api/v1/networks/%s/ips/%s", networkKey, ip))
}

// GetClusters returns all clusters with IP counts.
func (c *Client) GetClusters() ([]ClusterSummary, error) {
	var clusters []ClusterSummary
	if err := c.get("/api/v1/clusters", &clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}

// GetClusterInfo returns detailed IP info for a cluster.
func (c *Client) GetClusterInfo(name string) (*ClusterInfo, error) {
	var info ClusterInfo
	if err := c.get(fmt.Sprintf("/api/v1/clusters/%s", name), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// HTTP helpers

func (c *Client) get(path string, result interface{}) error {
	resp, err := c.HTTPClient.Get(c.URL + path)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s: status %d: %s", path, resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) post(path string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.URL+path, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("POST %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s: status %d: %s", path, resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) delete(path string) error {
	req, err := http.NewRequest("DELETE", c.URL+path, nil)
	if err != nil {
		return fmt.Errorf("DELETE %s: %w", path, err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DELETE %s: status %d: %s", path, resp.StatusCode, string(body))
	}
	return nil
}
