package opsworks

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceStaticWebLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeWeb,
		DefaultLayerName: "Static Web Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{},
	}

	return layerType.SchemaResource()
}
