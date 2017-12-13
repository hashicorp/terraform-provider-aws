package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-tls/tls"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: tls.Provider})
}
