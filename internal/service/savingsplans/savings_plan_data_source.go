// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package savingsplans

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_savingsplans_savings_plan", name="Savings Plan")
// @Tags
func newSavingsPlanDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &savingsPlanDataSource{}, nil
}

type savingsPlanDataSource struct {
	framework.DataSourceWithModel[savingsPlanDataSourceModel]
}

func (d *savingsPlanDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"commitment": schema.StringAttribute{
				Computed:    true,
				Description: "The hourly commitment, in USD.",
			},
			"currency": schema.StringAttribute{
				Computed:    true,
				Description: "The currency of the Savings Plan.",
			},
			names.AttrDescription: schema.StringAttribute{
				Computed:    true,
				Description: "The description.",
			},
			"ec2_instance_family": schema.StringAttribute{
				Computed:    true,
				Description: "The EC2 instance family for the Savings Plan.",
			},
			"end": schema.StringAttribute{
				Computed:    true,
				Description: "The end time of the Savings Plan.",
			},
			"offering_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the offering.",
			},
			"payment_option": schema.StringAttribute{
				Computed:    true,
				Description: "The payment option for the Savings Plan.",
			},
			"product_types": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "The product types.",
			},
			"purchase_time": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "The time at which to purchase the Savings Plan, in UTC format (YYYY-MM-DDTHH:MM:SSZ).",
			},
			"recurring_payment_amount": schema.StringAttribute{
				Computed:    true,
				Description: "The recurring payment amount.",
			},
			names.AttrRegion: schema.StringAttribute{
				Computed:    true,
				Description: "The AWS Region.",
			},
			"returnable_until": schema.StringAttribute{
				Computed:    true,
				Description: "The recurring payment amount.",
			},
			"savings_plan_arn": framework.ARNAttributeComputedOnly(),
			"savings_plan_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Savings Plan.",
			},
			"savings_plan_offering_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique ID of a Savings Plan offering.",
			},
			"savings_plan_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of Savings Plan.",
			},
			"start": schema.StringAttribute{
				Computed:    true,
				Description: "The start time of the Savings Plan.",
			},
			names.AttrState: schema.StringAttribute{
				Computed:    true,
				Description: "The current state of the Savings Plan.",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"term_duration_in_seconds": schema.Int64Attribute{
				Computed:    true,
				Description: "The duration of the term, in seconds.",
			},
			"upfront_payment_amount": schema.StringAttribute{
				Computed:    true,
				Description: "The up-front payment amount.",
			},
		},
	}
}

func (d *savingsPlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data savingsPlanDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SavingsPlansClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.SavingsPlanID)
	out, err := findSavingsPlanByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type savingsPlanDataSourceModel struct {
	Commitment             types.String         `tfsdk:"commitment"`
	Currency               types.String         `tfsdk:"currency"`
	Description            types.String         `tfsdk:"description"`
	EC2InstanceFamily      types.String         `tfsdk:"ec2_instance_family"`
	End                    types.String         `tfsdk:"end"`
	OfferingID             types.String         `tfsdk:"offering_id"`
	PaymentOption          types.String         `tfsdk:"payment_option"`
	ProductTypes           fwtypes.ListOfString `tfsdk:"product_types"`
	PurchaseTime           timetypes.RFC3339    `tfsdk:"purchase_time"`
	RecurringPaymentAmount types.String         `tfsdk:"recurring_payment_amount"`
	Region                 types.String         `tfsdk:"region"`
	ReturnableUntil        types.String         `tfsdk:"returnable_until"`
	SavingsPlanARN         types.String         `tfsdk:"savings_plan_arn"`
	SavingsPlanID          types.String         `tfsdk:"savings_plan_id"`
	SavingsPlanOfferingID  types.String         `tfsdk:"savings_plan_offering_id"`
	SavingsPlanType        types.String         `tfsdk:"savings_plan_type"`
	Start                  types.String         `tfsdk:"start"`
	State                  types.String         `tfsdk:"state"`
	Tags                   tftags.Map           `tfsdk:"tags"`
	TermDurationInSeconds  types.Int64          `tfsdk:"term_duration_in_seconds"`
	UpfrontPaymentAmount   types.String         `tfsdk:"upfront_payment_amount"`
}
