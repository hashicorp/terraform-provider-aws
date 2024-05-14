// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package {{ .ServicePackage }}

{{- if .IncludeComments }}
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
// it.{{- end }}

import (
{{- if .IncludeComments }}
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
{{- end }}
{{- if and .IncludeComments .AWSGoSDKV2 }}
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/{{ .ServicePackage }}/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
{{- end }}
	"context"
	"errors"
	"time"
{{ if .AWSGoSDKV2 }}
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/{{ .ServicePackage }}"
	awstypes "github.com/aws/aws-sdk-go-v2/service/{{ .ServicePackage }}/types"
{{- else }}
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/{{ .ServicePackage }}"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
{{- end }}
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
{{- if .AWSGoSDKV2 }}
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
{{- end }}
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
{{- if .IncludeTags }}
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
{{- end }}
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)
{{- if .IncludeComments }}
// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)
{{- end }}

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_{{ .ServicePackage }}_{{ .ResourceSnake }}", name="{{ .HumanResourceName }}")
{{- if .IncludeTags }}
// @Tags(identifierAttribute="arn")
{{- end }}
func newResource{{ .Resource }}(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resource{{ .Resource }}{}
	{{ if .IncludeComments }}
	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	{{- end }}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResName{{ .Resource }} = "{{ .HumanResourceName }}"
)

