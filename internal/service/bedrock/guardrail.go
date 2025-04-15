// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
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
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_guardrail", name="Guardrail")
// @Tags(identifierAttribute="guardrail_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrock;bedrock.GetGuardrailOutput")
// @Testing(importStateIdFunc="testAccGuardrailImportStateIDFunc")
// @Testing(importStateIdAttribute="guardrail_id")
func newResourceGuardrail(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGuardrail{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameGuardrail = "Guardrail"
)

type resourceGuardrail struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceGuardrail) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"blocked_input_messaging": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
				},
			},
			"blocked_outputs_messaging": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 200),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"guardrail_arn": framework.ARNAttributeComputedOnly(),
			"guardrail_id":  framework.IDAttribute(),
			names.AttrKMSKeyARN: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
					stringvalidator.RegexMatches(guardrailNameRegex, ""),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GuardrailStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"content_policy_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[contentPolicyConfig](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"filters_config": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[filtersConfig](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"input_strength": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GuardrailFilterStrength](),
									},
									"output_strength": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GuardrailFilterStrength](),
									},
									names.AttrType: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GuardrailContentFilterType](),
									},
								},
							},
						},
					},
				},
			},
			"contextual_grounding_policy_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[contextualGroundingPolicyConfig](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"filters_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[contextualGroundingFiltersConfig](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"threshold": schema.Float64Attribute{
										Required: true,
										Validators: []validator.Float64{
											float64validator.AtLeast(filtersConfigThresholdMin),
										},
									},
									names.AttrType: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GuardrailContextualGroundingFilterType](),
									},
								},
							},
						},
					},
				},
			},
			"sensitive_information_policy_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[sensitiveInformationPolicyConfig](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"pii_entities_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[piiEntitiesConfig](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrAction: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GuardrailSensitiveInformationAction](),
									},
									names.AttrType: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GuardrailPiiEntityType](),
									},
								},
							},
						},
						"regexes_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[regexesConfig](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrAction: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GuardrailSensitiveInformationAction](),
									},
									names.AttrDescription: schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 1000),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									names.AttrName: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 100),
										},
									},
									"pattern": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
									},
								},
							},
						},
					},
				},
			},
			"topic_policy_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[topicPolicyConfig](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"topics_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[topicsConfig](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"definition": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 200),
										},
									},
									"examples": schema.ListAttribute{
										ElementType: types.StringType,
										Optional:    true,
										Computed:    true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(0),
											listvalidator.ValueStringsAre(
												stringvalidator.LengthBetween(1, 100),
											),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
									names.AttrName: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 100),
											stringvalidator.RegexMatches(topicsConfigNameRegex, ""),
										},
									},
									names.AttrType: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GuardrailTopicType](),
									},
								},
							},
						},
					},
				},
			},
			"word_policy_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[wordPolicyConfig](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"managed_word_lists_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[managedWordListsConfig](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrType: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GuardrailManagedWordsType](),
									},
								},
							},
						},
						"words_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[wordsConfig](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"text": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
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

const (
	filtersConfigThresholdMin = 0.000000

	guardrailIDParts = 2
)

var (
	flexOpt = fwflex.WithFieldNameSuffix("Config")

	guardrailNameRegex    = regexache.MustCompile("^[0-9a-zA-Z-_]+$")
	topicsConfigNameRegex = regexache.MustCompile("^[0-9a-zA-Z-_ !?.]+$")
)

