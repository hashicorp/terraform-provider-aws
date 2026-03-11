// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ecs_task_definition")
func newTaskDefinitionResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceTaskDefinition{}
	l.SetResourceSchema(resourceTaskDefinition())
	return &l
}

var _ list.ListResource = &listResourceTaskDefinition{}

type listResourceTaskDefinition struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceTaskDefinition) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ECSClient(ctx)

	var query listTaskDefinitionModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing ECS (Elastic Container) Task Definition")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input ecs.ListTaskDefinitionsInput
		for arnStr, err := range listTaskDefinitions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), arnStr)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arnStr)
			rd.Set(names.AttrARN, arnStr)

			taskDefinition, tags, err := findTaskDefinitionByFamilyOrARN(ctx, conn, arnStr)
			if err != nil {
				tflog.Error(ctx, "Reading ECS (Elastic Container) Task Definition", map[string]any{
					names.AttrARN: arnStr,
					"err":         err.Error(),
				})
				continue
			}

			tflog.Info(ctx, "Reading ECS (Elastic Container) Task Definition")
			diags := resourceTaskDefinitionFlatten(ctx, rd, taskDefinition, tags)
			if diags.HasError() {
				tflog.Error(ctx, "Reading ECS (Elastic Container) Task Definition", map[string]any{
					"diags": sdkdiag.DiagnosticsString(diags),
				})
				continue
			}

			if rd.Id() == "" {
				tflog.Warn(ctx, "Resource disappeared during listing, skipping")
				continue
			}

			result.DisplayName = rd.Get(names.AttrFamily).(string)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listTaskDefinitionModel struct {
	framework.WithRegionModel
}

func listTaskDefinitions(ctx context.Context, conn *ecs.Client, input *ecs.ListTaskDefinitionsInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := ecs.NewListTaskDefinitionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield("", fmt.Errorf("listing ECS (Elastic Container) Task Definition resources: %w", err))
				return
			}

			for _, arnStr := range page.TaskDefinitionArns {
				if !yield(arnStr, nil) {
					return
				}
			}
		}
	}
}