type resource{{ .Resource }} struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resource{{ .Resource }}) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_{{ .ServicePackage }}_{{ .ResourceSnake }}"
}
{{ if .IncludeComments }}
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
{{- end }}
func (r *resource{{ .Resource }}) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				{{- if .IncludeComments }}
				// TIP: ==== PLAN MODIFIERS ====
				// Plan modifiers were introduced with Plugin-Framework to provide a mechanism
				// for adjusting planned changes prior to apply. The planmodifier subpackage
				// provides built-in modifiers for many common use cases such as 
				// requiring replacement on a value change ("ForceNew: true" in Plugin-SDK 
				// resources).
				//
				// See more:
				// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
				{{- end }}
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			{{- if .IncludeTags }}
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			{{- end }}
			"type": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"complex_argument": schema.ListNestedBlock{
				{{- if .IncludeComments }}
				// TIP: ==== LIST VALIDATORS ====
				// List and set validators take the place of MaxItems and MinItems in 
				// Plugin-Framework based resources. Use listvalidator.SizeAtLeast(1) to
				// make a nested object required. Similar to Plugin-SDK, complex objects 
				// can be represented as lists or sets with listvalidator.SizeAtMost(1).
				//
				// For a complete mapping of Plugin-SDK to Plugin-Framework schema fields, 
				// see:
				// https://developer.hashicorp.com/terraform/plugin/framework/migrating/attributes-blocks/blocks
				{{- end }}
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"nested_required": schema.StringAttribute{
							Required: true,
						},
						"nested_computed": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
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

func (r *resource{{ .Resource }}) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	{{- if .IncludeComments }}
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
	{{- end }}
	{{- if .IncludeComments }}

	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := r.Meta().{{ .Service }}{{ if .AWSGoSDKV2 }}Client(ctx){{ else }}Conn(ctx){{ end }}
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the plan
	{{- end }}
	var plan resource{{ .Resource }}Data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Populate a create input structure
	{{- end }}
	in := &{{ .ServicePackage }}.Create{{ .Resource }}Input{
		{{- if .IncludeComments }}
		// TIP: Mandatory or fields that will always be present can be set when
		// you create the Input structure. (Replace these with real fields.)
		{{- end }}
		{{ .Resource }}Name: aws.String(plan.Name.ValueString()),
		{{ .Resource }}Type: aws.String(plan.Type.ValueString()),
		{{- if .IncludeTags }}
		Tags:             getTagsIn(ctx),
		{{- end }}
	}

	if !plan.Description.IsNull() {
		{{- if .IncludeComments }}
		// TIP: Optional fields should be set based on whether or not they are
		// used.
		{{- end }}
		in.Description = aws.String(plan.Description.ValueString())
	}
	if !plan.ComplexArgument.IsNull() {
		{{- if .IncludeComments }}
		// TIP: Use an expander to assign a complex argument. The elements must be
		// deserialized into the appropriate struct before being passed to the expander.
		{{- end }}
		var tfList []complexArgumentData
		resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.ComplexArgument = expandComplexArgument(tfList)
	}
	{{ if .IncludeComments }}
	// TIP: -- 4. Call the AWS create function
	{{- end }}
	{{- if .AWSGoSDKV2 }}
	out, err := conn.Create{{ .Resource }}(ctx, in)
	{{- else }}
	out, err := conn.Create{{ .Resource }}WithContext(ctx, in)
	{{- end }}
	if err != nil {
		{{- if .IncludeComments }}
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		{{- end }}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionCreating, ResName{{ .Resource }}, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.{{ .Resource }} == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionCreating, ResName{{ .Resource }}, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 5. Using the output from the create function, set the minimum attributes
	{{- end }}
	plan.ARN = flex.StringToFramework(ctx, out.{{ .Resource }}.Arn)
	plan.ID = flex.StringToFramework(ctx, out.{{ .Resource }}.{{ .Resource }}Id)
	{{ if .IncludeComments }}
	// TIP: -- 6. Use a waiter to wait for create to complete
	{{- end }}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = wait{{ .Resource }}Created(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionWaitingForCreation, ResName{{ .Resource }}, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 7. Save the request plan to response state
	{{- end }}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resource{{ .Resource }}) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	{{- if .IncludeComments }}
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
	{{- end }}
	{{- if .IncludeComments }}

	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := r.Meta().{{ .Service }}{{ if .AWSGoSDKV2 }}Client(ctx){{ else }}Conn(ctx){{ end }}
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the state
	{{- end }}
	var state resource{{ .Resource }}Data
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	{{- end }}
	out, err := find{{ .Resource }}ByID(ctx, conn, state.ID.ValueString())
	{{- if .IncludeComments }}
	// TIP: -- 4. Remove resource from state if it is not found
	{{- end }}
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionSetting, ResName{{ .Resource }}, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	{{ if .IncludeComments }}
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
	{{- end }}
	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.ID = flex.StringToFramework(ctx, out.{{ .Resource }}Id)
	state.Name = flex.StringToFramework(ctx, out.{{ .Resource }}Name)
	state.Type = flex.StringToFramework(ctx, out.{{ .Resource }}Type)
	{{ if .IncludeComments }}
	// TIP: Setting a complex type.
	{{- end }}
	complexArgument, d := flattenComplexArgument(ctx, out.ComplexArgument)
	resp.Diagnostics.Append(d...)
	state.ComplexArgument = complexArgument
	{{ if .IncludeComments }}
	// TIP: -- 6. Set the state
	{{- end }}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resource{{ .Resource }}) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	{{- if .IncludeComments }}
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
	{{- end }}

	{{- if .IncludeComments }}
	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := r.Meta().{{ .Service }}{{ if .AWSGoSDKV2 }}Client(ctx){{ else }}Conn(ctx){{ end }}
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the plan
	{{- end }}
	var plan, state resource{{ .Resource }}Data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Populate a modify input structure and check for changes
	{{- end }}
	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.ComplexArgument.Equal(state.ComplexArgument) ||
		!plan.Type.Equal(state.Type) {

		in := &{{ .ServicePackage }}.Update{{ .Resource }}Input{
			{{- if .IncludeComments }}
			// TIP: Mandatory or fields that will always be present can be set when
			// you create the Input structure. (Replace these with real fields.)
			{{- end }}
			{{ .Resource }}Id:   aws.String(plan.ID.ValueString()),
			{{ .Resource }}Name: aws.String(plan.Name.ValueString()),
			{{ .Resource }}Type: aws.String(plan.Type.ValueString()),
		}

		if !plan.Description.IsNull() {
			{{- if .IncludeComments }}
			// TIP: Optional fields should be set based on whether or not they are
			// used.
			{{- end }}
			in.Description = aws.String(plan.Description.ValueString())
		}
		if !plan.ComplexArgument.IsNull() {
			{{- if .IncludeComments }}
			// TIP: Use an expander to assign a complex argument. The elements must be
			// deserialized into the appropriate struct before being passed to the expander.
			{{- end }}
			var tfList []complexArgumentData
			resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.ComplexArgument = expandComplexArgument(tfList)
		}
		{{ if .IncludeComments }}
		// TIP: -- 4. Call the AWS modify/update function
		{{- end }}
		{{- if .AWSGoSDKV2 }}
		out, err := conn.Update{{ .Resource }}(ctx, in)
		{{- else }}
		out, err := conn.Update{{ .Resource }}WithContext(ctx, in)
		{{- end }}
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionUpdating, ResName{{ .Resource }}, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.{{ .Resource }} == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionUpdating, ResName{{ .Resource }}, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		{{ if .IncludeComments }}
		// TIP: Using the output from the update function, re-set any computed attributes
		{{- end }}
		plan.ARN = flex.StringToFramework(ctx, out.{{ .Resource }}.Arn)
		plan.ID = flex.StringToFramework(ctx, out.{{ .Resource }}.{{ .Resource }}Id)
	}

	{{ if .IncludeComments }}
	// TIP: -- 5. Use a waiter to wait for update to complete
	{{- end }}
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := wait{{ .Resource }}Updated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionWaitingForUpdate, ResName{{ .Resource }}, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	{{ if .IncludeComments }}
	// TIP: -- 6. Save the request plan to response state
	{{- end }}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resource{{ .Resource }}) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	{{- if .IncludeComments }}
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
	{{- end }}

	{{- if .IncludeComments }}
	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := r.Meta().{{ .Service }}{{ if .AWSGoSDKV2 }}Client(ctx){{ else }}Conn(ctx){{ end }}
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the state
	{{- end }}
	var state resource{{ .Resource }}Data
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Populate a delete input structure
	{{- end }}
	in := &{{ .ServiceLower }}.Delete{{ .Resource }}Input{
		{{ .Resource }}Id: aws.String(state.ID.ValueString()),
	}
	{{ if .IncludeComments }}
	// TIP: -- 4. Call the AWS delete function
	{{- end }}
	{{- if .AWSGoSDKV2 }}
	_, err := conn.Delete{{ .Resource }}(ctx, in)
	{{- else }}
	_, err := conn.Delete{{ .Resource }}WithContext(ctx, in)
	{{- end }}
	{{- if .IncludeComments }}
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	{{- end }}
	if err != nil {
		{{- if .AWSGoSDKV2 }}
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		{{- else }}
		if tfawserr.ErrCodeEquals(err, {{ .ServiceLower }}.ErrCodeResourceNotFoundException) {
			return 
		}
		{{- end }}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionDeleting, ResName{{ .Resource }}, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 5. Use a waiter to wait for delete to complete
	{{- end }}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = wait{{ .Resource }}Deleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionWaitingForDeletion, ResName{{ .Resource }}, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}
{{ if .IncludeComments }}
// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
{{- end }}
func (r *resource{{ .Resource }}) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

