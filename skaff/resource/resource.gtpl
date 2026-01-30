// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package {{ .ServicePackage }}

{{ if .IncludeComments -}}
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
{{- end }}

import (
{{- if .IncludeComments }}
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/{{ .SDKPackage }}/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
{{- end }}
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}"
	awstypes "github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
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

{{ end -}}
// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("{{ .ProviderResourceName }}", name="{{ .HumanResourceName }}")
{{- if .IncludeTags }}
// @Tags(identifierAttribute="arn")
{{- end }}
{{ if .IncludeComments }}
// TIP: ==== RESOURCE IDENTITY ====
// Identify which attributes can be used to uniquely identify the resource.
// 
// * If the AWS APIs for the resource take the ARN as an identifier, use
// ARN Identity.
// * If the resource is a singleton (i.e., there is only one instance per region, or account for global resource types), use Singleton Identity.
// * Otherwise, use Parameterized Identity with one or more identity attributes.
//
// For more information about resource identity, see
// https://hashicorp.github.io/terraform-provider-aws/resource-identity/
//
// Keep one of the following sets of annotations as appropriate:
//
// * ARN Identity
// @ArnIdentity
// or
// @ArnIdentity("arn_attribute")
//
// * Singleton Identity
// @SingletonIdentity
//
// * Parameterized Identity
// @IdentityAttribute("id_attribute")
// // @IdentityAttribute("another_id_attribute")
//
// TIP: ==== GENERATED ACCEPTANCE TESTS ====
// Resource Identity and tagging make use of automatically generated acceptance tests.
// For more information about automatically generated acceptance tests, see
// https://hashicorp.github.io/terraform-provider-aws/acc-test-generation/
//
// Some common annotations are included below:
{{- end }}
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }};{{ .SDKPackage }}.Describe{{ .ResourceAWS }}Response")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importIgnore="...;...")
// @Testing(hasNoPreExistingResource=true)
func new{{ .Resource }}Resource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &{{ .ResourceLowerCamel }}Resource{}

	{{ if .IncludeComments -}}
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

