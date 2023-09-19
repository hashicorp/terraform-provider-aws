// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

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
	// using the services/lexv2models/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	// "time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexv2models"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexv2models/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
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
// @FrameworkResource(name="Bot")
func newResourceBot(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceBot{}

	// r.SetDefaultCreateTimeout(30 * time.Minute)
	// r.SetDefaultUpdateTimeout(30 * time.Minute)
	// r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameBot = "Bot"
)

type resourceBot struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceBot) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lexv2models_bot"
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
func (r *resourceBot) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"idle_session_ttl_in_seconds": schema.Int64Attribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				// TIP: ==== PLAN MODIFIERS ====
				// Plan modifiers were introduced with Plugin-Framework to provide a mechanism
				// for adjusting planned changes prior to apply. The planmodifier subpackage
				// provides built-in modifiers for many common use cases such as 
				// requiring replacement on a value change ("ForceNew: true" in Plugin-SDK 
				// resources).
				//
				// See more:
				// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"role_arn": schema.StringAttribute{
				Required: true,
			},
			"test_bot_alias_tags": schema.StringAttribute{
				Required: false,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"members": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alias_id": schema.StringAttribute{
							Required: true,
						},
						"alias_name": schema.StringAttribute{
							Required: true,
						},
						"id": schema.StringAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"version": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"data_privacy": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"child_directed": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			// "timeouts": timeouts.Block(ctx, timeouts.Opts{
			// 	Create: true,
			// 	Update: true,
			// 	Delete: true,
			// }),
		},
	}
}

