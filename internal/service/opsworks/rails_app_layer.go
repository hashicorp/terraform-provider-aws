// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_rails_app_layer", name="Rails App Layer")
// @Tags(identifierAttribute="arn")
func resourceRailsAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         awstypes.LayerTypeRailsApp,
		DefaultLayerName: "Rails App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"app_server": {
				AttrName: awstypes.LayerAttributesKeysRailsStack,
				Type:     schema.TypeString,
				Default:  "apache_passenger",
			},
			"bundler_version": {
				AttrName: awstypes.LayerAttributesKeysBundlerVersion,
				Type:     schema.TypeString,
				Default:  "1.5.3",
			},
			"manage_bundler": {
				AttrName: awstypes.LayerAttributesKeysManageBundler,
				Type:     schema.TypeBool,
				Default:  true,
			},
			"passenger_version": {
				AttrName: awstypes.LayerAttributesKeysPassengerVersion,
				Type:     schema.TypeString,
				Default:  "4.0.46",
			},
			"ruby_version": {
				AttrName: awstypes.LayerAttributesKeysRubyVersion,
				Type:     schema.TypeString,
				Default:  "2.0.0",
			},
			"rubygems_version": {
				AttrName: awstypes.LayerAttributesKeysRubygemsVersion,
				Type:     schema.TypeString,
				Default:  "2.2.2",
			},
		},
	}

	return layerType.resourceSchema()
}
