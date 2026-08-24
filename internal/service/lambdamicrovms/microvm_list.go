// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdamicrovms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkListResource("aws_lambdamicrovms_microvm")
func newMicroVMResourceAsListResource() list.ListResourceWithConfigure {
	return &microVMListResource{}
}

var _ list.ListResource = &microVMListResource{}

type microVMListResource struct {
	microVMResource
	framework.WithList
}

func (l *microVMListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"image_identifier": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter to list only MicroVMs running the specified image.",
			},
			"image_version": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter to list only MicroVMs running the specified image version.",
			},
		},
	}
}

func (l *microVMListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LambdaMicroVMsClient(ctx)

	var query listMicroVMModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input lambdamicrovms.ListMicrovmsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listMicroVMs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			microVMID := aws.ToString(item.MicrovmId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), microVMID)

			var output *lambdamicrovms.GetMicrovmOutput
			if request.IncludeResource {
				var err error
				output, err = findMicroVMByID(ctx, conn, microVMID)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data microVMResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.MicroVMID = fwflex.StringValueToFramework(ctx, microVMID)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = microVMID
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listMicroVMModel struct {
	framework.WithRegionModel
	ImageIdentifier types.String `tfsdk:"image_identifier"`
	ImageVersion    types.String `tfsdk:"image_version"`
}

func listMicroVMs(ctx context.Context, conn *lambdamicrovms.Client, input *lambdamicrovms.ListMicrovmsInput, optFns ...func(*lambdamicrovms.Options)) iter.Seq2[awstypes.MicrovmItem, error] {
	return tfiter.ConcatValuesWithError(listMicroVMPages(ctx, conn, input, optFns...))
}

func listMicroVMPages(ctx context.Context, conn *lambdamicrovms.Client, input *lambdamicrovms.ListMicrovmsInput, optFns ...func(*lambdamicrovms.Options)) iter.Seq2[[]awstypes.MicrovmItem, error] {
	return func(yield func([]awstypes.MicrovmItem, error) bool) {
		pages := lambdamicrovms.NewListMicrovmsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Lambda MicroVMs Micro VMs: %w", err))
				return
			}

			if !yield(page.Items, nil) {
				return
			}
		}
	}
}
