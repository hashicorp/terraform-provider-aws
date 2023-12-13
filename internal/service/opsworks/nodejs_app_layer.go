// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_nodejs_app_layer", name="NodeJS App Layer")
// @Tags(identifierAttribute="arn")
func ResourceNodejsAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeNodejsApp,
		DefaultLayerName: "Node.js App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"nodejs_version": {
				AttrName: opsworks.LayerAttributesKeysNodejsVersion,
				Type:     schema.TypeString,
				Default:  "0.10.38",
			},
		},
	}

	return layerType.resourceSchema()
}
