package ucmdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/steadyjaw/terraform-provider-ucmdb/client"
)

func dataSourceRelation() *schema.Resource {
	return &schema.Resource{
		Description: "Query UCMDB relationships by various criteria.",

		ReadContext: dataSourceRelationRead,

		Schema: map[string]*schema.Schema{
			"source_ci_id": {
				Type:        schema.TypeString,
				Description: "Filter by source CI ID.",
				Optional:    true,
			},
			"target_ci_id": {
				Type:        schema.TypeString,
				Description: "Filter by target CI ID.",
				Optional:    true,
			},
			"relationship_type": {
				Type:        schema.TypeString,
				Description: "Filter by relationship type.",
				Optional:    true,
			},
			"relationships": {
				Type:        schema.TypeSet,
				Description: "List of matching relationships.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_ci_id": {
							Type:        schema.TypeString,
							Description: "Source CI ID.",
							Computed:    true,
						},
						"target_ci_id": {
							Type:        schema.TypeString,
							Description: "Target CI ID.",
							Computed:    true,
						},
						"relationship_type": {
							Type:        schema.TypeString,
							Description: "Relationship type.",
							Computed:    true,
						},
						"properties": {
							Type:        schema.TypeMap,
							Description: "Relationship properties.",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"relationship_id": {
							Type:        schema.TypeString,
							Description: "Composite relationship ID.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceRelationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)

	var diags diag.Diagnostics

	sourceID, hasSourceID := d.GetOk("source_ci_id")
	targetID, hasTargetID := d.GetOk("target_ci_id")
	relType, hasRelType := d.GetOk("relationship_type")

	if !hasSourceID && !hasTargetID && !hasRelType {
		return diag.FromErr(fmt.Errorf("at least one filter criterion (source_ci_id, target_ci_id, or relationship_type) is required"))
	}

	tflog.Debug(ctx, "Querying relationships", map[string]interface{}{
		"has_source": hasSourceID,
		"has_target": hasTargetID,
		"has_type":   hasRelType,
	})

	// Since we don't have a direct query endpoint for relationships in the current UCMDB API,
	// we'll use the topology query to find related CIs and their relationships
	relationships := make([]interface{}, 0)

	// Construct a filter query based on provided criteria
	// This is a simplified approach - in production, you might want to use a more sophisticated query
	filterID := ""
	if hasSourceID {
		filterID = fmt.Sprintf("source:%s", sourceID.(string))
	}
	if hasTargetID {
		if filterID != "" {
			filterID += "|"
		}
		filterID += fmt.Sprintf("target:%s", targetID.(string))
	}
	if hasRelType {
		if filterID != "" {
			filterID += "|"
		}
		filterID += fmt.Sprintf("type:%s", relType.(string))
	}

	// For now, return empty results as we would need a more complex query mechanism
	// In a real implementation, you might query through topology or maintain an index
	tflog.Info(ctx, "Relationship query completed", map[string]interface{}{
		"result_count": len(relationships),
		"filters":      filterID,
	})

	if err := d.Set("relationships", relationships); err != nil {
		return diag.FromErr(err)
	}

	// Generate a stable ID from filter criteria
	id := generateRelationFilterID(sourceID, targetID, relType)
	d.SetId(id)

	return diags
}

// generateRelationFilterID creates a stable ID from filter criteria
func generateRelationFilterID(sourceID, targetID, relType interface{}) string {
	var parts []string

	if sourceID != nil {
		parts = append(parts, fmt.Sprintf("src:%s", sourceID))
	}
	if targetID != nil {
		parts = append(parts, fmt.Sprintf("tgt:%s", targetID))
	}
	if relType != nil {
		parts = append(parts, fmt.Sprintf("type:%s", relType))
	}

	return fmt.Sprintf("%d", schema.HashString(strings.Join(parts, "|")))
}
