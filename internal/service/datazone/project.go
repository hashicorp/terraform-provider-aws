// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

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
	// using the services/datazone/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	//"github.com/aws/smithy-go/middleware"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	//"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	//"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	//"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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
// @FrameworkResource("aws_datazone_project", name="Project")
func newResourceProject(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProject{}

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
	ResNameProject = "Project"
)

type resourceProject struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceProject) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_datazone_project"
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
func (r *resourceProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(2048),
				},
			},
			"domain_identifier": schema.StringAttribute{
				CustomType: fwtypes.RegexpType,
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^dzd[-_][a-zA-Z0-9_-]{1,36}$`), "must conform to: ^dzd[-_][a-zA-Z0-9_-]{1,36}$ "),
				},
			},
			"glossary_terms": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[types.String](ctx),
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 20)),
				},
			},
			names.AttrName: schema.StringAttribute{
				CustomType: fwtypes.RegexpType,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[\w -]+$`), "must conform to: ^[\\w -]+$ "),
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"failure_reasons": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[awstypes.ProjectDeletionError](ctx),
				Computed:   true,
			},
			"last_updated_at": schema.StringAttribute{
				Computed: true,
			},
			"project_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProjectStatus](),
				Computed:   true,
			},
			"result_metadata": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[middleware.Metadata](ctx),
				Computed:   true,
			},
		},
	}
}

