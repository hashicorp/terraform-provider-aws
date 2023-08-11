// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_memcached_layer", name="Memcached Layer")
// @Tags(identifierAttribute="arn")
func ResourceMemcachedLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeMemcached,
		DefaultLayerName: "Memcached",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"allocated_memory": {
				AttrName: opsworks.LayerAttributesKeysMemcachedMemory,
				Type:     schema.TypeInt,
				Default:  512,
			},
		},
	}

	return layerType.resourceSchema()
}
