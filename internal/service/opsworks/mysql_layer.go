// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_opsworks_mysql_layer", name="MySQL Layer")
// @Tags(identifierAttribute="arn")
func resourceMySQLLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         awstypes.LayerTypeDbMaster,
		DefaultLayerName: "MySQL",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"root_password": {
				AttrName:  awstypes.LayerAttributesKeysMysqlRootPassword,
				Type:      schema.TypeString,
				WriteOnly: true,
			},
			"root_password_on_all_instances": {
				AttrName: awstypes.LayerAttributesKeysMysqlRootPasswordUbiquitous,
				Type:     schema.TypeBool,
				Default:  true,
			},
		},
	}

	return layerType.resourceSchema()
}
