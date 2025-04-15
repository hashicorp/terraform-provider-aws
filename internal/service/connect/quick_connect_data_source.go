// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_quick_connect", name="Quick Connect")
// @Tags
func dataSourceQuickConnect() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQuickConnectRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "quick_connect_id"},
			},
			"quick_connect_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"phone_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"phone_number": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"queue_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"contact_flow_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"queue_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"quick_connect_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"contact_flow_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"quick_connect_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"quick_connect_id", names.AttrName},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQuickConnectRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.DescribeQuickConnectInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("quick_connect_id"); ok {
		input.QuickConnectId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		quickConnectSummary, err := findQuickConnectSummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Quick Connect (%s) summary: %s", name, err)
		}

		input.QuickConnectId = quickConnectSummary.Id
	}

	quickConnect, err := findQuickConnect(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Quick Connect: %s", err)
	}

	quickConnectID := aws.ToString(quickConnect.QuickConnectId)
	id := quickConnectCreateResourceID(instanceID, quickConnectID)
	d.SetId(id)
	d.Set(names.AttrARN, quickConnect.QuickConnectARN)
	d.Set(names.AttrDescription, quickConnect.Description)
	d.Set(names.AttrName, quickConnect.Name)
	if err := d.Set("quick_connect_config", flattenQuickConnectConfig(quickConnect.QuickConnectConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting quick_connect_config: %s", err)
	}
	d.Set("quick_connect_id", quickConnectID)

	setTagsOut(ctx, quickConnect.Tags)

	return diags
}

func findQuickConnectSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.QuickConnectSummary, error) {
	const maxResults = 60
	input := &connect.ListQuickConnectsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findQuickConnectSummary(ctx, conn, input, func(v *awstypes.QuickConnectSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findQuickConnectSummary(ctx context.Context, conn *connect.Client, input *connect.ListQuickConnectsInput, filter tfslices.Predicate[*awstypes.QuickConnectSummary]) (*awstypes.QuickConnectSummary, error) {
	output, err := findQuickConnectSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findQuickConnectSummaries(ctx context.Context, conn *connect.Client, input *connect.ListQuickConnectsInput, filter tfslices.Predicate[*awstypes.QuickConnectSummary]) ([]awstypes.QuickConnectSummary, error) {
	var output []awstypes.QuickConnectSummary

	pages := connect.NewListQuickConnectsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.QuickConnectSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
