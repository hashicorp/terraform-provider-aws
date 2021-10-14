package aws

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceStaticWebLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeWeb,
		DefaultLayerName: "Static Web Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{},
	}

	return layerType.SchemaResource()
}
