// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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

// @FrameworkResource(name="Data Share Authorization")
func newResourceDataShareAuthorization(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceDataShareAuthorization{}, nil
}

const (
	ResNameDataShareAuthorization = "Data Share Authorization"

	dataShareAuthorizationIDPartCount = 2
)

type resourceDataShareAuthorization struct {
	framework.ResourceWithConfigure
}

func (r *resourceDataShareAuthorization) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_redshift_data_share_authorization"
}

func (r *resourceDataShareAuthorization) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allow_writes": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"consumer_identifier": schema.StringAttribute{
				Required: true,
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

func (r *resourceDataShareAuthorization) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RedshiftConn(ctx)

	var plan resourceDataShareAuthorizationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataShareARN := plan.DataShareARN.ValueString()
	consumerIdentifier := plan.ConsumerIdentifier.ValueString()
	parts := []string{
		dataShareARN,
		consumerIdentifier,
	}

	id, err := intflex.FlattenResourceId(parts, dataShareAuthorizationIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionFlatteningResourceId, ResNameDataShareAuthorization, dataShareARN, err),
			err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(id)

	in := &redshift.AuthorizeDataShareInput{
		DataShareArn:       aws.String(dataShareARN),
		ConsumerIdentifier: aws.String(consumerIdentifier),
	}

	if !plan.AllowWrites.IsNull() {
		in.AllowWrites = aws.Bool(plan.AllowWrites.ValueBool())
	}

	out, err := conn.AuthorizeDataShareWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionCreating, ResNameDataShareAuthorization, id, err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionCreating, ResNameDataShareAuthorization, id, nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ManagedBy = flex.StringToFramework(ctx, out.ManagedBy)
	plan.ProducerARN = flex.StringToFrameworkARN(ctx, out.ProducerArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDataShareAuthorization) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RedshiftConn(ctx)

	var state resourceDataShareAuthorizationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts, err := intflex.ExpandResourceId(state.ID.ValueString(), dataShareAuthorizationIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionExpandingResourceId, ResNameDataShareAuthorization, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	// split ID and write constituent parts to state to support import
	state.DataShareARN = fwtypes.ARNValue(parts[0])
	state.ConsumerIdentifier = types.StringValue(parts[1])

	out, err := findDataShareAuthorizationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionSetting, ResNameDataShareAuthorization, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ManagedBy = flex.StringToFramework(ctx, out.ManagedBy)
	state.ProducerARN = flex.StringToFrameworkARN(ctx, out.ProducerArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDataShareAuthorization) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is a no-op
}

func (r *resourceDataShareAuthorization) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RedshiftConn(ctx)

	var state resourceDataShareAuthorizationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &redshift.DeauthorizeDataShareInput{
		DataShareArn:       aws.String(state.DataShareARN.ValueString()),
		ConsumerIdentifier: aws.String(state.ConsumerIdentifier.ValueString()),
	}

	_, err := conn.DeauthorizeDataShareWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeResourceNotFoundFault) ||
			tfawserr.ErrMessageContains(err, redshift.ErrCodeInvalidDataShareFault, "because the ARN doesn't exist.") ||
			tfawserr.ErrMessageContains(err, redshift.ErrCodeInvalidDataShareFault, "because you have already removed authorization from the data share.") {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionDeleting, ResNameDataShareAuthorization, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDataShareAuthorization) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findDataShareAuthorizationByID(ctx context.Context, conn *redshift.Redshift, id string) (*redshift.DataShare, error) {
	parts, err := intflex.ExpandResourceId(id, dataShareAuthorizationIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &redshift.DescribeDataSharesInput{
		DataShareArn: aws.String(parts[0]),
	}

	out, err := conn.DescribeDataSharesWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeResourceNotFoundFault) ||
		tfawserr.ErrMessageContains(err, redshift.ErrCodeInvalidDataShareFault, "because the ARN doesn't exist.") {
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

	// Verify a share with the expected consumer identifier is present and the
	// status is one of "AUTHORIZED" or "ACTIVE".
	share := out.DataShares[0]
	for _, assoc := range share.DataShareAssociations {
		if aws.StringValue(assoc.ConsumerIdentifier) == parts[1] {
			switch aws.StringValue(assoc.Status) {
			case redshift.DataShareStatusAuthorized, redshift.DataShareStatusActive:
				return share, nil
			}
		}
	}

	return nil, &retry.NotFoundError{
		LastError:   err,
		LastRequest: in,
	}
}

type resourceDataShareAuthorizationData struct {
	AllowWrites        types.Bool   `tfsdk:"allow_writes"`
	ConsumerIdentifier types.String `tfsdk:"consumer_identifier"`
	DataShareARN       fwtypes.ARN  `tfsdk:"data_share_arn"`
	ID                 types.String `tfsdk:"id"`
	ManagedBy          types.String `tfsdk:"managed_by"`
	ProducerARN        fwtypes.ARN  `tfsdk:"producer_arn"`
}
