// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Slot")
func newResourceSlot(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSlot{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameSlot = "Slot"
)

type resourceSlot struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceSlot) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lexv2models_slot"
}

// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
//   - If a user can provide a value ("configure a value") for an
//     attribute (e.g., instances = 5), we call the attribute an
//     "argument."
//   - You change the way users interact with attributes using:
//   - Required
//   - Optional
//   - Computed
//   - There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
//  2. Optional only - the user can configure or omit a value; do not
//     use Default or DefaultFunc
//
// Optional: true,
//
//  3. Computed only - the provider can provide a value but the user
//     cannot, i.e., read-only
//
// Computed: true,
//
//  4. Optional AND Computed - the provider or user can provide a value;
//     use this combination if you are using Default
//
// Optional: true,
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (r *resourceSlot) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	slotValueOverrideO := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"shape": types.StringType,
			"value": types.ObjectType{
				AttrTypes: map[string]attr.Types{
					"interpreted_value" : types.StringType,
				},
			},
		},
	}
	slotValueOverrideO.AttrTypes["values"] = slotValueOverrideO

	messageNBO := stypes.Object{
		Blocks: map[string]schema.Block{
			"custom_playload": schema.ListedNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"image_response_card": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"image_url": schema.StringAttribute{
							Optional: true,
						},
						"subtitle": schema.StringAttribute{
							Optional: true,
						},
						"title": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"button": schema.ListNestedBlock{
							Nestedobject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									
								}
							}
						}
					}
				},
			}

		}
		"message": schema.NestedBlockObject{
			Required: true,
			Blocks: map[string]schema.Block{


					Blocks: map[string]schema.Block{
						"buttons": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[strings]schema.Attribute{
									"text": schema.StringAttribute{
										Required: true,
									},
									"value": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
				"plain_text_message": schema.SingleNestedBlock{
					Attributes: map[strings]schema.Attribute{
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
				"ssml_message": schema.SingleNestedBlock{
					Attributes: map[strings]schema.Attribute{
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			// "arn": framework.ARNAttributeComputedOnly(),
			"bot_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bot_version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"intent_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"locale_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// "id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slot_type_id": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"value_elicitation_setting": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"slot_constraint": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							enum.FrameworkValidate[awstypes.SlotConstraint](),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"default_value_specification": schema.SingleNestedBlock{
						Blocks: map[string]schema.Block{
							"default_value_list": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"default_value": schema.StringAttribute{
											Required: true,
										},
									},
								},
							},
						},
					},
					"prompt_specification": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"max_retries": schema.Int32Attribute{
								Required: true,
							},
							"allow_interrupt": schema.BoolAttribute{
								Optional: true,
							},
							"message_selection_strategy": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									enum.FrameworkValidate[awstypes.MessageSelectionStrategy](),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"message_groups": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								Blocks: map[string]schema.Block{
									"message": schema.SingleNestedBlock{
										Required: true,
										Blocks: map[string]schema.Block{
											"custom_payload": schema.SingleNestedBlock{
												NestedObject: schema.NestedBlockObject{
													Attributes: map[string]schema.Attribute{
														"value": schema.StringAttribute{
															Required: true,
														},
													},
												},
											},
											"image_response_card": schema.SingleNestedBlock{
												Attributes: map[strings]schema.Attribute{
													"title": schema.StringAttribute{
														Required: true,
													},
													"image_url": schema.StringAttribute{
														Optional: true,
													},
													"subtitle": schema.StringAttribute{
														Optional: true,
													},
												},
												Blocks: map[string]schema.Block{
													"buttons": schema.ListNestedBlock{
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[strings]schema.Attribute{
																"text": schema.StringAttribute{
																	Required: true,
																},
																"value": schema.StringAttribute{
																	Required: true,
																},
															},
														},
													},
												},
											},
											"plain_text_message": schema.SingleNestedBlock{
												Attributes: map[strings]schema.Attribute{
													"value": schema.StringAttribute{
														Required: true,
													},
												},
											},
											"ssml_message": schema.SingleNestedBlock{
												Attributes: map[strings]schema.Attribute{
													"value": schema.StringAttribute{
														Required: true,
													},
												},
											},
										},
									"variations": schema.ListNestedBlock{
										Blocks: map[string]schema.Block{
											"custom_payload": schema.SingleNestedBlock{
												Attributes: map[string]schema.Attribute{
													"value": schema.StringAttribute{
														Required: true,
													},
												},
											},
											"image_response_card": schema.SingleNestedBlock{
												Attributes: map[strings]schema.Attribute{
													"title": schema.StringAttribute{
														Required: true,
													},
													"image_url": schema.StringAttribute{
														Optional: true,
													},
													"subtitle": schema.StringAttribute{
														Optional: true,
													},
												},
												Blocks: map[string]schema.Block{
													"buttons": schema.ListNestedBlock{
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[strings]schema.Attribute{
																"text": schema.StringAttribute{
																	Required: true,
																},
																"value": schema.StringAttribute{
																	Required: true,
																},
															},
														},
													},
												},
											},
											"plain_text_message": schema.SingleNestedBlock{
												Attributes: map[strings]schema.Attribute{
													"value": schema.StringAttribute{
														Required: true,
													},
												},
											},
											"ssml_message": schema.SingleNestedBlock{
												Attributes: map[strings]schema.Attribute{
													"value": schema.StringAttribute{
														Required: true,
													},
												},
											},
										},
									},
								},
							},
							"prompt_attempts_specification": schema.MapAttribute{
								Optional: true,
								Attributes: map[strings]schema.Attribute{
									"allow_interrupts": schema.BoolAttribute{
										Optional: true,
									},
								},
								Blocks: map[strings]schema.Block{
									"allow_input_types": schema.SingleNestedBlock{
										Attributes: map[strings]schema.Attribute{
											"allow_audio_input": schema.BoolAttribute{
												Required: true,
											},
											"allow_dtmf_input": schema.BoolAttribute{
												Required: true,
											},
										},
									},
									"audio_and_dtmf_input_specification": schema.SingleNestedBlock{
										Optional: true,
										Attributes: map[strings]schema.Attribute{
											"start_timeout_ms": schema.Int32Attribute{
												Required: true,
											},
										},
										"audio_specification": schema.SingleNestedBlock{
											Optional: true,
											Attributes: map[strings]schema.Attribute{
												"end_timeout_ms": schema.Int32Attribute{
													Required: true,
												},
												"max_length_ms": schema.Int32Attribute{
													Required: true,
												},
											},
										},
										"dtmf_specification": schema.SingleNestedBlock{
											Optional: true,
											Attributes: map[strings]schema.Attribute{
												"deletion_character": schema.StringAttribute{
													Required: true,
												},
												"end_character": schema.StringAttribute{
													Required: true,
												},
												"end_timeout_ms": schema.StringAttribute{
													Required: true,
												},
												"max_length": schema.Int32Attribute{
													Required: true,
												},
											},
										},
									},
									"text_input_specification": schema.SingleNestedBlock{
										Optional: true,
										Attributes: map[strings]schema.Attribute{
											"start_timeout_ms": schema.Int32Attribute{
												Required: true,
											},
										},
									},
								},
							},
						},
					},
					"sample_utterances": schema.ListNestedBlock{
						Optional: true,
						Attributes: map[strings]schema.Attribute{
							"utterance": schema.StringAttribute{
								Required: true,
							},
						},
					},
					"slot_capture_setting": schema.SingleNestedBlock{
						Optional: true,
						Blocks: map[string]schema.Block{
							"capture_conditional": schema.SingleNestedBlock{
								Optional: true,
								Blocks: map[string]schema.Block{
									"condition_specification": schema.SingleNestedBlock{
										Optional: true,
										Attributes: map[string]schema.Attribute{
											"active": schema.BoolAttribute{
												Required: true,
											},
										},
										Blocks: map[string]schema.Block{
											"conditional_branch": schema.ListNestedBlock{
												Required: true,
												NestedObject: schema.NestedBlockObject{
													Attributes: map[string]schema.Attribute{
														"name": schema.StringAttribute{
															Required: true,
														},
													},
													Blocks: map[string]schema.Block{
														"condition": schema.SingleNestedBlock{
															Validators: []validator.Object{
																objectvalidator.IsRequired(),
															},
															Attributes: map[string]schema.Attribute{
																"expression_string": schema.StringAttribute{
																	Required: true,
																},
															},
														},
														"next_step": schema.SingleNestedBlock{
															Validators: []validator.Object{
																objectvalidator.IsRequired(),
															},
															NestedObject: schema.NestedBlockObject{
																Blocks: map[string]schema.SingleNestedBlock{
																	"dialog_state": schema.SingleNestedBlock{
																		Optional: true,
																		Blocks: map[string]SingleNestedBlock{
																			"dialog_action": schema.SingleNestedBlock{
																				Optional: true,
																				Attributes: map[string]schema.Attribute{
																					"slot_to_elicit": schema.StringAttribute{
																						Optional: true,
																					},
																					"supress_next_message": schema.BoolAttribute{
																						Optional: true,
																					},
																				},
																				Blocks: map[string]schema.Block{
																					"type": schema.StringAttribute{
																						Required: true,
																						Validators: []validator.String{
																							enum.FrameworkValidate[awstypes.DialogActionType](),
																						},
																					},
																				},
																			},
																			"intent": schema.SingleNestedBlock{
																				Optional: true,
																				Attributes: map[string]schema.Attribute{
																					"intent_overide": schema.NestedBlockObject{
																						Attributes: map[string]schema.Attribute{
																							"name": schema.StringAttribute{
																								Optional: true,
																							},
																							NestedObject: schema.NestedBlockObject{
																								Blocks: map[string]schema.Block{
																								"slots": schema.MapAttribute{
																									NestedObject: slotValueOverrideO,
																								},
																							},
																						},
																					},
																				},
																			},
																			"session_attributes": schema.MapAttribute{
																				ElementType: types.StringType,
																			},
																		},
																	},
																},
																"response"
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
					},
				},
			},
		},
	},
}

func (r *resourceSlot) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TIP: ==== RESOURCE CREATE ====
	// Generally, the Create function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan
	// 3. Populate a create input structure
	// 4. Call the AWS create/put function
	// 5. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work, as well as any computed
	//    only attributes.
	// 6. Use a waiter to wait for create to complete
	// 7. Save the request plan to response state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().LexV2ModelsClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceSlotData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a create input structure
	in := &lexv2models.CreateSlotInput{
		// TIP: Mandatory or fields that will always be present can be set when
		// you create the Input structure. (Replace these with real fields.)
		SlotName: aws.String(plan.Name.ValueString()),
		SlotType: aws.String(plan.Type.ValueString()),
	}

	if !plan.Description.IsNull() {
		// TIP: Optional fields should be set based on whether or not they are
		// used.
		in.Description = aws.String(plan.Description.ValueString())
	}
	if !plan.ComplexArgument.IsNull() {
		// TIP: Use an expander to assign a complex argument. The elements must be
		// deserialized into the appropriate struct before being passed to the expander.
		var tfList []complexArgumentData
		resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.ComplexArgument = expandComplexArgument(tfList)
	}

	// TIP: -- 4. Call the AWS create function
	out, err := conn.CreateSlot(ctx, in)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameSlot, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Slot == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameSlot, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// TIP: -- 5. Using the output from the create function, set the minimum attributes
	plan.ARN = flex.StringToFramework(ctx, out.Slot.Arn)
	plan.ID = flex.StringToFramework(ctx, out.Slot.SlotId)

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitSlotCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForCreation, ResNameSlot, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSlot) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().LexV2ModelsClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceSlotData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findSlotByID(ctx, conn, state.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionSetting, ResNameSlot, state.ID.String(), err),
			err.Error(),
		)
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
	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.ID = flex.StringToFramework(ctx, out.SlotId)
	state.Name = flex.StringToFramework(ctx, out.SlotName)
	state.Type = flex.StringToFramework(ctx, out.SlotType)

	// TIP: Setting a complex type.
	complexArgument, d := flattenComplexArgument(ctx, out.ComplexArgument)
	resp.Diagnostics.Append(d...)
	state.ComplexArgument = complexArgument

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSlot) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	conn := r.Meta().LexV2ModelsClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceSlotData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a modify input structure and check for changes
	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.ComplexArgument.Equal(state.ComplexArgument) ||
		!plan.Type.Equal(state.Type) {

		in := &lexv2models.UpdateSlotInput{
			// TIP: Mandatory or fields that will always be present can be set when
			// you create the Input structure. (Replace these with real fields.)
			SlotId:   aws.String(plan.ID.ValueString()),
			SlotName: aws.String(plan.Name.ValueString()),
			SlotType: aws.String(plan.Type.ValueString()),
		}

		if !plan.Description.IsNull() {
			// TIP: Optional fields should be set based on whether or not they are
			// used.
			in.Description = aws.String(plan.Description.ValueString())
		}
		if !plan.ComplexArgument.IsNull() {
			// TIP: Use an expander to assign a complex argument. The elements must be
			// deserialized into the appropriate struct before being passed to the expander.
			var tfList []complexArgumentData
			resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.ComplexArgument = expandComplexArgument(tfList)
		}

		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateSlot(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameSlot, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Slot == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameSlot, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		// TIP: Using the output from the update function, re-set any computed attributes
		plan.ARN = flex.StringToFramework(ctx, out.Slot.Arn)
		plan.ID = flex.StringToFramework(ctx, out.Slot.SlotId)
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitSlotUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForUpdate, ResNameSlot, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceSlot) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().LexV2ModelsClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceSlotData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	in := &lexv2models.DeleteSlotInput{
		SlotId: aws.String(state.ID.ValueString()),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteSlot(ctx, in)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameSlot, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitSlotDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameSlot, state.ID.String(), err),
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
func (r *resourceSlot) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
func waitSlotCreated(ctx context.Context, conn *lexv2models.Client, id string, timeout time.Duration) (*lexv2models.Slot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusSlot(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexv2models.Slot); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitSlotUpdated(ctx context.Context, conn *lexv2models.Client, id string, timeout time.Duration) (*lexv2models.Slot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusSlot(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexv2models.Slot); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitSlotDeleted(ctx context.Context, conn *lexv2models.Client, id string, timeout time.Duration) (*lexv2models.Slot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusSlot(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexv2models.Slot); ok {
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
func statusSlot(ctx context.Context, conn *lexv2models.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findSlotByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findSlotByID(ctx context.Context, conn *lexv2models.Client, id string) (*lexv2models.Slot, error) {
	in := &lexv2models.GetSlotInput{
		Id: aws.String(id),
	}

	out, err := conn.GetSlot(ctx, in)
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

	if out == nil || out.Slot == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Slot, nil
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
func flattenComplexArgument(ctx context.Context, apiObject *lexv2models.ComplexArgument) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
		"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
	}
	objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
func flattenComplexArguments(ctx context.Context, apiObjects []*lexv2models.ComplexArgument) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		obj := map[string]attr.Value{
			"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
			"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
		}
		objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

// TIP: Remember, as mentioned above, expanders take a Terraform data structure
// and return something that you can send to the AWS API. In other words,
// expanders translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
func expandComplexArgument(tfList []complexArgumentData) *lexv2models.ComplexArgument {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &lexv2models.ComplexArgument{
		NestedRequired: aws.String(tfObj.NestedRequired.ValueString()),
	}
	if !tfObj.NestedOptional.IsNull() {
		apiObject.NestedOptional = aws.String(tfObj.NestedOptional.ValueString())
	}

	return apiObject
}

// TIP: Even when you have a list with max length of 1, this plural function
// works brilliantly. However, if the AWS API takes a structure rather than a
// slice of structures, you will not need it.
func expandComplexArguments(tfList []complexArgumentData) []*lexv2models.ComplexArgument {
	// TIP: The AWS API can be picky about whether you send a nil or zero-
	// length for an argument that should be cleared. For example, in some
	// cases, if you send a nil value, the AWS API interprets that as "make no
	// changes" when what you want to say is "remove everything." Sometimes
	// using a zero-length list will cause an error.
	//
	// As a result, here are two options. Usually, option 1, nil, will work as
	// expected, clearing the field. But, test going from something to nothing
	// to make sure it works. If not, try the second option.
	// TIP: Option 1: Returning nil for zero-length list
	if len(tfList) == 0 {
		return nil
	}

	var apiObject []*lexv2models.ComplexArgument
	// TIP: Option 2: Return zero-length list for zero-length list. If option 1 does
	// not work, after testing going from something to nothing (if that is
	// possible), uncomment out the next line and remove option 1.
	//
	// apiObject := make([]*lexv2models.ComplexArgument, 0)

	for _, tfObj := range tfList {
		item := &lexv2models.ComplexArgument{
			NestedRequired: aws.String(tfObj.NestedRequired.ValueString()),
		}
		if !tfObj.NestedOptional.IsNull() {
			item.NestedOptional = aws.String(tfObj.NestedOptional.ValueString())
		}

		apiObject = append(apiObject, item)
	}

	return apiObject
}

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type resourceSlotData struct {
	ARN             types.String   `tfsdk:"arn"`
	ComplexArgument types.List     `tfsdk:"complex_argument"`
	Description     types.String   `tfsdk:"description"`
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
	Type            types.String   `tfsdk:"type"`
}

type complexArgumentData struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}

var complexArgumentAttrTypes = map[string]attr.Type{
	"nested_required": types.StringType,
	"nested_optional": types.StringType,
}