{{- if .IncludeTags }}
func (r *resource{{ .Resource }}) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}
{{- end }}

{{ if .IncludeComments }}
// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
{{- end }}
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)
{{ if .IncludeComments }}
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
{{- end }}
{{- if .AWSGoSDKV2 }}
func wait{{ .Resource }}Created(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string, timeout time.Duration) (*awstypes.{{ .Resource }}, error) {
{{- else }}
func wait{{ .Resource }}Created(ctx context.Context, conn *{{ .ServiceLower }}.{{ .Service }}, id string, timeout time.Duration) (*{{ .ServiceLower }}.{{ .Resource }}, error) {
{{- end }}
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   status{{ .Resource }}(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*{{ .ServiceLower }}.{{ .Resource }}); ok {
		return out, err
	}

	return nil, err
}
{{ if .IncludeComments }}
// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
{{- end }}
{{- if .AWSGoSDKV2 }}
func wait{{ .Resource }}Updated(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string, timeout time.Duration) (*awstypes.{{ .Resource }}, error) {
{{- else }}
func wait{{ .Resource }}Updated(ctx context.Context, conn *{{ .ServiceLower }}.{{ .Service }}, id string, timeout time.Duration) (*{{ .ServiceLower }}.{{ .Resource }}, error) {
{{- end }}
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   status{{ .Resource }}(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*{{ .ServiceLower }}.{{ .Resource }}); ok {
		return out, err
	}

	return nil, err
}
{{ if .IncludeComments }}
// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
{{- end }}
{{- if .AWSGoSDKV2 }}
func wait{{ .Resource }}Deleted(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string, timeout time.Duration) (*awstypes.{{ .Resource }}, error) {
{{- else }}
func wait{{ .Resource }}Deleted(ctx context.Context, conn *{{ .ServiceLower }}.{{ .Service }}, id string, timeout time.Duration) (*{{ .ServiceLower }}.{{ .Resource }}, error) {
{{- end }}
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusDeleting, statusNormal},
		Target:                    []string{},
		Refresh:                   status{{ .Resource }}(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*{{ .ServiceLower }}.{{ .Resource }}); ok {
		return out, err
	}

	return nil, err
}
{{ if .IncludeComments }}
// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
{{- end }}
{{- if .AWSGoSDKV2 }}
func status{{ .Resource }}(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string) retry.StateRefreshFunc {
{{- else }}
func status{{ .Resource }}(ctx context.Context, conn *{{ .ServiceLower }}.{{ .Service }}, id string) retry.StateRefreshFunc {
{{- end }}
	return func() (interface{}, string, error) {
		out, err := find{{ .Resource }}ByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
{{ if .AWSGoSDKV2 }}
		return out, aws.ToString(out.Status), nil
{{- else }}
		return out, aws.StringValue(out.Status), nil
{{- end }}
	}
}
{{ if .IncludeComments }}
// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
{{- end }}
{{- if .AWSGoSDKV2 }}
func find{{ .Resource }}ByID(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string) (*awstypes.{{ .Resource }}, error) {
{{- else }}
func find{{ .Resource }}ByID(ctx context.Context, conn *{{ .ServiceLower }}.{{ .Service }}, id string) (*{{ .ServiceLower }}.{{ .Resource }}, error) {
{{- end }}
	in := &{{ .ServiceLower }}.Get{{ .Resource }}Input{
		Id: aws.String(id),
	}
	{{ if .AWSGoSDKV2 }}
	out, err := conn.Get{{ .Resource }}(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}
	{{- else }}
	out, err := conn.Get{{ .Resource }}WithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, {{ .ServiceLower }}.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}
	{{- end }}

	if out == nil || out.{{ .Resource }} == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.{{ .Resource }}, nil
}
{{ if .IncludeComments }}
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
{{- end }}
{{- if .AWSGoSDKV2 }}
func flattenComplexArgument(ctx context.Context, apiObject *awstypes.ComplexArgument) (types.List, diag.Diagnostics) {
{{- else }}
func flattenComplexArgument(ctx context.Context, apiObject *{{ .ServiceLower }}.ComplexArgument) (types.List, diag.Diagnostics) {
{{- end }}
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
{{ if .IncludeComments }}
// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
{{- end }}
{{- if .AWSGoSDKV2 }}
func flattenComplexArguments(ctx context.Context, apiObjects []*awstypes.ComplexArgument) (types.List, diag.Diagnostics) {
{{- else }}
func flattenComplexArguments(ctx context.Context, apiObjects []*{{ .ServiceLower }}.ComplexArgument) (types.List, diag.Diagnostics) {
{{- end }}
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
{{ if .IncludeComments }}
// TIP: Remember, as mentioned above, expanders take a Terraform data structure
// and return something that you can send to the AWS API. In other words,
// expanders translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
{{- end }}
{{- if .AWSGoSDKV2 }}
func expandComplexArgument(tfList []complexArgumentData) *awstypes.ComplexArgument {
{{- else }}
func expandComplexArgument(tfList []complexArgumentData) *{{ .ServiceLower }}.ComplexArgument {
{{- end }}
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
{{- if .AWSGoSDKV2 }}
	apiObject := &awstypes.ComplexArgument{
{{- else }}
	apiObject := &{{ .ServiceLower }}.ComplexArgument{
{{- end }}
		NestedRequired: aws.String(tfObj.NestedRequired.ValueString()),
	}
	if !tfObj.NestedOptional.IsNull() {
		apiObject.NestedOptional = aws.String(tfObj.NestedOptional.ValueString())
	}

	return apiObject
}
{{ if .IncludeComments }}
// TIP: Even when you have a list with max length of 1, this plural function
// works brilliantly. However, if the AWS API takes a structure rather than a
// slice of structures, you will not need it.
{{- end }}
func expandComplexArguments(tfList []complexArgumentData) []*{{ .ServiceLower }}.ComplexArgument {
	{{- if .IncludeComments }}
	// TIP: The AWS API can be picky about whether you send a nil or zero-
	// length for an argument that should be cleared. For example, in some
	// cases, if you send a nil value, the AWS API interprets that as "make no
	// changes" when what you want to say is "remove everything." Sometimes
	// using a zero-length list will cause an error.
	//
	// As a result, here are two options. Usually, option 1, nil, will work as
	// expected, clearing the field. But, test going from something to nothing
	// to make sure it works. If not, try the second option.
	{{- end }}
	{{- if .IncludeComments }}
	// TIP: Option 1: Returning nil for zero-length list
	{{- end }}
    if len(tfList) == 0 {
        return nil
    }

{{- if .AWSGoSDKV2 }}
    var apiObject []*awstypes.ComplexArgument
{{- else }}
    var apiObject []*{{ .ServiceLower }}.ComplexArgument
{{- end }}
	{{- if .IncludeComments }}
	// TIP: Option 2: Return zero-length list for zero-length list. If option 1 does
	// not work, after testing going from something to nothing (if that is
	// possible), uncomment out the next line and remove option 1.
	//
	{{- end }}
	// apiObject := make([]*{{ .ServiceLower }}.ComplexArgument, 0)

	for _, tfObj := range tfList {
		item := &{{ .ServiceLower }}.ComplexArgument{
			NestedRequired: aws.String(tfObj.NestedRequired.ValueString()),
		}
		if !tfObj.NestedOptional.IsNull() {
			item.NestedOptional = aws.String(tfObj.NestedOptional.ValueString())
		}

		apiObject = append(apiObject, item)
	}

	return apiObject
}
{{ if .IncludeComments }}
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
{{- end }}
type resource{{ .Resource }}Data struct {
	ARN             types.String   `tfsdk:"arn"`
	ComplexArgument types.List     `tfsdk:"complex_argument"`
	Description     types.String   `tfsdk:"description"`
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	{{- if .IncludeTags }}
	Tags            types.Map      `tfsdk:"tags"`
	TagsAll         types.Map      `tfsdk:"tags_all"`
	{{- end }}
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
