package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/nbaztec/terraform-provider-aws/aws"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: aws.Provider})
}
