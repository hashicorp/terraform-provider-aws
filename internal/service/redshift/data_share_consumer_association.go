// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Data Share Consumer Association")
func newResourceDataShareConsumerAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceDataShareConsumerAssociation{}, nil
}

const (
	ResNameDataShareConsumerAssociation = "Data Share Consumer Association"
)

type resourceDataShareConsumerAssociation struct {
	framework.ResourceWithConfigure
}

func (r *resourceDataShareConsumerAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_redshift_data_share_consumer_association"
}

func (r *resourceDataShareConsumerAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *resourceDataShareConsumerAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RedshiftConn(ctx)

	var plan resourceDataShareConsumerAssociationData
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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionFlatteningResourceId, ResNameDataShareConsumerAssociation, dataShareARN, err),
			err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(id)

	in := &redshift.AssociateDataShareConsumerInput{
		DataShareArn: aws.String(dataShareARN),
	}

	if !plan.AllowWrites.IsNull() {
		in.AllowWrites = aws.Bool(plan.AllowWrites.ValueBool())
	}
	if !plan.AssociateEntireAccount.IsNull() {
		in.AssociateEntireAccount = aws.Bool(plan.AssociateEntireAccount.ValueBool())
	}
	if !plan.ConsumerARN.IsNull() {
		in.ConsumerArn = aws.String(consumerARN)
	}
	if !plan.ConsumerRegion.IsNull() {
		in.ConsumerRegion = aws.String(consumerRegion)
	}

	out, err := conn.AssociateDataShareConsumerWithContext(ctx, in)
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

	plan.ProducerARN = flex.StringToFrameworkARN(ctx, out.ProducerArn)
	plan.ManagedBy = flex.StringToFramework(ctx, out.ManagedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDataShareConsumerAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RedshiftConn(ctx)

	var state resourceDataShareConsumerAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts, err := intflex.ExpandResourceId(state.ID.ValueString(), dataShareConsumerAssociationIDPartCount, true)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionExpandingResourceId, ResNameDataShareAuthorization, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	// split ID and write constituent parts to state to support import
	state.DataShareARN = fwtypes.ARNValue(parts[0])
	if parts[1] != "" {
		state.AssociateEntireAccount = types.BoolValue(parts[1] == "true")
	}
	if parts[2] != "" {
		state.ConsumerARN = fwtypes.ARNValue(parts[2])
	}
	if parts[3] != "" {
		state.ConsumerRegion = types.StringValue(parts[3])
	}

	out, err := findDataShareConsumerAssociationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
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

	state.ProducerARN = flex.StringToFrameworkARN(ctx, out.ProducerArn)
	state.ManagedBy = flex.StringToFramework(ctx, out.ManagedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDataShareConsumerAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is a no-op
}

func (r *resourceDataShareConsumerAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RedshiftConn(ctx)

	var state resourceDataShareConsumerAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &redshift.DisassociateDataShareConsumerInput{
		DataShareArn: aws.String(state.DataShareARN.ValueString()),
	}
	if !state.AssociateEntireAccount.IsNull() && state.AssociateEntireAccount.ValueBool() {
		in.DisassociateEntireAccount = aws.Bool(true)
	}
	if !state.ConsumerARN.IsNull() {
		in.ConsumerArn = aws.String(state.ConsumerARN.ValueString())
	}
	if !state.ConsumerRegion.IsNull() {
		in.ConsumerRegion = aws.String(state.ConsumerRegion.ValueString())
	}

	_, err := conn.DisassociateDataShareConsumerWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrMessageContains(err, redshift.ErrCodeInvalidDataShareFault, "because the ARN doesn't exist.") ||
			tfawserr.ErrMessageContains(err, redshift.ErrCodeInvalidDataShareFault, "either doesn't exist or isn't associated with this data consumer") {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionDeleting, ResNameDataShareConsumerAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDataShareConsumerAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}
func (r *resourceDataShareConsumerAssociation) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("associate_entire_account"),
			path.MatchRoot("consumer_arn"),
			path.MatchRoot("consumer_region"),
		),
	}
}

func findDataShareConsumerAssociationByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.DataShare, error) {
	parts, err := intflex.ExpandResourceId(id, dataShareConsumerAssociationIDPartCount, true)
	if err != nil {
		return nil, err
	}
	dataShareARN := parts[0]
	associateEntireAccount := parts[1]
	consumerARN := parts[2]
	consumerRegion := parts[3]

	in := &redshift.DescribeDataSharesInput{
		DataShareArn: aws.String(dataShareARN),
	}

	out, err := conn.DescribeDataSharesWithContext(ctx, in)
	if tfawserr.ErrMessageContains(err, redshift.ErrCodeInvalidDataShareFault, "because the ARN doesn't exist.") ||
		tfawserr.ErrMessageContains(err, redshift.ErrCodeInvalidDataShareFault, "either doesn't exist or isn't associated with this data consumer") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.DataShares) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}
	if len(out.DataShares) != 1 {
		return nil, tfresource.NewTooManyResultsError(len(out.DataShares), in)
	}

	share := out.DataShares[0]

	// The data share should include an association in an "ACTIVE" status where
	// one of the following is true:
	// - `associate_entire_account` is `true` and `ConsumerIdentifier` matches the
	//   account number of the data share ARN.
	// - `consumer_arn` is set and `ConsumerIdentifier` matches its value.
	// - `consumer_region` is set and `ConsumerRegion` matches its value.
	for _, assoc := range share.DataShareAssociations {
		if aws.StringValue(assoc.Status) == redshift.DataShareStatusActive {
			if associateEntireAccount == "true" && accountIDFromARN(dataShareARN) == aws.StringValue(assoc.ConsumerIdentifier) ||
				consumerARN != "" && consumerARN == aws.StringValue(assoc.ConsumerIdentifier) ||
				consumerRegion != "" && consumerRegion == aws.StringValue(assoc.ConsumerRegion) {
				return share, nil
			}
		}
	}

	return nil, &retry.NotFoundError{
		LastError:   err,
		LastRequest: in,
	}
}

type resourceDataShareConsumerAssociationData struct {
	AllowWrites            types.Bool   `tfsdk:"allow_writes"`
	AssociateEntireAccount types.Bool   `tfsdk:"associate_entire_account"`
	ConsumerARN            fwtypes.ARN  `tfsdk:"consumer_arn"`
	ConsumerRegion         types.String `tfsdk:"consumer_region"`
	DataShareARN           fwtypes.ARN  `tfsdk:"data_share_arn"`
	ID                     types.String `tfsdk:"id"`
	ManagedBy              types.String `tfsdk:"managed_by"`
	ProducerARN            fwtypes.ARN  `tfsdk:"producer_arn"`
}

// accountIDFromARN returns the account ID from the provided ARN string
//
// If the string is not a valid ARN, an empty string is returned, making
// this function safe for use in comparison operators.
func accountIDFromARN(s string) string {
	parsed, err := arn.Parse(s)
	if err != nil {
		return ""
	}
	return parsed.AccountID
}
