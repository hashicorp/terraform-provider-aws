// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_connect_quick_connect")
func DataSourceQuickConnect() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQuickConnectRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "quick_connect_id"},
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
				ExactlyOneOf: []string{"quick_connect_id", "name"},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQuickConnectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeQuickConnectInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("quick_connect_id"); ok {
		input.QuickConnectId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		quickConnectSummary, err := dataSourceGetQuickConnectSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Quick Connect Summary by name (%s): %s", name, err)
		}

		if quickConnectSummary == nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Quick Connect Summary by name (%s): not found", name)
		}

		input.QuickConnectId = quickConnectSummary.Id
	}

	resp, err := conn.DescribeQuickConnect(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Quick Connect: %s", err)
	}

	if resp == nil || resp.QuickConnect == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Quick Connect: empty response")
	}

	quickConnect := resp.QuickConnect

	d.Set("arn", quickConnect.QuickConnectARN)
	d.Set("description", quickConnect.Description)
	d.Set("name", quickConnect.Name)
	d.Set("quick_connect_id", quickConnect.QuickConnectId)

	if err := d.Set("quick_connect_config", flattenQuickConnectConfig(quickConnect.QuickConnectConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting quick_connect_config: %s", err)
	}

	if err := d.Set("tags", KeyValueTags(ctx, quickConnect.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.ToString(quickConnect.QuickConnectId)))

	return diags
}

func dataSourceGetQuickConnectSummaryByName(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.QuickConnectSummary, error) {
	var result *awstypes.QuickConnectSummary

	input := &connect.ListQuickConnectsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(ListQuickConnectsMaxResults),
	}

	pages := connect.NewListQuickConnectsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, cf := range page.QuickConnectSummaryList {
			if aws.ToString(cf.Name) == name {
				result = &cf
			}
		}
	}

	return result, nil
}
