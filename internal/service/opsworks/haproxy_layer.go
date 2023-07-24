// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_haproxy_layer", name="HAProxy Layer")
// @Tags(identifierAttribute="arn")
func ResourceHAProxyLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeLb,
		DefaultLayerName: "HAProxy",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"healthcheck_method": {
				AttrName: opsworks.LayerAttributesKeysHaproxyHealthCheckMethod,
				Type:     schema.TypeString,
				Default:  "OPTIONS",
			},
			"healthcheck_url": {
				AttrName: opsworks.LayerAttributesKeysHaproxyHealthCheckUrl,
				Type:     schema.TypeString,
				Default:  "/",
			},
			"stats_enabled": {
				AttrName: opsworks.LayerAttributesKeysEnableHaproxyStats,
				Type:     schema.TypeBool,
				Default:  true,
			},
			"stats_password": {
				AttrName:  opsworks.LayerAttributesKeysHaproxyStatsPassword,
				Type:      schema.TypeString,
				WriteOnly: true,
				Required:  true,
			},
			"stats_url": {
				AttrName: opsworks.LayerAttributesKeysHaproxyStatsUrl,
				Type:     schema.TypeString,
				Default:  "/haproxy?stats",
			},
			"stats_user": {
				AttrName: opsworks.LayerAttributesKeysHaproxyStatsUser,
				Type:     schema.TypeString,
				Default:  "opsworks",
			},
		},
	}

	return layerType.resourceSchema()
}
