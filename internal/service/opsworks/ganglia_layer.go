package opsworks

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceGangliaLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeMonitoringMaster,
		DefaultLayerName: "Ganglia",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"password": {
				AttrName:  opsworks.LayerAttributesKeysGangliaPassword,
				Type:      schema.TypeString,
				Required:  true,
				WriteOnly: true,
			},
			"url": {
				AttrName: opsworks.LayerAttributesKeysGangliaUrl,
				Type:     schema.TypeString,
				Default:  "/ganglia",
			},
			"username": {
				AttrName: opsworks.LayerAttributesKeysGangliaUser,
				Type:     schema.TypeString,
				Default:  "opsworks",
			},
		},
	}

	return layerType.resourceSchema()
}
