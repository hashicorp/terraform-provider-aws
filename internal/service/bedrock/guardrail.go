// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Guardrail")
// @Tags(identifierAttribute="arn")
func newResourceGuardrail(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGuardrail{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameGuardrail = "Guardrail"
)

type resourceGuardrail struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceGuardrail) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrock_guardrail"
}

func (r *resourceGuardrail) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"blocked_input_messaging": schema.StringAttribute{
				Description: "Messaging for when violations are detected in text",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
				},
			},
			"blocked_outputs_messaging": schema.StringAttribute{
				Description: "Messaging for when violations are detected in text",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Description: "Time Stamp",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the guardrail or its version",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 200),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": framework.IDAttribute(),
			"kms_key_arn": schema.StringAttribute{
				Description: "The KMS key with which the guardrail was encrypted at rest",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					stringvalidator.RegexMatches(regexp.MustCompile("^arn:aws(-[^:]+)?:kms:[a-zA-Z0-9-]*:[0-9]{12}:key/[a-zA-Z0-9-]{36}$"), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the guardrail",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
					stringvalidator.RegexMatches(regexp.MustCompile("^[0-9a-zA-Z-_]+$"), ""),
				},
			},
			"status": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.GuardrailStatus](),
				Description: "Status of the guardrail",
				Computed:    true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVersion: schema.StringAttribute{
				Description: "Guardrail version",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"content_policy_config": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[contentPolicyConfig](ctx),
				Description: "Word policy config for a guardrail.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"filters_config": schema.ListNestedBlock{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[filtersConfig](ctx),
							Description: "List of content filter configs in content policy.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"input_strength": schema.StringAttribute{
										Description: "Strength for filters",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												"NONE",
												"LOW",
												"MEDIUM",
												"HIGH",
											),
										},
									},
									// Property: OutputStrength
									"output_strength": schema.StringAttribute{
										Description: "Strength for filters",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												"NONE",
												"LOW",
												"MEDIUM",
												"HIGH",
											),
										},
									},
									// Property: Type
									"type": schema.StringAttribute{
										Description: "Type of filter in content policy",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												"SEXUAL",
												"VIOLENCE",
												"HATE",
												"INSULTS",
												"MISCONDUCT",
												"PROMPT_ATTACK",
											),
										},
									},
								},
							},
						},
					},
				},
			},
			"contextual_grounding_policy_config": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[contextualGroundingPolicyConfig](ctx),
				Description: "Contextual Grounding policy config for a guardrail.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"filters_config": schema.ListNestedBlock{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[contextualGroundingFiltersConfig](ctx),
							Description: "List of contextual grounding filter configs.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"threshold": schema.Float64Attribute{
										Description: "The threshold for this filter.",
										Required:    true,
										Validators: []validator.Float64{
											float64validator.AtLeast(0.000000),
										},
									},
									"type": schema.StringAttribute{
										Description: "Type of contextual grounding filter",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												"GROUNDING",
												"RELEVANCE",
											),
										},
									},
								},
							},
						},
					},
				},
			},
			"sensitive_information_policy_config": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[sensitiveInformationPolicyConfig](ctx),
				Description: "Sensitive information policy config for a guardrail.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"pii_entities_config": schema.ListNestedBlock{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[piiEntitiesConfig](ctx),
							Description: "List of entities.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"action": schema.StringAttribute{
										Description: "Options for sensitive information action.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												"BLOCK",
												"ANONYMIZE",
											),
										},
									},
									"type": schema.StringAttribute{
										Description: "The currently supported PII entities",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												"ADDRESS",
												"AGE",
												"AWS_ACCESS_KEY",
												"AWS_SECRET_KEY",
												"CA_HEALTH_NUMBER",
												"CA_SOCIAL_INSURANCE_NUMBER",
												"CREDIT_DEBIT_CARD_CVV",
												"CREDIT_DEBIT_CARD_EXPIRY",
												"CREDIT_DEBIT_CARD_NUMBER",
												"DRIVER_ID",
												"EMAIL",
												"INTERNATIONAL_BANK_ACCOUNT_NUMBER",
												"IP_ADDRESS",
												"LICENSE_PLATE",
												"MAC_ADDRESS",
												"NAME",
												"PASSWORD",
												"PHONE",
												"PIN",
												"SWIFT_CODE",
												"UK_NATIONAL_HEALTH_SERVICE_NUMBER",
												"UK_NATIONAL_INSURANCE_NUMBER",
												"UK_UNIQUE_TAXPAYER_REFERENCE_NUMBER",
												"URL",
												"USERNAME",
												"US_BANK_ACCOUNT_NUMBER",
												"US_BANK_ROUTING_NUMBER",
												"US_INDIVIDUAL_TAX_IDENTIFICATION_NUMBER",
												"US_PASSPORT_NUMBER",
												"US_SOCIAL_SECURITY_NUMBER",
												"VEHICLE_IDENTIFICATION_NUMBER",
											),
										},
									},
								},
							},
						},
						"regexes_config": schema.ListNestedBlock{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[regexesConfig](ctx),
							Description: "List of regex.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"action": schema.StringAttribute{
										Description: "Options for sensitive information action.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												"BLOCK",
												"ANONYMIZE",
											),
										},
									},
									// Property: Description
									"description": schema.StringAttribute{
										Description: "The regex description.",
										Optional:    true,
										Computed:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 1000),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									// Property: Name
									"name": schema.StringAttribute{
										Description: "The regex name.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 100),
										},
									},
									// Property: Pattern
									"pattern": schema.StringAttribute{
										Description: "The regex pattern.",
										Required:    true,
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
				CustomType:  fwtypes.NewListNestedObjectTypeOf[topicPolicyConfig](ctx),
				Description: "Topic policy config for a guardrail.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"topics_config": schema.ListNestedBlock{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[topicsConfig](ctx),
							Description: "List of topic configs in topic policy.",
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"definition": schema.StringAttribute{
										Description: "Definition of topic in topic policy",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 200),
										},
									},
									"examples": schema.ListAttribute{
										ElementType: types.StringType,
										Description: "List of text examples",
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
									"name": schema.StringAttribute{
										Description: "Name of topic in topic policy",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 100),
											stringvalidator.RegexMatches(regexp.MustCompile("^[0-9a-zA-Z-_ !?.]+$"), ""),
										},
									},
									"type": schema.StringAttribute{
										Description: "Type of topic in a policy",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												"DENY",
											),
										},
									},
								},
							},
						},
					},
				},
			},
			"word_policy_config": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[wordPolicyConfig](ctx),
				Description: "Word policy config for a guardrail.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"managed_word_lists_config": schema.ListNestedBlock{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[managedWordListsConfig](ctx),
							Description: "A config for the list of managed words.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Options for managed words.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												"PROFANITY",
											),
										},
									},
								},
							},
						},
						"words_config": schema.ListNestedBlock{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[wordsConfig](ctx),
							Description: "List of custom word configs.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"text": schema.StringAttribute{
										Description: "The custom word text.",
										Required:    true,
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
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

var flexConfig = fwflex.WithFieldNameSuffix("Config")

func (r *resourceGuardrail) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan resourceGuardrailData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &bedrock.CreateGuardrailInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in, flexConfig)...)
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
	plan.ID = fwflex.StringToFramework(ctx, out.GuardrailId)
	plan.Version = fwflex.StringToFramework(ctx, out.Version)
	plan.CreatedAt = fwflex.TimeToFramework(ctx, out.CreatedAt)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitGuardrailCreated(ctx, conn, plan.ID.ValueString(), plan.Version.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForCreation, ResNameGuardrail, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	output, err := findGuardrailByID(ctx, conn, plan.ID.ValueString(), plan.Version.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionSetting, ResNameGuardrail, plan.ID.String(), err),
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
	out, err := findGuardrailByID(ctx, conn, state.ID.ValueString(), state.Version.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionSetting, ResNameGuardrail, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state, flexConfig)...)
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
			GuardrailIdentifier: aws.String(plan.ID.ValueString()),
		}
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in, flexConfig)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateGuardrail(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Bedrock, create.ErrActionUpdating, ResNameGuardrail, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.GuardrailArn == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Bedrock, create.ErrActionUpdating, ResNameGuardrail, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.GuardrailArn = fwflex.StringToFramework(ctx, out.GuardrailArn)
		plan.ID = fwflex.StringToFramework(ctx, out.GuardrailId)
		output, err := findGuardrailByID(ctx, conn, plan.ID.ValueString(), plan.Version.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Bedrock, create.ErrActionSetting, ResNameGuardrail, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		plan.Status = fwtypes.StringEnumValue(output.Status)
	} else {
		plan.Status = state.Status
		plan.GuardrailArn = state.GuardrailArn
		plan.ID = state.ID
		plan.Version = state.Version
		plan.CreatedAt = state.CreatedAt
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitGuardrailUpdated(ctx, conn, plan.ID.ValueString(), state.Version.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForUpdate, ResNameGuardrail, plan.ID.String(), err),
			err.Error(),
		)
		return
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
		GuardrailIdentifier: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteGuardrail(ctx, in)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionDeleting, ResNameGuardrail, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitGuardrailDeleted(ctx, conn, state.ID.ValueString(), state.Version.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForDeletion, ResNameGuardrail, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceGuardrail) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {

	parts := strings.Split(request.ID, ":")
	if len(parts) != 2 {
		response.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "GuardrailId:Version"`, request.ID))
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrVersion), parts[1])...)
}

func (r *resourceGuardrail) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func waitGuardrailCreated(ctx context.Context, conn *bedrock.Client, id string, version string, timeout time.Duration) (*bedrock.GetGuardrailOutput, error) {
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

func waitGuardrailDeleted(ctx context.Context, conn *bedrock.Client, id string, version string, timeout time.Duration) (*bedrock.GetGuardrailOutput, error) {
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

func statusGuardrail(ctx context.Context, conn *bedrock.Client, id string, version string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findGuardrailByID(ctx, conn, id, version)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findGuardrailByID(ctx context.Context, conn *bedrock.Client, id string, version string) (*bedrock.GetGuardrailOutput, error) {
	in := &bedrock.GetGuardrailInput{
		GuardrailIdentifier: aws.String(id),
		GuardrailVersion:    aws.String(version),
	}

	out, err := conn.GetGuardrail(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceGuardrailData struct {
	GuardrailArn               types.String                                                      `tfsdk:"arn"`
	BlockedInputMessaging      types.String                                                      `tfsdk:"blocked_input_messaging"`
	BlockedOutputsMessaging    types.String                                                      `tfsdk:"blocked_outputs_messaging"`
	ContentPolicy              fwtypes.ListNestedObjectValueOf[contentPolicyConfig]              `tfsdk:"content_policy_config"`
	ContextualGroundingPolicy  fwtypes.ListNestedObjectValueOf[contextualGroundingPolicyConfig]  `tfsdk:"contextual_grounding_policy_config"`
	CreatedAt                  timetypes.RFC3339                                                 `tfsdk:"created_at"`
	Description                types.String                                                      `tfsdk:"description"`
	ID                         types.String                                                      `tfsdk:"id"`
	KmsKeyId                   types.String                                                      `tfsdk:"kms_key_arn"`
	Name                       types.String                                                      `tfsdk:"name"`
	SensitiveInformationPolicy fwtypes.ListNestedObjectValueOf[sensitiveInformationPolicyConfig] `tfsdk:"sensitive_information_policy_config"`
	Status                     fwtypes.StringEnum[awstypes.GuardrailStatus]                      `tfsdk:"status"`
	Tags                       types.Map                                                         `tfsdk:"tags"`
	TagsAll                    types.Map                                                         `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                                    `tfsdk:"timeouts"`
	TopicPolicy                fwtypes.ListNestedObjectValueOf[topicPolicyConfig]                `tfsdk:"topic_policy_config"`
	Version                    types.String                                                      `tfsdk:"version"`
	WordPolicy                 fwtypes.ListNestedObjectValueOf[wordPolicyConfig]                 `tfsdk:"word_policy_config"`
}

type contentPolicyConfig struct {
	Filters fwtypes.ListNestedObjectValueOf[filtersConfig] `tfsdk:"filters_config"`
}

type filtersConfig struct {
	InputStrength  types.String `tfsdk:"input_strength"`
	OutputStrength types.String `tfsdk:"output_strength"`
	Type           types.String `tfsdk:"type"`
}

type contextualGroundingPolicyConfig struct {
	Filters fwtypes.ListNestedObjectValueOf[contextualGroundingFiltersConfig] `tfsdk:"filters_config"`
}

type contextualGroundingFiltersConfig struct {
	Threshold types.Float64 `tfsdk:"threshold"`
	Type      types.String  `tfsdk:"type"`
}

type sensitiveInformationPolicyConfig struct {
	PIIEntities fwtypes.ListNestedObjectValueOf[piiEntitiesConfig] `tfsdk:"pii_entities_config"`
	Regexes     fwtypes.ListNestedObjectValueOf[regexesConfig]     `tfsdk:"regexes_config"`
}

type piiEntitiesConfig struct {
	Action types.String `tfsdk:"action"`
	Type   types.String `tfsdk:"type"`
}

type regexesConfig struct {
	Action      types.String `tfsdk:"action"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
	Pattern     types.String `tfsdk:"pattern"`
}

type topicPolicyConfig struct {
	Topics fwtypes.ListNestedObjectValueOf[topicsConfig] `tfsdk:"topics_config"`
}

type topicsConfig struct {
	Definition types.String                      `tfsdk:"definition"`
	Examples   fwtypes.ListValueOf[types.String] `tfsdk:"examples"`
	Name       types.String                      `tfsdk:"name"`
	Type       types.String                      `tfsdk:"type"`
}

type wordPolicyConfig struct {
	ManagedWordLists fwtypes.ListNestedObjectValueOf[managedWordListsConfig] `tfsdk:"managed_word_lists_config"`
	Words            fwtypes.ListNestedObjectValueOf[wordsConfig]            `tfsdk:"words_config"`
}

type managedWordListsConfig struct {
	Type types.String `tfsdk:"type"`
}

type wordsConfig struct {
	Text types.String `tfsdk:"text"`
}
