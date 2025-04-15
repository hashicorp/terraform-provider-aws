// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	autoflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	sourceAPIAssociationIDPartCount = 2
)

// @FrameworkResource("aws_appsync_source_api_association", name="Source API Association")
func newSourceAPIAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &sourceAPIAssociationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	resNameSourceAPIAssociation = "Source API Association"
)

type sourceAPIAssociationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *sourceAPIAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAssociationID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"merged_api_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"merged_api_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("merged_api_arn")),
				},
			},
			"source_api_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_api_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("source_api_arn")),
				},
			},
			"source_api_association_config": schema.ListAttribute{ // proto5 Optional+Computed nested block.
				CustomType: fwtypes.NewListNestedObjectTypeOf[sourceAPIAssociationConfigModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[sourceAPIAssociationConfigModel](ctx),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *sourceAPIAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var plan sourceAPIAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &appsync.AssociateSourceGraphqlApiInput{}

	response.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.SourceAPIId.IsNull() && !plan.SourceAPIId.IsUnknown() {
		in.SourceApiIdentifier = flex.StringFromFramework(ctx, plan.SourceAPIId)
	}

	if !plan.SourceAPIArn.IsNull() && !plan.SourceAPIArn.IsUnknown() {
		in.SourceApiIdentifier = flex.StringFromFramework(ctx, plan.SourceAPIArn)
	}

	if !plan.MergedAPIId.IsNull() && !plan.MergedAPIId.IsUnknown() {
		in.MergedApiIdentifier = flex.StringFromFramework(ctx, plan.MergedAPIId)
	}

	if !plan.MergedAPIArn.IsNull() && !plan.MergedAPIArn.IsUnknown() {
		in.MergedApiIdentifier = flex.StringFromFramework(ctx, plan.MergedAPIArn)
	}

	out, err := conn.AssociateSourceGraphqlApi(ctx, in)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionCreating, resNameSourceAPIAssociation, plan.MergedAPIId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.SourceApiAssociation == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionCreating, resNameSourceAPIAssociation, plan.MergedAPIId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.AssociationId = flex.StringToFramework(ctx, out.SourceApiAssociation.AssociationId)
	plan.MergedAPIId = flex.StringToFramework(ctx, out.SourceApiAssociation.MergedApiId)
	id, err := plan.setID()
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionFlatteningResourceId, resNameSourceAPIAssociation, plan.MergedAPIId.String(), err),
			err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitSourceAPIAssociationCreated(ctx, conn, plan.AssociationId.ValueString(), aws.ToString(out.SourceApiAssociation.MergedApiArn), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionWaitingForCreation, resNameSourceAPIAssociation, plan.MergedAPIId.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out.SourceApiAssociation, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *sourceAPIAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state sourceAPIAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := state.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	out, err := findSourceAPIAssociationByTwoPartKey(ctx, conn, state.AssociationId.ValueString(), state.MergedAPIId.ValueString())
	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionSetting, resNameSourceAPIAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *sourceAPIAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var plan, state sourceAPIAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.SourceAPIAssociationConfig.Equal(state.SourceAPIAssociationConfig) {
		in := &appsync.UpdateSourceApiAssociationInput{
			AssociationId:       flex.StringFromFramework(ctx, plan.AssociationId),
			MergedApiIdentifier: flex.StringFromFramework(ctx, plan.MergedAPIArn),
		}

		if !plan.Description.Equal(state.Description) {
			in.Description = flex.StringFromFramework(ctx, plan.Description)
		}

		if !plan.SourceAPIAssociationConfig.Equal(state.SourceAPIAssociationConfig) {
			var elements []sourceAPIAssociationConfigModel
			response.Diagnostics.Append(plan.SourceAPIAssociationConfig.ElementsAs(ctx, &elements, false)...)
			if response.Diagnostics.HasError() {
				return
			}
			if len(elements) == 1 {
				saac := &awstypes.SourceApiAssociationConfig{}
				flex.Expand(ctx, elements[0], saac)
				in.SourceApiAssociationConfig = saac
			}
		}

		out, err := conn.UpdateSourceApiAssociation(ctx, in)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AppSync, create.ErrActionUpdating, resNameSourceAPIAssociation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.SourceApiAssociation == nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AppSync, create.ErrActionUpdating, resNameSourceAPIAssociation, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitSourceAPIAssociationUpdated(ctx, conn, plan.AssociationId.ValueString(), plan.MergedAPIArn.ValueString(), updateTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionWaitingForUpdate, resNameSourceAPIAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *sourceAPIAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state sourceAPIAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := appsync.DisassociateSourceGraphqlApiInput{
		AssociationId:       state.AssociationId.ValueStringPointer(),
		MergedApiIdentifier: state.MergedAPIArn.ValueStringPointer(),
	}

	_, err := conn.DisassociateSourceGraphqlApi(ctx, &in)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionDeleting, resNameSourceAPIAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	_, err = waitSourceAPIAssociationDeleted(ctx, conn, state.AssociationId.ValueString(), state.MergedAPIArn.ValueString(), deleteTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionWaitingForDeletion, resNameSourceAPIAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findSourceAPIAssociationByTwoPartKey(ctx context.Context, conn *appsync.Client, associationID, mergedAPIID string) (*awstypes.SourceApiAssociation, error) {
	input := &appsync.GetSourceApiAssociationInput{
		AssociationId:       aws.String(associationID),
		MergedApiIdentifier: aws.String(mergedAPIID),
	}

	return findSourceAPIAssociation(ctx, conn, input)
}

func findSourceAPIAssociation(ctx context.Context, conn *appsync.Client, input *appsync.GetSourceApiAssociationInput) (*awstypes.SourceApiAssociation, error) {
	output, err := conn.GetSourceApiAssociation(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SourceApiAssociation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SourceApiAssociation, nil
}

func statusSourceAPIAssociation(ctx context.Context, conn *appsync.Client, associationID, mergedAPIID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSourceAPIAssociationByTwoPartKey(ctx, conn, associationID, mergedAPIID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.SourceApiAssociationStatus), nil
	}
}

func waitSourceAPIAssociationCreated(ctx context.Context, conn *appsync.Client, associationID, mergedAPIID string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SourceApiAssociationStatusMergeInProgress, awstypes.SourceApiAssociationStatusMergeScheduled),
		Target:  enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess),
		Refresh: statusSourceAPIAssociation(ctx, conn, associationID, mergedAPIID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.SourceApiAssociationStatusDetail)))

		return output, err
	}

	return nil, err
}