type {{ .ResourceLowerCamel }}Resource struct {
	framework.ResourceWithModel[{{ .ResourceLowerCamel }}ResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
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
func (r *{{ .ResourceLowerCamel }}Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			{{- if .IncludeComments }}
			// TIP: ==== "ID" ATTRIBUTE ====
			// When using the Terraform Plugin Framework, there is no required "id" attribute.
			// This is different from the Terraform Plugin SDK.
			//
			// Only include an "id" attribute if the AWS API has an "Id" field, such as "{{ .Resource }}Id"
			{{- end }}
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
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
				// TIP: ==== CUSTOM TYPES ====
				// Use a custom type to identify the model type of the tested object
				{{- end }}
				CustomType: fwtypes.NewListNestedObjectTypeOf[complexArgumentModel](ctx),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *{{ .ResourceLowerCamel }}Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	conn := r.Meta().{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the plan
	{{- end }}
	var plan {{ .ResourceLowerCamel }}ResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	{{ if .IncludeComments -}}
	// TIP: -- 3. Populate a Create input structure
	{{- end }}
	var input {{ .SDKPackage }}.Create{{ .ResourceAWS }}Input
	{{ if .IncludeComments -}}
	// TIP: Using a field name prefix allows mapping fields such as `ID` to `{{ .ResourceAWS }}Id`
	{{- end }}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("{{ .Resource }}")))
	if resp.Diagnostics.HasError() {
		return
	}
	{{ if .IncludeTags -}}
	input.Tags = getTagsIn(ctx)
	{{- end }}

	{{ if .IncludeComments -}}
	// TIP: -- 4. Call the AWS Create function
	{{- end }}
	out, err := conn.Create{{ .ResourceAWS }}(ctx, &input)
	if err != nil {
		{{- if .IncludeComments }}
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		{{- end }}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.{{ .ResourceAWS }} == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	{{ if .IncludeComments -}}
	// TIP: -- 5. Using the output from the create function, set attributes
	{{- end }}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	{{ if .IncludeComments -}}
	// TIP: -- 6. Use a waiter to wait for create to complete
	{{- end }}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = wait{{ .Resource }}Created(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 7. Save the request plan to response state
	{{- end }}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *{{ .ResourceLowerCamel }}Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the state
	{{- end }}
	var state {{ .ResourceLowerCamel }}ResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
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
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 5. Set the arguments and attributes
	{{- end }}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 6. Set the state
	{{- end }}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *{{ .ResourceLowerCamel }}Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	conn := r.Meta().{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the plan
	{{- end }}
	var plan, state {{ .ResourceLowerCamel }}ResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Get the difference between the plan and state, if any
	{{- end }}
	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input {{ .SDKPackage }}.Update{{ .ResourceAWS }}Input
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test")))
		if resp.Diagnostics.HasError() {
			return
		}
		{{ if .IncludeComments }}
		// TIP: -- 4. Call the AWS modify/update function
		{{- end }}
		out, err := conn.Update{{ .ResourceAWS }}(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil || out.{{ .ResourceAWS }} == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
			return
		}
		{{ if .IncludeComments }}
		// TIP: Using the output from the update function, re-set any computed attributes
		{{- end }}
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	{{ if .IncludeComments -}}
	// TIP: -- 5. Use a waiter to wait for update to complete
	{{- end }}
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := wait{{ .Resource }}Updated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	{{ if .IncludeComments -}}
	// TIP: -- 6. Save the request plan to response state
	{{- end }}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *{{ .ResourceLowerCamel }}Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the state
	{{- end }}
	var state {{ .ResourceLowerCamel }}ResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Populate a delete input structure
	{{- end }}
	input := {{ .ServiceLower }}.Delete{{ .ResourceAWS }}Input{
		{{ .ResourceAWS }}Id: state.ID.ValueStringPointer(),
	}
	{{ if .IncludeComments }}
	// TIP: -- 4. Call the AWS delete function
	{{- end }}
	_, err := conn.Delete{{ .ResourceAWS }}(ctx, &input)
	{{- if .IncludeComments }}
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	{{- end }}
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 5. Use a waiter to wait for delete to complete
	{{- end }}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = wait{{ .Resource }}Deleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}
{{ if .IncludeComments }}
// TIP: ==== TERRAFORM IMPORTING ====
// The built-in import function, and Import ID Handler, if any, should handle populating the required
// attributes from the Import ID or Resource Identity.
// In some cases, additional attributes must be set when importing.
// Adding a custom ImportState function can handle those.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/add-resource-identity-support/
{{- end }}
// func (r *{{ .ResourceLowerCamel }}Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	r.WithImportByIdentity.ImportState(ctx, req, resp)
// 
// 	// Set needed attribute values here
// }

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
func wait{{ .Resource }}Created(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string, timeout time.Duration) (*awstypes.{{ .ResourceAWS }}, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   status{{ .Resource }}(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.{{ .ResourceAWS }}); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}
{{ if .IncludeComments }}
// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
{{- end }}
func wait{{ .Resource }}Updated(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string, timeout time.Duration) (*awstypes.{{ .ResourceAWS }}, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   status{{ .Resource }}(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.{{ .ResourceAWS }}); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}
{{ if .IncludeComments }}
// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
{{- end }}
func wait{{ .Resource }}Deleted(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string, timeout time.Duration) (*awstypes.{{ .ResourceAWS }}, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: status{{ .Resource }}(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.{{ .ResourceAWS }}); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
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
func status{{ .Resource }}(conn *{{ .ServiceLower }}.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := find{{ .Resource }}ByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, aws.ToString(out.Status), nil
	}
}
{{ if .IncludeComments }}
// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
{{- end }}
func find{{ .Resource }}ByID(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string) (*awstypes.{{ .ResourceAWS }}, error) {
	input := {{ .ServiceLower }}.Get{{ .ResourceAWS }}Input{
		Id: aws.String(id),
	}

	out, err := conn.Get{{ .ResourceAWS }}(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.{{ .ResourceAWS }} == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.{{ .ResourceAWS }}, nil
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
type {{ .ResourceLowerCamel }}ResourceModel struct {
	framework.WithRegionModel
	ARN             types.String                                          `tfsdk:"arn"`
	ComplexArgument fwtypes.ListNestedObjectValueOf[complexArgumentModel] `tfsdk:"complex_argument"`
	Description     types.String                                          `tfsdk:"description"`
	ID              types.String                                          `tfsdk:"id"`
	Name            types.String                                          `tfsdk:"name"`
	{{- if .IncludeTags }}
	Tags            tftags.Map                                            `tfsdk:"tags"`
	TagsAll         tftags.Map                                            `tfsdk:"tags_all"`
	{{- end }}
	Timeouts        timeouts.Value                                        `tfsdk:"timeouts"`
	Type            types.String                                          `tfsdk:"type"`
}

type complexArgumentModel struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}

{{ if .IncludeComments }}
// TIP: ==== IMPORT ID HANDLER ====
// When a resource type has a Resource Identity with multiple attributes, it needs a handler to
// parse the Import ID used for the `terraform import` command or an `import` block with the `id` parameter.
//
// The parser takes the string value of the Import ID and returns:
// * A string value that is typically ignored. See documentation for more details.
// * A map of the resource attributes derived from the Import ID.
// * An error value if there are parsing errors.
//
// For more information, see https://hashicorp.github.io/terraform-provider-aws/resource-identity/#plugin-framework
{{- end }}
var (
	_ inttypes.ImportIDParser = {{ .ResourceLowerCamel }}ImportID{}
)

type {{ .ResourceLowerCamel }}ImportID struct{}

func ({{ .ResourceLowerCamel }}ImportID) Parse(id string) (string, map[string]string, error) {
	someValue, anotherValue, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id \"%s\" should be in the format <some-value>"+intflex.ResourceIdSeparator+"<another-value>", id)
	}

	result := map[string]string{
		"some-value":    someValue,
		"another-value": anotherValue,
	}

	return id, result, nil
}

{{ if .IncludeComments }}
// TIP: ==== SWEEPERS ====
// When acceptance testing resources, interrupted or failed tests may
// leave behind orphaned resources in an account. To facilitate cleaning
// up lingering resources, each resource implementation should include
// a corresponding "sweeper" function.
//
// The sweeper function lists all resources of a given type and sets the
// appropriate identifers required to delete the resource via the Delete
// method implemented above.
//
// Once the sweeper function is implemented, register it in sweep.go
// as follows:
//
//  awsv2.Register("{{ .ProviderResourceName }}", sweep{{ .Resource }}s)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
{{- end }}
func sweep{{ .Resource }}s(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := {{ .ServiceLower }}.List{{ .ResourceAWS }}sInput{}
	conn := client.{{ .Service }}Client(ctx)
	var sweepResources []sweep.Sweepable

	pages := {{ .ServiceLower }}.NewList{{ .ResourceAWS }}sPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.{{ .ResourceAWS }}s {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(new{{ .Resource }}Resource, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.{{ .ResourceAWS }}Id))),
			)
		}
	}

	return sweepResources, nil
}
