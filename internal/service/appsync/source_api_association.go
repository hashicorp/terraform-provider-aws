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

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_appsync_source_api_association", name="Source API Association")
func newResourceSourceAPIAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSourceAPIAssociation{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameSourceAPIAssociation = "Source API Association"
)

type resourceSourceAPIAssociation struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceSourceAPIAssociation) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_appsync_source_api_association"
}

func (r *resourceSourceAPIAssociation) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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

func (r *resourceSourceAPIAssociation) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var plan resourceSourceAPIAssociationModel
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
			create.ProblemStandardMessage(names.AppSync, create.ErrActionCreating, ResNameSourceAPIAssociation, plan.MergedAPIId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.SourceApiAssociation == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionCreating, ResNameSourceAPIAssociation, plan.MergedAPIId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.AssociationId = flex.StringToFramework(ctx, out.SourceApiAssociation.AssociationId)
	plan.MergedAPIId = flex.StringToFramework(ctx, out.SourceApiAssociation.MergedApiId)
	plan.setID()

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitSourceAPIAssociationCreated(ctx, conn, plan.AssociationId.ValueString(), aws.ToString(out.SourceApiAssociation.MergedApiArn), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionWaitingForCreation, ResNameSourceAPIAssociation, plan.MergedAPIId.String(), err),
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

func (r *resourceSourceAPIAssociation) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state resourceSourceAPIAssociationModel
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
			create.ProblemStandardMessage(names.AppSync, create.ErrActionSetting, ResNameSourceAPIAssociation, state.ID.String(), err),
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

func (r *resourceSourceAPIAssociation) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var plan, state resourceSourceAPIAssociationModel
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
				create.ProblemStandardMessage(names.AppSync, create.ErrActionUpdating, ResNameSourceAPIAssociation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.SourceApiAssociation == nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AppSync, create.ErrActionUpdating, ResNameSourceAPIAssociation, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitSourceAPIAssociationUpdated(ctx, conn, plan.AssociationId.ValueString(), plan.MergedAPIArn.ValueString(), updateTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionWaitingForUpdate, ResNameSourceAPIAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceSourceAPIAssociation) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state resourceSourceAPIAssociationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &appsync.DisassociateSourceGraphqlApiInput{
		AssociationId:       aws.String(state.AssociationId.ValueString()),
		MergedApiIdentifier: aws.String(state.MergedAPIArn.ValueString()),
	}

	_, err := conn.DisassociateSourceGraphqlApi(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionDeleting, ResNameSourceAPIAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	_, err = waitSourceAPIAssociationDeleted(ctx, conn, state.AssociationId.ValueString(), state.MergedAPIArn.ValueString(), deleteTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionWaitingForDeletion, ResNameSourceAPIAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceSourceAPIAssociation) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

func waitSourceAPIAssociationCreated(ctx context.Context, conn *appsync.Client, associationId, mergedAPIIdentifier string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.SourceApiAssociationStatusMergeInProgress, awstypes.SourceApiAssociationStatusMergeScheduled),
		Target:                    enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess),
		Refresh:                   statusSourceAPIAssociation(ctx, conn, associationId, mergedAPIIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		return out, err
	}

	return nil, err
}

func waitSourceAPIAssociationUpdated(ctx context.Context, conn *appsync.Client, associationId, mergedAPIIdentifier string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.SourceApiAssociationStatusMergeInProgress, awstypes.SourceApiAssociationStatusMergeScheduled),
		Target:                    enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess),
		Refresh:                   statusSourceAPIAssociation(ctx, conn, associationId, mergedAPIIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		return out, err
	}

	return nil, err
}

func waitSourceAPIAssociationDeleted(ctx context.Context, conn *appsync.Client, associationId, mergedAPIIdentifier string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess, awstypes.SourceApiAssociationStatusDeletionInProgress, awstypes.SourceApiAssociationStatusDeletionScheduled),
		Target:  []string{},
		Refresh: statusSourceAPIAssociation(ctx, conn, associationId, mergedAPIIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		return out, err
	}

	return nil, err
}

func statusSourceAPIAssociation(ctx context.Context, conn *appsync.Client, associationId, mergedAPIIdentifier string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findSourceAPIAssociationByTwoPartKey(ctx, conn, associationId, mergedAPIIdentifier)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out.SourceApiAssociationStatusDetail, string(out.SourceApiAssociationStatus), nil
	}
}

func findSourceAPIAssociationByTwoPartKey(ctx context.Context, conn *appsync.Client, associationId, mergedAPIIdentifier string) (*awstypes.SourceApiAssociation, error) {
	in := &appsync.GetSourceApiAssociationInput{
		AssociationId:       aws.String(associationId),
		MergedApiIdentifier: aws.String(mergedAPIIdentifier),
	}

	out, err := conn.GetSourceApiAssociation(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.SourceApiAssociation == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.SourceApiAssociation, nil
}

type resourceSourceAPIAssociationModel struct {
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

func (m *resourceSourceAPIAssociationModel) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := autoflex.ExpandResourceId(id, sourceAPIAssociationIDPartCount, false)

	if err != nil {
		return err
	}

	m.MergedAPIId = types.StringValue(parts[0])
	m.AssociationId = types.StringValue(parts[1])
	return nil
}

func (m *resourceSourceAPIAssociationModel) setID() {
	m.ID = types.StringValue(errs.Must(autoflex.FlattenResourceId([]string{m.MergedAPIId.ValueString(), m.AssociationId.ValueString()}, sourceAPIAssociationIDPartCount, false)))
}
