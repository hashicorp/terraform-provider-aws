// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_efs_access_points")
func DataSourceAccessPoints() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAccessPointsRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"file_system_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAccessPointsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn(ctx)

	fileSystemID := d.Get("file_system_id").(string)
	input := &efs.DescribeAccessPointsInput{
		FileSystemId: aws.String(fileSystemID),
	}

	output, err := findAccessPointDescriptions(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Access Points: %s", err)
	}

	var accessPointIDs, arns []string

	for _, v := range output {
		accessPointIDs = append(accessPointIDs, aws.StringValue(v.AccessPointId))
		arns = append(arns, aws.StringValue(v.AccessPointArn))
	}

	d.SetId(fileSystemID)
	d.Set("arns", arns)
	d.Set("ids", accessPointIDs)

	return diags
}

func findAccessPointDescriptions(ctx context.Context, conn *efs.EFS, input *efs.DescribeAccessPointsInput) ([]*efs.AccessPointDescription, error) {
	var output []*efs.AccessPointDescription

	err := conn.DescribeAccessPointsPagesWithContext(ctx, input, func(page *efs.DescribeAccessPointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AccessPoints {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
