# UCMDB REST API Client

This package provides a Go client for the OpenText UCMDB (Universal Discovery and Configuration Management Database) REST API.

## API Reference

The client implements endpoints from the OpenText UCMDB REST API:

- **Documentation**: [UCMDB 25.4 REST API](https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1)
- **OpenAPI Specification**: [OpenAPI 3.0 JSON](https://docs.microfocus.com/artifacts/UCMDB/25.4/UCMDB_REST_API_v1.json)
- **OpenText Docs Portal**: [https://docs.opentext.com/](https://docs.opentext.com/)

## Supported UCMDB Versions

- UCMDB 25.4 (standard deployment)
- UCMDB 26.1+ (containerized - requires `/ucmdb-server/rest-api` URL prefix)

## Usage

### Creating a Client

```go
import "github.com/steadyjaw/terraform-provider-ucmdb/client"

ctx := context.Background()
c, err := client.NewClient(ctx, "https://ucmdb.example.com:8080", "username", "password")
if err != nil {
    panic(err)
}
```

### Creating Configuration Items (CIs)

```go
td := &client.TopologyData{
    CIS: []client.DataInConfigurationItem{
        {
            UcmdbId:    "ci-1",
            Type:       "Host",
            Properties: map[string]interface{}{
                "name": "server1",
                "description": "Production server",
            },
        },
    },
}

result, err := c.Ucmdb.CreateDataModel(ctx, td)
if err != nil {
    panic(err)
}
```

### Retrieving a CI

```go
ci, err := c.Ucmdb.GetConfigurationItem(ctx, "ci-1")
if err != nil {
    panic(err)
}

fmt.Printf("CI Type: %s\n", ci.Type)
fmt.Printf("CI Properties: %v\n", ci.Properties)
```

### Updating a CI

```go
dic := client.DataInConfigurationItem{
    UcmdbId:    "ci-1",
    Type:       "Host",
    Properties: map[string]interface{}{
        "name": "server1-updated",
    },
}

result, err := c.Ucmdb.UpdateConfigurationItem(ctx, "ci-1", dic)
if err != nil {
    panic(err)
}
```

### Deleting a CI

```go
_, err := c.Ucmdb.DeleteConfigurationItem(ctx, "ci-1")
if err != nil {
    panic(err)
}
```

### Executing Topology Queries

```go
query := &client.TopologyQuery{
    Nodes: []client.Node{
        {
            Type:            "Host",
            QueryIdentifier: "Host",
            Visible:         true,
            IncludeSubtypes: true,
            Layout:          []string{"name"},
            AttributeConditions: []client.AttributeConditions{
                {
                    Attribute: "name",
                    Operator:  "in",
                    Value:     []string{"server1", "server2"},
                },
            },
        },
    },
}

result, err := c.Ucmdb.ExecuteQuery(ctx, query)
if err != nil {
    panic(err)
}

for _, ci := range result.CIS {
    fmt.Printf("Found CI: %s (%s)\n", ci.UcmdbId, ci.Type)
}
```

### Creating Relationships

```go
relation := &client.Relation{
    Id1:    "ci-1",
    Id2:    "ci-2",
    Type:   "running_on",
    Properties: map[string]interface{}{
        // Optional relationship properties
    },
}

created, err := c.Ucmdb.CreateRelation(ctx, relation)
if err != nil {
    panic(err)
}
```

### Retrieving Relationships

```go
relation, err := c.Ucmdb.GetRelation(ctx, "relation-id")
if err != nil {
    panic(err)
}

fmt.Printf("Relation: %s -> %s (type: %s)\n", relation.Id1, relation.Id2, relation.Type)
```

### Updating Relationships

```go
relation.Properties = map[string]interface{}{
    "status": "active",
}

updated, err := c.Ucmdb.UpdateRelation(ctx, "relation-id", relation)
if err != nil {
    panic(err)
}
```

### Deleting Relationships

```go
err := c.Ucmdb.DeleteRelation(ctx, "relation-id")
if err != nil {
    panic(err)
}
```

## Authentication

The client uses HTTP Basic Authentication as specified in the UCMDB REST API documentation. Credentials are automatically encoded and sent with each request.

## Error Handling

All API calls return standard Go `error` interfaces. HTTP errors are wrapped with meaningful error messages indicating the HTTP status code and response body.

```go
if err != nil {
    fmt.Printf("API Error: %v\n", err)
}
```

## API Endpoints Summary

| Method | Endpoint | Function |
| ------ | -------- | -------- |
| POST | `/dataModel` | Create CIs |
| GET | `/dataModel/ci/{id}` | Retrieve CI details |
| PUT | `/dataModel/ci` | Update CIs |
| DELETE | `/dataModel/ci/{id}` | Delete CI |
| POST | `/dataModel/relation` | Create relationship |
| GET | `/dataModel/relation/{id}` | Retrieve relationship details |
| PUT | `/dataModel/relation/{id}` | Update relationship |
| DELETE | `/dataModel/relation/{id}` | Delete relationship |
| POST | `/topology` | Execute queries |

## Reference

For detailed information about each endpoint, parameters, and response formats, refer to:

- [UCMDB 25.4 REST API Documentation](https://docs.microfocus.com/api/Universal_Discovery_and_CMDB/25.4/UCMDB%20REST%20API/v1)
- [OpenAPI Specification (OAS 3.0)](https://docs.microfocus.com/artifacts/UCMDB/25.4/UCMDB_REST_API_v1.json)
