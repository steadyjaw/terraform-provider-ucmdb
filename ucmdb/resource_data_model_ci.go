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

func resourceDataModelCi() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a UCMDB Configuration Item (CI) data model.",

		CreateContext: resourceDataModelCiCreate,
		ReadContext:   resourceDataModelCiRead,
		UpdateContext: resourceDataModelCiUpdate,
		DeleteContext: resourceDataModelCiDelete,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Description:  "The type of CI to manage.",
				ValidateFunc: validation.StringIsNotEmpty,
				Required:     true,
				ForceNew:     true,
			},
			"properties": {
				Type:        schema.TypeSet,
				Description: "Properties of the CI as key-value pairs.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the CI.",
							Required:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Optional description of the CI.",
							Optional:    true,
						},
					},
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

func resourceDataModelCiCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)

	ci_type := d.Get("type").(string)
	ci_prop_set := d.Get("properties").(*schema.Set)
	ci_props := (ci_prop_set.List())[0].(map[string]interface{})

	tflog.Debug(ctx, "Creating CI", map[string]interface{}{
		"type": ci_type,
		"name": ci_props["name"],
	})

	td := &client.TopologyData{
		CIS: []client.DataInConfigurationItem{
			{
				UcmdbId:    ci_type,
				Type:       ci_type,
				Properties: ci_props,
			},
		},
	}

	dmc, err := conn.Ucmdb.CreateDataModel(ctx, td)
	if err != nil {
		tflog.Error(ctx, "Failed to create CI", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(err)
	}

	if len(dmc.AddedCis) == 1 {
		d.SetId(dmc.AddedCis[0])
		tflog.Info(ctx, "CI created successfully", map[string]interface{}{"id": dmc.AddedCis[0]})
	} else if len(dmc.AddedCis) == 0 && len(dmc.UpdatedCis) == 1 {
		d.SetId(dmc.UpdatedCis[0])
		tflog.Warn(ctx, "CI already existed, updated instead", map[string]interface{}{"id": dmc.UpdatedCis[0]})
	} else if len(dmc.AddedCis) == 0 && len(dmc.IgnoredCis) == 1 {
		tflog.Warn(ctx, "CI creation resulted in ignored CI", map[string]interface{}{"id": dmc.IgnoredCis[0]})
		return diag.FromErr(fmt.Errorf("CI creation was ignored by UCMDB"))
	} else {
		tflog.Error(ctx, "Unexpected response from UCMDB", map[string]interface{}{"response": dmc})
		return diag.FromErr(fmt.Errorf("unexpected response from UCMDB: %v", dmc))
	}

	return resourceDataModelCiRead(ctx, d, meta)
}

func resourceDataModelCiRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)

	var diags diag.Diagnostics

	ucmdbid := d.Id()
	tflog.Debug(ctx, "Reading CI", map[string]interface{}{"id": ucmdbid})

	dci, err := conn.Ucmdb.GetConfigurationItem(ctx, ucmdbid)
	if err != nil {
		tflog.Error(ctx, "Failed to read CI", map[string]interface{}{"id": ucmdbid, "error": err.Error()})
		return diag.FromErr(err)
	}

	d.Set("type", dci.Type)

	props := make(map[string]interface{})
	props["name"] = dci.Properties["name"].(string)
	if desc, ok := dci.Properties["description"].(string); ok {
		props["description"] = desc
	}

	if err := d.Set("properties", []interface{}{props}); err != nil {
		tflog.Error(ctx, "Failed to set properties", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "CI read successfully", map[string]interface{}{"id": ucmdbid})
	return diags
}

func resourceDataModelCiUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)
	id := d.Id()

	if d.HasChanges("type", "properties") {
		tflog.Debug(ctx, "Updating CI", map[string]interface{}{"id": id})

		ci_type := d.Get("type").(string)
		ci_prop_set := d.Get("properties").(*schema.Set)
		ci_props := (ci_prop_set.List())[0].(map[string]interface{})

		dic := client.DataInConfigurationItem{
			UcmdbId:    id,
			Type:       ci_type,
			Properties: ci_props,
		}

		_, err := conn.Ucmdb.UpdateConfigurationItem(ctx, id, dic)
		if err != nil {
			tflog.Error(ctx, "Failed to update CI", map[string]interface{}{"id": id, "error": err.Error()})
			return diag.FromErr(err)
		}

		tflog.Info(ctx, "CI updated successfully", map[string]interface{}{"id": id})
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceDataModelCiRead(ctx, d, meta)
}

func resourceDataModelCiDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*client.Client)

	var diags diag.Diagnostics

	id := d.Id()
	tflog.Debug(ctx, "Deleting CI", map[string]interface{}{"id": id})

	_, err := conn.Ucmdb.DeleteConfigurationItem(ctx, id)
	if err != nil {
		tflog.Error(ctx, "Failed to delete CI", map[string]interface{}{"id": id, "error": err.Error()})
		return diag.FromErr(err)
	}

	d.SetId("")
	tflog.Info(ctx, "CI deleted successfully", map[string]interface{}{"id": id})

	return diags
}