func waitSourceAPIAssociationUpdated(ctx context.Context, conn *appsync.Client, associationID, mergedAPIID string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SourceApiAssociationStatusMergeInProgress, awstypes.SourceApiAssociationStatusMergeScheduled),
		Target:  enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess),
		Refresh: statusSourceAPIAssociation(ctx, conn, associationID, mergedAPIID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.SourceApiAssociationStatusDetail)))

		return output, err
	}

	return nil, err
}

func waitSourceAPIAssociationDeleted(ctx context.Context, conn *appsync.Client, associationID, mergedAPIID string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess, awstypes.SourceApiAssociationStatusDeletionInProgress, awstypes.SourceApiAssociationStatusDeletionScheduled),
		Target:  []string{},
		Refresh: statusSourceAPIAssociation(ctx, conn, associationID, mergedAPIID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.SourceApiAssociationStatusDetail)))

		return output, err
	}

	return nil, err
}

type sourceAPIAssociationResourceModel struct {
	AssociationArn             types.String                                                     `tfsdk:"arn"`
	AssociationId              types.String                                                     `tfsdk:"association_id"`
	ID                         types.String                                                     `tfsdk:"id"`
	Description                types.String                                                     `tfsdk:"description"`
	MergedAPIArn               fwtypes.ARN                                                      `tfsdk:"merged_api_arn"`
	MergedAPIId                types.String                                                     `tfsdk:"merged_api_id"`
	SourceAPIArn               fwtypes.ARN                                                      `tfsdk:"source_api_arn"`
	SourceAPIAssociationConfig fwtypes.ListNestedObjectValueOf[sourceAPIAssociationConfigModel] `tfsdk:"source_api_association_config"`
	SourceAPIId                types.String                                                     `tfsdk:"source_api_id"`
	Timeouts                   timeouts.Value                                                   `tfsdk:"timeouts"`
}

type sourceAPIAssociationConfigModel struct {
	MergeType fwtypes.StringEnum[awstypes.MergeType] `tfsdk:"merge_type"`
}

func (m *sourceAPIAssociationResourceModel) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := autoflex.ExpandResourceId(id, sourceAPIAssociationIDPartCount, false)

	if err != nil {
		return err
	}

	m.MergedAPIId = types.StringValue(parts[0])
	m.AssociationId = types.StringValue(parts[1])
	return nil
}

func (m *sourceAPIAssociationResourceModel) setID() (string, error) {
	parts := []string{
		m.MergedAPIId.ValueString(),
		m.AssociationId.ValueString(),
	}

	return autoflex.FlattenResourceId(parts, sourceAPIAssociationIDPartCount, false)
}
