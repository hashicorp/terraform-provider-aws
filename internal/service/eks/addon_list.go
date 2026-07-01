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
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_eks_addon")
func newAddonResourceAsListResource() inttypes.ListResourceForSDK {
	l := addonListResource{}
	l.SetResourceSchema(resourceAddon())
	return &l
}

var _ list.ListResource = &addonListResource{}

type addonListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *addonListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"cluster_name": listschema.StringAttribute{
				Required:    true,
				Description: "Name of the EKS Cluster to list add-ons for.",
			},
		},
	}
}

func (l *addonListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EKSClient(ctx)

	var query listAddonModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	clusterName := query.ClusterName.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("cluster_name"): clusterName,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := eks.ListAddonsInput{
			ClusterName: aws.String(clusterName),
		}
		for item, err := range listAddons(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			addonName := item
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("addon_name"), addonName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()

			rd.SetId(addonCreateResourceID(clusterName, addonName))
			rd.Set("cluster_name", clusterName)
			rd.Set("addon_name", addonName)

			if request.IncludeResource {

				addon, err := findAddonByTwoPartKey(ctx, conn, clusterName, addonName)
				if err != nil {
					tflog.Error(ctx, "Reading EKS Add-On", map[string]any{
						"error": err.Error(),
					})
					continue
				}
				if err := resourceAddonFlatten(ctx, l.Meta(), addon, rd); err != nil {
					tflog.Error(ctx, "Reading EKS (Elastic Kubernetes) Addon", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = addonName

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

type listAddonModel struct {
	framework.WithRegionModel
	ClusterName types.String `tfsdk:"cluster_name"`
}

func listAddons(ctx context.Context, conn *eks.Client, input *eks.ListAddonsInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := eks.NewListAddonsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[string](), fmt.Errorf("listing EKS (Elastic Kubernetes) Addon resources: %w", err))
				return
			}

			for _, item := range page.Addons {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
