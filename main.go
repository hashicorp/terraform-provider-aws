package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/frc9/terraform-provider-aws/aws"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: aws.Provider})
}
