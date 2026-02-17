package ucmdb

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/steadyjaw/terraform-provider-ucmdb/client"
)

// Provider returns a terraform ResourceProvider interface
func Provider() *schema.Provider {
	return &schema.Provider{
		// setting up shared configuration objects, e.g. addresses, secrets, access keys
		Schema: map[string]*schema.Schema{
			"target_env": {
				Type:         schema.TypeString,
				Description:  "A value which represents the UCMDB Target Environment. Valid values: CMS, OPSB, APM",
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"CMS", "OPSB", "APM"}, true),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"ucmdb_data_model_ci": resourceDataModelCi(),
			"ucmdb_relation":      resourceRelation(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"ucmdb_list":     dataSourceUcmdbList(),
			"ucmdb_relation": dataSourceRelation(),
		},

		// initialize shared configuration objects - the SDK client which makes API requests to UCMDB
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	i, err := readConfiguration(ctx, d.Get("target_env").(string))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	conn, err := client.NewClient(ctx, i["address"], i["user"], i["password"])
	if err != nil {
		return nil, diag.FromErr(err)
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully connected to UCMDB at %s", i["address"]))
	return conn, diags
}

func readConfiguration(ctx context.Context, target_env string) (map[string]string, error) {
	address_env_name := fmt.Sprintf("UCMDB_%s_ADDRESS", strings.ToUpper(target_env))
	user_env_name := fmt.Sprintf("UCMDB_%s_API_USER", strings.ToUpper(target_env))
	password_env_name := fmt.Sprintf("UCMDB_%s_API_PASSWORD", strings.ToUpper(target_env))

	address, address_exists := os.LookupEnv(address_env_name)
	user, user_exists := os.LookupEnv(user_env_name)
	password, password_exists := os.LookupEnv(password_env_name)

	if !(address_exists && user_exists && password_exists) {
		return nil, fmt.Errorf("the following environment variables must be set: %s, %s, %s", address_env_name, user_env_name, password_env_name)
	}

	if strings.TrimSpace(address) == "" || strings.TrimSpace(user) == "" || strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("the following environment variables must not have empty values: %s, %s, %s", address_env_name, user_env_name, password_env_name)
	}

	config := make(map[string]string)
	config["address"] = address
	config["user"] = user
	config["password"] = password

	tflog.Debug(ctx, "Configuration loaded successfully", map[string]interface{}{
		"address": address,
	})

	return config, nil
}
