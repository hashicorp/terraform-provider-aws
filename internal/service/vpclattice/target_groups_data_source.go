// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_vpclattice_target_groups", name="Target Groups")
func DataSourceTargetGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTargetGroupsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(3, 128),
			},
			"tags": tftags.TagsSchemaComputed(),
			"type": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.TargetGroupType](),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

const (
	DSNameTargetGroups = "Target Groups Data Source"
)

func dataSourceTargetGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(ctx, d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	filter := func(x types.TargetGroupSummary) bool {
		if v, ok := d.GetOk("name_prefix"); ok {
			if !strings.Contains(aws.ToString(x.Name), v.(string)) {
				return false
			}
		}

		if v, ok := d.GetOk("type"); ok {
			if x.Type != types.TargetGroupType(v.(string)) {
				return false
			}
		}

		if v, ok := d.GetOk("vpc_id"); ok {
			if aws.ToString(x.VpcIdentifier) != v.(string) {
				return false
			}
		}

		return true
	}

	targetGroups, err := findTargetGroups(ctx, conn, filter)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	targetGroupIds := []string{}

	if len(tagsToMatch) > 0 {
		for _, targetGroup := range targetGroups {
			arn := aws.ToString(targetGroup.Arn)
			tags, err := listTags(ctx, conn, arn)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing tags for VPC Lattice Target Group (%s): %s", arn, err)
			}

			if !tags.ContainsAll(tagsToMatch) {
				continue
			}

			targetGroupIds = append(targetGroupIds, aws.ToString(targetGroup.Id))
		}

	} else {
		for _, targetGroup := range targetGroups {
			targetGroupIds = append(targetGroupIds, aws.ToString(targetGroup.Id))
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", targetGroupIds)

	return nil
}
