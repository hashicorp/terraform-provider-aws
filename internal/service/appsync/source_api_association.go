// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package appsync

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	autoflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
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

type sourceAPIAssociationResource struct {
	framework.ResourceWithModel[sourceAPIAssociationResourceModel]
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
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	in := &appsync.AssociateSourceGraphqlApiInput{}

	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Expand(ctx, plan, in))
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.SourceAPIID.IsNull() && !plan.SourceAPIID.IsUnknown() {
		in.SourceApiIdentifier = flex.StringFromFramework(ctx, plan.SourceAPIID)
	}

	if !plan.SourceAPIARN.IsNull() && !plan.SourceAPIARN.IsUnknown() {
		in.SourceApiIdentifier = flex.StringFromFramework(ctx, plan.SourceAPIARN)
	}

	if !plan.MergedAPIID.IsNull() && !plan.MergedAPIID.IsUnknown() {
		in.MergedApiIdentifier = flex.StringFromFramework(ctx, plan.MergedAPIID)
	}

	if !plan.MergedAPIARN.IsNull() && !plan.MergedAPIARN.IsUnknown() {
		in.MergedApiIdentifier = flex.StringFromFramework(ctx, plan.MergedAPIARN)
	}

	out, err := conn.AssociateSourceGraphqlApi(ctx, in)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}
	if out == nil || out.SourceApiAssociation == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"))
		return
	}

	plan.AssociationID = flex.StringToFramework(ctx, out.SourceApiAssociation.AssociationId)
	plan.MergedAPIID = flex.StringToFramework(ctx, out.SourceApiAssociation.MergedApiId)
	id, err := plan.setID()
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}
	plan.ID = types.StringValue(id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitSourceAPIAssociationCreated(ctx, conn, plan.AssociationID.ValueString(), aws.ToString(out.SourceApiAssociation.MergedApiArn), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Flatten(ctx, out.SourceApiAssociation, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *sourceAPIAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state sourceAPIAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	if err := state.InitFromID(); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)

		return
	}

	out, err := findSourceAPIAssociationByTwoPartKey(ctx, conn, state.AssociationID.ValueString(), state.MergedAPIID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Flatten(ctx, out, &state))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *sourceAPIAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var plan, state sourceAPIAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.SourceAPIAssociationConfig.Equal(state.SourceAPIAssociationConfig) {
		in := &appsync.UpdateSourceApiAssociationInput{
			AssociationId:       flex.StringFromFramework(ctx, plan.AssociationID),
			MergedApiIdentifier: flex.StringFromFramework(ctx, plan.MergedAPIARN),
		}

		if !plan.Description.Equal(state.Description) {
			in.Description = flex.StringFromFramework(ctx, plan.Description)
		}

		if !plan.SourceAPIAssociationConfig.Equal(state.SourceAPIAssociationConfig) {
			var elements []sourceAPIAssociationConfigModel
			smerr.AddEnrich(ctx, &response.Diagnostics, plan.SourceAPIAssociationConfig.ElementsAs(ctx, &elements, false))
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
			smerr.AddError(ctx, &response.Diagnostics, err)
			return
		}
		if out == nil || out.SourceApiAssociation == nil {
			smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"))
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitSourceAPIAssociationUpdated(ctx, conn, plan.AssociationID.ValueString(), plan.MergedAPIARN.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *sourceAPIAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state sourceAPIAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	in := appsync.DisassociateSourceGraphqlApiInput{
		AssociationId:       state.AssociationID.ValueStringPointer(),
		MergedApiIdentifier: state.MergedAPIARN.ValueStringPointer(),
	}

	_, err := conn.DisassociateSourceGraphqlApi(ctx, &in)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	_, err = waitSourceAPIAssociationDeleted(ctx, conn, state.AssociationID.ValueString(), state.MergedAPIARN.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
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
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.SourceApiAssociation == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.SourceApiAssociation, nil
}

func statusSourceAPIAssociation(conn *appsync.Client, associationID, mergedAPIID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSourceAPIAssociationByTwoPartKey(ctx, conn, associationID, mergedAPIID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.SourceApiAssociationStatus), nil
	}
}

func waitSourceAPIAssociationCreated(ctx context.Context, conn *appsync.Client, associationID, mergedAPIID string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SourceApiAssociationStatusMergeInProgress, awstypes.SourceApiAssociationStatusMergeScheduled),
		Target:  enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess),
		Refresh: statusSourceAPIAssociation(conn, associationID, mergedAPIID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.SourceApiAssociationStatusDetail)))

		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitSourceAPIAssociationUpdated(ctx context.Context, conn *appsync.Client, associationID, mergedAPIID string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SourceApiAssociationStatusMergeInProgress, awstypes.SourceApiAssociationStatusMergeScheduled),
		Target:  enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess),
		Refresh: statusSourceAPIAssociation(conn, associationID, mergedAPIID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.SourceApiAssociationStatusDetail)))

		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitSourceAPIAssociationDeleted(ctx context.Context, conn *appsync.Client, associationID, mergedAPIID string, timeout time.Duration) (*awstypes.SourceApiAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SourceApiAssociationStatusMergeSuccess, awstypes.SourceApiAssociationStatusDeletionInProgress, awstypes.SourceApiAssociationStatusDeletionScheduled),
		Target:  []string{},
		Refresh: statusSourceAPIAssociation(conn, associationID, mergedAPIID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SourceApiAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.SourceApiAssociationStatusDetail)))

		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type sourceAPIAssociationResourceModel struct {
	framework.WithRegionModel
	AssociationARN             types.String                                                     `tfsdk:"arn"`
	AssociationID              types.String                                                     `tfsdk:"association_id"`
	ID                         types.String                                                     `tfsdk:"id"`
	Description                types.String                                                     `tfsdk:"description"`
	MergedAPIARN               fwtypes.ARN                                                      `tfsdk:"merged_api_arn"`
	MergedAPIID                types.String                                                     `tfsdk:"merged_api_id"`
	SourceAPIARN               fwtypes.ARN                                                      `tfsdk:"source_api_arn"`
	SourceAPIAssociationConfig fwtypes.ListNestedObjectValueOf[sourceAPIAssociationConfigModel] `tfsdk:"source_api_association_config"`
	SourceAPIID                types.String                                                     `tfsdk:"source_api_id"`
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

	m.MergedAPIID = types.StringValue(parts[0])
	m.AssociationID = types.StringValue(parts[1])
	return nil
}

func (m *sourceAPIAssociationResourceModel) setID() (string, error) {
	parts := []string{
		m.MergedAPIID.ValueString(),
		m.AssociationID.ValueString(),
	}

	return autoflex.FlattenResourceId(parts, sourceAPIAssociationIDPartCount, false)
}
