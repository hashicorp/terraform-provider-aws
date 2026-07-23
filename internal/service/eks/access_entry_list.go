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
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_eks_access_entry")
func newAccessEntryResourceAsListResource() inttypes.ListResourceForSDK {
	l := accessEntryListResource{}
	l.SetResourceSchema(resourceAccessEntry())
	return &l
}

var _ list.ListResource = &accessEntryListResource{}

type accessEntryListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *accessEntryListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"associated_policy_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Optional:    true,
				Description: "Only access entries associated to this access policy are returned.",
			},
			names.AttrClusterName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the cluster to list access entries from.",
			},
		},
	}
}

func (l *accessEntryListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EKSClient(ctx)

	var query listAccessEntryModel
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
		input := eks.ListAccessEntriesInput{
			AssociatedPolicyArn: fwflex.StringFromFramework(ctx, query.AssociatedPolicyARN),
			ClusterName:         aws.String(clusterName),
		}
		for item, err := range listAccessEntries(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), item)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(accessEntryCreateResourceID(clusterName, item))
			rd.Set(names.AttrClusterName, clusterName)
			rd.Set("principal_arn", item)

			if request.IncludeResource {
				accessEntry, err := findAccessEntryByTwoPartKey(ctx, conn, clusterName, item)
				if err != nil {
					tflog.Error(ctx, "Reading EKS Access Entry", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourceAccessEntryFlatten(ctx, accessEntry, rd); err != nil {
					tflog.Error(ctx, "Reading EKS Access Entry", map[string]any{
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

type listAccessEntryModel struct {
	framework.WithRegionModel
	AssociatedPolicyARN fwtypes.ARN  `tfsdk:"associated_policy_arn"`
	ClusterName         types.String `tfsdk:"cluster_name"`
}

func listAccessEntries(ctx context.Context, conn *eks.Client, input *eks.ListAccessEntriesInput, optFns ...func(*eks.Options)) iter.Seq2[string, error] {
	return tfiter.ConcatValuesWithError(listAccessEntryPages(ctx, conn, input, optFns...))
}

func listAccessEntryPages(ctx context.Context, conn *eks.Client, input *eks.ListAccessEntriesInput, optFns ...func(*eks.Options)) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		pages := eks.NewListAccessEntriesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing EKS Access Entries: %w", err))
				return
			}

			if !yield(page.AccessEntries, nil) {
				return
			}
		}
	}
}
