// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
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
			"cluster_arn": listschema.StringAttribute{
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
		input := &ecs.ListDaemonsInput{
			ClusterArn: query.ClusterArn.ValueStringPointer(),
		}

		for summary, err := range listDaemonSummaries(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data daemonResourceModel
			r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				daemon, err := findDaemonByARN(ctx, conn, aws.ToString(summary.DaemonArn))
				if err != nil {
					result.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon (%s)", aws.ToString(summary.DaemonArn)), err.Error())
					return
				}

				data.CapacityProviderArns = fwflex.FlattenFrameworkStringValueListOfString(ctx, []string{})

				result.Diagnostics.Append(fwflex.Flatten(ctx, daemon, &data)...)
				data.DaemonName = daemonNameFromARN(data.DaemonArn.ValueString())
				if result.Diagnostics.HasError() {
					return
				}

				if len(daemon.CurrentRevisions) > 0 && daemon.CurrentRevisions[0].Arn != nil {
					revision, err := findDaemonRevisionByARN(ctx, conn, aws.ToString(daemon.CurrentRevisions[0].Arn))
					if err != nil {
						result.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Revision (%s)", aws.ToString(daemon.CurrentRevisions[0].Arn)), err.Error())
						return
					}
					flattenDaemonRevision(ctx, revision, daemon.CurrentRevisions[0], &data)
				}

				setTagsOut(ctx, nil)

				if summary.DaemonArn != nil {
					result.DisplayName = daemonNameFromARN(aws.ToString(summary.DaemonArn)).ValueString()
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
	}
}

func listDaemonSummaries(ctx context.Context, conn *ecs.Client, input *ecs.ListDaemonsInput) iter.Seq2[awstypes.DaemonSummary, error] {
	return func(yield func(awstypes.DaemonSummary, error) bool) {
		err := listDaemonsPages(ctx, conn, input, func(page *ecs.ListDaemonsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}
			for _, summary := range page.DaemonSummariesList {
				if !yield(summary, nil) {
					return false
				}
			}
			return true
		})
		if err != nil {
			yield(awstypes.DaemonSummary{}, err)
		}
	}
}

type daemonListModel struct {
	framework.WithRegionModel
	ClusterArn types.String `tfsdk:"cluster_arn"`
}
