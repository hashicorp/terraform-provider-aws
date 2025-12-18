// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"
	"strconv"

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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_redshift_data_share_consumer_association", name="Data Share Consumer Association")
func newDataShareConsumerAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &dataShareConsumerAssociationResource{}, nil
}

const (
	ResNameDataShareConsumerAssociation = "Data Share Consumer Association"
)

type dataShareConsumerAssociationResource struct {
	framework.ResourceWithModel[dataShareConsumerAssociationResourceModel]
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

const dataShareConsumerAssociationIDPartCount = 4

func (r *dataShareConsumerAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var plan dataShareConsumerAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataShareARN := plan.DataShareARN.ValueString()
	associateEntireAccountString := ""
	if plan.AssociateEntireAccount.ValueBool() {
		associateEntireAccountString = strconv.FormatBool(plan.AssociateEntireAccount.ValueBool())
	}
	consumerARN := plan.ConsumerARN.ValueString()
	consumerRegion := plan.ConsumerRegion.ValueString()
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
	plan.ID = types.StringValue(id)

	in := &redshift.AssociateDataShareConsumerInput{
		DataShareArn: aws.String(dataShareARN),
	}

	if !plan.AllowWrites.IsNull() {
		in.AllowWrites = plan.AllowWrites.ValueBoolPointer()
	}
	if !plan.AssociateEntireAccount.IsNull() {
		in.AssociateEntireAccount = plan.AssociateEntireAccount.ValueBoolPointer()
	}
	if !plan.ConsumerARN.IsNull() {
		in.ConsumerArn = aws.String(consumerARN)
	}
	if !plan.ConsumerRegion.IsNull() {
		in.ConsumerRegion = aws.String(consumerRegion)
	}

	out, err := conn.AssociateDataShareConsumer(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionCreating, ResNameDataShareConsumerAssociation, id, err),
			err.Error(),
		)
		return
	}
	if out == nil || len(out.DataShareAssociations) == 0 {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionCreating, ResNameDataShareConsumerAssociation, id, nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ProducerARN = fwflex.StringToFrameworkARN(ctx, out.ProducerArn)
	plan.ManagedBy = fwflex.StringToFramework(ctx, out.ManagedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *dataShareConsumerAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var state dataShareConsumerAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionSetting, ResNameDataShareConsumerAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

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

func (r *dataShareConsumerAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is a no-op
}

func (r *dataShareConsumerAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var state dataShareConsumerAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &redshift.DisassociateDataShareConsumerInput{
		DataShareArn: state.DataShareARN.ValueStringPointer(),
	}
	if !state.AssociateEntireAccount.IsNull() && state.AssociateEntireAccount.ValueBool() {
		in.DisassociateEntireAccount = aws.Bool(true)
	}
	if !state.ConsumerARN.IsNull() {
		in.ConsumerArn = state.ConsumerARN.ValueStringPointer()
	}
	if !state.ConsumerRegion.IsNull() {
		in.ConsumerRegion = state.ConsumerRegion.ValueStringPointer()
	}

	_, err := conn.DisassociateDataShareConsumer(ctx, in)
	if err != nil {
		if errs.IsAErrorMessageContains[*awstypes.InvalidDataShareFault](err, "because the ARN doesn't exist.") ||
			errs.IsAErrorMessageContains[*awstypes.InvalidDataShareFault](err, "either doesn't exist or isn't associated with this data consumer") {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionDeleting, ResNameDataShareConsumerAssociation, state.ID.String(), err),
			err.Error(),
		)
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
