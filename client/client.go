package client

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NewClient creates a new UCMDB REST API client.
//
// The client automatically handles HTTP Basic Authentication and content negotiation.
// For more information about UCMDB REST API authentication, see:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
func NewClient(ctx context.Context, baseURL, username, password string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	client := &Client{
		BaseURL:    baseURL,
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}

	client.Ucmdb = &UcmdbService{client: client}

	return client, nil
}

// CreateDataModel creates one or more CIs in UCMDB.
//
// This method calls the POST /dataModel endpoint documented in the OpenAPI spec:
// https://docs.microfocus.com/artifacts/UCMDB/25.4/UCMDB_REST_API_v1.json
func (s *UcmdbService) CreateDataModel(ctx context.Context, td *TopologyData) (*TopologyData, error) {
	return s.submitTopologyData(ctx, "/ucmdb-rest-svc/rest/topology/create", td)
}

// GetConfigurationItem retrieves a specific CI by its UCMDB ID.
//
// This method calls the GET /dataModel/ci/{id} endpoint as documented in:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
func (s *UcmdbService) GetConfigurationItem(ctx context.Context, ucmdbID string) (*ConfigurationItem, error) {
	if ucmdbID == "" {
		return nil, fmt.Errorf("ucmdbID cannot be empty")
	}

	url := fmt.Sprintf("%s/ucmdb-rest-svc/rest/topology/ci/%s", s.client.BaseURL, ucmdbID)
	
	resp, err := s.client.makeRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get CI: %w", err)
	}

	var ci ConfigurationItem
	if err := json.Unmarshal(resp, &ci); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CI response: %w", err)
	}

	return &ci, nil
}

// UpdateConfigurationItem updates an existing CI.
//
// This method calls the PUT /dataModel/ci endpoint as documented in:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
func (s *UcmdbService) UpdateConfigurationItem(ctx context.Context, ucmdbID string, ci DataInConfigurationItem) (*TopologyData, error) {
	if ucmdbID == "" {
		return nil, fmt.Errorf("ucmdbID cannot be empty")
	}

	td := &TopologyData{
		CIS: []DataInConfigurationItem{ci},
	}

	return s.submitTopologyData(ctx, "/ucmdb-rest-svc/rest/topology/update", td)
}

// DeleteConfigurationItem deletes a CI by its ID.
//
// This method calls the DELETE /dataModel/ci/{id} endpoint as documented in:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
func (s *UcmdbService) DeleteConfigurationItem(ctx context.Context, ucmdbID string) (*TopologyData, error) {
	if ucmdbID == "" {
		return nil, fmt.Errorf("ucmdbID cannot be empty")
	}

	url := fmt.Sprintf("%s/ucmdb-rest-svc/rest/topology/ci/%s", s.client.BaseURL, ucmdbID)
	
	_, err := s.client.makeRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to delete CI: %w", err)
	}

	return &TopologyData{}, nil
}

// ExecuteQuery executes a topology query and returns matching CIs.
//
// This method calls the POST /topology endpoint as documented in:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
func (s *UcmdbService) ExecuteQuery(ctx context.Context, query *TopologyQuery) (*TopologyData, error) {
	if query == nil {
		return nil, fmt.Errorf("query cannot be nil")
	}

	url := fmt.Sprintf("%s/ucmdb-rest-svc/rest/topology/query", s.client.BaseURL)
	
	payload, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	resp, err := s.client.makeRequest(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var td TopologyData
	if err := json.Unmarshal(resp, &td); err != nil {
		return nil, fmt.Errorf("failed to unmarshal query response: %w", err)
	}

	return &td, nil
}

// CreateRelation creates a relationship between two CIs.
//
// This method calls the POST /dataModel/relation endpoint as documented in:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
func (s *UcmdbService) CreateRelation(ctx context.Context, relation *Relation) (*Relation, error) {
	if relation == nil {
		return nil, fmt.Errorf("relation cannot be nil")
	}

	url := fmt.Sprintf("%s/ucmdb-rest-svc/rest/dataModel/relation", s.client.BaseURL)
	
	payload, err := json.Marshal(relation)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal relation: %w", err)
	}

	resp, err := s.client.makeRequest(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create relation: %w", err)
	}

	var result Relation
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relation response: %w", err)
	}

	return &result, nil
}

// GetRelation retrieves a relationship by its ID.
//
// This method calls the GET /dataModel/relation/{id} endpoint as documented in:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
func (s *UcmdbService) GetRelation(ctx context.Context, relationID string) (*Relation, error) {
	if relationID == "" {
		return nil, fmt.Errorf("relationID cannot be empty")
	}

	url := fmt.Sprintf("%s/ucmdb-rest-svc/rest/dataModel/relation/%s", s.client.BaseURL, relationID)
	
	resp, err := s.client.makeRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get relation: %w", err)
	}

	var relation Relation
	if err := json.Unmarshal(resp, &relation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relation response: %w", err)
	}

	return &relation, nil
}

// UpdateRelation updates an existing relationship.
//
// This method calls the PUT /dataModel/relation/{id} endpoint as documented in:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
func (s *UcmdbService) UpdateRelation(ctx context.Context, relationID string, relation *Relation) (*Relation, error) {
	if relationID == "" {
		return nil, fmt.Errorf("relationID cannot be empty")
	}
	if relation == nil {
		return nil, fmt.Errorf("relation cannot be nil")
	}

	url := fmt.Sprintf("%s/ucmdb-rest-svc/rest/dataModel/relation/%s", s.client.BaseURL, relationID)
	
	payload, err := json.Marshal(relation)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal relation: %w", err)
	}

	resp, err := s.client.makeRequest(ctx, http.MethodPut, url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to update relation: %w", err)
	}

	var result Relation
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relation response: %w", err)
	}

	return &result, nil
}

// DeleteRelation deletes a relationship.
//
// This method calls the DELETE /dataModel/relation/{id} endpoint as documented in:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
func (s *UcmdbService) DeleteRelation(ctx context.Context, relationID string) error {
	if relationID == "" {
		return fmt.Errorf("relationID cannot be empty")
	}

	url := fmt.Sprintf("%s/ucmdb-rest-svc/rest/dataModel/relation/%s", s.client.BaseURL, relationID)
	
	_, err := s.client.makeRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}

	return nil
}

// submitTopologyData submits topology data to the specified UCMDB endpoint
func (s *UcmdbService) submitTopologyData(ctx context.Context, endpoint string, td *TopologyData) (*TopologyData, error) {
	url := fmt.Sprintf("%s%s", s.client.BaseURL, endpoint)
	
	payload, err := json.Marshal(td)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal topology data: %w", err)
	}

	resp, err := s.client.makeRequest(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to submit topology data: %w", err)
	}

	var result TopologyData
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// makeRequest performs an HTTP request to UCMDB with proper authentication.
//
// All requests use HTTP Basic Authentication as specified in:
// https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
//
// Note: For containerized UCMDB (version 26.1+), the endpoint URLs should include
// the /ucmdb-server/rest-api prefix.
func (c *Client) makeRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	auth := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	httpResp, err := c.HTTPClient.(*http.Client).Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("UCMDB API error: HTTP %d - %s", httpResp.StatusCode, string(respBody))
	}

	return respBody, nil
}
