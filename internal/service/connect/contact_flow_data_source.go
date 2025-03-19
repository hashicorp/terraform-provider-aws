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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_contact_flow", name="Contact Flow")
// @Tags
func dataSourceContactFlow() *schema.Resource {
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

func dataSourceContactFlowRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.DescribeContactFlowInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("contact_flow_id"); ok {
		input.ContactFlowId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		contactFlowSummary, err := findContactFlowSummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Contact Flow (%s) summary: %s", name, err)
		}

		input.ContactFlowId = contactFlowSummary.Id
	}

	contactFlow, err := findContactFlow(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Contact Flow: %s", err)
	}

	contactFlowID := aws.ToString(contactFlow.Id)
	id := contactFlowCreateResourceID(instanceID, contactFlowID)
	d.SetId(id)
	d.Set(names.AttrARN, contactFlow.Arn)
	d.Set("contact_flow_id", contactFlowID)
	d.Set(names.AttrContent, contactFlow.Content)
	d.Set(names.AttrDescription, contactFlow.Description)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, contactFlow.Name)
	d.Set(names.AttrType, contactFlow.Type)

	setTagsOut(ctx, contactFlow.Tags)

	return diags
}

func findContactFlowSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.ContactFlowSummary, error) {
	const maxResults = 60
	input := &connect.ListContactFlowsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findContactFlowSummary(ctx, conn, input, func(v *awstypes.ContactFlowSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findContactFlowSummary(ctx context.Context, conn *connect.Client, input *connect.ListContactFlowsInput, filter tfslices.Predicate[*awstypes.ContactFlowSummary]) (*awstypes.ContactFlowSummary, error) {
	output, err := findContactFlowSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findContactFlowSummaries(ctx context.Context, conn *connect.Client, input *connect.ListContactFlowsInput, filter tfslices.Predicate[*awstypes.ContactFlowSummary]) ([]awstypes.ContactFlowSummary, error) {
	var output []awstypes.ContactFlowSummary

	pages := connect.NewListContactFlowsPaginator(conn, input)
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

		for _, v := range page.ContactFlowSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
