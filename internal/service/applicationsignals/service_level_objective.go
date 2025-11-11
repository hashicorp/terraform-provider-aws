// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationsignals

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationsignals"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationsignals/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_applicationsignals_service_level_objective", name="Service Level Objective")
func newResourceServiceLevelObjective(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceServiceLevelObjective{}

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
	ResNameServiceLevelObjective = "Service Level Objective"
)

type resourceServiceLevelObjective struct {
	framework.ResourceWithModel[resourceServiceLevelObjectiveModel]
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
func (r *resourceServiceLevelObjective) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreatedTime: schema.StringAttribute{
				Computed: true,
			},
			"last_updated_time": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"metric_source_type": schema.StringAttribute{
				Computed: true,
			},
			"evaluation_type": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"goal": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[goalModel](ctx),
				Attributes: map[string]schema.Attribute{
					"attainment_goal":   schema.Float64Attribute{Computed: true},
					"warning_threshold": schema.Float64Attribute{Computed: true},
				},
				Blocks: map[string]schema.Block{
					"interval": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[intervalModel](ctx),
						Blocks: map[string]schema.Block{
							"calendar_interval": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[calendarIntervalModel](ctx),
								Attributes: map[string]schema.Attribute{
									"duration":      schema.Int32Attribute{Computed: true},
									"duration_unit": schema.StringAttribute{Computed: true},
									"start_time":    schema.StringAttribute{Computed: true},
								},
							},
							"rolling_interval": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[rollingIntervalModel](ctx),
								Attributes: map[string]schema.Attribute{
									"duration":      schema.Int32Attribute{Computed: true},
									"duration_unit": schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
			"burn_rate_configurations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[burnRateConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"look_back_window_minutes": schema.Int32Attribute{Computed: true},
					},
				},
			},
			"request_based_sli": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[requestBasedSliModel](ctx),
				Attributes: map[string]schema.Attribute{
					"metric_threshold":    schema.Float64Attribute{Computed: true},
					"comparison_operator": schema.StringAttribute{Computed: true},
				},
				Blocks: map[string]schema.Block{
					"request_based_sli_metric": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[requestBasedSliMetricModel](ctx),
						Attributes: map[string]schema.Attribute{
							"dependency_config": schema.StringAttribute{Computed: true},
							"key_attributes":    schema.MapAttribute{CustomType: fwtypes.MapOfStringType, ElementType: types.StringType, Computed: true},
							"metric_type":       schema.StringAttribute{Computed: true},
							"operation_name":    schema.StringAttribute{Computed: true},
						},
						Blocks: map[string]schema.Block{
							"total_request_count_metric": metricDataQueriesBlock(ctx),
						},
					},
				},
			},
			"sli": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[sliModel](ctx),
				Attributes: map[string]schema.Attribute{
					"metric_threshold":    schema.Float64Attribute{Computed: true},
					"comparison_operator": schema.StringAttribute{Computed: true},
				},
				Blocks: map[string]schema.Block{
					"sli_metric": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[sliMetricModel](ctx),
						Attributes: map[string]schema.Attribute{
							"dependency_config": schema.StringAttribute{Computed: true},
							"key_attributes":    schema.MapAttribute{CustomType: fwtypes.MapOfStringType, ElementType: types.StringType, Computed: true},
							"metric_type":       schema.StringAttribute{Computed: true},
							"operation_name":    schema.StringAttribute{Computed: true},
						},
						Blocks: map[string]schema.Block{
							"metric_data_queries": metricDataQueriesBlock(ctx),
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

func metricDataQueriesBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[metricDataQueryModel](ctx),
		NestedObject: schema.NestedBlockObject{
			CustomType: fwtypes.NewObjectTypeOf[metricDataQueryModel](ctx),
			Attributes: map[string]schema.Attribute{
				"id":          schema.StringAttribute{Computed: true},
				"account_id":  schema.StringAttribute{Computed: true},
				"expression":  schema.StringAttribute{Computed: true},
				"label":       schema.StringAttribute{Computed: true},
				"period":      schema.Int32Attribute{Computed: true},
				"return_data": schema.BoolAttribute{Computed: true},
			},
			Blocks: map[string]schema.Block{
				"metric_stat": schema.SingleNestedBlock{
					CustomType: fwtypes.NewObjectTypeOf[metricStatModel](ctx),
					Attributes: map[string]schema.Attribute{
						"period": schema.Int32Attribute{Computed: true},
						"stat":   schema.StringAttribute{Computed: true},
						"unit":   schema.StringAttribute{Computed: true},
					},
					Blocks: map[string]schema.Block{
						"metric": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[metricModel](ctx),
							Attributes: map[string]schema.Attribute{
								"metric_name": schema.StringAttribute{Computed: true},
								"namespace":   schema.StringAttribute{Computed: true},
							},
							Blocks: map[string]schema.Block{
								"dimensions": schema.ListNestedBlock{
									CustomType: fwtypes.NewListNestedObjectTypeOf[dimensionModel](ctx),
									NestedObject: schema.NestedBlockObject{
										CustomType: fwtypes.NewObjectTypeOf[dimensionModel](ctx),
										Attributes: map[string]schema.Attribute{
											"name":  schema.StringAttribute{Computed: true},
											"value": schema.StringAttribute{Computed: true},
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

func (r *resourceServiceLevelObjective) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	conn := r.Meta().ApplicationSignalsClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	var input applicationsignals.CreateServiceLevelObjectiveInput
	// TIP: Using a field name prefix allows mapping fields such as `ID` to `ServiceLevelObjectiveId`
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("ServiceLevelObjective")))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 4. Call the AWS Create function
	out, err := conn.CreateServiceLevelObjective(ctx, &input)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.ServiceLevelObjective == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	// TIP: -- 5. Using the output from the create function, set attributes
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitServiceLevelObjectiveCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	// TIP: -- 7. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceServiceLevelObjective) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().ApplicationSignalsClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findServiceLevelObjectiveByID(ctx, conn, state.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// TIP: -- 5. Set the arguments and attributes
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Set the state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceServiceLevelObjective) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	conn := r.Meta().ApplicationSignalsClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceServiceLevelObjectiveModel
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
		var input applicationsignals.UpdateServiceLevelObjectiveInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test")))
		if resp.Diagnostics.HasError() {
			return
		}

		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateServiceLevelObjective(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil || out.ServiceLevelObjective == nil {
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
	_, err := waitServiceLevelObjectiveUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	// TIP: -- 6. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceServiceLevelObjective) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().ApplicationSignalsClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := applicationsignals.DeleteServiceLevelObjectiveInput{
		ServiceLevelObjectiveId: state.ID.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteServiceLevelObjective(ctx, &input)
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
	_, err = waitServiceLevelObjectiveDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
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
func (r *resourceServiceLevelObjective) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
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
func waitServiceLevelObjectiveCreated(ctx context.Context, conn *applicationsignals.Client, id string, timeout time.Duration) (*awstypes.ServiceLevelObjective, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusServiceLevelObjective(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*applicationsignals.ServiceLevelObjective); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitServiceLevelObjectiveUpdated(ctx context.Context, conn *applicationsignals.Client, id string, timeout time.Duration) (*awstypes.ServiceLevelObjective, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusServiceLevelObjective(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*applicationsignals.ServiceLevelObjective); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitServiceLevelObjectiveDeleted(ctx context.Context, conn *applicationsignals.Client, id string, timeout time.Duration) (*awstypes.ServiceLevelObjective, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusServiceLevelObjective(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*applicationsignals.ServiceLevelObjective); ok {
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
func statusServiceLevelObjective(ctx context.Context, conn *applicationsignals.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findServiceLevelObjectiveByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, aws.ToString(out.Status), nil
	}
}

func findServiceLevelObjectiveByID(ctx context.Context, conn *applicationsignals.Client, id string) (*awstypes.ServiceLevelObjective, error) {
	input := applicationsignals.GetServiceLevelObjectiveInput{
		Id: aws.String(id),
	}

	out, err := conn.GetServiceLevelObjective(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Slo == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out.Slo, nil
}

var _ flex.Flattener = &intervalModel{}

func (m *intervalModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.CalendarInterval = fwtypes.NewObjectValueOfNull[calendarIntervalModel](ctx)
	m.RollingInterval = fwtypes.NewObjectValueOfNull[rollingIntervalModel](ctx)

	switch t := v.(type) {

	case awstypes.IntervalMemberCalendarInterval:
		var model calendarIntervalModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		if !diags.HasError() {
			m.CalendarInterval = fwtypes.NewObjectValueOfMust(ctx, &model)
		}

	case awstypes.IntervalMemberRollingInterval:
		var model rollingIntervalModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		if !diags.HasError() {
			m.RollingInterval = fwtypes.NewObjectValueOfMust(ctx, &model)
		}
	}

	return diags
}

type resourceServiceLevelObjectiveModel struct {
	framework.WithRegionModel
	ID                     types.String                                                `tfsdk:"id"`
	ARN                    types.String                                                `tfsdk:"arn"`
	CreatedTime            types.String                                                `tfsdk:"created_time"`
	BurnRateConfigurations fwtypes.ListNestedObjectValueOf[burnRateConfigurationModel] `tfsdk:"burn_rate_configurations"`
	LastUpdatedTime        types.String                                                `tfsdk:"last_updated_time"`
	Name                   types.String                                                `tfsdk:"name"`
	Description            types.String                                                `tfsdk:"description"`
	MetricSourceType       types.String                                                `tfsdk:"metric_source_type"`
	EvaluationType         types.String                                                `tfsdk:"evaluation_type"`
	Goal                   fwtypes.ObjectValueOf[goalModel]                            `tfsdk:"goal"`
	Sli                    fwtypes.ObjectValueOf[sliModel]                             `tfsdk:"sli"`
	RequestBasedSli        fwtypes.ObjectValueOf[requestBasedSliModel]                 `tfsdk:"request_based_sli"`
	Timeouts               timeouts.Value                                              `tfsdk:"timeouts"`
	Type                   types.String                                                `tfsdk:"type"`
}

type goalModel struct {
	AttainmentGoal   types.Float64                        `tfsdk:"attainment_goal"`
	WarningThreshold types.Float64                        `tfsdk:"warning_threshold"`
	Interval         fwtypes.ObjectValueOf[intervalModel] `tfsdk:"interval"`
}

type intervalModel struct {
	CalendarInterval fwtypes.ObjectValueOf[calendarIntervalModel] `tfsdk:"calendar_interval"`
	RollingInterval  fwtypes.ObjectValueOf[rollingIntervalModel]  `tfsdk:"rolling_interval"`
}

type calendarIntervalModel struct {
	Duration     types.Int32  `tfsdk:"duration"`
	DurationUnit types.String `tfsdk:"duration_unit"`
	StartTime    types.String `tfsdk:"start_time"`
}

type rollingIntervalModel struct {
	Duration     types.Int32  `tfsdk:"duration"`
	DurationUnit types.String `tfsdk:"duration_unit"`
}

type sliModel struct {
	ComparisonOperator types.String                          `tfsdk:"comparison_operator"`
	MetricThreshold    types.Float64                         `tfsdk:"metric_threshold"`
	SliMetric          fwtypes.ObjectValueOf[sliMetricModel] `tfsdk:"sli_metric"`
}

type requestBasedSliModel struct {
	RequestBasedSliMetric fwtypes.ObjectValueOf[requestBasedSliMetricModel] `tfsdk:"request_based_sli_metric"`
	ComparisonOperator    types.String                                      `tfsdk:"comparison_operator"`
	MetricThreshold       types.Float64                                     `tfsdk:"metric_threshold"`
}

type burnRateConfigurationModel struct {
	LookBackWindowMinutes types.Int32 `tfsdk:"look_back_window_minutes"`
}

type requestBasedSliMetricModel struct {
	TotalRequestCountMetric fwtypes.ListNestedObjectValueOf[metricDataQueryModel] `tfsdk:"total_request_count_metric"`
	DependencyConfig        fwtypes.ObjectValueOf[dependencyConfigModel]          `tfsdk:"dependency_config"`
	KeyAttributes           fwtypes.MapOfString                                   `tfsdk:"key_attributes"`
	MetricType              types.String                                          `tfsdk:"metric_type"`
	OperationName           types.String                                          `tfsdk:"operation_name"`
}

type sliMetricModel struct {
	MetricDataQueries fwtypes.ListNestedObjectValueOf[metricDataQueryModel] `tfsdk:"metric_data_queries"`
	DependencyConfig  fwtypes.ObjectValueOf[dependencyConfigModel]          `tfsdk:"dependency_config"`
	KeyAttributes     fwtypes.MapOfString                                   `tfsdk:"key_attributes"`
	MetricType        types.String                                          `tfsdk:"metric_type"`
	OperationName     types.String                                          `tfsdk:"operation_name"`
}

type metricDataQueryModel struct {
	Id         types.String                           `tfsdk:"id"`
	AccountId  types.String                           `tfsdk:"account_id"`
	Expression types.String                           `tfsdk:"expression"`
	Label      types.String                           `tfsdk:"label"`
	MetricStat fwtypes.ObjectValueOf[metricStatModel] `tfsdk:"metric_stat"`
	Period     types.Int32                            `tfsdk:"period"`
	ReturnData types.Bool                             `tfsdk:"return_data"`
}

type metricStatModel struct {
	Metric fwtypes.ObjectValueOf[metricModel] `tfsdk:"metric"`
	Period types.Int32                        `tfsdk:"period"`
	Stat   types.String                       `tfsdk:"stat"`
	Unit   types.String                       `tfsdk:"unit"`
}

type metricModel struct {
	Dimensions fwtypes.ListNestedObjectValueOf[dimensionModel] `tfsdk:"dimensions"`
	MetricName types.String                                    `tfsdk:"metric_name"`
	Namespace  types.String                                    `tfsdk:"namespace"`
}

type dimensionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type dependencyConfigModel struct {
	DependencyKeyAttributes types.String `tfsdk:"dependency_key_attributes"`
	DependencyOperationName types.String `tfsdk:"dependency_operation_name"`
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
//	awsv2.Register("aws_applicationsignals_service_level_objective", sweepServiceLevelObjectives)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepServiceLevelObjectives(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := applicationsignals.ListServiceLevelObjectivesInput{}
	conn := client.ApplicationSignalsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := applicationsignals.NewListServiceLevelObjectivesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ServiceLevelObjectives {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceServiceLevelObjective, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.ServiceLevelObjectiveId))),
			)
		}
	}

	return sweepResources, nil
}
