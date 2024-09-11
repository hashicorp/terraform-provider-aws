// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudwatch_event_source", name="Source")
func dataSourceSource() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSourceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrNamePrefix: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	input := &eventbridge.ListEventSourcesInput{}

	if v, ok := d.GetOk(names.AttrNamePrefix); ok {
		input.NamePrefix = aws.String(v.(string))
	}

	es, err := findEventSource(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EventBridge Source", err))
	}

	d.SetId(aws.ToString(es.Name))
	d.Set(names.AttrARN, es.Arn)
	d.Set("created_by", es.CreatedBy)
	d.Set(names.AttrName, es.Name)
	d.Set(names.AttrState, es.State)

	return diags
}

func findEventSource(ctx context.Context, conn *eventbridge.Client, input *eventbridge.ListEventSourcesInput) (*types.EventSource, error) { // nosemgrep:ci.events-in-func-name
	output, err := findEventSources(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEventSources(ctx context.Context, conn *eventbridge.Client, input *eventbridge.ListEventSourcesInput) ([]types.EventSource, error) { // nosemgrep:ci.events-in-func-name
	var output []types.EventSource

	err := listEventSourcesPages(ctx, conn, input, func(page *eventbridge.ListEventSourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.EventSources...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
