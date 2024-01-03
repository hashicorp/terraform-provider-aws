// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_ganglia_layer", name="Ganglia Layer")
// @Tags(identifierAttribute="arn")
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
