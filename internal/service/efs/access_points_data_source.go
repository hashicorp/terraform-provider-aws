// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_efs_access_points", name="Access Point")
func dataSourceAccessPoints() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAccessPointsRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFileSystemID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAccessPointsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	fileSystemID := d.Get(names.AttrFileSystemID).(string)
	input := &efs.DescribeAccessPointsInput{
		FileSystemId: aws.String(fileSystemID),
	}

	output, err := findAccessPointDescriptions(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Access Points: %s", err)
	}

	var accessPointIDs, arns []string

	for _, v := range output {
		accessPointIDs = append(accessPointIDs, aws.ToString(v.AccessPointId))
		arns = append(arns, aws.ToString(v.AccessPointArn))
	}

	d.SetId(fileSystemID)
	d.Set(names.AttrARNs, arns)
	d.Set(names.AttrIDs, accessPointIDs)

	return diags
}

func findAccessPointDescriptions(ctx context.Context, conn *efs.Client, input *efs.DescribeAccessPointsInput) ([]awstypes.AccessPointDescription, error) {
	var output []awstypes.AccessPointDescription

	pages := efs.NewDescribeAccessPointsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.AccessPoints...)
	}

	return output, nil
}
