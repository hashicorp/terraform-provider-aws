// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsagent/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_devopsagent_agent_space")
func newAgentSpaceResourceAsListResource() list.ListResourceWithConfigure {
	return &agentSpaceListResource{}
}

var _ list.ListResource = &agentSpaceListResource{}

type agentSpaceListResource struct {
	agentSpaceResource
	framework.WithList
}

func (l *agentSpaceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().DevOpsAgentClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input devopsagent.ListAgentSpacesInput
		for item, err := range listAgentSpaces(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}
			id := aws.ToString(item.AgentSpaceId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("agent_space_id"), id)

			result := request.NewListResult(ctx)
			var data agentSpaceResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data), smerr.ID, id)
					if result.Diagnostics.HasError() {
						return
					}
				} else {
					data.AgentSpaceID = types.StringValue(id)
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listAgentSpaces(ctx context.Context, conn *devopsagent.Client, input *devopsagent.ListAgentSpacesInput) iter.Seq2[awstypes.AgentSpace, error] {
	return func(yield func(awstypes.AgentSpace, error) bool) {
		pages := devopsagent.NewListAgentSpacesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AgentSpace{}, smarterr.NewError(err))
				return
			}

			for _, item := range page.AgentSpaces {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
