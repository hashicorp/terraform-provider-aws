// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_redshift_data_share_consumer_association", name="Data Share Consumer Association")
func newDataShareConsumerAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &dataShareConsumerAssociationResource{}, nil
}

const (
	dataShareConsumerAssociationIDPartCount = 4
)

type dataShareConsumerAssociationResource struct {
	framework.ResourceWithModel[dataShareConsumerAssociationResourceModel]
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *dataShareConsumerAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allow_writes": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"associate_entire_account": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"consumer_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"consumer_region": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.AWSRegion(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_share_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"managed_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"producer_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *dataShareConsumerAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dataShareConsumerAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	dataShareARN := fwflex.StringValueFromFramework(ctx, plan.DataShareARN)
	var associateEntireAccountString string
	if v := fwflex.BoolValueFromFramework(ctx, plan.AssociateEntireAccount); v {
		associateEntireAccountString = intflex.BoolValueToStringValue(v)
	}
	consumerARN := fwflex.StringValueFromFramework(ctx, plan.ConsumerARN)
	consumerRegion := fwflex.StringValueFromFramework(ctx, plan.ConsumerRegion)
	parts := []string{
		dataShareARN,
		associateEntireAccountString,
		consumerARN,
		consumerRegion,
	}
	id, err := intflex.FlattenResourceId(parts, dataShareConsumerAssociationIDPartCount, true)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewCreatingResourceIDErrorDiagnostic(err))
		return
	}

	in := redshift.AssociateDataShareConsumerInput{
		AllowWrites:            fwflex.BoolFromFramework(ctx, plan.AllowWrites),
		AssociateEntireAccount: fwflex.BoolFromFramework(ctx, plan.AssociateEntireAccount),
		DataShareArn:           aws.String(dataShareARN),
	}

	if !plan.ConsumerARN.IsNull() {
		in.ConsumerArn = aws.String(consumerARN)
	}
	if !plan.ConsumerRegion.IsNull() {
		in.ConsumerRegion = aws.String(consumerRegion)
	}

	out, err := conn.AssociateDataShareConsumer(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating Redshift Data Share Consumer Association (%s)", id), err.Error())
		return
	}

	// Set values for unknowns.
	plan.ID = fwflex.StringValueToFramework(ctx, id)
	plan.ManagedBy = fwflex.StringToFramework(ctx, out.ManagedBy)
	plan.ProducerARN = fwflex.StringToFrameworkARN(ctx, out.ProducerArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *dataShareConsumerAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dataShareConsumerAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)
	parts, err := intflex.ExpandResourceId(id, dataShareConsumerAssociationIDPartCount, true)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	dataShareARN, associateEntireAccount, consumerARN, consumerRegion := parts[0], parts[1], parts[2], parts[3]
	out, err := findDataShareConsumerAssociationByFourPartKey(ctx, conn, dataShareARN, associateEntireAccount, consumerARN, consumerRegion)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading Redshift Data Share Consumer Association (%s)", id), err.Error())
		return
	}

	// Set attributes for import.
	if associateEntireAccount != "" {
		state.AssociateEntireAccount = fwflex.BoolValueToFramework(ctx, intflex.StringValueToBoolValue(associateEntireAccount))
	}
	if consumerARN != "" {
		state.ConsumerARN = fwtypes.ARNValue(consumerARN)
	}
	if consumerRegion != "" {
		state.ConsumerRegion = fwflex.StringValueToFramework(ctx, consumerRegion)
	}
	state.DataShareARN = fwtypes.ARNValue(dataShareARN)
	state.ProducerARN = fwflex.StringToFrameworkARN(ctx, out.ProducerArn)
	state.ManagedBy = fwflex.StringToFramework(ctx, out.ManagedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dataShareConsumerAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dataShareConsumerAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)
	parts, err := intflex.ExpandResourceId(id, dataShareConsumerAssociationIDPartCount, true)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	dataShareARN, associateEntireAccount, consumerARN, consumerRegion := parts[0], parts[1], parts[2], parts[3]

	in := redshift.DisassociateDataShareConsumerInput{
		DataShareArn: aws.String(dataShareARN),
	}
	if associateEntireAccount != "" {
		in.DisassociateEntireAccount = aws.Bool(true)
	}
	if consumerARN != "" {
		in.ConsumerArn = aws.String(consumerARN)
	}
	if consumerRegion != "" {
		in.ConsumerRegion = aws.String(consumerRegion)
	}
	_, err = conn.DisassociateDataShareConsumer(ctx, &in)
	if errs.IsAErrorMessageContains[*awstypes.InvalidDataShareFault](err, "because the ARN doesn't exist.") ||
		errs.IsAErrorMessageContains[*awstypes.InvalidDataShareFault](err, "either doesn't exist or isn't associated with this data consumer") {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting Redshift Data Share Consumer Association (%s)", id), err.Error())
		return
	}
}

func (r *dataShareConsumerAssociationResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("associate_entire_account"),
			path.MatchRoot("consumer_arn"),
			path.MatchRoot("consumer_region"),
		),
	}
}

type dataShareConsumerAssociationResourceModel struct {
	framework.WithRegionModel
	AllowWrites            types.Bool   `tfsdk:"allow_writes"`
	AssociateEntireAccount types.Bool   `tfsdk:"associate_entire_account"`
	ConsumerARN            fwtypes.ARN  `tfsdk:"consumer_arn"`
	ConsumerRegion         types.String `tfsdk:"consumer_region"`
	DataShareARN           fwtypes.ARN  `tfsdk:"data_share_arn"`
	ID                     types.String `tfsdk:"id"`
	ManagedBy              types.String `tfsdk:"managed_by"`
	ProducerARN            fwtypes.ARN  `tfsdk:"producer_arn"`
}
