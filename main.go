package main

import (
	"github.com/adamdecaf/terraform-provider-namecheap/namecheap"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: namecheap.Provider,
	})
}
