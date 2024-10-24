// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_haproxy_layer", name="HAProxy Layer")
// @Tags(identifierAttribute="arn")
func resourceHAProxyLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         awstypes.LayerTypeLb,
		DefaultLayerName: "HAProxy",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"healthcheck_method": {
				AttrName: awstypes.LayerAttributesKeysHaproxyHealthCheckMethod,
				Type:     schema.TypeString,
				Default:  "OPTIONS",
			},
			"healthcheck_url": {
				AttrName: awstypes.LayerAttributesKeysHaproxyHealthCheckUrl,
				Type:     schema.TypeString,
				Default:  "/",
			},
			"stats_enabled": {
				AttrName: awstypes.LayerAttributesKeysEnableHaproxyStats,
				Type:     schema.TypeBool,
				Default:  true,
			},
			"stats_password": {
				AttrName:  awstypes.LayerAttributesKeysHaproxyStatsPassword,
				Type:      schema.TypeString,
				WriteOnly: true,
				Required:  true,
			},
			"stats_url": {
				AttrName: awstypes.LayerAttributesKeysHaproxyStatsUrl,
				Type:     schema.TypeString,
				Default:  "/haproxy?stats",
			},
			"stats_user": {
				AttrName: awstypes.LayerAttributesKeysHaproxyStatsUser,
				Type:     schema.TypeString,
				Default:  "opsworks",
			},
		},
	}

	return layerType.resourceSchema()
}
