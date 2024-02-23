// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub

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
	// using the services/costoptimizationhub/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"time"

	//"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costoptimizationhub/types"
	//"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	//"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	//"github.com/hashicorp/terraform-plugin-framework/attr"
	//"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	//"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	//"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	//"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	//"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	//"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	//"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
// @FrameworkResource(name="Enrollment")
func newResourceEnrollment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEnrollment{}

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
	ResNameEnrollment = "Enrollment"
)

type resourceEnrollment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceEnrollment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_costoptimizationhub_enrollment"
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
func (r *resourceEnrollment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Description: "Represents the enrollment of the AWS account in Cost Optimization Hub.\n" +
			"The IncludeMemberAccounts attribute is optional and defaults to false. It can be set to true only on Management Accounts. \n" +
			"If set to true on a management account, the member accounts (current and any added later) will also be enrolled into Cost Optimization Hub and cannot unenroll by themselves.",
		MarkdownDescription: "Represents the enrollment status of the AWS account in Cost Optimization Hub.\n" +
			"The `IncludeMemberAccounts` attribute is optional and defaults to `false`. It can be set to `true` only on Management Accounts. \n" +
			"If set to `true` on a management account, the member accounts (current and any added later) will also be enrolled into Cost Optimization Hub and cannot unenroll by themselves.",
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"include_member_accounts": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			//The existence of this Enrollment resource means that the status will be "Active". The non-existence of this resource will mean status will be "Inactive"
			//"status": schema.StringAttribute{
			//	Computed: true,
			//},
			"member_account_discount_visibility": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.MemberAccountDiscountVisibilityAll)),
				Validators: []validator.String{
					stringvalidator.OneOf(getStringArray(awstypes.MemberAccountDiscountVisibilityAll.Values())...),
				},
			},
			"savings_estimation_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.SavingsEstimationModeBeforeDiscounts)),
				Validators: []validator.String{
					stringvalidator.OneOf(getStringArray(awstypes.SavingsEstimationModeBeforeDiscounts.Values())...),
					//stringvalidator.OneOf([]string{string(awstypes.SavingsEstimationModeBeforeDiscounts), string(awstypes.SavingsEstimationModeAfterDiscounts)}...),
				},
			},
		},
	}
}

