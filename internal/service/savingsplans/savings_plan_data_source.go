// Copyright IBM Corp. 2014, 2025
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_savingsplans_plan", name="Savings Plan")
// @Tags
func newDataSourceSavingsPlan(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceSavingsPlan{}, nil
}

type dataSourceSavingsPlan struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceSavingsPlan) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed:    true,
				Description: "The ARN of the Savings Plan.",
			},
			names.AttrID: schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Savings Plan.",
			},
			names.AttrState: schema.StringAttribute{
				Computed:    true,
				Description: "The current state of the Savings Plan.",
			},
			names.AttrStart: schema.StringAttribute{
				Computed:    true,
				Description: "The start time of the Savings Plan.",
			},
			"end": schema.StringAttribute{
				Computed:    true,
				Description: "The end time of the Savings Plan.",
			},
			"savings_plan_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of Savings Plan.",
			},
			"payment_option": schema.StringAttribute{
				Computed:    true,
				Description: "The payment option for the Savings Plan.",
			},
			"currency": schema.StringAttribute{
				Computed:    true,
				Description: "The currency of the Savings Plan.",
			},
			"commitment": schema.StringAttribute{
				Computed:    true,
				Description: "The hourly commitment, in USD.",
			},
			"upfront_payment_amount": schema.StringAttribute{
				Computed:    true,
				Description: "The up-front payment amount.",
			},
			"recurring_payment_amount": schema.StringAttribute{
				Computed:    true,
				Description: "The recurring payment amount.",
			},
			"term_duration_in_seconds": schema.Int64Attribute{
				Computed:    true,
				Description: "The duration of the term, in seconds.",
			},
			"ec2_instance_family": schema.StringAttribute{
				Computed:    true,
				Description: "The EC2 instance family for the Savings Plan.",
			},
			names.AttrRegion: schema.StringAttribute{
				Computed:    true,
				Description: "The AWS Region.",
			},
			"offering_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the offering.",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *dataSourceSavingsPlan) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SavingsPlansClient(ctx)

	var data dataSourceSavingsPlanModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSavingsPlanByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenSavingsPlanDataSource(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	// Set tags
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)
	tags := KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	data.Tags = tftags.FlattenStringValueMap(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func flattenSavingsPlanDataSource(ctx context.Context, sp *awstypes.SavingsPlan, model *dataSourceSavingsPlanModel) error {
	model.ARN = flex.StringToFramework(ctx, sp.SavingsPlanArn)
	model.ID = flex.StringToFramework(ctx, sp.SavingsPlanId)
	model.State = types.StringValue(string(sp.State))
	model.SavingsPlanType = types.StringValue(string(sp.SavingsPlanType))
	model.PaymentOption = types.StringValue(string(sp.PaymentOption))
	model.Currency = types.StringValue(string(sp.Currency))
	model.Commitment = flex.StringToFramework(ctx, sp.Commitment)
	model.UpfrontPaymentAmount = flex.StringToFramework(ctx, sp.UpfrontPaymentAmount)
	model.RecurringPaymentAmount = flex.StringToFramework(ctx, sp.RecurringPaymentAmount)
	model.TermDurationInSeconds = types.Int64PointerValue(sp.TermDurationInSeconds)
	model.EC2InstanceFamily = flex.StringToFramework(ctx, sp.Ec2InstanceFamily)
	model.Region = flex.StringToFramework(ctx, sp.Region)
	model.OfferingID = flex.StringToFramework(ctx, sp.OfferingId)

	if sp.Start != nil {
		model.Start = types.StringValue(sp.Start.Format("2006-01-02T15:04:05Z"))
	}
	if sp.End != nil {
		model.End = types.StringValue(sp.End.Format("2006-01-02T15:04:05Z"))
	}

	return nil
}

type dataSourceSavingsPlanModel struct {
	ARN                    types.String `tfsdk:"arn"`
	Commitment             types.String `tfsdk:"commitment"`
	Currency               types.String `tfsdk:"currency"`
	EC2InstanceFamily      types.String `tfsdk:"ec2_instance_family"`
	End                    types.String `tfsdk:"end"`
	ID                     types.String `tfsdk:"id"`
	OfferingID             types.String `tfsdk:"offering_id"`
	PaymentOption          types.String `tfsdk:"payment_option"`
	RecurringPaymentAmount types.String `tfsdk:"recurring_payment_amount"`
	Region                 types.String `tfsdk:"region"`
	SavingsPlanType        types.String `tfsdk:"savings_plan_type"`
	Start                  types.String `tfsdk:"start"`
	State                  types.String `tfsdk:"state"`
	Tags                   tftags.Map   `tfsdk:"tags"`
	TermDurationInSeconds  types.Int64  `tfsdk:"term_duration_in_seconds"`
	UpfrontPaymentAmount   types.String `tfsdk:"upfront_payment_amount"`
}

// KeyValueTags creates tftags.KeyValueTags from savingsplans service tags.
func KeyValueTags(ctx context.Context, tags map[string]string) tftags.KeyValueTags {
	return tftags.New(ctx, tags)
}
