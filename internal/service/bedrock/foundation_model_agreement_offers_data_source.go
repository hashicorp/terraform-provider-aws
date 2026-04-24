// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrock

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bedrock_foundation_model_agreement_offers", name="Foundation Model Agreement Offers")
func newFoundationModelAgreementOffersDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &foundationModelAgreementOffersDataSource{}, nil
}

const (
	DSNameFoundationModelAgreementOffers = "Foundation Model Agreement Offers Data Source"
)

type foundationModelAgreementOffersDataSource struct {
	framework.DataSourceWithModel[foundationModelAgreementOffersDataSourceModel]
}

func (d *foundationModelAgreementOffersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"model_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([a-z0-9-]{1,63}[.]){0,2}[a-z0-9-]{1,63}([:][a-z0-9-]{1,63}){0,2}(/[a-z0-9]{12}|)$`), ""),
				},
			},
			"offer_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OfferType](),
				Optional:   true,
			},
			"offers": framework.DataSourceComputedListOfObjectAttribute[foundationModelOfferModel](ctx),
		},
	}
}

func (d *foundationModelAgreementOffersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().BedrockClient(ctx)

	var data foundationModelAgreementOffersDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := &bedrock.ListFoundationModelAgreementOffersInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.ListFoundationModelAgreementOffers(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("listing Bedrock Foundation Model Agreement Offers", err.Error())

		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(d.Meta().Region(ctx))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type foundationModelAgreementOffersDataSourceModel struct {
	framework.WithRegionModel
	ID        types.String                                               `tfsdk:"id"`
	OfferType fwtypes.StringEnum[awstypes.OfferType]                     `tfsdk:"offer_type"`
	ModelID   types.String                                               `tfsdk:"model_id"`
	Offers    fwtypes.ListNestedObjectValueOf[foundationModelOfferModel] `tfsdk:"offers"`
}

type foundationModelOfferModel struct {
	TermDetails fwtypes.ListNestedObjectValueOf[foundationModelTermDetailsModel] `tfsdk:"term_details"`
}

type foundationModelTermDetailsModel struct {
	LegalTerm             fwtypes.ListNestedObjectValueOf[foundationModelLegalTermModel]             `tfsdk:"legal_term"`
	SupportTerm           fwtypes.ListNestedObjectValueOf[foundationModelSupportTermModel]           `tfsdk:"support_term"`
	UsageBasedPricingTerm fwtypes.ListNestedObjectValueOf[foundationModelUsageBasedPricingTermModel] `tfsdk:"usage_based_pricing_term"`
	ValidityTerm          fwtypes.ListNestedObjectValueOf[foundationModelValidityTermModel]          `tfsdk:"validity_term"`
}

type foundationModelLegalTermModel struct {
	URL types.String `tfsdk:"url"`
}

type foundationModelSupportTermModel struct {
	RefundPolicyDescription types.String `tfsdk:"refund_policy_description"`
}

type foundationModelUsageBasedPricingTermModel struct {
	RateCard fwtypes.ListNestedObjectValueOf[foundationModelRateCardModel] `tfsdk:"rate_card"`
}

type foundationModelRateCardModel struct {
	Description types.String `tfsdk:"description"`
	Dimension   types.String `tfsdk:"dimension"`
	Price       types.String `tfsdk:"price"`
	Unit        types.String `tfsdk:"unit"`
}

type foundationModelValidityTermModel struct {
	AgreementDuration types.String `tfsdk:"agreement_duration"`
}