func (r *resourceProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	conn := r.Meta().DataZoneClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceProjectData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a create input structure
	in := &datazone.CreateProjectInput{
		Name: aws.String(plan.Name.ValueString()),
	}
	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}
	if !plan.DomainIdentifier.IsNull() {
		in.DomainIdentifier = aws.String(plan.DomainIdentifier.ValueString())
	}
	if !plan.GlossaryTerms.IsNull() {
		// do later
	}

	// TIP: -- 4. Call the AWS create function
	out, err := conn.CreateProject(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameProject, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || !(out.FailureReasons == nil) {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameProject, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// TIP: -- 5. Using the output from the create function, set the minimum attributes

	// can i autoflex this?
	resp.Diagnostics.Append(flex.Flatten(ctx, &out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitProjectCreated(ctx, conn, plan.DomainIdentifier.ValueString(),plan.ID.ValueString(),createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForCreation, ResNameProject, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().DataZoneClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceProjectData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	var input datazone.GetProjectInput
	if !state.DomainIdentifier.IsNull(){
		input.DomainIdentifier = state.DomainIdentifier.ValueStringPointer()
	}
	if !state.ID.IsNull(){
		input.Identifier = state.ID.ValueStringPointer()
	}
	out, err := conn.GetProject(ctx,&input)
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameProject, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, &out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	conn := r.Meta().DataZoneClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceProjectData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...) /// get the plan
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // get the state
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.DomainIdentifier != state.DomainIdentifier { // Do i have to error check this
		return // add error here
	}

	// TIP: -- 3. Populate a modify input structure and check for changes
	/*/
	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.ComplexArgument.Equal(state.ComplexArgument) ||
		!plan.Type.Equal(state.Type) {
*/

	in := &datazone.UpdateProjectInput{
		DomainIdentifier: aws.String(plan.DomainIdentifier.ValueString()), // This can't change
		Identifier: aws.String(plan.ID.ValueString()),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.GlossaryTerms.IsNull() {
		in.GlossaryTerms = aws.String(plan.Description.ValueString()) // don't know how to do 
	}

	if !plan.Name.IsNull() {
		in.Name = aws.String(plan.Name.ValueString())
	}



	// TIP: -- 4. Call the AWS modify/update function
	out, err := conn.UpdateProject(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameProject, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || !(out.FailureReasons == nil) {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameProject, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, &out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err = waitProjectUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForUpdate, ResNameProject, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().DataZoneClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceProjectData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	in := &datazone.DeleteProjectInput{
		DomainIdentifier: aws.String((state.DomainIdentifier.ValueString())),
		Identifier:       aws.String((state.ID.ValueString())),
	}
	if !state.SkipDeletionCheck.IsNull() {
		in.SkipDeletionCheck = state.SkipDeletionCheck.ValueBoolPointer()
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteProject(ctx, in)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameProject, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitProjectDeleted(ctx, conn,state.DomainIdentifier.ValueString(),state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForDeletion, ResNameProject, state.ID.String(), err),
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
func (r *resourceProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
const (
	statusChangePending = "PENDING" // I changes these to be more inline with aws types
	statusUpdated       = "UPDATED" // ^
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
func waitProjectCreated(ctx context.Context, conn *datazone.Client, domain string, identifier string, timeout time.Duration) (*datazone.GetProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{string(awstypes.ProjectStatusActive)},
		Refresh:                   statusProject(ctx, conn, domain,identifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetProjectOutput); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitProjectUpdated(ctx context.Context, conn *datazone.Client, domain string, identifier string, timeout time.Duration) (*datazone.GetProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusProject(ctx, conn, domain,identifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetProjectOutput); ok { // not too sure if this is correct
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitProjectDeleted(ctx context.Context, conn *datazone.Client, domain string, identifier string, timeout time.Duration) (*datazone.GetProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.ProjectStatusDeleting), string(awstypes.ProjectStatusActive)},
		Target:  []string{},
		Refresh: statusProject(ctx, conn, domain,identifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetProjectOutput); ok { // not too sure if this is correct
		return out, err
	}

	return nil, err
}

func statusProject(ctx context.Context, conn *datazone.Client, domain string, identifier string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findProjectByID(ctx, conn, domain,identifier)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.ProjectStatus)), nil
	}
}

// TIP: ==== FINDERS ====
func findProjectByID(ctx context.Context, conn *datazone.Client, domain string, identifier string) (*datazone.GetProjectOutput, error) {
	in := &datazone.GetProjectInput{
		DomainIdentifier: aws.String(domain),
		Identifier: aws.String(identifier),
	}

	out, err := conn.GetProject(ctx, in) // add finder here later
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || !(out.FailureReasons == nil) {
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
func flattenComplexArgument(ctx context.Context, apiObject *awstypes.ComplexArgument) (types.List, diag.Diagnostics) {
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
/*/
func flattenComplexArguments(ctx context.Context, apiObjects []*awstypes.ComplexArgument) (types.List, diag.Diagnostics) {
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
*/
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
type resourceProjectData struct {
	Description       types.String                                                   `tfsdk:"description"`
	DomainIdentifier  types.String                                                   `tfsdk:"domain_identifier"`
	GlossaryTerms     fwtypes.ListNestedObjectValueOf[types.String]                  `tfsdk:"glossary_terms"`
	Name              types.String                                                   `tfsdk:"name"`
	CreatedBy         types.String                                                   `tfsdk:"created_by"`
	ID                types.String                                                   `tfsdk:"id"`
	CreatedAt         timetypes.RFC3339                                              `tfsdk:"created_at"`
	FailureReasons    fwtypes.ListNestedObjectValueOf[awstypes.ProjectDeletionError] `tfsdk:"failure_reasons"`
	ProjectStatus     fwtypes.StringEnum[awstypes.ProjectStatus]                     `tfsdk:"project_status"`
	ResultMetadata    middleware.Metadata                                            `tfsdk:"result_metadata"`
	Timeouts          timeouts.Value                                                 `tfsdk:"timeouts"`
	SkipDeletionCheck types.Bool                                                     `tfsdk:"skip_deletion_check"`
}

// DomainId             types.String                                  `tfsdk:"domain_id"` // is this the same thing as domainIdent

type complexArgumentData struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}

var complexArgumentAttrTypes = map[string]attr.Type{
	"nested_required": types.StringType,
	"nested_optional": types.StringType,
}
