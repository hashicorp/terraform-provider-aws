// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrockagentcore_registry")
func newRegistryResourceAsListResource() list.ListResourceWithConfigure {
	return &registryListResource{}
}

var _ list.ListResource = &registryListResource{}

type registryListResource struct {
	registryResource
	framework.WithList
}

func (l *registryListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		DeprecationMessage: "This resource is deprecated and will continue to work until September 17, 2026.",
	}
}

func (l *registryListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockAgentCoreClient(ctx)

	var query listRegistryModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input bedrockagentcorecontrol.ListRegistriesInput
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
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
				if result.Diagnostics.HasError() {
					return
				}

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

func listRegistries(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListRegistriesInput) iter.Seq2[awstypes.RegistrySummary, error] {
	return func(yield func(awstypes.RegistrySummary, error) bool) {
		pages := bedrockagentcorecontrol.NewListRegistriesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.RegistrySummary](), fmt.Errorf("listing Bedrock AgentCore Registries: %w", err))
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
