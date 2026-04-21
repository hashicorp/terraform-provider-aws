// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"iter"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_lambda_layer_version")
func newLayerVersionResourceAsListResource() inttypes.ListResourceForSDK {
	l := layerVersionListResource{}
	l.SetResourceSchema(resourceLayerVersion())
	return &l
}

var _ list.ListResource = &layerVersionListResource{}

type layerVersionListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type layerVersionListResourceModel struct {
	framework.WithRegionModel
}

func (l *layerVersionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.LambdaClient(ctx)

	var query layerVersionListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Lambda Layer Versions")

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listLayerVersionsAll(ctx, conn) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.layerVersion.LayerVersionArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set("layer_name", item.layerName)
			rd.Set(names.AttrVersion, strconv.FormatInt(item.layerVersion.Version, 10))

			if request.IncludeResource {
				tflog.Info(ctx, "Reading Lambda Layer Version")
				diags := resourceLayerVersionRead(ctx, rd, awsClient)
				if diags.HasError() {
					tflog.Error(ctx, "Reading Lambda Layer Version", map[string]any{"error": fmt.Sprintf("reading Lambda Layer Version (%s)", id)})
					continue
				}
				if rd.Id() == "" {
					continue
				}
			}

			result.DisplayName = id

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

type layerVersionListItem struct {
	layerName    string
	layerVersion awstypes.LayerVersionsListItem
}

func listLayerVersionsAll(ctx context.Context, conn *lambda.Client) iter.Seq2[layerVersionListItem, error] {
	return func(yield func(layerVersionListItem, error) bool) {
		pages := lambda.NewListLayersPaginator(conn, &lambda.ListLayersInput{})
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[layerVersionListItem](), fmt.Errorf("listing Lambda Layers: %w", err))
				return
			}

			for _, layer := range page.Layers {
				layerName := aws.ToString(layer.LayerName)
				versionPages := lambda.NewListLayerVersionsPaginator(conn, &lambda.ListLayerVersionsInput{
					LayerName: layer.LayerName,
				})
				for versionPages.HasMorePages() {
					versionPage, err := versionPages.NextPage(ctx)
					if err != nil {
						yield(inttypes.Zero[layerVersionListItem](), fmt.Errorf("listing Lambda Layer (%s) Versions: %w", layerName, err))
						return
					}

					for _, v := range versionPage.LayerVersions {
						if !yield(layerVersionListItem{layerName: layerName, layerVersion: v}, nil) {
							return
						}
					}
				}
			}
		}
	}
}
