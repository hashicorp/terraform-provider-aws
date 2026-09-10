// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package eks

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// @FrameworkDataSource("aws_eks_access_policies", name="Access Policies")
func newAccessPoliciesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &accessPoliciesDataSource{}, nil
}

type accessPoliciesDataSource struct {
	framework.DataSourceWithModel[accessPoliciesDataSourceModel]
}

func (d *accessPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_policies": framework.DataSourceComputedListOfObjectAttribute[accessPolicyModel](ctx),
		},
	}
}

func (d *accessPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data accessPoliciesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EKSClient(ctx)

	var input eks.ListAccessPoliciesInput
	output, err := findAccessPolicies(ctx, conn, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			"listing EKS Access Policies",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.AccessPolicies)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findAccessPolicies(ctx context.Context, conn *eks.Client, input *eks.ListAccessPoliciesInput, optFns ...tfslices.FinderOptionsFunc[awstypes.AccessPolicy]) ([]awstypes.AccessPolicy, error) {
	return tfslices.CollectAndConcatWithError(listAccessPolicyPages(ctx, conn, input), optFns...)
}

func listAccessPolicyPages(ctx context.Context, conn *eks.Client, input *eks.ListAccessPoliciesInput, optFns ...func(*eks.Options)) iter.Seq2[[]awstypes.AccessPolicy, error] {
	return func(yield func([]awstypes.AccessPolicy, error) bool) {
		pages := eks.NewListAccessPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing EKS Access Policies: %w", err))
				return
			}

			if !yield(page.AccessPolicies, nil) {
				return
			}
		}
	}
}

type accessPoliciesDataSourceModel struct {
	framework.WithRegionModel
	AccessPolicies fwtypes.ListNestedObjectValueOf[accessPolicyModel] `tfsdk:"access_policies"`
}

type accessPolicyModel struct {
	ARN  fwtypes.ARN  `tfsdk:"arn"`
	Name types.String `tfsdk:"name"`
}
