// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ssm_document")
func newDocumentResourceAsListResource() inttypes.ListResourceForSDK {
	l := documentListResource{}
	l.SetResourceSchema(resourceDocument())

	return &l
}

type documentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type documentListResourceModel struct {
	framework.WithRegionModel
	Filters fwtypes.ListNestedObjectValueOf[documentFiltersModel] `tfsdk:"filter"`
}

type documentFiltersModel struct {
	Key    types.String         `tfsdk:"key"`
	Values fwtypes.ListOfString `tfsdk:"values"`
}

func (l *documentListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Blocks: map[string]listschema.Block{
			names.AttrFilter: listschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[documentFiltersModel](ctx),
				NestedObject: listschema.NestedBlockObject{
					Attributes: map[string]listschema.Attribute{
						names.AttrKey: listschema.StringAttribute{
							Required: true,
						},
						names.AttrValues: listschema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func (l *documentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SSMClient(ctx)

	var query documentListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ssm.ListDocumentsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// if there are no filters set, default to returning documents that are owned by self
	if len(input.Filters) == 0 {
		input.Filters = append(input.Filters, awstypes.DocumentKeyValuesFilter{
			Key:    aws.String("Owner"),
			Values: []string{"Self"},
		})
	}

	tflog.Info(ctx, "Listing SSM documents")

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listDocuments(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				doc, err := findDocumentByName(ctx, conn, name)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					result = fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading SSM Document (%s): %w", name, err))
					yield(result)
					return
				}

				if diags := resourceDocumentFlatten(ctx, conn, rd, awsClient, doc); diags.HasError() {
					tflog.Error(ctx, "Flatten SSM Document", map[string]any{
						"diags": sdkdiag.DiagnosticsString(diags),
					})
					continue
				}
			}

			result.DisplayName = name

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

func listDocuments(ctx context.Context, conn *ssm.Client, input *ssm.ListDocumentsInput) iter.Seq2[awstypes.DocumentIdentifier, error] {
	return func(yield func(awstypes.DocumentIdentifier, error) bool) {
		pages := ssm.NewListDocumentsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.DocumentIdentifier{}, fmt.Errorf("listing SSM Documents: %w", err))
				return
			}

			for _, item := range page.DocumentIdentifiers {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
