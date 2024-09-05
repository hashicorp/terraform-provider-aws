// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	sourceApiAssociationIDPartCount = 2
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_appsync_source_api_association", name="Source Api Association")
func newResourceSourceApiAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSourceApiAssociation{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameSourceApiAssociation = "Source Api Association"
)

type resourceSourceApiAssociation struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceSourceApiAssociation) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_appsync_source_api_association"
}

func (r *resourceSourceApiAssociation) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"association_id": schema.StringAttribute{
				Computed: true,
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
				},
			},
			"merged_api_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				},
			},
			"source_api_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("source_api_arn")),
				},
			},
			"source_api_association_config": schema.ListAttribute{ // proto5 Optional+Computed nested block.
				CustomType: fwtypes.NewListNestedObjectTypeOf[sourceApiAssociationConfigModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[sourceApiAssociationConfigModel](ctx),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceSourceApiAssociation) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var plan resourceSourceApiAssociationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &appsync.AssociateSourceGraphqlApiInput{}

	response.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.SourceApiId.IsNull() && !plan.SourceApiId.IsUnknown() {
		in.SourceApiIdentifier = flex.StringFromFramework(ctx, plan.SourceApiId)
	}

	if !plan.SourceApiArn.IsNull() && !plan.SourceApiArn.IsUnknown() {
		in.SourceApiIdentifier = flex.StringFromFramework(ctx, plan.SourceApiArn)
	}

	if !plan.MergedApiId.IsNull() && !plan.MergedApiId.IsUnknown() {
		in.MergedApiIdentifier = flex.StringFromFramework(ctx, plan.MergedApiId)
	}

	if !plan.MergedApiArn.IsNull() && !plan.MergedApiArn.IsUnknown() {
		in.MergedApiIdentifier = flex.StringFromFramework(ctx, plan.MergedApiArn)
	}

	out, err := conn.AssociateSourceGraphqlApi(ctx, in)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionCreating, ResNameSourceApiAssociation, plan.MergedApiId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.SourceApiAssociation == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionCreating, ResNameSourceApiAssociation, plan.MergedApiId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.AssociationId = flex.StringToFramework(ctx, out.SourceApiAssociation.AssociationId)
	plan.MergedApiId = flex.StringToFramework(ctx, out.SourceApiAssociation.MergedApiId)
	plan.setID()

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitSourceApiAssociationCreated(ctx, conn, plan.AssociationId.ValueString(), aws.ToString(out.SourceApiAssociation.MergedApiArn), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionWaitingForCreation, ResNameSourceApiAssociation, plan.MergedApiId.String(), err),
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

func (r *resourceSourceApiAssociation) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state resourceSourceApiAssociationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := state.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	out, err := findSourceApiAssociationByTwoPartKey(ctx, conn, state.AssociationId.ValueString(), state.MergedApiId.ValueString())
	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionSetting, ResNameSourceApiAssociation, state.ID.String(), err),
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

