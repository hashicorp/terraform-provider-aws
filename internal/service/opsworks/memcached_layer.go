// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_memcached_layer", name="Memcached Layer")
// @Tags(identifierAttribute="arn")
func resourceMemcachedLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         awstypes.LayerTypeMemcached,
		DefaultLayerName: "Memcached",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"allocated_memory": {
				AttrName: awstypes.LayerAttributesKeysMemcachedMemory,
				Type:     schema.TypeInt,
				Default:  512,
			},
		},
	}

	return layerType.resourceSchema()
}
