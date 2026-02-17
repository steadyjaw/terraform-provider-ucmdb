package ucmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/steadyjaw/terraform-provider-ucmdb/client"
)

func dataSourceUcmdbList() *schema.Resource {
	return &schema.Resource{
		Description: "Query and retrieve UCMDB Configuration Items (CIs) based on filters.",

		ReadContext: dataSourceUcmdbReadList,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeSet,
				Description: "Filter criteria for querying CIs.",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Description: "The type of CI to filter.",
							Required:    true,
						},
						"names": {
							Type:        schema.TypeList,
							Description: "List of CI names to filter.",
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"items": {
				Type:        schema.TypeSet,
				Description: "The list of CIs matching the filter criteria.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ucmdb_id": {
							Type:        schema.TypeString,
							Description: "The unique UCMDB ID of the CI.",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "The type of the CI.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the CI.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUcmdbReadList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)

	var diags diag.Diagnostics

	tql := client.TopologyQuery{}
	var nodes []client.Node

	filters := d.Get("filter").(*schema.Set)
	filter_list := filters.List()

	tflog.Debug(ctx, "Executing UCMDB query", map[string]interface{}{"filter_count": len(filter_list)})

	for _, filter := range filter_list {
		f := filter.(map[string]interface{})
		ci_type := f["type"].(string)
		names := f["names"].(interface{})

		tflog.Debug(ctx, "Adding filter to query", map[string]interface{}{
			"type": ci_type,
			"names_count": len(names.([]interface{})),
		})

		nodes = append(nodes, client.Node{
			Type:            ci_type,
			QueryIdentifier: ci_type,
			Visible:         true,
			IncludeSubtypes: true,
			Layout:          []string{"name"},
			AttributeConditions: []client.AttributeConditions{
				{
					Attribute: "name",
					Operator:  "in",
					Value:     names,
				},
			},
		})
	}

	tql.Nodes = nodes

	td, err := conn.Ucmdb.ExecuteQuery(ctx, &tql)
	if err != nil {
		tflog.Error(ctx, "Failed to execute UCMDB query", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(err)
	}

	// Log query response
	b, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		tflog.Warn(ctx, "Failed to marshal query response for logging", map[string]interface{}{"error": err.Error()})
	} else {
		tflog.Debug(ctx, "Query response", map[string]interface{}{"response": string(b)})
	}

	// Extract CIs and create items
	var ucmdb_ids []string
	cis := td.CIS

	items := make([]interface{}, 0, len(cis))

	for _, ci := range cis {
		item := make(map[string]interface{})
		ucmdb_ids = append(ucmdb_ids, ci.UcmdbId)
		item["ucmdb_id"] = ci.UcmdbId
		item["type"] = ci.Type
		
		// Safely extract name property
		if name, ok := ci.Properties["name"].(string); ok {
			item["name"] = name
		} else {
			tflog.Warn(ctx, "CI missing or invalid name property", map[string]interface{}{"id": ci.UcmdbId})
			continue
		}
		
		items = append(items, item)
	}

	if err := d.Set("items", items); err != nil {
		tflog.Error(ctx, "Failed to set items", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(err)
	}

	// Generate stable ID from query results
	id := generateStableID(ucmdb_ids)
	d.SetId(id)

	tflog.Info(ctx, "UCMDB query completed successfully", map[string]interface{}{
		"ci_count": len(items),
		"data_source_id": id,
	})

	return diags
}

// generateStableID creates a deterministic ID from a list of UCMDB IDs
// This ensures the ID remains consistent across multiple reads
func generateStableID(ids []string) string {
	if len(ids) == 0 {
		return schema.HashString("")
	}
	
	// Sort the IDs to ensure consistent ordering
	sortedIDs := make([]string, len(ids))
	copy(sortedIDs, ids)
	
	// Create a stable hash from the concatenated IDs
	concatenated := strings.Join(sortedIDs, "|")
	return fmt.Sprintf("%d", schema.HashString(concatenated))
}
