// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkListResource("aws_ecs_daemon")
func daemonResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceDaemon{}
}

var _ list.ListResource = &listResourceDaemon{}

type listResourceDaemon struct {
	daemonResource
	framework.WithList
}

func (r *listResourceDaemon) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"cluster": listschema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *listResourceDaemon) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query daemonListModel

	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := r.Meta()
	conn := awsClient.ECSClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ecs.ListDaemonsInput{}
		if !query.ClusterArn.IsNull() {
			input.ClusterArn = aws.String(query.ClusterArn.ValueString())
		}

		for {
			output, err := conn.ListDaemons(ctx, input)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing ECS Daemons: %w", err))
				yield(result)
				return
			}

			for _, summary := range output.DaemonSummariesList {
				result := request.NewListResult(ctx)

				var data daemonResourceModel
				r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
					daemon, err := findDaemonByARN(ctx, conn, aws.ToString(summary.DaemonArn))
					if err != nil {
						result.Diagnostics.AddError("reading ECS Daemon ("+aws.ToString(summary.DaemonArn)+")", err.Error())
						return
					}

					data.CapacityProviderArns = fwflex.FlattenFrameworkStringValueListOfString(ctx, []string{})

					result.Diagnostics.Append(flattenDaemon(ctx, conn, daemon, &data)...)
					if result.Diagnostics.HasError() {
						return
					}

					setTagsOut(ctx, nil)

					if data.EnableECSManagedTags.IsNull() {
						data.EnableECSManagedTags = types.BoolValue(false)
					}
					if data.EnableExecuteCommand.IsNull() {
						data.EnableExecuteCommand = types.BoolValue(false)
					}

					// Display name: extract daemon name from ARN
					if summary.DaemonArn != nil {
						arnParts := strings.Split(aws.ToString(summary.DaemonArn), "/")
						if len(arnParts) >= 3 {
							result.DisplayName = arnParts[len(arnParts)-1]
						}
					}
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

type daemonListModel struct {
	framework.WithRegionModel
	ClusterArn types.String `tfsdk:"cluster"`
}
