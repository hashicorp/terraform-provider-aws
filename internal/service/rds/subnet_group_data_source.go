// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_db_subnet_group", name="DB Subnet Group")
func dataSourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSubnetGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_network_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	v, err := findDBSubnetGroupByName(ctx, conn, d.Get("name").(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("RDS DB Subnet Group", err))
	}

	d.SetId(aws.ToString(v.DBSubnetGroupName))
	d.Set("arn", v.DBSubnetGroupArn)
	d.Set("description", v.DBSubnetGroupDescription)
	d.Set("name", v.DBSubnetGroupName)
	d.Set("status", v.SubnetGroupStatus)
	d.Set("subnet_ids", tfslices.ApplyToAll(v.Subnets, func(v types.Subnet) string {
		return aws.ToString(v.SubnetIdentifier)
	}))
	d.Set("supported_network_types", v.SupportedNetworkTypes)
	d.Set("vpc_id", v.VpcId)

	return diags
}
