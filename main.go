package main

import (
	"log"
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-aws/aws"
)

func main() {
	scriptTracer, err := aws.NewScriptTracer()
	if err != nil {
		log.Println("Warning: unable to initialize AWS request tracing:", err)
	} else {
		defer scriptTracer.Close()
	}

	plugin.Serve(&plugin.ServeOpts{ProviderFunc: aws.Provider})
}
