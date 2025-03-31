// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight
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
	// using the services/quicksight/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
// @FrameworkResource("aws_quicksight_account_settings", name="Account Settings")
func newResourceAccountSettings(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAccountSettings{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameAccountSettings = "Account Settings"
)

type resourceAccountSettings struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
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
func (r *resourceAccountSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_name": schema.StringAttribute{
				Computed: true,
			},
			"default_namespace": schema.StringAttribute{
				Optional: true,
				Default:  stringdefault.StaticString("default"),
			},
			"edition": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttributeDeprecatedNoReplacement(),
			"notification_email": schema.StringAttribute{
				Optional: true,
			},
			"public_sharing_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"termination_protection_enabled": schema.BoolAttribute{
				Optional: true,
			},
			// "name": schema.StringAttribute{
			// 	Required: true,
			// 	// TIP: ==== PLAN MODIFIERS ====
			// 	// Plan modifiers were introduced with Plugin-Framework to provide a mechanism
			// 	// for adjusting planned changes prior to apply. The planmodifier subpackage
			// 	// provides built-in modifiers for many common use cases such as 
			// 	// requiring replacement on a value change ("ForceNew: true" in Plugin-SDK 
			// 	// resources).
			// 	//
			// 	// See more:
			// 	// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
			// 	PlanModifiers: []planmodifier.String{
			// 		stringplanmodifier.RequiresReplace(),
			// 	},
			// },
			// "type": schema.StringAttribute{
			// 	Required: true,
			// },
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceAccountSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	conn := r.Meta().QuickSightClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceAccountSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	awsAccountId := r.Meta().AccountID(ctx)
	input := quicksight.UpdateAccountSettingsInput{
		AwsAccountId: &awsAccountId,
	}

	if !plan.DefaultNamespace.IsNull() {
		input.DefaultNamespace = plan.DefaultNamespace.ValueStringPointer()
	}

	if !plan.NotificationEmail.IsNull() {
		input.NotificationEmail = plan.NotificationEmail.ValueStringPointer()
	}

	if !plan.TerminationProtectionEnabled.IsNull() {
		input.TerminationProtectionEnabled = plan.TerminationProtectionEnabled.ValueBool()
	}

	// TIP: Using a field name prefix allows mapping fields such as `ID` to `AccountSettingsId`
	// resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AccountSettings"))...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }


	// TIP: -- 4. Call the AWS Create function
	// out, err := conn.UpdateAccountSettings(ctx, &input)
	// if err != nil {
	// 	// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
	// 	// in error messages at this point.
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameAccountSettings, plan.AccountName.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }
	// if out == nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameAccountSettings, plan.AccountName.String(), nil),
	// 		errors.New("empty output").Error(),
	// 	)
	// 	return
	// }

	// TIP: -- 5. Using the output from the create function, set attributes


	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	// _, err = waitAccountSettingsCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForCreation, ResNameAccountSettings, plan.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }
	output, err := tfresource.RetryGWhen(ctx, createTimeout,
		func() (*quicksight.UpdateAccountSettingsOutput, error) {
			return conn.UpdateAccountSettings(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "You don't have access to this item.\n  The provided credentials couldn't be validated.\n  You might not be authorized to carry out the request.\n  Make sure that your account is authorized to use the Amazon QuickSight service, that your policies have the correct permissions, and that you are using the correct credentials.") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.InternalFailureException](err, "An internal failure occurred.") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "One or more parameters has a value that isn't valid.") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ResourceNotFoundException](err, "One or more resources can't be found."){
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ResourceUnavailableException](err, "This resource is currently unavailable."){
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ThrottlingException](err, "Access is throttled."){
				return true, err
			}
			return false, err
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("creating Quicksight Account", err.Error())
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAccountSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().QuickSightClient(ctx)
	awsAccountID := r.Meta().AccountID(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceAccountSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findAccountSettingsByID(ctx, conn, awsAccountID)
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameAccountSettings, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Set the arguments and attributes
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAccountSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	conn := r.Meta().QuickSightClient(ctx)


	// TIP: -- 2. Fetch the plan
	var plan, state resourceAccountSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the difference between the plan and state, if any
	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		awsAccountID := r.Meta().AccountID(ctx)
		input := quicksight.UpdateAccountSettingsInput{
			AwsAccountId: &awsAccountID,
		}
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.DefaultNamespace.IsNull() {
			input.DefaultNamespace = plan.DefaultNamespace.ValueStringPointer()
		}

		if !plan.NotificationEmail.IsNull() {
			input.NotificationEmail = plan.NotificationEmail.ValueStringPointer()
		}

		if !plan.TerminationProtectionEnabled.IsNull() {
			input.TerminationProtectionEnabled = plan.TerminationProtectionEnabled.ValueBool()
		}

		// TIP: -- 4. Call the AWS modify/update function
		// out, err := conn.UpdateAccountSettings(ctx, &input)
		// if err != nil {
		// 	resp.Diagnostics.AddError(
		// 		create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameAccountSettings, plan.AccountName.String(), err),
		// 		err.Error(),
		// 	)
		// 	return
		// }
		// if out == nil {
		// 	resp.Diagnostics.AddError(
		// 		create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameAccountSettings, plan.AccountName.String(), nil),
		// 		errors.New("empty output").Error(),
		// 	)
		// 	return
		// }

		createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
		output, err := tfresource.RetryGWhen(ctx, createTimeout,
			func() (*quicksight.UpdateAccountSettingsOutput, error) {
				return conn.UpdateAccountSettings(ctx, &input)
			},
			func(err error) (bool, error) {
				if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "You don't have access to this item.\n  The provided credentials couldn't be validated.\n  You might not be authorized to carry out the request.\n  Make sure that your account is authorized to use the Amazon QuickSight service, that your policies have the correct permissions, and that you are using the correct credentials.") {
					return true, err
				}
				if errs.IsAErrorMessageContains[*awstypes.InternalFailureException](err, "An internal failure occurred.") {
					return true, err
				}
				if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "One or more parameters has a value that isn't valid.") {
					return true, err
				}
				if errs.IsAErrorMessageContains[*awstypes.ResourceNotFoundException](err, "One or more resources can't be found."){
					return true, err
				}
				if errs.IsAErrorMessageContains[*awstypes.ResourceUnavailableException](err, "This resource is currently unavailable."){
					return true, err
				}
				if errs.IsAErrorMessageContains[*awstypes.ThrottlingException](err, "Access is throttled."){
					return true, err
				}
				return false, err
			},
		)
		if err != nil {
			resp.Diagnostics.AddError("creating Quicksight Account", err.Error())
			return
		}

		// TIP: Using the output from the update function, re-set any computed attributes
		resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	// updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	// _, err := waitAccountSettingsUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForUpdate, ResNameAccountSettings, plan.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAccountSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	// TIP: -- 2. Fetch the state
	var state resourceAccountSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
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
func (r *resourceAccountSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("account_name"), req, resp)
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
// func waitAccountSettingsCreated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*awstypes.AccountSettings, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{},
// 		Target:                    []string{statusNormal},
// 		Refresh:                   statusAccountSettings(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*quicksight.AccountSettings); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
// func waitAccountSettingsUpdated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*awstypes.AccountSettings, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{statusChangePending},
// 		Target:                    []string{statusUpdated},
// 		Refresh:                   statusAccountSettings(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*quicksight.AccountSettings); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
// func waitAccountSettingsDeleted(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*awstypes.AccountSettings, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{statusDeleting, statusNormal},
// 		Target:                    []string{},
// 		Refresh:                   statusAccountSettings(ctx, conn, id),
// 		Timeout:                   timeout,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*quicksight.AccountSettings); ok {
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
func statusAccountSettings(ctx context.Context, conn *quicksight.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findAccountSettingsByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.AccountName), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findAccountSettingsByID(ctx context.Context, conn *quicksight.Client, id string) (*awstypes.AccountSettings, error) {
	input := quicksight.DescribeAccountSettingsInput{
		AwsAccountId: aws.String(id),
	}

	out, err := conn.DescribeAccountSettings(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.AccountSettings == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.AccountSettings, nil
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
type resourceAccountSettingsModel struct {
	AccountName types.String `tfsdk:"account_name"`
	DefaultNamespace types.String `tfsdk:"default_namespace"`
	Edition types.String `tfsdk:"edition"`
	ID              types.String                                          `tfsdk:"id"`
	NotificationEmail types.String `tfsdk:"notification_email"`
	PublicSharingEnabled types.Bool `tfsdk:"public_sharing_enabled"`
	TerminationProtectionEnabled types.Bool `tfsdk:"termination_protection_enabled"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}