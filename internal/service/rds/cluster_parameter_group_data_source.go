// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_cluster_parameter_group")
func DataSourceClusterParameterGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterParameterGroupRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrFamily: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceClusterParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	groupName := d.Get(names.AttrName).(string)

	input := rds.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(groupName),
	}

	output, err := conn.DescribeDBClusterParameterGroupsWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Parameter Groups (%s): %s", d.Get(names.AttrName).(string), err)
	}

	if len(output.DBClusterParameterGroups) != 1 || aws.StringValue(output.DBClusterParameterGroups[0].DBClusterParameterGroupName) != groupName {
		return sdkdiag.AppendErrorf(diags, "RDS DB Parameter Group not found (%#v): %s", output, err)
	}

	d.SetId(aws.StringValue(output.DBClusterParameterGroups[0].DBClusterParameterGroupName))
	d.Set(names.AttrName, output.DBClusterParameterGroups[0].DBClusterParameterGroupName)
	d.Set(names.AttrARN, output.DBClusterParameterGroups[0].DBClusterParameterGroupArn)
	d.Set(names.AttrFamily, output.DBClusterParameterGroups[0].DBParameterGroupFamily)
	d.Set(names.AttrDescription, output.DBClusterParameterGroups[0].Description)

	return diags
}
