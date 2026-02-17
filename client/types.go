// Package client provides a Go client for the OpenText UCMDB REST API.
//
// This client implements the OpenText Universal Discovery and CMDB REST API.
// For the latest API documentation and OpenAPI specification, see:
//
// - UCMDB 25.4 REST API: https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
// - OpenAPI Specification (OAS 3.0): https://docs.microfocus.com/artifacts/UCMDB/25.4/UCMDB_REST_API_v1.json
// - OpenText Documentation Portal: https://docs.opentext.com/
//
// Key endpoints used by the Terraform provider:
// - POST /dataModel - Create configuration items
// - GET /dataModel/ci/{id} - Retrieve CI details
// - PUT /dataModel/ci - Update CIs
// - DELETE /dataModel/ci/{id} - Delete CI
// - POST /topology - Execute topology queries
package client

import "net/http"

// Client represents an UCMDB REST API client
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
	Ucmdb      *UcmdbService
}

// UcmdbService provides methods for UCMDB operations
type UcmdbService struct {
	client *Client
}

// TopologyData represents a topology data model for CI operations
type TopologyData struct {
	CIS        []DataInConfigurationItem `json:"cis"`
	Relations  []interface{}             `json:"relations,omitempty"`
	AddedCis   []string                  `json:"addedCis,omitempty"`
	UpdatedCis []string                  `json:"updatedCis,omitempty"`
	IgnoredCis []string                  `json:"ignoredCis,omitempty"`
}

// DataInConfigurationItem represents a Configuration Item (CI)
type DataInConfigurationItem struct {
	UcmdbId    string                 `json:"ucmdbId"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// ConfigurationItem represents a CI retrieved from UCMDB
type ConfigurationItem struct {
	UcmdbId    string                 `json:"ucmdbId"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// TopologyQuery represents a topology query for searching CIs
type TopologyQuery struct {
	Nodes []Node `json:"nodes"`
}

// Node represents a query node with filter criteria
type Node struct {
	Type                string                `json:"type"`
	QueryIdentifier     string                `json:"queryIdentifier"`
	Visible             bool                  `json:"visible"`
	IncludeSubtypes     bool                  `json:"includeSubtypes"`
	Layout              []string              `json:"layout,omitempty"`
	AttributeConditions []AttributeConditions `json:"attributeConditions,omitempty"`
}

// AttributeConditions represents filter conditions for a query
type AttributeConditions struct {
	Attribute string      `json:"attribute"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
}

// Relation represents a relationship between two CIs
// See: https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1
type Relation struct {
	Id1  string                 `json:"id1"`
	Id2  string                 `json:"id2"`
	Type string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// RelationData represents relation data for create/update operations
type RelationData struct {
	Relations []Relation `json:"relations"`
}

// Response represents a generic UCMDB API response
type Response struct {
	Status string      `json:"status"`
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

// QueryResult represents the result of a topology query
type QueryResult struct {
	CIS []ConfigurationItem `json:"cis"`
}
