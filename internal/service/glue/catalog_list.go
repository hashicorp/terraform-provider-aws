// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_glue_catalog")
func newCatalogResourceAsListResource() list.ListResourceWithConfigure {
	return &catalogListResource{}
}

var _ list.ListResource = &catalogListResource{}

type catalogListResource struct {
	catalogResource
	framework.WithList
}

type catalogListModel struct {
	framework.WithRegionModel
}

func (r *catalogListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := r.Meta()
	conn := awsClient.GlueClient(ctx)

	var query catalogListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		for catalog, err := range listCatalogs(ctx, conn, &glue.GetCatalogsInput{}) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(catalog.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			var data catalogResourceModel
			r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				var diags diag.Diagnostics
				readCatalogIntoModel(ctx, &catalog, &data, false, &diags)
				result.Diagnostics.Append(diags...)
				if result.Diagnostics.HasError() {
					return
				}

				tags, err := listTags(ctx, conn, data.ARN.ValueString())
				if err != nil {
					tflog.Error(ctx, "Listing Glue Catalog tags", map[string]any{
						names.AttrName: name,
						"error":        err.Error(),
					})
				} else {
					setTagsOut(ctx, tags.Map())
				}

				result.DisplayName = name
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

func listCatalogs(ctx context.Context, conn *glue.Client, input *glue.GetCatalogsInput) iter.Seq2[awstypes.Catalog, error] {
	return func(yield func(awstypes.Catalog, error) bool) {
		for {
			out, err := conn.GetCatalogs(ctx, input)
			if err != nil {
				yield(awstypes.Catalog{}, fmt.Errorf("listing Glue Catalogs: %w", err))
				return
			}
			for _, catalog := range out.CatalogList {
				if !yield(catalog, nil) {
					return
				}
			}
			if out.NextToken == nil {
				return
			}
			input.NextToken = out.NextToken
		}
	}
}
