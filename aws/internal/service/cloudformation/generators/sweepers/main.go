// +build ignore

package main

import (
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/generators/sweepers"
)

const serviceName = "cloudformation"

type ResourceType struct {
	ListerFunction       string
	ListerOutputType     string
	ListerPageField      string
	ResourceNameFunction string
}

var resourceTypes = map[string]sweepers.ResourceType{
	"StackSet": {
		ListerFunction:       "ListAllStackSetsPages",
		ListerOutputType:     "ListStackSetsOutput",
		ListerPageField:      "Summaries",
		ResourceNameFunction: "StackSetName",
	},
}

type TemplateData struct {
	Package       string
	ServiceName   string
	ResourceTypes map[string]ResourceType
}

func main() {
	sweepers.Run(serviceName, resourceTypes)
}