func (r *resourceEnrollment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	conn := r.Meta().CostOptimizationHubClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceEnrollmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create two input structures as the single Terraform resource will have to invoke two API calls

	//Input for UpdateEnrollmentStatus
	ues_in := &costoptimizationhub.UpdateEnrollmentStatusInput{
		Status: awstypes.EnrollmentStatus("Active"),
	}

	up_in := &costoptimizationhub.UpdatePreferencesInput{}

	if !plan.IncludeMemberAccounts.IsNull() {
		ues_in.IncludeMemberAccounts = plan.IncludeMemberAccounts.ValueBoolPointer()
	}

	if !plan.MemberAccountDiscountVisibility.IsNull() {
		up_in.MemberAccountDiscountVisibility = awstypes.MemberAccountDiscountVisibility(plan.MemberAccountDiscountVisibility.ValueString())
	}

	if !plan.SavingsEstimationMode.IsNull() {
		up_in.SavingsEstimationMode = awstypes.SavingsEstimationMode(plan.SavingsEstimationMode.ValueString())
	}

	// TIP: -- 4. Call the AWS create function
	ues_out, ues_err := conn.UpdateEnrollmentStatus(ctx, ues_in)
	if ues_err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, "UpdateEnrollmentStatus", ues_err),
			ues_err.Error(),
		)
		return
	}
	if ues_out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, "UpdateEnrollmentStatus", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	up_out, up_err := conn.UpdatePreferences(ctx, up_in)
	if up_err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, "UpdatePreferences", up_err),
			up_err.Error(),
		)
		return
	}
	if up_out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, "UpdatePreferences", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// TIP: -- 5. Using the output from the create functions, set the attributes
	plan.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID)
	//plan.Status = flex.StringToFramework(ctx, ues_out.Status)
	//The UpdateEnrollmentStatus function does not return the include-member-accounts value. Hence we cannot use that directly.
	//plan.IncludeMemberAccounts = flex.BoolToFramework(ctx, plan.IncludeMemberAccounts.ValueBoolPointer())
	plan.MemberAccountDiscountVisibility = flex.StringValueToFramework(ctx, up_out.MemberAccountDiscountVisibility)
	plan.SavingsEstimationMode = flex.StringValueToFramework(ctx, up_out.SavingsEstimationMode)

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEnrollment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().CostOptimizationHubClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceEnrollmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	les_in := &costoptimizationhub.ListEnrollmentStatusesInput{
		IncludeOrganizationInfo: false, //Pass in false to get only this account's status
	}

	les_out, les_err := conn.ListEnrollmentStatuses(ctx, les_in)
	if les_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionSetting, ResNameEnrollment, state.ID.String(), les_err),
			les_err.Error(),
		)
		return
	}
	//For this Enrollment resource, The non-existence of this resource will mean status will be "Inactive"
	//So if that is the case, remove the resource from state
	if les_out.Items[0].Status != "Active" {
		resp.State.RemoveResource(ctx)
		return
	}

	gp_in := &costoptimizationhub.GetPreferencesInput{}

	gp_out, gp_err := conn.GetPreferences(ctx, gp_in)
	if gp_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionSetting, ResNameEnrollment, state.ID.String(), gp_err),
			gp_err.Error(),
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
	state.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID)
	//state.Status = flex.StringValueToFramework(ctx, les_out.Items[0].Status)
	state.IncludeMemberAccounts = flex.BoolToFramework(ctx, les_out.IncludeMemberAccounts)
	state.MemberAccountDiscountVisibility = flex.StringValueToFramework(ctx, gp_out.MemberAccountDiscountVisibility)
	state.SavingsEstimationMode = flex.StringValueToFramework(ctx, gp_out.SavingsEstimationMode)

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEnrollment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	conn := r.Meta().CostOptimizationHubClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceEnrollmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.IncludeMemberAccounts.Equal(state.IncludeMemberAccounts) {
		//Input for UpdateEnrollmentStatus
		ues_in := &costoptimizationhub.UpdateEnrollmentStatusInput{
			//Status is a mandatory parameter. Hence has to be passed in.
			Status:                awstypes.EnrollmentStatus("Active"),
			IncludeMemberAccounts: plan.IncludeMemberAccounts.ValueBoolPointer(),
		}

		ues_out, ues_err := conn.UpdateEnrollmentStatus(ctx, ues_in)
		if ues_err != nil {
			// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
			// in error messages at this point.
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, plan.ID.String(), ues_err),
				ues_err.Error(),
			)
			return
		}
		if ues_out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		//The UpdateEnrollmentStatus function does not return the include-member-accounts value. Hence we cannot use that directly.
		plan.ID = state.ID
		//plan.Status = state.Status
	}

	// Save the request plan to response state so that even if the next API fails, this update is recorded in the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	// TIP: -- 3. Populate a modify input structure and check for changes
	if !plan.MemberAccountDiscountVisibility.Equal(state.MemberAccountDiscountVisibility) ||
		!plan.SavingsEstimationMode.Equal(state.SavingsEstimationMode) {

		up_in := &costoptimizationhub.UpdatePreferencesInput{}

		if !plan.MemberAccountDiscountVisibility.IsNull() {
			up_in.MemberAccountDiscountVisibility = awstypes.MemberAccountDiscountVisibility(plan.MemberAccountDiscountVisibility.ValueString())
		}

		if !plan.SavingsEstimationMode.IsNull() {
			up_in.SavingsEstimationMode = awstypes.SavingsEstimationMode(plan.SavingsEstimationMode.ValueString())
		}

		// TIP: -- 4. Call the AWS create function
		up_out, up_err := conn.UpdatePreferences(ctx, up_in)
		if up_err != nil {
			// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
			// in error messages at this point.
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, plan.ID.String(), up_err),
				up_err.Error(),
			)
			return
		}
		if up_out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		// TIP: -- 5. Using the output from the create functions, set the attributes
		//The UpdateEnrollmentStatus function does not return the include-member-accounts value. Hence we cannot use that directly.
		plan.ID = state.ID
		//plan.Status = state.Status
		plan.MemberAccountDiscountVisibility = flex.StringValueToFramework(ctx, up_out.MemberAccountDiscountVisibility)
		plan.SavingsEstimationMode = flex.StringValueToFramework(ctx, up_out.SavingsEstimationMode)
	}

	// Save the request plan to response state so that even if the next API fails, this update is recorded in the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEnrollment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().CostOptimizationHubClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceEnrollmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	ues_in := &costoptimizationhub.UpdateEnrollmentStatusInput{
		Status: awstypes.EnrollmentStatus("Inactive"),
	}

	// TIP: -- 4. Call the AWS delete function
	_, ues_err := conn.UpdateEnrollmentStatus(ctx, ues_in)
	if ues_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionDeleting, ResNameEnrollment, state.ID.String(), ues_err),
			ues_err.Error(),
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
func (r *resourceEnrollment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to convert an array of types from the Go SDK to an array of strings.
func getStringArray[T awstypes.MemberAccountDiscountVisibility | awstypes.SavingsEstimationMode](attrValArray []T) []string {
	results := make([]string, 0, len(attrValArray))
	for _, value := range attrValArray {
		results = append(results, string(value))
	}
	return results
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
type resourceEnrollmentData struct {
	ID types.String `tfsdk:"id"`
	//Status                          types.String `tfsdk:"status"`
	IncludeMemberAccounts           types.Bool   `tfsdk:"include_member_accounts"`
	MemberAccountDiscountVisibility types.String `tfsdk:"member_account_discount_visibility"`
	SavingsEstimationMode           types.String `tfsdk:"savings_estimation_mode"`
}
