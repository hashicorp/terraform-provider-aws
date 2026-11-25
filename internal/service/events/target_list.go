// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_cloudwatch_event_target")
func newTargetResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceTarget(), framework.SDKListSpec[awstypes.Target, listTargetModel]{
		ConfigSchema: listTargetConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listTargetModel) (iter.Seq2[awstypes.Target, error], diag.Diagnostics) {
			input := eventbridge.ListTargetsByRuleInput{
				EventBusName: query.EventBusName.ValueStringPointer(),
				Rule:         query.Rule.ValueStringPointer(),
				Limit:        aws.Int32(100),
			}

			return listTargets(ctx, c.EventsClient(ctx), &input), nil
		},
		ID: func(t awstypes.Target, query listTargetModel) string {
			return targetCreateResourceID(query.EventBusName.ValueString(), query.Rule.ValueString(), aws.ToString(t.Id))
		},
		Prepare: func(_ context.Context, t awstypes.Target, query listTargetModel, rd *schema.ResourceData) {
			// The bus and rule come from the query, and together with the
			// target make up this resource's identity.
			rd.Set("event_bus_name", query.EventBusName.ValueString())
			rd.Set(names.AttrRule, query.Rule.ValueString())
			rd.Set("target_id", aws.ToString(t.Id))
		},
		Flatten:     resourceTargetFlatten,
		DisplayName: func(t awstypes.Target) string { return aws.ToString(t.Id) },
	})
}

type listTargetModel struct {
	framework.WithRegionModel
	EventBusName types.String `tfsdk:"event_bus_name"`
	Rule         types.String `tfsdk:"rule"`
}

func listTargets(ctx context.Context, conn *eventbridge.Client, input *eventbridge.ListTargetsByRuleInput) iter.Seq2[awstypes.Target, error] {
	return func(yield func(awstypes.Target, error) bool) {
		var stopped bool
		err := listTargetsByRulePages(ctx, conn, input, func(page *eventbridge.ListTargetsByRuleOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}
			for _, item := range page.Targets {
				if !yield(item, nil) {
					stopped = true
					return false
				}
			}
			return !lastPage
		})
		if !stopped && err != nil {
			yield(inttypes.Zero[awstypes.Target](), fmt.Errorf("listing EventBridge Targets: %w", err))
			return
		}
	}
}

func listTargetConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"event_bus_name": listschema.StringAttribute{
				Required:    true,
				Description: "Name or ARN of the event bus associated with the rule.",
			},
			names.AttrRule: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the rule.",
			},
		},
	}
}
