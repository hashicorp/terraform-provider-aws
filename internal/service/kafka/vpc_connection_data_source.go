// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_msk_vpc_connection", name="VPC Connection")
// @Tags
func dataSourceVPCConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCConnectionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Required: true,
			},
			"authentication": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_subnets": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"target_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	arn := d.Get(names.AttrARN).(string)
	output, err := findVPCConnectionByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK VPC Connection (%s): %s", arn, err)
	}

	d.SetId(aws.ToString(output.VpcConnectionArn))
	d.Set(names.AttrARN, output.VpcConnectionArn)
	d.Set("authentication", output.Authentication)
	d.Set("client_subnets", flex.FlattenStringValueSet(output.Subnets))
	d.Set(names.AttrSecurityGroups, flex.FlattenStringValueSet(output.SecurityGroups))
	d.Set("target_cluster_arn", output.TargetClusterArn)
	d.Set(names.AttrVPCID, output.VpcId)

	setTagsOut(ctx, output.Tags)

	return diags
}
