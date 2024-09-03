// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudwatch_log_groups")
func dataSourceGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupsRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"log_group_name_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"log_group_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	input := &cloudwatchlogs.DescribeLogGroupsInput{}

	if v, ok := d.GetOk("log_group_name_prefix"); ok {
		input.LogGroupNamePrefix = aws.String(v.(string))
	}

	var output []types.LogGroup

	pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudWatch Log Groups: %s", err)
		}

		output = append(output, page.LogGroups...)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, logGroupNames []string

	for _, r := range output {
		arns = append(arns, TrimLogGroupARNWildcardSuffix(aws.ToString(r.Arn)))
		logGroupNames = append(logGroupNames, aws.ToString(r.LogGroupName))
	}

	d.Set(names.AttrARNs, arns)
	d.Set("log_group_names", logGroupNames)

	return diags
}
