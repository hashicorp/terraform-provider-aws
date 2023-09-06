// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_custom_layer", name="Custom Layer")
// @Tags(identifierAttribute="arn")
func ResourceCustomLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:        opsworks.LayerTypeCustom,
		CustomShortName: true,

		// The "custom" layer type has no additional attributes.
		Attributes: map[string]*opsworksLayerTypeAttribute{},
	}

	return layerType.resourceSchema()
}
