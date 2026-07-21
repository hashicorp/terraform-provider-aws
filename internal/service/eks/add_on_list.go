// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_eks_addon")
func newAddonResourceAsListResource() inttypes.ListResourceForSDK {
	l := addOnListResource{}
	l.SetResourceSchema(resourceAddon())
	return &l
}

var _ list.ListResource = &addOnListResource{}

type addOnListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *addOnListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrClusterName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the cluster to list add-ons from.",
			},
		},
	}
}

func (l *addOnListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EKSClient(ctx)

	var query listAddOnModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	clusterName := fwflex.StringValueFromFramework(ctx, query.ClusterName)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey(names.AttrClusterName): clusterName,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := eks.ListAddonsInput{
			ClusterName: aws.String(clusterName),
		}
		for item, err := range listAddOns(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), item)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(addonCreateResourceID(clusterName, item))
			rd.Set("addon_name", item)
			rd.Set(names.AttrClusterName, clusterName)

			if request.IncludeResource {
				addon, err := findAddonByTwoPartKey(ctx, conn, clusterName, item)
				if err != nil {
					tflog.Error(ctx, "Reading EKS Add-On", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourceAddonFlatten(ctx, conn, addon, rd); err != nil {
					tflog.Error(ctx, "Reading EKS Add-On", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = item

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

type listAddOnModel struct {
	framework.WithRegionModel
	ClusterName types.String `tfsdk:"cluster_name"`
}

func listAddOns(ctx context.Context, conn *eks.Client, input *eks.ListAddonsInput, optFns ...func(*eks.Options)) iter.Seq2[string, error] {
	return tfiter.ConcatValuesWithError(listAddonPages(ctx, conn, input, optFns...))
}

func listAddonPages(ctx context.Context, conn *eks.Client, input *eks.ListAddonsInput, optFns ...func(*eks.Options)) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		pages := eks.NewListAddonsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing EKS Add-Ons: %w", err))
				return
			}

			if !yield(page.Addons, nil) {
				return
			}
		}
	}
}
