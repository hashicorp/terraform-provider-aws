// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_rails_app_layer", name="Rails App Layer")
// @Tags(identifierAttribute="arn")
func ResourceRailsAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeRailsApp,
		DefaultLayerName: "Rails App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"app_server": {
				AttrName: opsworks.LayerAttributesKeysRailsStack,
				Type:     schema.TypeString,
				Default:  "apache_passenger",
			},
			"bundler_version": {
				AttrName: opsworks.LayerAttributesKeysBundlerVersion,
				Type:     schema.TypeString,
				Default:  "1.5.3",
			},
			"manage_bundler": {
				AttrName: opsworks.LayerAttributesKeysManageBundler,
				Type:     schema.TypeBool,
				Default:  true,
			},
			"passenger_version": {
				AttrName: opsworks.LayerAttributesKeysPassengerVersion,
				Type:     schema.TypeString,
				Default:  "4.0.46",
			},
			"ruby_version": {
				AttrName: opsworks.LayerAttributesKeysRubyVersion,
				Type:     schema.TypeString,
				Default:  "2.0.0",
			},
			"rubygems_version": {
				AttrName: opsworks.LayerAttributesKeysRubygemsVersion,
				Type:     schema.TypeString,
				Default:  "2.2.2",
			},
		},
	}

	return layerType.resourceSchema()
}
