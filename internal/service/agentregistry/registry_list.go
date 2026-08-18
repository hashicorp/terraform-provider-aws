// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_agentregistry_registry")
func newRegistryResourceAsListResource() list.ListResourceWithConfigure {
	return &registryListResource{}
}

var _ list.ListResource = &registryListResource{}

type registryListResource struct {
	registryResource
	framework.WithList
}

func (l *registryListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().AgentRegistryClient(ctx)

	var query listRegistryModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input agentregistrycontrol.ListRegistriesInput
		for item, err := range listRegistries(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			registryID := aws.ToString(item.RegistryId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), aws.ToString(item.RegistryArn))

			output, err := findRegistryByID(ctx, conn, registryID)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data registryResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.Name = types.StringValue(aws.ToString(output.Name))
				data.Description = types.StringPointerValue(output.Description)
				data.RegistryARN = types.StringValue(aws.ToString(output.RegistryArn))
				data.RegistryID = types.StringValue(aws.ToString(output.RegistryId))
				data.Status = fwtypes.StringEnumValue(output.Status)

				result.Diagnostics.Append(flattenApprovalConfiguration(ctx, output.ApprovalConfiguration, &data)...)
				result.Diagnostics.Append(flattenDiscoveryConfiguration(ctx, output.DiscoveryConfiguration, &data)...)

				result.DisplayName = aws.ToString(item.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listRegistryModel struct {
	framework.WithRegionModel
}

func listRegistries(ctx context.Context, conn *agentregistrycontrol.Client, input *agentregistrycontrol.ListRegistriesInput) iter.Seq2[awstypes.RegistrySummary, error] {
	return func(yield func(awstypes.RegistrySummary, error) bool) {
		pages := agentregistrycontrol.NewListRegistriesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.RegistrySummary](), fmt.Errorf("listing Agent Registry Registries: %w", err))
				return
			}

			for _, item := range page.Registries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
