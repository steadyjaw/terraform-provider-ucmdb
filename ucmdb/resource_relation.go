package ucmdb

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/steadyjaw/terraform-provider-ucmdb/client"
)

func resourceRelation() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a UCMDB relationship between two Configuration Items (CIs).",

		CreateContext: resourceRelationCreate,
		ReadContext:   resourceRelationRead,
		UpdateContext: resourceRelationUpdate,
		DeleteContext: resourceRelationDelete,

		Schema: map[string]*schema.Schema{
			"source_ci_id": {
				Type:         schema.TypeString,
				Description:  "The UCMDB ID of the source CI.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"target_ci_id": {
				Type:         schema.TypeString,
				Description:  "The UCMDB ID of the target CI.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"relationship_type": {
				Type:         schema.TypeString,
				Description:  "The type of relationship (e.g., 'running_on', 'depends_on').",
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"properties": {
				Type:        schema.TypeMap,
				Description: "Optional relationship properties.",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"last_updated": {
				Type:        schema.TypeString,
				Description: "Timestamp of the last update.",
				Optional:    true,
				Computed:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceRelationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)

	sourceID := d.Get("source_ci_id").(string)
	targetID := d.Get("target_ci_id").(string)
	relationType := d.Get("relationship_type").(string)

	tflog.Debug(ctx, "Creating relationship", map[string]interface{}{
		"source_ci_id": sourceID,
		"target_ci_id": targetID,
		"type":         relationType,
	})

	props := make(map[string]interface{})
	if propsRaw, ok := d.GetOk("properties"); ok {
		propsMap := propsRaw.(map[string]interface{})
		for k, v := range propsMap {
			props[k] = v
		}
	}

	relation := &client.Relation{
		Id1:        sourceID,
		Id2:        targetID,
		Type:       relationType,
		Properties: props,
	}

	created, err := conn.Ucmdb.CreateRelation(ctx, relation)
	if err != nil {
		tflog.Error(ctx, "Failed to create relationship", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(err)
	}

	// Use a composite ID to uniquely identify the relationship
	resourceID := fmt.Sprintf("%s:%s:%s", sourceID, targetID, relationType)
	d.SetId(resourceID)

	tflog.Info(ctx, "Relationship created successfully", map[string]interface{}{
		"id": resourceID,
	})

	return resourceRelationRead(ctx, d, meta)
}

func resourceRelationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)

	var diags diag.Diagnostics

	resourceID := d.Id()

	tflog.Debug(ctx, "Reading relationship", map[string]interface{}{"id": resourceID})

	// Parse composite ID
	sourceID, targetID, relationType, err := parseRelationID(resourceID)
	if err != nil {
		tflog.Error(ctx, "Failed to parse relationship ID", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(err)
	}

	relation := &client.Relation{
		Id1:  sourceID,
		Id2:  targetID,
		Type: relationType,
	}

	// Retrieve relationship to verify it exists
	retrieved, err := conn.Ucmdb.GetRelation(ctx, resourceID)
	if err != nil {
		tflog.Error(ctx, "Failed to read relationship", map[string]interface{}{"id": resourceID, "error": err.Error()})
		// If not found, remove from state
		d.SetId("")
		return diag.FromErr(err)
	}

	d.Set("source_ci_id", retrieved.Id1)
	d.Set("target_ci_id", retrieved.Id2)
	d.Set("relationship_type", retrieved.Type)

	if len(retrieved.Properties) > 0 {
		props := make(map[string]interface{})
		for k, v := range retrieved.Properties {
			props[k] = v
		}
		if err := d.Set("properties", props); err != nil {
			tflog.Error(ctx, "Failed to set properties", map[string]interface{}{"error": err.Error()})
			return diag.FromErr(err)
		}
	}

	tflog.Debug(ctx, "Relationship read successfully", map[string]interface{}{"id": resourceID})
	return diags
}

func resourceRelationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)

	resourceID := d.Id()

	if d.HasChanges("relationship_type", "properties") {
		tflog.Debug(ctx, "Updating relationship", map[string]interface{}{"id": resourceID})

		sourceID := d.Get("source_ci_id").(string)
		targetID := d.Get("target_ci_id").(string)
		relationType := d.Get("relationship_type").(string)

		props := make(map[string]interface{})
		if propsRaw, ok := d.GetOk("properties"); ok {
			propsMap := propsRaw.(map[string]interface{})
			for k, v := range propsMap {
				props[k] = v
			}
		}

		relation := &client.Relation{
			Id1:        sourceID,
			Id2:        targetID,
			Type:       relationType,
			Properties: props,
		}

		_, err := conn.Ucmdb.UpdateRelation(ctx, resourceID, relation)
		if err != nil {
			tflog.Error(ctx, "Failed to update relationship", map[string]interface{}{"id": resourceID, "error": err.Error()})
			return diag.FromErr(err)
		}

		tflog.Info(ctx, "Relationship updated successfully", map[string]interface{}{"id": resourceID})
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceRelationRead(ctx, d, meta)
}

func resourceRelationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)

	var diags diag.Diagnostics

	resourceID := d.Id()

	tflog.Debug(ctx, "Deleting relationship", map[string]interface{}{"id": resourceID})

	err := conn.Ucmdb.DeleteRelation(ctx, resourceID)
	if err != nil {
		tflog.Error(ctx, "Failed to delete relationship", map[string]interface{}{"id": resourceID, "error": err.Error()})
		return diag.FromErr(err)
	}

	d.SetId("")

	tflog.Info(ctx, "Relationship deleted successfully", map[string]interface{}{"id": resourceID})

	return diags
}

// parseRelationID parses a composite relationship ID in the format "sourceID:targetID:type"
func parseRelationID(id string) (string, string, string, error) {
	// Simple parsing - assumes no colons in IDs
	// In production, consider using a more robust format
	var sourceID, targetID, relationType string
	_, err := fmt.Sscanf(id, "%s:%s:%s", &sourceID, &targetID, &relationType)
	if err != nil {
		// Try alternative parsing with regex or other method if needed
		return "", "", "", fmt.Errorf("invalid relationship ID format: %s", id)
	}
	return sourceID, targetID, relationType, nil
}
