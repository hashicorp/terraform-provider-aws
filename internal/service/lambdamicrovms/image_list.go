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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_lambdamicrovms_image")
func newImageResourceAsListResource() list.ListResourceWithConfigure {
	return &imageListResource{}
}

var _ list.ListResource = &imageListResource{}

type imageListResource struct {
	imageResource
	framework.WithList
}

func (l *imageListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"name_filter": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter to list only MicroVM images whose name contains the specified string.",
			},
		},
	}
}

func (l *imageListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LambdaMicroVMsClient(ctx)

	var query listImageModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input lambdamicrovms.ListMicrovmImagesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listImages(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			arn := aws.ToString(item.ImageArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var output *lambdamicrovms.GetMicrovmImageOutput
			if request.IncludeResource {
				output, err = findImageByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data imageResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
					if result.Diagnostics.HasError() {
						return
					}

					setTagsOut(ctx, output.Tags)
				} else {
					result.Diagnostics.Append(l.flattenSummary(ctx, &item, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = data.Name.ValueString()
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

type listImageModel struct {
	framework.WithRegionModel
	NameFilter types.String `tfsdk:"name_filter"`
}

func listImages(ctx context.Context, conn *lambdamicrovms.Client, input *lambdamicrovms.ListMicrovmImagesInput, optFns ...func(*lambdamicrovms.Options)) iter.Seq2[awstypes.MicrovmImageSummary, error] {
	return func(yield func(awstypes.MicrovmImageSummary, error) bool) {
		pages := lambdamicrovms.NewListMicrovmImagesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(awstypes.MicrovmImageSummary{}, fmt.Errorf("listing Lambda MicroVMs Images: %w", err))
				return
			}

			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

func (r *imageResource) flattenSummary(ctx context.Context, image *awstypes.MicrovmImageSummary, data *imageResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, image, data, fwflex.WithFieldNamePrefix("Image"))...)
	return diags
}
