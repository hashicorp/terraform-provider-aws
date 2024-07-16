// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/bedrock/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"regexache"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"google.golang.org/grpc/balancer/grpclb/state"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrock_evaluation_job", name="Evaluation Job")
func newResourceEvaluationJob(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEvaluationJob{}
	
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
	ResNameEvaluationJob = "Evaluation Job"
)

type resourceEvaluationJob struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceEvaluationJob) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrock_evaluation_job"
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
// * If a user can provide a value ("configure a value") for an
//   attribute (e.g., instances = 5), we call the attribute an
//   "argument."
// * You change the way users interact with attributes using:
//     - Required
//     - Optional
//     - Computed
// * There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
// 2. Optional only - the user can configure or omit a value; do not
//    use Default or DefaultFunc
// Optional: true,
//
// 3. Computed only - the provider can provide a value but the user
//    cannot, i.e., read-only
// Computed: true,
//
// 4. Optional AND Computed - the provider or user can provide a value;
//    use this combination if you are using Default
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
func (r *resourceEvaluationJob) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_request_token": schema.StringAttribute{ 
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1,256),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9](-*[a-zA-Z0-9])*$"),"client_request_token must conform to ^[a-zA-Z0-9](-*[a-zA-Z0-9])*$"),
				},
			},
			"customer_encryption_key_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1,2048),
					stringvalidator.RegexMatches(regexache.MustCompile("^arn:aws(-[^:]+)?:kms:[a-zA-Z0-9-]*:[0-9]{12}:((key/[a-zA-Z0-9-]{36})|(alias/[a-zA-Z0-9-_/]+))$"),"customer_encryption_key_id must conform to ^arn:aws(-[^:]+)?:kms:[a-zA-Z0-9-]*:[0-9]{12}:((key/[a-zA-Z0-9-]{36})|(alias/[a-zA-Z0-9-_/]+))$"),
				},
			},
			names.AttrDescription: schema.StringAttribute{ // job description
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1,200),
					stringvalidator.RegexMatches(regexache.MustCompile("^.+$"),"description must conform to ^.+$"),
				},
			},
			names.AttrARN: schema.StringAttribute{
				Required: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1,63),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-z0-9](-*[a-z0-9]){0,62}$"),"arn must conform to ^[a-z0-9](-*[a-z0-9]){0,62}$"),
				},
			},
			"role_arn": schema.StringAttribute{
				Computed: true,
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0,2048),
					stringvalidator.RegexMatches(regexache.MustCompile("^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),"role_arn must conform to ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),
				},
			},
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
			"output_data_config": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dOutputDataConfig](ctx),
			},
			names.AttrStatus :schema.StringAttribute{
				Computed: true,
			},
			"failure_messages": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed: true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0,2048),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-z0-9](-*[a-z0-9]){0,62}$"),"name must conform to ^^[a-z0-9](-*[a-z0-9]){0,62}$"),
				},

			},
		},
		Blocks: map[string] schema.Block {
			"evaluation_config": schema.SingleNestedBlock{
				Blocks: map[string]schema.Block{
					"Automated": schema.SingleNestedBlock{
						Blocks: map[string]schema.Block{
							"automated_evaluation_config": schema.ListNestedBlock{
								//CustomType: fwtypes.NewListNestedObjectTypeOf[dsEvaluationDataset](ctx),
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"metric_names":schema.ListAttribute{
											Required: true,
											Validators: []validator.List{
												listvalidator.SizeBetween(1, 10),
												listvalidator.ValueStringsAre(stringvalidator.LengthBetween(1,63)),
												listvalidator.ValueStringsAre(stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-zA-Z-_.]+$`), "metric_names must conform to: ^[0-9a-zA-Z-_.]+$ ")),
											},
										},
										"task_type":schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.LengthBetween(0,2048),
												stringvalidator.RegexMatches(regexache.MustCompile("^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),"role_arn must conform to ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),
												stringvalidator.AtLeastOneOf(path.Expressions{
													path.MatchRoot("Summarization"),path.MatchRoot("Classification"),path.MatchRoot("QuestionAndAnswer"),path.MatchRoot("Generation"),path.MatchRoot("Custom"),
												}...),
											},
										},
									},
									Blocks: map[string]schema.Block{ // ds evaluation dataset
										"data_set": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(0,63),
														stringvalidator.RegexMatches(regexache.MustCompile("^[0-9a-zA-Z-_.]+$"),"role_arn must conform to ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),
														stringvalidator.AtLeastOneOf(path.Expressions{
															path.MatchRoot("Builtin.Bold"),path.MatchRoot("Builtin.BoolQ"),path.MatchRoot("Builtin.NaturalQuestions"),path.MatchRoot("Builtin.Gigaword"),path.MatchRoot("Builtin.RealToxicityPrompts"), path.MatchRoot("Builtin.TriviaQa"), path.MatchRoot("Builtin.WomensEcommerceClothingReviews"), path.MatchRoot("Builtin.Wikitext2"),
														}...),
													},
														},
											},
											Blocks: map[string]schema.Block{
												"evaluation_dataset_location": schema.SingleNestedBlock{
													Attributes: map[string]schema.Attribute{
														"s3_uri": schema.StringAttribute{
															Optional: true,
															Validators: []validator.String {
																stringvalidator.LengthBetween(1,1024),
																stringvalidator.RegexMatches(regexache.MustCompile("^s3://[a-z0-9][\\.\\-a-z0-9]{1,61}[a-z0-9](/.*)?$"),"role_arn must conform to ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),
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
			"inference_config": schema.SingleNestedBlock{
				Blocks: map[string]schema.Block{
					"bedrock_model": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Blocks: map[string]schema.Block{
								"evaluation_model_config": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"inference_params": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.LengthBetween(1,1023),	
											},
										},
										"model_identifier": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.LengthBetween(1,2048),
												stringvalidator.RegexMatches(regexache.MustCompile("^arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}(([:][a-z0-9-]{1,63}){0,2})?/[a-z0-9]{12})|(:foundation-model/([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})))|(([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2}))|(([0-9a-zA-Z][_-]?)+)$"),"model_identifier must match ^arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}(([:][a-z0-9-]{1,63}){0,2})?/[a-z0-9]{12})|(:foundation-model/([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})))|(([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2}))|(([0-9a-zA-Z][_-]?)+)$"),
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
}

func (r *resourceEvaluationJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)
	var plan dEvaluationJob
	
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	in := &bedrock.CreateEvaluationJobInput{}
	resp.Diagnostics.Append(flex.Expand(ctx,plan,in)...)
	if resp.Diagnostics.HasError(){
		return
	}
	
	out, err := conn.CreateEvaluationJob(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameEvaluationJob, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.JobArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameEvaluationJob, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}



	/*	
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitEvaluationJobCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForCreation, ResNameEvaluationJob, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	*/
	
	plan.JobArn = flex.StringToFramework(ctx,out.JobArn)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEvaluationJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dEvaluationJob
	resp.Diagnostics.Append(req.State.Get(ctx,&state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)
	out, _ := conn.GetEvaluationJob(ctx,&bedrock.GetEvaluationJobInput{
		JobIdentifier: state.Arn.ValueStringPointer(),
	})

	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Get the resource from AWS
	// 4. Remove resource from state if it is not found
	// 5. Set the arguments and attributes
	// 6. Set the state



	
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
	state.ID = flex.StringToFramework(ctx, out.EvaluationJobId)
	state.Name = flex.StringToFramework(ctx, out.EvaluationJobName)
	state.Type = flex.StringToFramework(ctx, out.EvaluationJobType)
	
	// TIP: Setting a complex type.
	complexArgument, d := flattenComplexArgument(ctx, out.ComplexArgument)
	resp.Diagnostics.Append(d...)
	state.ComplexArgument = complexArgument
	
	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEvaluationJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	var plan, state resourceEvaluationJobData
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

		in := &bedrock.UpdateEvaluationJobInput{
			// TIP: Mandatory or fields that will always be present can be set when
			// you create the Input structure. (Replace these with real fields.)
			EvaluationJobId:   aws.String(plan.ID.ValueString()),
			EvaluationJobName: aws.String(plan.Name.ValueString()),
			EvaluationJobType: aws.String(plan.Type.ValueString()),
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
		out, err := conn.UpdateEvaluationJob(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Bedrock, create.ErrActionUpdating, ResNameEvaluationJob, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.EvaluationJob == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Bedrock, create.ErrActionUpdating, ResNameEvaluationJob, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		
		// TIP: Using the output from the update function, re-set any computed attributes
		plan.ARN = flex.StringToFramework(ctx, out.EvaluationJob.Arn)
		plan.ID = flex.StringToFramework(ctx, out.EvaluationJob.EvaluationJobId)
	}

	
	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitEvaluationJobUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForUpdate, ResNameEvaluationJob, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	
	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEvaluationJob) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
		conn := r.Meta().BedrockClient(ctx)
	
		var state dEvaluationJob

		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		
		in := &bedrock.StopEvaluationJobInput{
			JobIdentifier: state.Arn.ValueStringPointer(),
		}
	
		_, err := conn.StopEvaluationJob(ctx,in)

		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
				return
			}
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, state.Name.ValueString() , state.Arn.ValueString(), err),
				err.Error(),
			)
			return
		}
		
	/*
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitEvaluationJobDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForDeletion, ResNameEvaluationJob, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	*/
}

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceEvaluationJob) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}


// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
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
func waitEvaluationJobCreated(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*awstypes.EvaluationJob, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusEvaluationJob(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.EvaluationJob); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitEvaluationJobUpdated(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*awstypes.EvaluationJob, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusEvaluationJob(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.EvaluationJob); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitEvaluationJobDeleted(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*awstypes.EvaluationJob, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusDeleting, statusNormal},
		Target:                    []string{},
		Refresh:                   statusEvaluationJob(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.EvaluationJob); ok {
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
func statusEvaluationJob(ctx context.Context, conn *bedrock.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findEvaluationJobByID(ctx, conn, id)
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
func findEvaluationJobByID(ctx context.Context, conn *bedrock.Client, id string) (*awstypes.EvaluationJob, error) {
	in := &bedrock.GetEvaluationJobInput{
		Id: aws.String(id),
	}
	
	out, err := conn.GetEvaluationJob(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.EvaluationJob == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.EvaluationJob, nil
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
// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.

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
type dEvaluationJob struct {
	ClientRequestToken types.String  `tfsdk:"client_request_token"`
	CustomerEncryptionKeyId types.String  `tfsdk:"customer_encryption_key_id"`
	EvaluationConfig fwtypes.ObjectValueOf[dAutomatedEvaluationConfig]  `tfsdk:"evaluation_config"`
	InferenceConfig fwtypes.ObjectValueOf[dEvaluationInferenceConfig]  `tfsdk:"inference_config"`
	Description types.String  `tfsdk:"description"`
	Name types.String  `tfsdk:"name"`
	// tags idk lol
	OutputDataConfig fwtypes.ObjectValueOf[dOutputDataConfig] `tfsdk:"output_data_config"`
	RoleArn types.String `tfsdk:"role_arn"`
	Arn types.String `tfsdk:"arn"`

	// computed
	CreationTime timetypes.RFC3339 `tfsdk:"creation_time"`
	JobType fwtypes.StringEnum[awstypes.EvaluationJobType] `tfsdk:"type"`
	OutOutputDataConfig fwtypes.ListNestedObjectValueOf[dEvaluationDatasetLocation] `tfsdk:"output_data_config"`
	Status fwtypes.StringEnum[awstypes.EvaluationJobStatus] `tfsdk:"status"`
	FailureMessages fwtypes.ListValueOf[types.String] `tfsdk:"failure_messages"`
	LastModifiedTime timetypes.RFC3339 `tfsdk:"last_modified_time"`



}

type temp struct {
	EvaluationConfig types.String `tfsdk:"name"` // help
	InferenceConfig types.String `tfsdk:"name"`// help
	Arn types.String `tfsdk:"arn"` // combine
	Name types.String `tfsdk:"name"` // combine
	FailureMessages fwtypes.ListValueOf[types.String] `tfsdk:"failure_messages"`
}

// start of evaluation_config
type dEvaluationConfig struct {
	Automated             fwtypes.ObjectValueOf[dAutomatedEvaluationConfig]   `tfsdk:"automated"`
	// Human             fwtypes.ListNestedObjectValueOf[dsHumanEvaluationConfig]   `tfsdk:"human"` 
}
type dAutomatedEvaluationConfig struct {
	EvaluationDataSet fwtypes.ListNestedObjectValueOf[dEvaluationDatasetMetricConfig] `tfsdk:"evaluation_data_set"` // this is a list make sure its a list
	/*
		Array Members: Minimum number of 1 item. Maximum number of 5 items.
	*/
}
type dEvaluationDatasetMetricConfig struct {

	Dataset fwtypes.ObjectValueOf[dEvaluationDataset] `tfsdk:"dataset"`
	MetricNames types.List `tfsdk:"metric_names"` // array members; min 1 max 10 items ; min length of items 1 max 63 ; pattern ^[0-9a-zA-Z-_.]+$ ; required yes
	TaskType types.String `tfsdk:"task_type"` // min length of 1 max 63 ; pattern ^[A-Za-z0-9]+$ ; Valid values Summarization | Classification | QuestionAndAnswer | Generation | Custom ; required yes
}
type dEvaluationDataset struct {
	name types.String `tfsdk:"name"` 
	/*
	Used to specify supported built-in prompt datasets. Valid values are Builtin.Bold, Builtin.BoolQ, Builtin.NaturalQuestions, Builtin.Gigaword, Builtin.RealToxicityPrompts, Builtin.TriviaQa, Builtin.T-Rex, Builtin.WomensEcommerceClothingReviews and Builtin.Wikitext2.

	Type: String

	Length Constraints: Minimum length of 1. Maximum length of 63.

	Pattern: ^[0-9a-zA-Z-_.]+$

	Required: Yes	
	*/
	EvaluationDatasetLocation fwtypes.ObjectValueOf[dEvaluationDatasetLocation] `tfsdk:"evaluation_dataset_location"`
	/*
	Type: EvaluationDatasetLocation object

	Note: This object is a Union. Only one member of this object can be specified or returned.

	Required: No
	*/	
}
type dEvaluationDatasetLocation struct {
	S3Uri types.String `tfsdk:"s3_uri"` 
	/*
	Length Constraints: Minimum length of 1. Maximum length of 1024.

	Pattern: ^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$

	Required: No

	*/
}
// end of evaluation_config


// start of inference_config
type dEvaluationInferenceConfig struct {
	Models fwtypes.ListNestedObjectValueOf[dEvaluationModelConfig] `tfsdk:"bedrock_model"`
	/*
	Array Members: Minimum number of 1 item. Maximum number of 2 items.

	Required: No
	*/
}

type dEvaluationModelConfig struct {
	EvaluationModelConfig fwtypes.ObjectValueOf[dEvaluationBedrockModel] `tfsdk:"evaluation_model_config"`
}

type dEvaluationBedrockModel struct {
	InferenceParams types.String `tfsdk:"inference_params"`
	/*
	Each Amazon Bedrock support different inference parameters that change how the model behaves during inference.

	Type: String

	Length Constraints: Minimum length of 1. Maximum length of 1023.

	Required: Yes
	*/
	ModelIdentifiers types.String `tfsdk:"model_identifier"`
	/*

	The ARN of the Amazon Bedrock model specified.

	Type: String

	Length Constraints: Minimum length of 1. Maximum length of 2048.

	Pattern: ^arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}(([:][a-z0-9-]{1,63}){0,2})?/[a-z0-9]{12})|(:foundation-model/([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})))|(([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2}))|(([0-9a-zA-Z][_-]?)+)$

	Required: Yes

	*/
}
// end of evaluation_config

type dOutputDataConfig struct {
	S3Uri types.String  `tfsdk:"s3_uri"`
}

