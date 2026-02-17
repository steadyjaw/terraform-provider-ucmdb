package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/steadyjaw/terraform-provider-ucmdb/ucmdb"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run provider with support for debugger")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return ucmdb.Provider()
		},
	}

	if debugMode {
		err := plugin.Debug(context.Background(), "hashicorp/ucmdb", opts)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	plugin.Serve(opts)
}
