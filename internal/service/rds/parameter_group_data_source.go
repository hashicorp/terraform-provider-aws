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

// @SDKDataSource("aws_db_parameter_group")
func DataSourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceParameterGroupRead,
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

func dataSourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	groupName := d.Get(names.AttrName).(string)

	input := rds.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(groupName),
	}

	output, err := conn.DescribeDBParameterGroupsWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Parameter Groups (%s): %s", d.Get(names.AttrName).(string), err)
	}

	if len(output.DBParameterGroups) != 1 || aws.StringValue(output.DBParameterGroups[0].DBParameterGroupName) != groupName {
		return sdkdiag.AppendErrorf(diags, "RDS DB Parameter Group not found (%#v): %s", output, err)
	}

	d.SetId(aws.StringValue(output.DBParameterGroups[0].DBParameterGroupName))
	d.Set(names.AttrName, output.DBParameterGroups[0].DBParameterGroupName)
	d.Set(names.AttrARN, output.DBParameterGroups[0].DBParameterGroupArn)
	d.Set(names.AttrFamily, output.DBParameterGroups[0].DBParameterGroupFamily)
	d.Set(names.AttrDescription, output.DBParameterGroups[0].Description)

	return diags
}
