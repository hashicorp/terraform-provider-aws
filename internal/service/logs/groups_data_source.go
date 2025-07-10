// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudwatch_log_groups", name="Log Groups")
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

func dataSourceGroupsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	input := cloudwatchlogs.DescribeLogGroupsInput{}
	if v, ok := d.GetOk("log_group_name_prefix"); ok {
		input.LogGroupNamePrefix = aws.String(v.(string))
	}

	output, err := findLogGroups(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.LogGroup]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Log Groups: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	var arns, logGroupNames []string
	for _, v := range output {
		arns = append(arns, trimLogGroupARNWildcardSuffix(aws.ToString(v.Arn)))
		logGroupNames = append(logGroupNames, aws.ToString(v.LogGroupName))
	}
	d.Set(names.AttrARNs, arns)
	d.Set("log_group_names", logGroupNames)

	return diags
}
