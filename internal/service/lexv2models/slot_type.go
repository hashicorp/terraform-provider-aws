// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lexv2models_slot_type", name="Slot Type")
func newResourceSlotType(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSlotType{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameSlotType = "Slot Type"

	slotTypeIDPartCount = 4
)

type resourceSlotType struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceSlotType) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	subSlotTypeCompositionLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[SubSlotTypeComposition](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Required: true,
				},
				"slot_type_id": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bot_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bot_version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"locale_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slot_type_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"parent_slot_type_signature": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"composite_slot_type_setting": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[CompositeSlotTypeSetting](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"sub_slots": subSlotTypeCompositionLNB,
					},
				},
			},
			"external_source_setting": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ExternalSourceSetting](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"grammar_slot_type_setting": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[GrammarSlotTypeSetting](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									names.AttrSource: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[Source](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrS3BucketName: schema.StringAttribute{
													Required: true,
												},
												"s3_object_key": schema.StringAttribute{
													Required: true,
												},
												names.AttrKMSKeyARN: schema.StringAttribute{
													Required: true,
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
			"slot_type_values": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[SlotTypeValues](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"sample_value": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[SampleValue](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrValue: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"synonyms": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrValue: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"value_selection_setting": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[ValueSelectionSetting](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"resolution_strategy": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.SlotValueResolutionStrategy](),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"advanced_recognition_setting": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[AdvancedRecognitionSetting](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"audio_recognition_strategy": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.AudioRecognitionStrategy](),
										},
									},
								},
							},
						},
						"regex_filter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[RegexFilter](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"pattern": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

var slotTypeFlexOpt = flex.WithFieldNamePrefix(ResNameSlotType)

func (r *resourceSlotType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan resourceSlotTypeData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.CreateSlotTypeInput{
		SlotTypeName: plan.Name.ValueStringPointer(),
	}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in, slotTypeFlexOpt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateSlotType(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameSlotType, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameSlotType, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	idParts := []string{
		aws.ToString(out.BotId),
		aws.ToString(out.BotVersion),
		aws.ToString(out.LocaleId),
		aws.ToString(out.SlotTypeId),
	}
	id, err := intflex.FlattenResourceId(idParts, slotTypeIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameSlotType, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan, slotTypeFlexOpt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSlotType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceSlotTypeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindSlotTypeByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionSetting, ResNameSlotType, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, slotTypeFlexOpt)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSlotType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan, state resourceSlotTypeData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if slotTypeHasChanges(ctx, plan, state) {
		input := &lexmodelsv2.UpdateSlotTypeInput{}

		resp.Diagnostics.Append(flex.Expand(ctx, plan, input, slotTypeFlexOpt)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateSlotType(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameSlotType, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameSlotType, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, input, &plan, slotTypeFlexOpt)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceSlotType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceSlotTypeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.DeleteSlotTypeInput{
		BotId:      state.BotID.ValueStringPointer(),
		BotVersion: state.BotVersion.ValueStringPointer(),
		LocaleId:   state.LocaleID.ValueStringPointer(),
		SlotTypeId: state.SlotTypeID.ValueStringPointer(),
	}

	_, err := conn.DeleteSlotType(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if errs.IsAErrorMessageContains[*awstypes.PreconditionFailedException](err, "does not exist") {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameSlotType, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func FindSlotTypeByID(ctx context.Context, conn *lexmodelsv2.Client, id string) (*lexmodelsv2.DescribeSlotTypeOutput, error) {
	parts, err := intflex.ExpandResourceId(id, slotTypeIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &lexmodelsv2.DescribeSlotTypeInput{
		BotId:      aws.String(parts[0]),
		BotVersion: aws.String(parts[1]),
		LocaleId:   aws.String(parts[2]),
		SlotTypeId: aws.String(parts[3]),
	}

	out, err := conn.DescribeSlotType(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceSlotTypeData struct {
	BotID                    types.String                                              `tfsdk:"bot_id"`
	BotVersion               types.String                                              `tfsdk:"bot_version"`
	LocaleID                 types.String                                              `tfsdk:"locale_id"`
	Description              types.String                                              `tfsdk:"description"`
	ID                       types.String                                              `tfsdk:"id"`
	SlotTypeID               types.String                                              `tfsdk:"slot_type_id"`
	Name                     types.String                                              `tfsdk:"name"`
	CompositeSlotTypeSetting fwtypes.ListNestedObjectValueOf[CompositeSlotTypeSetting] `tfsdk:"composite_slot_type_setting"`
	ExternalSourceSetting    fwtypes.ListNestedObjectValueOf[ExternalSourceSetting]    `tfsdk:"external_source_setting"`
	SlotTypeValues           fwtypes.ListNestedObjectValueOf[SlotTypeValues]           `tfsdk:"slot_type_values"`
	ValueSelectionSetting    fwtypes.ListNestedObjectValueOf[ValueSelectionSetting]    `tfsdk:"value_selection_setting"`
	ParentSlotTypeSignature  types.String                                              `tfsdk:"parent_slot_type_signature"`
	Timeouts                 timeouts.Value                                            `tfsdk:"timeouts"`
}

type SubSlotTypeComposition struct {
	Name       types.String `tfsdk:"name"`
	SlotTypeID types.String `tfsdk:"slot_type_id"`
}

type CompositeSlotTypeSetting struct {
	SubSlots fwtypes.ListNestedObjectValueOf[SubSlotTypeComposition] `tfsdk:"sub_slots"`
}

type Source struct {
	S3BucketName types.String `tfsdk:"s3_bucket_name"`
	S3ObjectKey  types.String `tfsdk:"s3_object_key"`
	KmsKeyARN    types.String `tfsdk:"kms_key_arn"`
}

type GrammarSlotTypeSetting struct {
	Source fwtypes.ListNestedObjectValueOf[Source] `tfsdk:"source"`
}

type ExternalSourceSetting struct {
	GrammarSlotTypeSetting fwtypes.ListNestedObjectValueOf[GrammarSlotTypeSetting] `tfsdk:"grammar_slot_type_setting"`
}

type SampleValue struct {
	Value types.String `tfsdk:"value"`
}

type SlotTypeValues struct {
	SampleValue fwtypes.ListNestedObjectValueOf[SampleValue] `tfsdk:"sample_value"`
	Synonyms    fwtypes.ListNestedObjectValueOf[SampleValue] `tfsdk:"synonyms"`
}

type AdvancedRecognitionSetting struct {
	AudioRecognitionStrategy types.String `tfsdk:"audio_recognition_strategy"`
}

type RegexFilter struct {
	Pattern types.String `tfsdk:"pattern"`
}

type ValueSelectionSetting struct {
	ResolutionStrategy         fwtypes.StringEnum[awstypes.SlotValueResolutionStrategy]    `tfsdk:"resolution_strategy"`
	AdvancedRecognitionSetting fwtypes.ListNestedObjectValueOf[AdvancedRecognitionSetting] `tfsdk:"advanced_recognition_setting"`
	RegexFilter                fwtypes.ListNestedObjectValueOf[RegexFilter]                `tfsdk:"regex_filter"`
}

func slotTypeHasChanges(_ context.Context, plan, state resourceSlotTypeData) bool {
	return !plan.Description.Equal(state.Description) ||
		!plan.ValueSelectionSetting.Equal(state.ValueSelectionSetting) ||
		!plan.ExternalSourceSetting.Equal(state.ExternalSourceSetting) ||
		!plan.CompositeSlotTypeSetting.Equal(state.CompositeSlotTypeSetting) ||
		!plan.SlotTypeValues.Equal(state.SlotTypeValues) ||
		!plan.ParentSlotTypeSignature.Equal(state.ParentSlotTypeSignature)
}
