// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_java_app_layer", name="Java App Layer")
// @Tags(identifierAttribute="arn")
func resourceJavaAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         awstypes.LayerTypeJavaApp,
		DefaultLayerName: "Java App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"app_server": {
				AttrName: awstypes.LayerAttributesKeysJavaAppServer,
				Type:     schema.TypeString,
				Default:  "tomcat",
			},
			"app_server_version": {
				AttrName: awstypes.LayerAttributesKeysJavaAppServerVersion,
				Type:     schema.TypeString,
				Default:  "7",
			},
			"jvm_options": {
				AttrName: awstypes.LayerAttributesKeysJvmOptions,
				Type:     schema.TypeString,
				Default:  "",
			},
			"jvm_type": {
				AttrName: awstypes.LayerAttributesKeysJvm,
				Type:     schema.TypeString,
				Default:  "openjdk",
			},
			"jvm_version": {
				AttrName: awstypes.LayerAttributesKeysJvmVersion,
				Type:     schema.TypeString,
				Default:  "7",
			},
		},
	}

	return layerType.resourceSchema()
}
