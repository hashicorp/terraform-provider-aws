// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_php_app_layer", name="PHP App Layer")
// @Tags(identifierAttribute="arn")
func resourcePHPAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         awstypes.LayerTypePhpApp,
		DefaultLayerName: "PHP App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{},
	}

	return layerType.resourceSchema()
}
