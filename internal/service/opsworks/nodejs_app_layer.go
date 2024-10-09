// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_nodejs_app_layer", name="NodeJS App Layer")
// @Tags(identifierAttribute="arn")
func resourceNodejsAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         awstypes.LayerTypeNodejsApp,
		DefaultLayerName: "Node.js App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"nodejs_version": {
				AttrName: awstypes.LayerAttributesKeysNodejsVersion,
				Type:     schema.TypeString,
				Default:  "0.10.38",
			},
		},
	}

	return layerType.resourceSchema()
}
