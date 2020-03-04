package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsOpsworksPhpAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         "php-app",
		DefaultLayerName: "PHP App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{},
	}

	return layerType.SchemaResource()
}
