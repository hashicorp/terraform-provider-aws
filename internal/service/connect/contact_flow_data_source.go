// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_contact_flow")
func DataSourceContactFlow() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceContactFlowRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"contact_flow_id", names.AttrName},
			},
			names.AttrContent: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "contact_flow_id"},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceContactFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.DescribeContactFlowInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("contact_flow_id"); ok {
		input.ContactFlowId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		contactFlowSummary, err := dataSourceGetContactFlowSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Contact Flow Summary by name (%s): %s", name, err)
		}

		if contactFlowSummary == nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Contact Flow Summary by name (%s): not found", name)
		}

		input.ContactFlowId = contactFlowSummary.Id
	}

	resp, err := conn.DescribeContactFlowWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Contact Flow: %s", err)
	}

	if resp == nil || resp.ContactFlow == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Contact Flow: empty response")
	}

	contactFlow := resp.ContactFlow

	d.Set(names.AttrARN, contactFlow.Arn)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("contact_flow_id", contactFlow.Id)
	d.Set(names.AttrName, contactFlow.Name)
	d.Set(names.AttrDescription, contactFlow.Description)
	d.Set(names.AttrContent, contactFlow.Content)
	d.Set(names.AttrType, contactFlow.Type)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, contactFlow.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(contactFlow.Id)))

	return diags
}

func dataSourceGetContactFlowSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.ContactFlowSummary, error) {
	var result *connect.ContactFlowSummary

	input := &connect.ListContactFlowsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListContactFlowsMaxResults),
	}

	err := conn.ListContactFlowsPagesWithContext(ctx, input, func(page *connect.ListContactFlowsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.ContactFlowSummaryList {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf.Name) == name {
				result = cf
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