func (r *resourceGuardrail) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan resourceGuardrailData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &bedrock.CreateGuardrailInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in, flexOpt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.Tags = getTagsIn(ctx)
	out, err := conn.CreateGuardrail(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameGuardrail, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.GuardrailArn = fwflex.StringToFramework(ctx, out.GuardrailArn)
	plan.GuardrailID = fwflex.StringToFramework(ctx, out.GuardrailId)
	plan.Version = fwflex.StringToFramework(ctx, out.Version)
	plan.CreatedAt = fwflex.TimeToFramework(ctx, out.CreatedAt)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitGuardrailCreated(ctx, conn, plan.GuardrailID.ValueString(), plan.Version.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForCreation, ResNameGuardrail, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	output, err := findGuardrailByTwoPartKey(ctx, conn, plan.GuardrailID.ValueString(), plan.Version.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionSetting, ResNameGuardrail, plan.GuardrailID.String(), err),
			err.Error(),
		)
		return
	}
	plan.Status = fwtypes.StringEnumValue(output.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceGuardrail) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state resourceGuardrailData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := findGuardrailByTwoPartKey(ctx, conn, state.GuardrailID.ValueString(), state.Version.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionSetting, ResNameGuardrail, state.GuardrailID.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state, flexOpt)...)
	state.KmsKeyId = fwflex.StringToFramework(ctx, out.KmsKeyArn)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceGuardrail) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan, state resourceGuardrailData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.BlockedInputMessaging.Equal(state.BlockedInputMessaging) ||
		!plan.BlockedOutputsMessaging.Equal(state.BlockedOutputsMessaging) ||
		!plan.KmsKeyId.Equal(state.KmsKeyId) ||
		!plan.ContentPolicy.Equal(state.ContentPolicy) ||
		!plan.ContextualGroundingPolicy.Equal(state.ContextualGroundingPolicy) ||
		!plan.SensitiveInformationPolicy.Equal(state.SensitiveInformationPolicy) ||
		!plan.TopicPolicy.Equal(state.TopicPolicy) ||
		!plan.WordPolicy.Equal(state.WordPolicy) ||
		!plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) {
		in := &bedrock.UpdateGuardrailInput{
			GuardrailIdentifier: plan.GuardrailID.ValueStringPointer(),
		}
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in, flexOpt)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateGuardrail(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Bedrock, create.ErrActionUpdating, ResNameGuardrail, plan.GuardrailID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.GuardrailArn == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Bedrock, create.ErrActionUpdating, ResNameGuardrail, plan.GuardrailID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		plan.GuardrailArn = fwflex.StringToFramework(ctx, out.GuardrailArn)
		plan.GuardrailID = fwflex.StringToFramework(ctx, out.GuardrailId)

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		if _, err := waitGuardrailUpdated(ctx, conn, plan.GuardrailID.ValueString(), state.Version.ValueString(), updateTimeout); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForUpdate, ResNameGuardrail, plan.GuardrailID.String(), err),
				err.Error(),
			)
			return
		}

		output, err := findGuardrailByTwoPartKey(ctx, conn, plan.GuardrailID.ValueString(), plan.Version.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Bedrock, create.ErrActionSetting, ResNameGuardrail, plan.GuardrailID.String(), err),
				err.Error(),
			)
			return
		}
		plan.Status = fwtypes.StringEnumValue(output.Status)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceGuardrail) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state resourceGuardrailData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &bedrock.DeleteGuardrailInput{
		GuardrailIdentifier: state.GuardrailID.ValueStringPointer(),
	}
	if _, err := conn.DeleteGuardrail(ctx, in); err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionDeleting, ResNameGuardrail, state.GuardrailID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if _, err := waitGuardrailDeleted(ctx, conn, state.GuardrailID.ValueString(), state.Version.ValueString(), deleteTimeout); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForDeletion, ResNameGuardrail, state.GuardrailID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceGuardrail) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, guardrailIDParts, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: guardrail_id,version. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guardrail_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrVersion), parts[1])...)
}

func waitGuardrailCreated(ctx context.Context, conn *bedrock.Client, id string, version string, timeout time.Duration) (*bedrock.GetGuardrailOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GuardrailStatusCreating),
		Target:                    enum.Slice(awstypes.GuardrailStatusReady),
		Refresh:                   statusGuardrail(ctx, conn, id, version),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.GetGuardrailOutput); ok {
		return out, err
	}

	return nil, err
}

func waitGuardrailUpdated(ctx context.Context, conn *bedrock.Client, id string, version string, timeout time.Duration) (*bedrock.GetGuardrailOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GuardrailStatusUpdating),
		Target:                    enum.Slice(awstypes.GuardrailStatusReady),
		Refresh:                   statusGuardrail(ctx, conn, id, version),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.GetGuardrailOutput); ok {
		return out, err
	}

	return nil, err
}

func waitGuardrailDeleted(ctx context.Context, conn *bedrock.Client, id string, version string, timeout time.Duration) (*bedrock.GetGuardrailOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GuardrailStatusDeleting, awstypes.GuardrailStatusReady),
		Target:  []string{},
		Refresh: statusGuardrail(ctx, conn, id, version),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.GetGuardrailOutput); ok {
		return out, err
	}

	return nil, err
}

