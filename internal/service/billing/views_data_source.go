// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package billing

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/billing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/billing/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_billing_views", name="Views")
func newDataSourceViews(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceViews{}, nil
}

const (
	DSNameViews = "Views Data Source"
)

type dataSourceViews struct {
	framework.DataSourceWithModel[dataSourceViewsModel]
}

func (d *dataSourceViews) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"billing_view_types": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringEnumType[awstypes.BillingViewType](),
				Optional:    true,
				ElementType: types.StringType,
			},
			"billing_view": framework.ResourceComputedListOfObjectsAttribute[dataSourceBillingViewModel](ctx, nil, nil),
		},
	}
}

func (d *dataSourceViews) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().BillingClient(ctx)

	var data dataSourceViewsModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var billingViewTypes []awstypes.BillingViewType
	smerr.AddEnrich(ctx, &resp.Diagnostics, data.BillingViewTypes.ElementsAs(ctx, &billingViewTypes, false))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findViewsByViewTypes(ctx, conn, billingViewTypes)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.BillingViewTypes.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.BillingView, flex.WithFieldNamePrefix("Views")), smerr.ID, data.BillingViewTypes.String())
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.BillingViewTypes.String())
}

func findViewsByViewTypes(ctx context.Context, conn *billing.Client, billingViewTypes []awstypes.BillingViewType) ([]awstypes.BillingViewListElement, error) {
	input := billing.ListBillingViewsInput{}
	if len(billingViewTypes) > 0 {
		input.BillingViewTypes = billingViewTypes
	}

	return findViews(ctx, conn, &input)
}

func findViews(ctx context.Context, conn *billing.Client, input *billing.ListBillingViewsInput) ([]awstypes.BillingViewListElement, error) {
	var results []awstypes.BillingViewListElement

	paginator := billing.NewListBillingViewsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, page.BillingViews...)
	}

	return results, nil
}

type dataSourceViewsModel struct {
	BillingViewTypes fwtypes.ListOfStringEnum[awstypes.BillingViewType]          `tfsdk:"billing_view_types"`
	BillingView      fwtypes.ListNestedObjectValueOf[dataSourceBillingViewModel] `tfsdk:"billing_view"`
}

type dataSourceBillingViewModel struct {
	ARN             types.String `tfsdk:"arn"`
	BillingViewType types.String `tfsdk:"billing_view_type"`
	Description     types.String `tfsdk:"description"`
	Name            types.String `tfsdk:"name"`
	OwnerAccountId  types.String `tfsdk:"owner_account_id"`
}