func (r *resourceBot) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().LexV2ModelsClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceBotData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var dp []dataPrivacyData
	resp.Diagnostics.Append(plan.DataPrivacy.ElementsAs(ctx, &dp, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dpInput, d := expandDataPrivacy(ctx, dp)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a create input structure
	in := &lexv2models.CreateBotInput{
		// TIP: Mandatory or fields that will always be present can be set when
		// you create the Input structure. (Replace these with real fields.)
		Name:                    aws.String(plan.Name.ValueString()),
		Type:                    aws.String(plan.Type.ValueString()),
		DataPrivacy:             dpInput,
		IdleSessionTTLInSeconds: aws.Int64(plan.IdleSessionTTLInSeconds.ValueInt64()),
		RoleARN:                 aws.String(plan.RoleARN.ValueString()),
		Tags:                    getTagsIn(ctx),
	}

	if !plan.Description.IsNull() {
		// TIP: Optional fields should be set based on whether or not they are
		// used.
		in.Description = aws.String(plan.Description.ValueString())
	}

	// TIP: -- 4. Call the AWS create function
	out, err := conn.CreateBot(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBot, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.BotId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBot, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// TIP: -- 5. Using the output from the create function, set the minimum attributes
	plan.ID = flex.StringToFramework(ctx, out.BotId)

	// // TIP: -- 6. Use a waiter to wait for create to complete
	// createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	// _, err = waitBotCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForCreation, ResNameBot, plan.Name.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// TIP: -- 7. Save the request plan to response state
	state := plan
	// resp.Diagnostics.Append(state.refreshFromOutput(ctx, out.BotId)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceBot) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	var state resourceBotData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := FindBotByID(ctx, conn, state.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionSetting, ResNameBot, state.ID.String(), err),
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

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, out)...)

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceBot) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	var plan, state resourceBotData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a modify input structure and check for changes
	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.IdleSessionTTLInSeconds.Equal(state.IdleSessionTTLInSeconds) ||
		!plan.RoleARN.Equal(state.RoleARN) ||
		!plan.TestBotAliasTags.Equal(state.TestBotAliasTags) ||
		!plan.DataPrivacy.Equal(state.DataPrivacy) ||
		!plan.Type.Equal(state.Type) {

		var dp []dataPrivacyData
		resp.Diagnostics.Append(plan.DataPrivacy.ElementsAs(ctx, &dp, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		dpInput, d := expandDataPrivacy(ctx, dp)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := &lexv2models.UpdateBotInput{
			// TIP: Mandatory or fields that will always be present can be set when
			// you create the Input structure. (Replace these with real fields.)
			Id:                      aws.String(plan.ID.ValueString()),
			Name:                    aws.String(plan.Name.ValueString()),
			Type:                    aws.String(plan.Type.ValueString()),
			IdleSessionTTLInSeconds: aws.Int64(plan.IdleSessionTTLInSeconds.ValueInt64()),
			DataPrivacy:             dpInput,
			RoleARN:                 aws.String(plan.RoleARN.ValueString()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}
		if !plan.Type.IsNull() {
			in.Type = aws.String(plan.Type.ValueString())
		}
		if !plan.Members.IsNull() {
			// TIP: Use an expander to assign a complex argument. The elements must be
			// deserialized into the appropriate struct before being passed to the expander.
			var tfList []membersData
			resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.ComplexArgument, d = expandMembers(ctx, tfList)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateBot(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBot, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.BotId == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBot, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		// TIP: Using the output from the update function, re-set any computed attributes
		state.refreshFromOutput(ctx, out.BotId)
	}

	
	// TIP: -- 5. Use a waiter to wait for update to complete
	// updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	// _, err := waitBotUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForUpdate, ResNameBot, plan.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceBot) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	var state resourceBotData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	in := &lexv2models.DeleteBotInput{
		BotId: aws.String(state.ID.ValueString()),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteBot(ctx, in)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameBot, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	// deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	// _, err = waitBotDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameBot, state.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }
}

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceBot) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
// func waitBotCreated(ctx context.Context, conn *lexv2models.Client, id string, timeout time.Duration) (*lexv2models.Bot, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{},
// 		Target:                    []string{statusNormal},
// 		Refresh:                   statusBot(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*lexv2models.Bot); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
// func waitBotUpdated(ctx context.Context, conn *lexv2models.Client, id string, timeout time.Duration) (*lexv2models.Bot, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{statusChangePending},
// 		Target:                    []string{statusUpdated},
// 		Refresh:                   statusBot(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*lexv2models.Bot); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
// func waitBotDeleted(ctx context.Context, conn *lexv2models.Client, id string, timeout time.Duration) (*lexv2models.Bot, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{statusDeleting, statusNormal},
// 		Target:                    []string{},
// 		Refresh:                   statusBot(ctx, conn, id),
// 		Timeout:                   timeout,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*lexv2models.Bot); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
// func statusBot(ctx context.Context, conn *lexv2models.Client, id string) retry.StateRefreshFunc {
// 	return func() (interface{}, string, error) {
// 		out, err := findBotByID(ctx, conn, id)
// 		if tfresource.NotFound(err) {
// 			return nil, "", nil
// 		}

// 		if err != nil {
// 			return nil, "", err
// 		}

// 		return out, aws.ToString(out.Status), nil
// 	}
// }

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func FindBotByID(ctx context.Context, conn *lexv2models.Client, id string) (*awstypes.Bot, error) {
	in := &lexv2models.GetBotInput{
		BotId: aws.String(id),
	}

	out, err := conn.ListBot(ctx, in)
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

	if out == nil || out.BotId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Bot, nil
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
func flattenDataPrivacy(ctx context.Context, apiObject *awstypes.DataPrivacy) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataPrivacyAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"child_directed": flex.StringValueToFramework(ctx, apiObject.ChildDirected),
	}
	objVal, d := types.ObjectValue(dataPrivacyAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

// TIP: Remember, as mentioned above, expanders take a Terraform data structure
// and return something that you can send to the AWS API. In other words,
// expanders translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
func expandDataPrivacy(ctx context.Context, tfList []dataPrivacyData) (*awstypes.DataPrivacy, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(tfList) == 0 {
		return nil
	}

	dp := tfList[0]
	return &awstypes.DataPrivacy{
		ChildDirected: aws.String(dp.ChildDirected.ValueString()),
	}, diags
}

// TIP: Even when you have a list with max length of 1, this plural function
// works brilliantly. However, if the AWS API takes a structure rather than a
// slice of structures, you will not need it.
func expandMembers(ctx context.Context, tfList []membersData) ([]*awstypes.Members, diag.Diagnostics) {
	var diags diag.Diagnostics

    if len(tfList) == 0 {
        return nil
    }

	mb := tfList[0]
	return &awstypes.DataPrivacy{
		AliasID: aws.String(mb.AliasID.ValueString()),
		AliasName: aws.String(mb.AliasName.ValueString()),
		ID: aws.String(mb.ID.ValueString()),
		Name: aws.String(mb.Name.ValueString()),
		Version: aws.String(mb.Version.ValueString()),
	}, diags
}

func (rd *resourceBotData) refreshFromOutput(ctx context.Context, out *awstypes.Bot) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}
	rd.RoleARN = flex.StringToFramework(ctx, out.RoleARN)
	rd.ID = flex.StringToFramework(ctx, out.BotId)
	rd.Name = flex.StringToFramework(ctx, out.BotName)
	rd.Type = flex.StringToFramework(ctx, out.BotType)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.Type = flex.StringToFramework(ctx, out.Type)
	rd.IdleSessionTTLInSeconds = flex.Int64ToFramework(ctx, out.IdleSessionTTLInSeconds)

	// TIP: Setting a complex type.
	datap, d := flattenDataPrivacy(ctx, out.DataPrivacy)
	diags.Append(d...)
	rd.DataPrivacy = datap
	setTagsOut(ctx, out.Tags)

	return diags
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
type resourceBotData struct {
	DataPrivacy             types.List     `tfsdk:"data_privacy"`
	Description             types.String   `tfsdk:"description"`
	ID                      types.String   `tfsdk:"id"`
	IdleSessionTTLInSeconds types.Int64 `tfsdk:idle_session_ttl_in_seconds`
	Name                    types.String   `tfsdk:"name"`
	Members                 types.List `tfsdk:"members"`
	RoleARN                 types.String `tfsdk:"role_arn"`
	tags                    types.Map `tfsdk:"tags"`
	TestBotAliasTags        types.Map `tfsdk:"test_bot_alias_tags"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
	Type                    types.String   `tfsdk:"type"`
}

type dataPrivacyData struct {
	ChildDirected types.String `tfsdk:"child_directed"`
}

type membersData struct {
	AliasID    types.String `tfsdk:"alias_id"`
	AliasName  types.String `tfsdk:"alias_name"`
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Version    types.String `tfsdk:"version"`

}

var dataPrivacyAttrTypes = map[string]attr.Type{
	"child_directed": types.StringType,
}