func statusGuardrail(ctx context.Context, conn *bedrock.Client, id, version string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findGuardrailByTwoPartKey(ctx, conn, id, version)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findGuardrailByTwoPartKey(ctx context.Context, conn *bedrock.Client, id, version string) (*bedrock.GetGuardrailOutput, error) {
	input := &bedrock.GetGuardrailInput{
		GuardrailIdentifier: aws.String(id),
		GuardrailVersion:    aws.String(version),
	}

	output, err := conn.GetGuardrail(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type resourceGuardrailData struct {
	BlockedInputMessaging      types.String                                                      `tfsdk:"blocked_input_messaging"`
	BlockedOutputsMessaging    types.String                                                      `tfsdk:"blocked_outputs_messaging"`
	ContentPolicy              fwtypes.ListNestedObjectValueOf[contentPolicyConfig]              `tfsdk:"content_policy_config"`
	ContextualGroundingPolicy  fwtypes.ListNestedObjectValueOf[contextualGroundingPolicyConfig]  `tfsdk:"contextual_grounding_policy_config"`
	CreatedAt                  timetypes.RFC3339                                                 `tfsdk:"created_at"`
	Description                types.String                                                      `tfsdk:"description"`
	GuardrailArn               types.String                                                      `tfsdk:"guardrail_arn"`
	GuardrailID                types.String                                                      `tfsdk:"guardrail_id"`
	KmsKeyId                   types.String                                                      `tfsdk:"kms_key_arn"`
	Name                       types.String                                                      `tfsdk:"name"`
	SensitiveInformationPolicy fwtypes.ListNestedObjectValueOf[sensitiveInformationPolicyConfig] `tfsdk:"sensitive_information_policy_config"`
	Status                     fwtypes.StringEnum[awstypes.GuardrailStatus]                      `tfsdk:"status"`
	Tags                       tftags.Map                                                        `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                        `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                                    `tfsdk:"timeouts"`
	TopicPolicy                fwtypes.ListNestedObjectValueOf[topicPolicyConfig]                `tfsdk:"topic_policy_config"`
	Version                    types.String                                                      `tfsdk:"version"`
	WordPolicy                 fwtypes.ListNestedObjectValueOf[wordPolicyConfig]                 `tfsdk:"word_policy_config"`
}

type contentPolicyConfig struct {
	Filters fwtypes.SetNestedObjectValueOf[filtersConfig] `tfsdk:"filters_config"`
}

type filtersConfig struct {
	InputStrength  fwtypes.StringEnum[awstypes.GuardrailFilterStrength]    `tfsdk:"input_strength"`
	OutputStrength fwtypes.StringEnum[awstypes.GuardrailFilterStrength]    `tfsdk:"output_strength"`
	Type           fwtypes.StringEnum[awstypes.GuardrailContentFilterType] `tfsdk:"type"`
}

type contextualGroundingPolicyConfig struct {
	Filters fwtypes.ListNestedObjectValueOf[contextualGroundingFiltersConfig] `tfsdk:"filters_config"`
}

type contextualGroundingFiltersConfig struct {
	Threshold types.Float64                                                       `tfsdk:"threshold"`
	Type      fwtypes.StringEnum[awstypes.GuardrailContextualGroundingFilterType] `tfsdk:"type"`
}

type sensitiveInformationPolicyConfig struct {
	PIIEntities fwtypes.ListNestedObjectValueOf[piiEntitiesConfig] `tfsdk:"pii_entities_config"`
	Regexes     fwtypes.ListNestedObjectValueOf[regexesConfig]     `tfsdk:"regexes_config"`
}

type piiEntitiesConfig struct {
	Action fwtypes.StringEnum[awstypes.GuardrailSensitiveInformationAction] `tfsdk:"action"`
	Type   fwtypes.StringEnum[awstypes.GuardrailPiiEntityType]              `tfsdk:"type"`
}

type regexesConfig struct {
	Action      fwtypes.StringEnum[awstypes.GuardrailSensitiveInformationAction] `tfsdk:"action"`
	Description types.String                                                     `tfsdk:"description"`
	Name        types.String                                                     `tfsdk:"name"`
	Pattern     types.String                                                     `tfsdk:"pattern"`
}

type topicPolicyConfig struct {
	Topics fwtypes.ListNestedObjectValueOf[topicsConfig] `tfsdk:"topics_config"`
}

type topicsConfig struct {
	Definition types.String                                    `tfsdk:"definition"`
	Examples   fwtypes.ListValueOf[types.String]               `tfsdk:"examples"`
	Name       types.String                                    `tfsdk:"name"`
	Type       fwtypes.StringEnum[awstypes.GuardrailTopicType] `tfsdk:"type"`
}

type wordPolicyConfig struct {
	ManagedWordLists fwtypes.ListNestedObjectValueOf[managedWordListsConfig] `tfsdk:"managed_word_lists_config"`
	Words            fwtypes.ListNestedObjectValueOf[wordsConfig]            `tfsdk:"words_config"`
}

type managedWordListsConfig struct {
	Type fwtypes.StringEnum[awstypes.GuardrailManagedWordsType] `tfsdk:"type"`
}

type wordsConfig struct {
	Text types.String `tfsdk:"text"`
}
