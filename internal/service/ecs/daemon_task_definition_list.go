// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @FrameworkListResource("aws_ecs_daemon_task_definition")
func daemonTaskDefinitionResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceDaemonTaskDefinition{}
}

var _ list.ListResource = &listResourceDaemonTaskDefinition{}

type listResourceDaemonTaskDefinition struct {
	daemonTaskDefinitionResource
	framework.WithList
}

func (r *listResourceDaemonTaskDefinition) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{},
	}
}

func (r *listResourceDaemonTaskDefinition) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query daemonTaskDefinitionListModel

	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := r.Meta()
	conn := awsClient.ECSClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ecs.ListDaemonTaskDefinitionsInput{}

		for {
			output, err := conn.ListDaemonTaskDefinitions(ctx, input)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing ECS Daemon Task Definitions: %w", err))
				yield(result)
				return
			}

			for _, summary := range output.DaemonTaskDefinitions {
				result := request.NewListResult(ctx)

				var data daemonTaskDefinitionResourceModel
				r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
					dtd, err := findDaemonTaskDefinitionByARN(ctx, conn, aws.ToString(summary.Arn))
					if err != nil {
						result.Diagnostics.AddError("reading ECS Daemon Task Definition ("+aws.ToString(summary.Arn)+")", err.Error())
						return
					}

					result.Diagnostics.Append(flattenDaemonTaskDefinition(ctx, dtd, &data)...)
					if result.Diagnostics.HasError() {
						return
					}

					setTagsOut(ctx, nil)
					result.DisplayName = aws.ToString(summary.Arn)
				})

				if result.Diagnostics.HasError() {
					result = list.ListResult{Diagnostics: result.Diagnostics}
					yield(result)
					return
				}

				if !yield(result) {
					return
				}
			}

			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}
	}
}

type daemonTaskDefinitionListModel struct {
	framework.WithRegionModel
}
