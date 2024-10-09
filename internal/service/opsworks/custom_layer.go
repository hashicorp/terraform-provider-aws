// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_custom_layer", name="Custom Layer")
// @Tags(identifierAttribute="arn")
func resourceCustomLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:        awstypes.LayerTypeCustom,
		CustomShortName: true,

		// The "custom" layer type has no additional attributes.
		Attributes: map[string]*opsworksLayerTypeAttribute{},
	}

	return layerType.resourceSchema()
}
