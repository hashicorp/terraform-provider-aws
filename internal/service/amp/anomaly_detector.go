// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package amp

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
	// using the services/amp/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
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
// @FrameworkResource("aws_amp_anomaly_detector", name="Anomaly Detector")

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
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amp;amp.DescribeAnomalyDetectorResponse")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importIgnore="...;...")
// @Testing(hasNoPreExistingResource=true)
func newAnomalyDetectorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &anomalyDetectorResource{}

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
	ResNameAnomalyDetector = "Anomaly Detector"
)

type anomalyDetectorResource struct {
	framework.ResourceWithModel[anomalyDetectorResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
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
func (r *anomalyDetectorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAlias: schema.StringAttribute{
				Required: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				Computed: true,
			},
			"evaluation_interval_in_seconds": schema.Int32Attribute{
				Optional: true,
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"labels": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"workspace_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[anomalyDetectorConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"random_cut_forest": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[randomCutForestConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"query": schema.StringAttribute{
										Required: true,
									},
									"sample_size": schema.Int32Attribute{
										Optional: true,
										Computed: true,
									},
									"shingle_size": schema.Int32Attribute{
										Optional: true,
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"ignore_near_expected_from_above": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[ignoreNearExpectedModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"amount": schema.Float64Attribute{
													Optional: true,
												},
												"ratio": schema.Float64Attribute{
													Optional: true,
												},
											},
										},
									},
									"ignore_near_expected_from_below": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[ignoreNearExpectedModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"amount": schema.Float64Attribute{
													Optional: true,
												},
												"ratio": schema.Float64Attribute{
													Optional: true,
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
			"missing_data_action": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[anomalyDetectorMissingDataActionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"mark_as_anomaly": schema.BoolAttribute{
							Optional: true,
						},
						"skip": schema.BoolAttribute{
							Optional: true,
						},
					},
				},
			},
			"status": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[anomalyDetectorStatusModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"status_code": schema.StringAttribute{
							Computed: true,
						},
						"status_reason": schema.StringAttribute{
							Computed: true,
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

func (r *anomalyDetectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	conn := r.Meta().AMPClient(ctx)

	var plan anomalyDetectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input amp.CreateAnomalyDetectorInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields not covered by AutoFlex
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	// TIP: -- 4. Call the AWS Create function
	out, err := conn.CreateAnomalyDetector(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Alias.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Alias.String())
		return
	}

	// Set computed values
	plan.ID = fwflex.StringToFramework(ctx, out.AnomalyDetectorId)
	plan.ARN = fwflex.StringToFramework(ctx, out.Arn)

	detector, err := waitAnomalyDetectorCreated(ctx, conn, plan.ID.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.State.SetAttribute(ctx, path.Root(names.AttrID), plan.ID)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, detector, &plan, fwflex.WithFieldNamePrefix("AnomalyDetector")))
    if resp.Diagnostics.HasError() {
        return
    }

	// TIP: -- 7. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *anomalyDetectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().AMPClient(ctx)

	// TIP: -- 2. Fetch the state
	var state anomalyDetectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findAnomalyDetectorByID(ctx, conn, state.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// TIP: -- 5. Set the arguments and attributes
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Set the state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *anomalyDetectorResource) flatten(ctx context.Context, anomalyDetector *awstypes.AnomalyDetector, data *anomalyDetectorResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, anomalyDetector, data)...)
	return diags
}

func (r *anomalyDetectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	conn := r.Meta().AMPClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state anomalyDetectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the difference between the plan and state, if any
	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input amp.UpdateAnomalyDetectorInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test")))
		if resp.Diagnostics.HasError() {
			return
		}

		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateAnomalyDetector(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil || out.AnomalyDetector == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
			return
		}

		// TIP: Using the output from the update function, re-set any computed attributes
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitAnomalyDetectorUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	// TIP: -- 6. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *anomalyDetectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().AMPClient(ctx)

	// TIP: -- 2. Fetch the state
	var state anomalyDetectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := amp.DeleteAnomalyDetectorInput{
		AnomalyDetectorId: state.ID.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteAnomalyDetector(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAnomalyDetectorDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

// TIP: ==== TERRAFORM IMPORTING ====
// The built-in import function, and Import ID Handler, if any, should handle populating the required
// attributes from the Import ID or Resource Identity.
// In some cases, additional attributes must be set when importing.
// Adding a custom ImportState function can handle those.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/add-resource-identity-support/
// func (r *anomalyDetectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	r.WithImportByIdentity.ImportState(ctx, req, resp)
//
// 	// Set needed attribute values here
// }

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
func waitAnomalyDetectorCreated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.AnomalyDetector, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusAnomalyDetector(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AnomalyDetector); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitAnomalyDetectorUpdated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.AnomalyDetector, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusAnomalyDetector(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AnomalyDetector); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitAnomalyDetectorDeleted(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.AnomalyDetector, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusAnomalyDetector(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AnomalyDetector); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusAnomalyDetector(conn *amp.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findAnomalyDetectorByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, aws.ToString(out.Status), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findAnomalyDetectorByID(ctx context.Context, conn *amp.Client, id string) (*awstypes.AnomalyDetector, error) {
	input := amp.GetAnomalyDetectorInput{
		Id: aws.String(id),
	}

	out, err := conn.GetAnomalyDetector(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.AnomalyDetector == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.AnomalyDetector, nil
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

type anomalyDetectorResourceModel struct {
	framework.WithRegionModel
	Alias                       types.String                                                           `tfsdk:"alias"`
	ARN                         types.String                                                           `tfsdk:"arn"`
	Configuration               fwtypes.ListNestedObjectValueOf[anomalyDetectorConfigurationModel]     `tfsdk:"configuration"`
	CreatedAt                   timetypes.RFC3339                                                      `tfsdk:"created_at"`
	EvaluationIntervalInSeconds types.Int32                                                            `tfsdk:"evaluation_interval_in_seconds"`
	ID                          types.String                                                           `tfsdk:"id"`
	Labels                      fwtypes.MapOfString                                                    `tfsdk:"labels"`
	MissingDataAction           fwtypes.ListNestedObjectValueOf[anomalyDetectorMissingDataActionModel] `tfsdk:"missing_data_action"`
	Status                      fwtypes.ListNestedObjectValueOf[anomalyDetectorStatusModel]            `tfsdk:"status"`
	Tags                        tftags.Map                                                             `tfsdk:"tags"`
	TagsAll                     tftags.Map
	Timeouts                	timeouts.Value                                                		   `tfsdk:"timeouts"`                                                            `tfsdk:"tags_all"`
	WorkspaceID                 types.String                                                           `tfsdk:"workspace_id"`
}

var (
	_ fwflex.Expander  = anomalyDetectorConfigurationModel{}
	_ fwflex.Flattener = &anomalyDetectorConfigurationModel{}
)

type anomalyDetectorConfigurationModel struct {
	RandomCutForest fwtypes.ListNestedObjectValueOf[randomCutForestConfigurationModel] `tfsdk:"random_cut_forest"`
}

func (m anomalyDetectorConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.RandomCutForest.IsNull():
		data, d := m.RandomCutForest.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.AnomalyDetectorConfigurationMemberRandomCutForest
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

func (m *anomalyDetectorConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.AnomalyDetectorConfigurationMemberRandomCutForest:
		var data randomCutForestConfigurationModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}

		m.RandomCutForest = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	}
	return diags
}

type randomCutForestConfigurationModel struct {
	Query                       types.String                                             `tfsdk:"query"`
	IgnoreNearExpectedFromAbove fwtypes.ListNestedObjectValueOf[ignoreNearExpectedModel] `tfsdk:"ignore_near_expected_from_above"`
	IgnoreNearExpectedFromBelow fwtypes.ListNestedObjectValueOf[ignoreNearExpectedModel] `tfsdk:"ignore_near_expected_from_below"`
	SampleSize                  types.Int32                                              `tfsdk:"sample_size"`
	ShingleSize                 types.Int32                                              `tfsdk:"shingle_size"`
}

var (
	_ fwflex.Expander  = ignoreNearExpectedModel{}
	_ fwflex.Flattener = &ignoreNearExpectedModel{}
)

type ignoreNearExpectedModel struct {
	Amount types.Float64 `tfsdk:"amount"`
	Ratio  types.Float64 `tfsdk:"ratio"`
}

func (m ignoreNearExpectedModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Amount.IsNull():
		return &awstypes.IgnoreNearExpectedMemberAmount{
			Value: m.Amount.ValueFloat64(),
		}, diags
	case !m.Ratio.IsNull():
		return &awstypes.IgnoreNearExpectedMemberRatio{
			Value: m.Ratio.ValueFloat64(),
		}, diags
	}
	return nil, diags
}

func (m *ignoreNearExpectedModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case *awstypes.IgnoreNearExpectedMemberAmount:
		m.Amount = types.Float64Value(t.Value)
	case *awstypes.IgnoreNearExpectedMemberRatio:
		m.Ratio = types.Float64Value(t.Value)
	}
	return diags
}

var (
	_ fwflex.Expander  = anomalyDetectorMissingDataActionModel{}
	_ fwflex.Flattener = &anomalyDetectorMissingDataActionModel{}
)

type anomalyDetectorMissingDataActionModel struct {
	MarkAsAnomaly types.Bool `tfsdk:"mark_as_anomaly"`
	Skip          types.Bool `tfsdk:"skip"`
}

func (m anomalyDetectorMissingDataActionModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.MarkAsAnomaly.IsNull():
		return &awstypes.AnomalyDetectorMissingDataActionMemberMarkAsAnomaly{
			Value: m.MarkAsAnomaly.ValueBool(),
		}, diags
	case !m.Skip.IsNull():
		return &awstypes.AnomalyDetectorMissingDataActionMemberSkip{
			Value: m.Skip.ValueBool(),
		}, diags
	}
	return nil, diags
}

func (m *anomalyDetectorMissingDataActionModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case *awstypes.AnomalyDetectorMissingDataActionMemberMarkAsAnomaly:
		m.MarkAsAnomaly = fwflex.BoolValueToFramework(ctx, t.Value)
	case *awstypes.AnomalyDetectorMissingDataActionMemberSkip:
		m.Skip = fwflex.BoolValueToFramework(ctx, t.Value)
	}
	return diags
}

type anomalyDetectorStatusModel struct {
	StatusCode   fwtypes.StringEnum[awstypes.AnomalyDetectorStatusCode] `tfsdk:"status_code"`
	StatusReason types.String                                           `tfsdk:"status_reason"`
}

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
var (
	_ inttypes.ImportIDParser = anomalyDetectorImportID{}
)

type anomalyDetectorImportID struct{}

func (anomalyDetectorImportID) Parse(id string) (string, map[string]string, error) {
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
//	awsv2.Register("aws_amp_anomaly_detector", sweepAnomalyDetectors)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepAnomalyDetectors(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := amp.ListAnomalyDetectorsInput{}
	conn := client.AMPClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := amp.NewListAnomalyDetectorsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.AnomalyDetectors {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newAnomalyDetectorResource, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.AnomalyDetectorId))),
			)
		}
	}

	return sweepResources, nil
}