func (r *resourceSourceApiAssociation) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var plan, state resourceSourceApiAssociationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.SourceApiAssociationConfig.Equal(state.SourceApiAssociationConfig) {

		in := &appsync.UpdateSourceApiAssociationInput{
			AssociationId:       flex.StringFromFramework(ctx, plan.AssociationId),
			MergedApiIdentifier: flex.StringFromFramework(ctx, plan.MergedApiArn),
		}

		if !plan.Description.Equal(state.Description) {
			in.Description = flex.StringFromFramework(ctx, plan.Description)
		}

		if !plan.SourceApiAssociationConfig.Equal(state.SourceApiAssociationConfig) {
			var elements []sourceApiAssociationConfigModel
			response.Diagnostics.Append(plan.SourceApiAssociationConfig.ElementsAs(ctx, &elements, false)...)
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
				create.ProblemStandardMessage(names.AppSync, create.ErrActionUpdating, ResNameSourceApiAssociation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.SourceApiAssociation == nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AppSync, create.ErrActionUpdating, ResNameSourceApiAssociation, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitSourceApiAssociationUpdated(ctx, conn, plan.AssociationId.ValueString(), plan.MergedApiArn.ValueString(), updateTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionWaitingForUpdate, ResNameSourceApiAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceSourceApiAssociation) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state resourceSourceApiAssociationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &appsync.DisassociateSourceGraphqlApiInput{
		AssociationId:       aws.String(state.AssociationId.ValueString()),
		MergedApiIdentifier: aws.String(state.MergedApiArn.ValueString()),
	}

	_, err := conn.DisassociateSourceGraphqlApi(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionDeleting, ResNameSourceApiAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	_, err = waitSourceApiAssociationDeleted(ctx, conn, state.AssociationId.ValueString(), state.MergedApiArn.ValueString(), deleteTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppSync, create.ErrActionWaitingForDeletion, ResNameSourceApiAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceSourceApiAssociation) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

func waitSourceApiAssociationCreated(ctx context.Context, conn *appsync.Client, associationId, mergedApiIdentifier string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.SourceApiAssociationStatusMergeInProgress, awstypes.SourceApiAssociationStatusMergeScheduled),
		Target:                    enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess),
		Refresh:                   statusSourceApiAssociation(ctx, conn, associationId, mergedApiIdentifier),
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

func waitSourceApiAssociationUpdated(ctx context.Context, conn *appsync.Client, associationId, mergedApiIdentifier string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.SourceApiAssociationStatusMergeInProgress, awstypes.SourceApiAssociationStatusMergeScheduled),
		Target:                    enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess),
		Refresh:                   statusSourceApiAssociation(ctx, conn, associationId, mergedApiIdentifier),
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

func waitSourceApiAssociationDeleted(ctx context.Context, conn *appsync.Client, associationId, mergedApiIdentifier string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess, awstypes.SourceApiAssociationStatusDeletionInProgress, awstypes.SourceApiAssociationStatusDeletionScheduled),
		Target:  []string{},
		Refresh: statusSourceApiAssociation(ctx, conn, associationId, mergedApiIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		return out, err
	}

	return nil, err
}

func statusSourceApiAssociation(ctx context.Context, conn *appsync.Client, associationId, mergedApiIdentifier string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findSourceApiAssociationByTwoPartKey(ctx, conn, associationId, mergedApiIdentifier)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out.SourceApiAssociationStatusDetail, string(out.SourceApiAssociationStatus), nil
	}
}

func findSourceApiAssociationByTwoPartKey(ctx context.Context, conn *appsync.Client, associationId, mergedApiIdentifier string) (*awstypes.SourceApiAssociation, error) {
	in := &appsync.GetSourceApiAssociationInput{
		AssociationId:       aws.String(associationId),
		MergedApiIdentifier: aws.String(mergedApiIdentifier),
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

type resourceSourceApiAssociationModel struct {
	AssociationArn             types.String                                                     `tfsdk:"arn"`
	AssociationId              types.String                                                     `tfsdk:"association_id"`
	ID                         types.String                                                     `tfsdk:"id"`
	Description                types.String                                                     `tfsdk:"description"`
	MergedApiArn               fwtypes.ARN                                                      `tfsdk:"merged_api_arn"`
	MergedApiId                types.String                                                     `tfsdk:"merged_api_id"`
	SourceApiArn               fwtypes.ARN                                                      `tfsdk:"source_api_arn"`
	SourceApiAssociationConfig fwtypes.ListNestedObjectValueOf[sourceApiAssociationConfigModel] `tfsdk:"source_api_association_config"`
	SourceApiId                types.String                                                     `tfsdk:"source_api_id"`
	Timeouts                   timeouts.Value                                                   `tfsdk:"timeouts"`
}

type sourceApiAssociationConfigModel struct {
	MergeType fwtypes.StringEnum[awstypes.MergeType] `tfsdk:"merge_type"`
}

func (m *resourceSourceApiAssociationModel) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := autoflex.ExpandResourceId(id, sourceApiAssociationIDPartCount, false)

	if err != nil {
		return err
	}

	m.MergedApiId = types.StringValue(parts[0])
	m.AssociationId = types.StringValue(parts[1])
	return nil
}

func (m *resourceSourceApiAssociationModel) setID() {
	m.ID = types.StringValue(errs.Must(autoflex.FlattenResourceId([]string{m.MergedApiId.ValueString(), m.AssociationId.ValueString()}, sourceApiAssociationIDPartCount, false)))
}
