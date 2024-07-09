// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Guardrail")
func newResourceGuardrail(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGuardrail{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
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
			"created_at": schema.StringAttribute{
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
					stringvalidator.LengthBetween(1, 200),
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
				Description: "Status of the guardrail",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"version": schema.StringAttribute{
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
									"input_strength": schema.StringAttribute{ /*START ATTRIBUTE*/
										Description: "Strength for filters",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.OneOf(
												"NONE",
												"LOW",
												"MEDIUM",
												"HIGH",
											),
										}, /*END VALIDATORS*/
									}, /*END ATTRIBUTE*/
									// Property: OutputStrength
									"output_strength": schema.StringAttribute{ /*START ATTRIBUTE*/
										Description: "Strength for filters",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.OneOf(
												"NONE",
												"LOW",
												"MEDIUM",
												"HIGH",
											),
										}, /*END VALIDATORS*/
									}, /*END ATTRIBUTE*/
									// Property: Type
									"type": schema.StringAttribute{ /*START ATTRIBUTE*/
										Description: "Type of filter in content policy",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.OneOf(
												"SEXUAL",
												"VIOLENCE",
												"HATE",
												"INSULTS",
												"MISCONDUCT",
												"PROMPT_ATTACK",
											),
										}, /*END VALIDATORS*/
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
									"action": schema.StringAttribute{ /*START ATTRIBUTE*/
										Description: "Options for sensitive information action.",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.OneOf(
												"BLOCK",
												"ANONYMIZE",
											),
										}, /*END VALIDATORS*/
									}, /*END ATTRIBUTE*/
									// Property: Type
									"type": schema.StringAttribute{ /*START ATTRIBUTE*/
										Description: "The currently supported PII entities",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
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
										}, /*END VALIDATORS*/
									},
								},
							},
						},
						"regexes_config": schema.ListNestedBlock{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[regexesConfig](ctx),
							Description: "List of regex.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"action": schema.StringAttribute{ /*START ATTRIBUTE*/
										Description: "Options for sensitive information action.",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.OneOf(
												"BLOCK",
												"ANONYMIZE",
											),
										}, /*END VALIDATORS*/
									}, /*END ATTRIBUTE*/
									// Property: Description
									"description": schema.StringAttribute{ /*START ATTRIBUTE*/
										Description: "The regex description.",
										Optional:    true,
										Computed:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.LengthBetween(1, 1000),
										}, /*END VALIDATORS*/
										PlanModifiers: []planmodifier.String{ /*START PLAN MODIFIERS*/
											stringplanmodifier.UseStateForUnknown(),
										}, /*END PLAN MODIFIERS*/
									}, /*END ATTRIBUTE*/
									// Property: Name
									"name": schema.StringAttribute{ /*START ATTRIBUTE*/
										Description: "The regex name.",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.LengthBetween(1, 100),
										}, /*END VALIDATORS*/
									}, /*END ATTRIBUTE*/
									// Property: Pattern
									"pattern": schema.StringAttribute{ /*START ATTRIBUTE*/
										Description: "The regex pattern.",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.LengthAtLeast(1),
										}, /*END VALIDATORS*/
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
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.LengthBetween(1, 200),
										}, /*END VALIDATORS*/
									},
									"examples": schema.ListAttribute{
										ElementType: types.StringType,
										Description: "List of text examples",
										Optional:    true,
										Computed:    true,
										Validators: []validator.List{ /*START VALIDATORS*/
											listvalidator.SizeAtLeast(0),
											listvalidator.ValueStringsAre(
												stringvalidator.LengthBetween(1, 100),
											),
										}, /*END VALIDATORS*/
										PlanModifiers: []planmodifier.List{ /*START PLAN MODIFIERS*/
											listplanmodifier.UseStateForUnknown(),
										}, /*END PLAN MODIFIERS*/
									}, /*END ATTRIBUTE*/
									// Property: Name
									"name": schema.StringAttribute{
										Description: "Name of topic in topic policy",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.LengthBetween(1, 100),
											stringvalidator.RegexMatches(regexp.MustCompile("^[0-9a-zA-Z-_ !?.]+$"), ""),
										}, /*END VALIDATORS*/
									}, /*END ATTRIBUTE*/
									// Property: Type
									"type": schema.StringAttribute{
										Description: "Type of topic in a policy",
										Required:    true,
										Validators: []validator.String{ /*START VALIDATORS*/
											stringvalidator.OneOf(
												"DENY",
											),
										}, /*END VALIDATORS*/
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

func (r *resourceGuardrail) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan resourceGuardrailData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a create input structure
	in := &bedrock.CreateGuardrailInput{
		// TIP: Mandatory or fields that will always be present can be set when
		// you create the Input structure. (Replace these with real fields.)
		Name:                    aws.String(plan.Name.ValueString()),
		BlockedInputMessaging:   aws.String(plan.BlockedInputMessaging.ValueString()),
		BlockedOutputsMessaging: aws.String(plan.BlockedOutputsMessaging.ValueString()),
		Tags:                    getTagsIn(ctx),
	}
	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}
	if !plan.KMSKeyArn.IsNull() {
		in.KmsKeyId = aws.String(plan.KMSKeyArn.ValueString())
	}

	if !plan.ContentPolicyConfig.IsNull() {
		var contentPolicyConfig []contentPolicyConfig
		resp.Diagnostics.Append(plan.ContentPolicyConfig.ElementsAs(ctx, &contentPolicyConfig, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		policyConfigInput, d := expandContentPolicyConfig(ctx, contentPolicyConfig)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.ContentPolicyConfig = policyConfigInput
	}

	if !plan.SensitiveInformationPolicyConfig.IsNull() {
		var sensitiveInformationPolicyConfig []sensitiveInformationPolicyConfig
		resp.Diagnostics.Append(plan.SensitiveInformationPolicyConfig.ElementsAs(ctx, &sensitiveInformationPolicyConfig, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		sensitiveInformationPolicyConfigInput, d := expandSensitiveInformationPolicyConfig(ctx, sensitiveInformationPolicyConfig)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.SensitiveInformationPolicyConfig = sensitiveInformationPolicyConfigInput
	}
	if !plan.TopicPolicyConfig.IsNull() {
		var topicPolicyConfigData []topicPolicyConfig
		resp.Diagnostics.Append(plan.TopicPolicyConfig.ElementsAs(ctx, &topicPolicyConfigData, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		topicPolicyConfigInput, d := expandTopicPolicyConfig(ctx, topicPolicyConfigData)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.TopicPolicyConfig = topicPolicyConfigInput
	}
	if !plan.WordPolicyConfig.IsNull() {
		var wordPolicyConfigData []wordPolicyConfig
		resp.Diagnostics.Append(plan.WordPolicyConfig.ElementsAs(ctx, &wordPolicyConfigData, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		wordPolicyConfigInput, d := expandWordPolicyConfig(ctx, wordPolicyConfigData)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.WordPolicyConfig = wordPolicyConfigInput
	}

	// TIP: -- 4. Call the AWS create function
	out, err := conn.CreateGuardrail(ctx, in)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameGuardrail, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Using the output from the create function, set the minimum attributes
	plan.ARN = fwflex.StringToFramework(ctx, out.GuardrailArn)
	plan.ID = fwflex.StringToFramework(ctx, out.GuardrailId)
	plan.Version = fwflex.StringToFramework(ctx, out.Version)

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitGuardrailCreated(ctx, conn, plan.ID.ValueString(), plan.Version.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForCreation, ResNameGuardrail, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceGuardrail) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Get the resource from AWS
	// 4. Remove resource from state if it is not found
	// 5. Set the arguments and attributes
	// 6. Set the state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().BedrockClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceGuardrailData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findGuardrailByID(ctx, conn, state.ID.ValueString(), state.Version.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
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
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 5. Set the arguments and attributes
	//
	// For simple data types (i.e., schema.StringAttribute, schema.BoolAttribute,
	// schema.Int64Attribute, and schema.Float64Attribue), simply setting the
	// appropriate data struct field is sufficient. The flex package implements
	// helpers for converting between Go and Plugin-Framework types seamlessly. No
	// error or nil checking is necessary.
	//
	// However, there are some situations where more handling is needed such as
	// complex data types (e.g., schema.ListAttribute, schema.SetAttribute). In
	// these cases the flatten function may have a diagnostics return value, which
	// should be appended to resp.Diagnostics.
	// state.ARN = flex.StringToFramework(ctx, out.GuardrailArn)
	// state.ID = flex.StringToFramework(ctx, out.GuardrailId)

	// state.BlockedInputMessaging = flex.StringToFramework(ctx, out.BlockedInputMessaging)
	// state.BlockedOutputsMessaging = flex.StringToFramework(ctx, out.BlockedOutputsMessaging)

	// contentPolicyConfig, d := flattenContentPolicyConfig(ctx, out.ContentPolicy)
	// resp.Diagnostics.Append(d...)
	// state.ContentPolicyConfig = contentPolicyConfig

	// // TIP: Setting a complex type.
	// complexArgument, d := flattenComplexArgument(ctx, out.ComplexArgument)
	// resp.Diagnostics.Append(d...)
	// state.ComplexArgument = complexArgument

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceGuardrail) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TIP: ==== RESOURCE UPDATE ====
	// Not all resources have Update functions. There are a few reasons:
	// a. The AWS API does not support changing a resource
	// b. All arguments have RequiresReplace() plan modifiers
	// c. The AWS API uses a create call to modify an existing resource
	//
	// In the cases of a. and b., the resource will not have an update method
	// defined. In the case of c., Update and Create can be refactored to call
	// the same underlying function.
	//
	// The rest of the time, there should be an Update function and it should
	// do the following things. Make sure there is a good reason if you don't
	// do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan and state
	// 3. Populate a modify input structure and check for changes
	// 4. Call the AWS modify/update function
	// 5. Use a waiter to wait for update to complete
	// 6. Save the request plan to response state
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().BedrockClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceGuardrailData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a modify input structure and check for changes
	if !plan.BlockedInputMessaging.Equal(state.BlockedInputMessaging) ||
		!plan.BlockedOutputsMessaging.Equal(state.BlockedOutputsMessaging) ||
		!plan.KMSKeyArn.Equal(state.KMSKeyArn) ||
		!plan.ContentPolicyConfig.Equal(state.ContentPolicyConfig) ||
		!plan.SensitiveInformationPolicyConfig.Equal(state.SensitiveInformationPolicyConfig) ||
		!plan.TopicPolicyConfig.Equal(state.TopicPolicyConfig) ||
		!plan.WordPolicyConfig.Equal(state.WordPolicyConfig) ||
		!plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) {

		in := &bedrock.UpdateGuardrailInput{
			// TIP: Mandatory or fields that will always be present can be set when
			// you create the Input structure. (Replace these with real fields.)
			Name:                    aws.String(plan.Name.ValueString()),
			BlockedInputMessaging:   aws.String(plan.BlockedInputMessaging.ValueString()),
			BlockedOutputsMessaging: aws.String(plan.BlockedOutputsMessaging.ValueString()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}
		if !plan.KMSKeyArn.IsNull() {
			in.KmsKeyId = aws.String(plan.KMSKeyArn.ValueString())
		}

		if !plan.ContentPolicyConfig.IsNull() {
			var contentPolicyConfig []contentPolicyConfig
			resp.Diagnostics.Append(plan.ContentPolicyConfig.ElementsAs(ctx, &contentPolicyConfig, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			policyConfigInput, d := expandContentPolicyConfig(ctx, contentPolicyConfig)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.ContentPolicyConfig = policyConfigInput
		}

		if !plan.SensitiveInformationPolicyConfig.IsNull() {
			var sensitiveInformationPolicyConfig []sensitiveInformationPolicyConfig
			resp.Diagnostics.Append(plan.SensitiveInformationPolicyConfig.ElementsAs(ctx, &sensitiveInformationPolicyConfig, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sensitiveInformationPolicyConfigInput, d := expandSensitiveInformationPolicyConfig(ctx, sensitiveInformationPolicyConfig)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.SensitiveInformationPolicyConfig = sensitiveInformationPolicyConfigInput
		}
		if !plan.TopicPolicyConfig.IsNull() {
			var topicPolicyConfigData []topicPolicyConfig
			resp.Diagnostics.Append(plan.TopicPolicyConfig.ElementsAs(ctx, &topicPolicyConfigData, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			topicPolicyConfigInput, d := expandTopicPolicyConfig(ctx, topicPolicyConfigData)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.TopicPolicyConfig = topicPolicyConfigInput
		}
		if !plan.WordPolicyConfig.IsNull() {
			var wordPolicyConfigData []wordPolicyConfig
			resp.Diagnostics.Append(plan.WordPolicyConfig.ElementsAs(ctx, &wordPolicyConfigData, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			wordPolicyConfigInput, d := expandWordPolicyConfig(ctx, wordPolicyConfigData)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.WordPolicyConfig = wordPolicyConfigInput
		}

		// TIP: -- 4. Call the AWS modify/update function
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

		// TIP: Using the output from the update function, re-set any computed attributes
		plan.ARN = fwflex.StringToFramework(ctx, out.GuardrailArn)
		plan.ID = fwflex.StringToFramework(ctx, out.GuardrailId)
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitGuardrailUpdated(ctx, conn, plan.ID.ValueString(), state.Version.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForUpdate, ResNameGuardrail, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceGuardrail) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TIP: ==== RESOURCE DELETE ====
	// Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Populate a delete input structure
	// 4. Call the AWS delete function
	// 5. Use a waiter to wait for delete to complete
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().BedrockClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceGuardrailData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	in := &bedrock.DeleteGuardrailInput{
		GuardrailIdentifier: aws.String(state.ID.ValueString()),
		GuardrailVersion:    aws.String(state.Version.ValueString()),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteGuardrail(ctx, in)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
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

	// TIP: -- 5. Use a waiter to wait for delete to complete
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

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceGuardrail) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
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

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
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

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
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

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
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

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
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

// TIP: ==== FLEX ====
// Flatteners and expanders ("flex" functions) help handle complex data
// types. Flatteners take an API data type and return the equivalent Plugin-Framework
// type. In other words, flatteners translate from AWS -> Terraform.
//
// On the other hand, expanders take a Terraform data structure and return
// something that you can send to the AWS API. In other words, expanders
// translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
// func flattenComplexArgument(ctx context.Context, apiObject *bedrock.ComplexArgument) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

// 	if apiObject == nil {
// 		return types.ListNull(elemType), diags
// 	}

// 	obj := map[string]attr.Value{
// 		"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
// 		"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
// 	}
// 	objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
// 	diags.Append(d...)

// 	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
// 	diags.Append(d...)

// 	return listVal, diags
// }

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
// func flattenComplexArguments(ctx context.Context, apiObjects []*bedrock.ComplexArgument) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

// 	if len(apiObjects) == 0 {
// 		return types.ListNull(elemType), diags
// 	}

// 	elems := []attr.Value{}
// 	for _, apiObject := range apiObjects {
// 		if apiObject == nil {
// 			continue
// 		}

// 		obj := map[string]attr.Value{
// 			"nested_required": fwflex.StringValueToFramework(ctx, apiObject.NestedRequired),
// 			"nested_optional": fwflex.StringValueToFramework(ctx, apiObject.NestedOptional),
// 		}
// 		objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
// 		diags.Append(d...)

// 		elems = append(elems, objVal)
// 	}

// 	listVal, d := types.ListValue(elemType, elems)
// 	diags.Append(d...)

// 	return listVal, diags
// }

// func flattenContentPolicyConfig(ctx context.Context, apiObject *awstypes.GuardrailContentPolicy) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	var filtersConfigAttrTypes = map[string]attr.Type{
// 		"input_strength":  types.StringType,
// 		"output_strength": types.StringType,
// 		"type":            types.StringType,
// 	}
// 	var contentPolicyConfigAttrTypes = map[string]attr.Type{
// 		"filters_config": types.SetType{ElemType: types.ObjectType{AttrTypes: filtersConfigAttrTypes}},
// 	}

// 	elemType := types.ObjectType{AttrTypes: contentPolicyConfigAttrTypes}

// 	if apiObject == nil {
// 		return types.ListValueMust(elemType, []attr.Value{}), diags
// 	}

// 	filtersConfig, d := flattenFilters(ctx, apiObject.Filters)
// 	diags.Append(d...)

// 	obj := map[string]attr.Value{
// 		"filters_config": filtersConfig,
// 	}
// 	objVal, d := types.ObjectValue(contentPolicyConfigAttrTypes, obj)
// 	diags.Append(d...)

// 	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
// 	diags.Append(d...)

// 	return listVal, diags
// }

// func flattenFilters(ctx context.Context, apiObject []awstypes.GuardrailContentFilter) (types.Set, diag.Diagnostics) { // nosemgrep:ci.aws-in-func-name
// 	var diags diag.Diagnostics
// 	var filtersConfigAttrTypes = map[string]attr.Type{
// 		"input_strength":  types.StringType,
// 		"output_strength": types.StringType,
// 		"type":            types.StringType,
// 	}
// 	elemType := types.ObjectType{AttrTypes: filtersConfigAttrTypes}

// 	if apiObject == nil {
// 		return types.SetValueMust(elemType, []attr.Value{}), diags
// 	}

// 	elems := []attr.Value{}
// 	for _, filterConfig := range apiObject {
// 		obj := map[string]attr.Value{
// 			"input_strength":  flex.StringToFramework(ctx, (*string)(&filterConfig.InputStrength)),
// 			"output_strength": flex.StringToFramework(ctx, (*string)(&filterConfig.OutputStrength)),
// 			"type":            flex.StringToFramework(ctx, (*string)(&filterConfig.Type)),
// 		}
// 		objVal, d := types.ObjectValue(filtersConfigAttrTypes, obj)
// 		diags.Append(d...)

// 		elems = append(elems, objVal)
// 	}
// 	setVal, d := types.SetValue(elemType, elems)
// 	diags.Append(d...)

// 	return setVal, diags
// }

func expandContentPolicyConfig(ctx context.Context, tfList []contentPolicyConfig) (*awstypes.GuardrailContentPolicyConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	contentPolicyConfig := tfList[0]

	var filtersConfig []filtersConfig
	diags.Append(contentPolicyConfig.FiltersConfig.ElementsAs(ctx, &filtersConfig, false)...)

	return &awstypes.GuardrailContentPolicyConfig{
		FiltersConfig: expandFiltersConfig(filtersConfig),
	}, diags
}

func expandFiltersConfig(tfList []filtersConfig) []awstypes.GuardrailContentFilterConfig { // nosemgrep:ci.aws-in-func-name
	var filtersConfig []awstypes.GuardrailContentFilterConfig
	for _, item := range tfList {
		new := awstypes.GuardrailContentFilterConfig{
			InputStrength:  awstypes.GuardrailFilterStrength(item.InputStrength.ValueString()),
			OutputStrength: awstypes.GuardrailFilterStrength(item.OutputStrength.ValueString()),
			Type:           awstypes.GuardrailContentFilterType(item.Type.ValueString()),
		}
		filtersConfig = append(filtersConfig, new)
	}
	return filtersConfig
}

func expandSensitiveInformationPolicyConfig(ctx context.Context, tfList []sensitiveInformationPolicyConfig) (*awstypes.GuardrailSensitiveInformationPolicyConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	sensitiveInformationPolicyConfigData := tfList[0]

	var piiEntitiesConfigData []piiEntitiesConfig
	diags.Append(sensitiveInformationPolicyConfigData.PIIEntitiesConfig.ElementsAs(ctx, &piiEntitiesConfigData, false)...)
	var regexesConfigData []regexesConfig
	diags.Append(sensitiveInformationPolicyConfigData.RegexesConfig.ElementsAs(ctx, &regexesConfigData, false)...)

	return &awstypes.GuardrailSensitiveInformationPolicyConfig{
		PiiEntitiesConfig: expandPIIEntitiesConfig(piiEntitiesConfigData),
		RegexesConfig:     expandRegexesConfig(regexesConfigData),
	}, diags
}

func expandPIIEntitiesConfig(tfList []piiEntitiesConfig) []awstypes.GuardrailPiiEntityConfig { // nosemgrep:ci.aws-in-func-name
	var piiEntitiesConfig []awstypes.GuardrailPiiEntityConfig
	for _, item := range tfList {
		new := awstypes.GuardrailPiiEntityConfig{
			Action: awstypes.GuardrailSensitiveInformationAction(item.Action.ValueString()),
			Type:   awstypes.GuardrailPiiEntityType(item.Type.ValueString()),
		}
		piiEntitiesConfig = append(piiEntitiesConfig, new)
	}
	return piiEntitiesConfig
}

func expandRegexesConfig(tfList []regexesConfig) []awstypes.GuardrailRegexConfig { // nosemgrep:ci.aws-in-func-name
	var regexesConfig []awstypes.GuardrailRegexConfig
	for _, item := range tfList {
		new := awstypes.GuardrailRegexConfig{
			Action:      awstypes.GuardrailSensitiveInformationAction(item.Action.ValueString()),
			Name:        aws.String(item.Name.ValueString()),
			Pattern:     aws.String(item.Pattern.ValueString()),
			Description: aws.String(item.Description.ValueString()),
		}
		regexesConfig = append(regexesConfig, new)
	}
	return regexesConfig
}

func expandTopicPolicyConfig(ctx context.Context, tfList []topicPolicyConfig) (*awstypes.GuardrailTopicPolicyConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	topicPolicyConfigData := tfList[0]

	var topicsConfigData []topicsConfig
	diags.Append(topicPolicyConfigData.TopicsConfig.ElementsAs(ctx, &topicsConfigData, false)...)

	return &awstypes.GuardrailTopicPolicyConfig{
		TopicsConfig: expandTopicsConfig(topicsConfigData),
	}, diags
}

func expandTopicsConfig(tfList []topicsConfig) []awstypes.GuardrailTopicConfig { // nosemgrep:ci.aws-in-func-name
	var topicsConfig []awstypes.GuardrailTopicConfig
	for _, item := range tfList {
		new := awstypes.GuardrailTopicConfig{
			Definition: aws.String(item.Definition.ValueString()),
			Type:       awstypes.GuardrailTopicType(item.Type.ValueString()),
			Name:       aws.String(item.Name.ValueString()),
			//TODO SASI CHECK THIS BEFORE MERGE
			// Examples:   item.Examples,
		}
		topicsConfig = append(topicsConfig, new)
	}
	return topicsConfig
}

func expandWordPolicyConfig(ctx context.Context, tfList []wordPolicyConfig) (*awstypes.GuardrailWordPolicyConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	wordPolicyConfigData := tfList[0]

	var managedWordListsConfigData []managedWordListsConfig
	diags.Append(wordPolicyConfigData.ManagedWordListsConfig.ElementsAs(ctx, &managedWordListsConfigData, false)...)
	var wordsConfigData []wordsConfig
	diags.Append(wordPolicyConfigData.WordsConfig.ElementsAs(ctx, &wordsConfigData, false)...)

	return &awstypes.GuardrailWordPolicyConfig{
		ManagedWordListsConfig: expandManagedWordListsConfig(managedWordListsConfigData),
		WordsConfig:            expandsWordsConfig(wordsConfigData),
	}, diags
}

func expandManagedWordListsConfig(tfList []managedWordListsConfig) []awstypes.GuardrailManagedWordsConfig { // nosemgrep:ci.aws-in-func-name
	var managedWordListsConfig []awstypes.GuardrailManagedWordsConfig
	for _, item := range tfList {
		new := awstypes.GuardrailManagedWordsConfig{
			Type: awstypes.GuardrailManagedWordsType(item.Type.ValueString()),
		}
		managedWordListsConfig = append(managedWordListsConfig, new)
	}
	return managedWordListsConfig
}

func expandsWordsConfig(tfList []wordsConfig) []awstypes.GuardrailWordConfig { // nosemgrep:ci.aws-in-func-name
	var wordsConfig []awstypes.GuardrailWordConfig
	for _, item := range tfList {
		new := awstypes.GuardrailWordConfig{
			Text: aws.String(item.Text.ValueString()),
		}
		wordsConfig = append(wordsConfig, new)
	}
	return wordsConfig
}

type resourceGuardrailData struct {
	ARN                              types.String                                                      `tfsdk:"arn"`
	BlockedInputMessaging            types.String                                                      `tfsdk:"blocked_input_messaging"`
	BlockedOutputsMessaging          types.String                                                      `tfsdk:"blocked_outputs_messaging"`
	ContentPolicyConfig              fwtypes.ListNestedObjectValueOf[contentPolicyConfig]              `tfsdk:"content_policy_config"`
	CreatedAt                        types.String                                                      `tfsdk:"created_at"`
	Description                      types.String                                                      `tfsdk:"description"`
	ID                               types.String                                                      `tfsdk:"id"`
	KMSKeyArn                        types.String                                                      `tfsdk:"kms_key_arn"`
	Name                             types.String                                                      `tfsdk:"name"`
	SensitiveInformationPolicyConfig fwtypes.ListNestedObjectValueOf[sensitiveInformationPolicyConfig] `tfsdk:"sensitive_information_policy_config"`
	Status                           types.String                                                      `tfsdk:"status"`
	Tags                             types.Map                                                         `tfsdk:"tags"`
	Timeouts                         timeouts.Value                                                    `tfsdk:"timeouts"`
	TopicPolicyConfig                fwtypes.ListNestedObjectValueOf[topicPolicyConfig]                `tfsdk:"topic_policy_config"`
	UpdatedAt                        types.String                                                      `tfsdk:"updated_at"`
	Version                          types.String                                                      `tfsdk:"version"`
	WordPolicyConfig                 fwtypes.ListNestedObjectValueOf[wordPolicyConfig]                 `tfsdk:"word_policy_config"`
}

type contentPolicyConfig struct {
	FiltersConfig fwtypes.ListNestedObjectValueOf[filtersConfig] `tfsdk:"filters_config"`
}

type filtersConfig struct {
	InputStrength  types.String `tfsdk:"input_strength"`
	OutputStrength types.String `tfsdk:"output_strength"`
	Type           types.String `tfsdk:"type"`
}

type sensitiveInformationPolicyConfig struct {
	PIIEntitiesConfig fwtypes.ListNestedObjectValueOf[piiEntitiesConfig] `tfsdk:"pii_entities_config"`
	RegexesConfig     fwtypes.ListNestedObjectValueOf[regexesConfig]     `tfsdk:"regexes_config"`
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
	TopicsConfig fwtypes.ListNestedObjectValueOf[topicsConfig] `tfsdk:"topics_config"`
}

type topicsConfig struct {
	Definition types.String                      `tfsdk:"definition"`
	Examples   fwtypes.ListValueOf[types.String] `tfsdk:"examples"`
	Name       types.String                      `tfsdk:"name"`
	Type       types.String                      `tfsdk:"type"`
}

type wordPolicyConfig struct {
	ManagedWordListsConfig fwtypes.ListNestedObjectValueOf[managedWordListsConfig] `tfsdk:"managed_word_lists_config"`
	WordsConfig            fwtypes.ListNestedObjectValueOf[wordsConfig]            `tfsdk:"words_config"`
}

type managedWordListsConfig struct {
	Type types.String `tfsdk:"type"`
}

type wordsConfig struct {
	Text types.String `tfsdk:"text"`
}

type complexArgumentData struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}

var complexArgumentAttrTypes = map[string]attr.Type{
	"nested_required": types.StringType,
	"nested_optional": types.StringType,
}
