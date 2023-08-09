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
)

// @SDKDataSource("aws_msk_vpc_connection", name="VPC Connection")
func DataSourceVpcConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVpcConnectionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"target_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVpcConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	arn := d.Get("arn").(string)
	out, err := FindVPCConnectionByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK VPC Connection (%s): %s", arn, err)
	}

	d.SetId(aws.ToString(out.VpcConnectionArn))
	d.Set("arn", out.VpcConnectionArn)
	d.Set("authentication", out.Authentication)
	d.Set("client_subnets", flex.FlattenStringValueSet(out.Subnets))
	d.Set("security_groups", flex.FlattenStringValueSet(out.SecurityGroups))
	d.Set("target_cluster_arn", out.TargetClusterArn)
	d.Set("vpc_id", out.VpcId)

	return nil
}
