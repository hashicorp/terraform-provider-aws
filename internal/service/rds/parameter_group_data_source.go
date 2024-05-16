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
)

// @SDKDataSource("aws_db_parameter_group")
func DataSourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceParameterGroupRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"family": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	groupName := d.Get("name").(string)

	input := rds.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(groupName),
	}

	output, err := conn.DescribeDBParameterGroupsWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Parameter Groups (%s): %s", d.Get("name").(string), err)
	}

	if len(output.DBParameterGroups) != 1 || aws.StringValue(output.DBParameterGroups[0].DBParameterGroupName) != groupName {
		return sdkdiag.AppendErrorf(diags, "RDS DB Parameter Group not found (%#v): %s", output, err)
	}

	d.SetId(aws.StringValue(output.DBParameterGroups[0].DBParameterGroupName))
	d.Set("name", output.DBParameterGroups[0].DBParameterGroupName)
	d.Set("arn", output.DBParameterGroups[0].DBParameterGroupArn)
	d.Set("family", output.DBParameterGroups[0].DBParameterGroupFamily)
	d.Set("description", output.DBParameterGroups[0].Description)

	return nil
}
