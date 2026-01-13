// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flowLogSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deliver_cross_account_role": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_options": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_format": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"hive_compatible_partitions": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"per_hour_partition": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"eni_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrIAMRoleARN: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"log_destination": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"log_destination_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"log_format": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			names.AttrLogGroupName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"max_aggregation_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"traffic_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTransitGatewayAttachmentID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTransitGatewayID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func flowLogStateUpgradeV0(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	delete(rawState, names.AttrLogGroupName)

	return rawState, nil
}
