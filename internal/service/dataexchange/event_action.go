// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dataexchange_event_action", name="Event Action")
func ResourceEventAction(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEventAction{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameEventAction = "Event Action"
)

type resourceEventAction struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceEventAction) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"action_export_revision_to_s3": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[actionExportRevisionToS3Model](ctx),
				Blocks: map[string]schema.Block{
					"encryption": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[actionS3Encryption](ctx),
						Attributes: map[string]schema.Attribute{
							names.AttrKMSKeyARN: schema.StringAttribute{
								CustomType: fwtypes.ARNType,
								Optional:   true,
								Validators: []validator.String{
									validators.ARN(),
								},
							},
							names.AttrType: schema.StringAttribute{
								Optional:   true,
								CustomType: fwtypes.StringEnumType[awstypes.ServerSideEncryptionTypes](),
							},
						},
					},
					"revision_destination": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[actionRevisionDestination](ctx),
						Attributes: map[string]schema.Attribute{
							names.AttrBucket: schema.StringAttribute{
								Required: true,
							},
							"key_pattern": schema.StringAttribute{
								Optional: true,
								Computed: true,
								Default:  stringdefault.StaticString("${Revision.CreatedAt}/${Asset.Name}"),
							},
						},
					},
				},
			},
			"event_revision_published": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[eventRevisionPublishedModel](ctx),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"data_set_id": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
	}
}

func (r *resourceEventAction) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var plan resourceEventActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := dataexchange.CreateEventActionInput{}
	diags := plan.Expand(ctx, &input)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	out, err := conn.CreateEventAction(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameEventAction, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameEventAction, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	data.ARN = types.StringPointerValue(out.Arn)
	data.ID = types.StringPointerValue(out.Id)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEventAction) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state resourceEventActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindEventActionByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameEventAction, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.Flatten(ctx, out)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEventAction) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var plan, state resourceEventActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := dataexchange.UpdateEventActionInput{}
	diags := plan.Expand(ctx, &input)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	out, err := conn.UpdateEventAction(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionUpdating, ResNameEventAction, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionUpdating, ResNameEventAction, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEventAction) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state resourceEventActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := dataexchange.DeleteEventActionInput{
		EventActionId: state.ID.ValueStringPointer(),
	}
	_, err := conn.DeleteEventAction(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionDeleting, ResNameEventAction, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEventAction) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindEventActionByID(ctx context.Context, conn *dataexchange.Client, id string) (*dataexchange.GetEventActionOutput, error) {
	in := &dataexchange.GetEventActionInput{
		EventActionId: aws.String(id),
	}

	out, err := conn.GetEventAction(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceEventActionModel struct {
	ARN                      types.String                                         `tfsdk:"arn"`
	ID                       types.String                                         `tfsdk:"id"`
	ActionExportRevisionToS3 fwtypes.ObjectValueOf[actionExportRevisionToS3Model] `tfsdk:"action_export_revision_to_s3"`
	EventRevisionPublished   fwtypes.ObjectValueOf[eventRevisionPublishedModel]   `tfsdk:"event_revision_published"`
}

type actionExportRevisionToS3Model struct {
	RevisionDestination fwtypes.ObjectValueOf[actionRevisionDestination] `tfsdk:"revision_destination"`
	Encryption          fwtypes.ObjectValueOf[actionS3Encryption]        `tfsdk:"encryption"`
}

type actionRevisionDestination struct {
	Bucket     types.String `tfsdk:"bucket"`
	KeyPattern types.String `tfsdk:"key_pattern"`
}

type actionS3Encryption struct {
	Type      fwtypes.StringEnum[awstypes.ServerSideEncryptionTypes] `tfsdk:"type"`
	KmsKeyArn types.String                                           `tfsdk:"kms_key_arn"`
}

type eventRevisionPublishedModel struct {
	DataSetId types.String `tfsdk:"data_set_id"`
}

func (m resourceEventActionModel) Expand(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Use switch variable assignment to eliminate type assertions in cases
	switch apiInput := v.(type) {
	case *dataexchange.CreateEventActionInput:
		var eventApiInput awstypes.RevisionPublished
		eventModel, _ := m.EventRevisionPublished.ToPtr(ctx)
		diags.Append(flex.Expand(ctx, eventModel, &eventApiInput)...)
		apiInput.Event = &awstypes.Event{RevisionPublished: &eventApiInput}

		actionModel, _ := m.ActionExportRevisionToS3.ToPtr(ctx)
		var actionApiInput awstypes.AutoExportRevisionToS3RequestDetails
		diags.Append(flex.Expand(ctx, actionModel, &actionApiInput)...)
		apiInput.Action = &awstypes.Action{ExportRevisionToS3: &actionApiInput}

	case *dataexchange.UpdateEventActionInput:
		apiInput.EventActionId = m.ID.ValueStringPointer()
		actionModel, _ := m.ActionExportRevisionToS3.ToPtr(ctx)
		var actionApiInput awstypes.AutoExportRevisionToS3RequestDetails
		diags.Append(flex.Expand(ctx, actionModel, &actionApiInput)...)
		apiInput.Action = &awstypes.Action{ExportRevisionToS3: &actionApiInput}
	}

	return diags
}

func (m *resourceEventActionModel) Flatten(ctx context.Context, v *dataexchange.GetEventActionOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(flex.Flatten(ctx, v, m)...)

	if v.Action != nil {
		var actionModel actionExportRevisionToS3Model
		diags.Append(flex.Flatten(ctx, v.Action.ExportRevisionToS3, &actionModel)...)
		m.ActionExportRevisionToS3, _ = fwtypes.NewObjectValueOf(ctx, &actionModel)
	}

	if v.Event != nil {
		var eventModel eventRevisionPublishedModel
		diags.Append(flex.Flatten(ctx, v.Event.RevisionPublished, &eventModel)...)
		m.EventRevisionPublished, _ = fwtypes.NewObjectValueOf(ctx, &eventModel)
	}

	return diags
}
