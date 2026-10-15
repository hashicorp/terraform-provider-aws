// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package savingsplans

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/savingsplans"
	awstypes "github.com/aws/aws-sdk-go-v2/service/savingsplans/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_savingsplans_offerings", name="Offerings")
func newOfferingsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &offeringsDataSource{}, nil
}

type offeringsDataSource struct {
	framework.DataSourceWithModel[offeringsDataSourceModel]
}

func (d *offeringsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"currencies": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				CustomType:  fwtypes.SetOfStringEnumType[awstypes.CurrencyCode](),
			},
			"descriptions": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				CustomType:  fwtypes.SetOfStringType,
			},
			"durations": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				CustomType:  fwtypes.SetOfInt64Type,
			},
			"offerings": framework.DataSourceComputedListOfObjectAttribute[offeringModel](ctx),
			"offering_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				CustomType:  fwtypes.SetOfStringType,
			},
			"operations": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				CustomType:  fwtypes.SetOfStringType,
			},
			"payment_options": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				CustomType:  fwtypes.SetOfStringEnumType[awstypes.SavingsPlanPaymentOption](),
			},
			"plan_types": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				CustomType:  fwtypes.SetOfStringEnumType[awstypes.SavingsPlanType](),
			},
			"product_type": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.SavingsPlanProductType](),
			},
			"service_codes": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				CustomType:  fwtypes.SetOfStringType,
			},
			"usage_types": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				CustomType:  fwtypes.SetOfStringType,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[filterModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.SavingsPlanOfferingFilterAttribute](),
						},
						names.AttrValues: schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							CustomType:  fwtypes.ListOfStringType,
						},
					},
				},
			},
		},
	}
}

func (d *offeringsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SavingsPlansClient(ctx)

	var data offeringsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input savingsplans.DescribeSavingsPlansOfferingsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findOfferings(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.Offerings))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findOfferings(ctx context.Context, conn *savingsplans.Client, input *savingsplans.DescribeSavingsPlansOfferingsInput) ([]awstypes.SavingsPlanOffering, error) {
	var output []awstypes.SavingsPlanOffering
	err := describeSavingsPlansOfferingsPages(ctx, conn, input, func(page *savingsplans.DescribeSavingsPlansOfferingsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.SearchResults...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

type offeringsDataSourceModel struct {
	Currencies     fwtypes.SetOfStringEnum[awstypes.CurrencyCode]             `tfsdk:"currencies"`
	Descriptions   fwtypes.SetOfString                                        `tfsdk:"descriptions"`
	Durations      fwtypes.SetOfInt64                                         `tfsdk:"durations"`
	Filters        fwtypes.ListNestedObjectValueOf[filterModel]               `tfsdk:"filter" flex:"name=Filters"`
	OfferingIds    fwtypes.SetOfString                                        `tfsdk:"offering_ids"`
	Operations     fwtypes.SetOfString                                        `tfsdk:"operations"`
	PaymentOptions fwtypes.SetOfStringEnum[awstypes.SavingsPlanPaymentOption] `tfsdk:"payment_options"`
	PlanTypes      fwtypes.SetOfStringEnum[awstypes.SavingsPlanType]          `tfsdk:"plan_types"`
	ProductType    fwtypes.StringEnum[awstypes.SavingsPlanProductType]        `tfsdk:"product_type"`
	ServiceCodes   fwtypes.SetOfString                                        `tfsdk:"service_codes"`
	UsageTypes     fwtypes.SetOfString                                        `tfsdk:"usage_types"`
	Offerings      fwtypes.ListNestedObjectValueOf[offeringModel]             `tfsdk:"offerings"`
}

type filterModel struct {
	Name   fwtypes.StringEnum[awstypes.SavingsPlanOfferingFilterAttribute] `tfsdk:"name"`
	Values fwtypes.ListOfString                                            `tfsdk:"values"`
}

type offeringModel struct {
	Currency        fwtypes.StringEnum[awstypes.CurrencyCode]                `tfsdk:"currency"`
	Description     types.String                                             `tfsdk:"description"`
	DurationSeconds types.Int64                                              `tfsdk:"duration_seconds"`
	OfferingId      types.String                                             `tfsdk:"offering_id"`
	Operation       types.String                                             `tfsdk:"operation"`
	PaymentOption   fwtypes.StringEnum[awstypes.SavingsPlanPaymentOption]    `tfsdk:"payment_option"`
	PlanType        fwtypes.StringEnum[awstypes.SavingsPlanType]             `tfsdk:"plan_type"`
	ProductTypes    fwtypes.SetOfStringEnum[awstypes.SavingsPlanProductType] `tfsdk:"product_types"`
	Properties      fwtypes.ListNestedObjectValueOf[propertyModel]           `tfsdk:"properties"`
	ServiceCode     types.String                                             `tfsdk:"service_code"`
	UsageType       types.String                                             `tfsdk:"usage_type"`
}

type propertyModel struct {
	Name  fwtypes.StringEnum[awstypes.SavingsPlanOfferingPropertyKey] `tfsdk:"name"`
	Value types.String                                                `tfsdk:"value"`
}
