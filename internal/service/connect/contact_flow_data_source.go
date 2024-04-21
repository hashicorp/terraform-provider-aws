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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_connect_contact_flow")
func DataSourceContactFlow() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceContactFlowRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"contact_flow_id", "name"},
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "contact_flow_id"},
			},
			"tags": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceContactFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeContactFlowInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("contact_flow_id"); ok {
		input.ContactFlowId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		contactFlowSummary, err := dataSourceGetContactFlowSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Contact Flow Summary by name (%s): %s", name, err)
		}

		input.ContactFlowId = contactFlowSummary.Id
	}

	resp, err := conn.DescribeContactFlow(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Contact Flow: %s", err)
	}

	if resp == nil || resp.ContactFlow == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Contact Flow: empty response")
	}

	contactFlow := resp.ContactFlow

	d.Set("arn", contactFlow.Arn)
	d.Set("instance_id", instanceID)
	d.Set("contact_flow_id", contactFlow.Id)
	d.Set("name", contactFlow.Name)
	d.Set("description", contactFlow.Description)
	d.Set("content", contactFlow.Content)
	d.Set("type", contactFlow.Type)

	if err := d.Set("tags", KeyValueTags(ctx, contactFlow.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.ToString(contactFlow.Id)))

	return diags
}

func dataSourceGetContactFlowSummaryByName(ctx context.Context, conn *connect.Client, instanceID, name string) (awstypes.ContactFlowSummary, error) {
	var result awstypes.ContactFlowSummary

	input := &connect.ListContactFlowsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(ListContactFlowsMaxResults),
	}

	pages := connect.NewListContactFlowsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return result, err
		}

		for _, cf := range page.ContactFlowSummaryList {
			if aws.ToString(cf.Name) == name {
				result = cf
			}
		}
	}

	return result, nil
}
