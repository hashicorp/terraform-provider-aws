// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
func newEventActionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceEventAction{}, nil
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
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrAction: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[actionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"export_revision_to_s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[autoExportRevisionToS3RequestDetailsModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"encryption": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[actionS3Encryption](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
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
									},
									"revision_destination": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[autoExportRevisionDestinationEntryModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
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
							},
						},
					},
				},
			},
			"event": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eventModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"revision_published": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[revisionPublishedModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"data_set_id": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
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
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
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

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
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

	out, err := findEventActionByID(ctx, conn, state.ID.ValueString())
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

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
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
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("EventAction"))...)
	if resp.Diagnostics.HasError() {
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

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
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

func findEventActionByID(ctx context.Context, conn *dataexchange.Client, id string) (*dataexchange.GetEventActionOutput, error) {
	input := dataexchange.GetEventActionInput{
		EventActionId: aws.String(id),
	}

	out, err := conn.GetEventAction(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out, nil
}

type resourceEventActionModel struct {
	Action    fwtypes.ListNestedObjectValueOf[actionModel] `tfsdk:"action"`
	ARN       types.String                                 `tfsdk:"arn"`
	CreatedAt timetypes.RFC3339                            `tfsdk:"created_at"`
	Event     fwtypes.ListNestedObjectValueOf[eventModel]  `tfsdk:"event"`
	ID        types.String                                 `tfsdk:"id"`
	UpdatedAt timetypes.RFC3339                            `tfsdk:"updated_at"`
}

type actionModel struct {
	ExportRevisionToS3 fwtypes.ListNestedObjectValueOf[autoExportRevisionToS3RequestDetailsModel] `tfsdk:"export_revision_to_s3"`
}

type autoExportRevisionToS3RequestDetailsModel struct {
	Encryption          fwtypes.ListNestedObjectValueOf[actionS3Encryption]                      `tfsdk:"encryption"`
	RevisionDestination fwtypes.ListNestedObjectValueOf[autoExportRevisionDestinationEntryModel] `tfsdk:"revision_destination"`
}

type autoExportRevisionDestinationEntryModel struct {
	Bucket     types.String `tfsdk:"bucket"`
	KeyPattern types.String `tfsdk:"key_pattern"`
}

type actionS3Encryption struct {
	KmsKeyArn types.String                                           `tfsdk:"kms_key_arn"`
	Type      fwtypes.StringEnum[awstypes.ServerSideEncryptionTypes] `tfsdk:"type"`
}

type eventModel struct {
	RevisionPublished fwtypes.ListNestedObjectValueOf[revisionPublishedModel] `tfsdk:"revision_published"`
}

type revisionPublishedModel struct {
	DataSetId types.String `tfsdk:"data_set_id"`
}
