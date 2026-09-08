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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
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

	stream.Results = func(yield func(list.ListResult) bool) {
		var input agentregistrycontrol.ListRegistriesInput
		for item, err := range listRegistries(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			registryID := aws.ToString(item.RegistryId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), registryID)

			var output *agentregistrycontrol.GetRegistryOutput
			if request.IncludeResource {
				var err error
				output, err = findRegistryByID(ctx, conn, registryID)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data registryResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.RegistryID = fwflex.StringValueToFramework(ctx, registryID)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}
			if !yield(result) {
				return
			}
		}
	}
}

type listRegistryModel struct {
	framework.WithRegionModel
}

func listRegistries(ctx context.Context, conn *agentregistrycontrol.Client, input *agentregistrycontrol.ListRegistriesInput, optFns ...func(*agentregistrycontrol.Options)) iter.Seq2[awstypes.RegistrySummary, error] {
	return tfiter.ConcatValuesWithError(listRegistryPages(ctx, conn, input, optFns...))
}

func listRegistryPages(ctx context.Context, conn *agentregistrycontrol.Client, input *agentregistrycontrol.ListRegistriesInput, optFns ...func(*agentregistrycontrol.Options)) iter.Seq2[[]awstypes.RegistrySummary, error] {
	return func(yield func([]awstypes.RegistrySummary, error) bool) {
		pages := agentregistrycontrol.NewListRegistriesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Agent Registry Registries: %w", err))
				return
			}

			if !yield(page.Registries, nil) {
				return
			}
		}
	}
}
