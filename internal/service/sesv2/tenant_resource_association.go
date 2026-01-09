// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"

	//	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	//	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	//"github.com/hashicorp/terraform-plugin-go/tftypes"

	//	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	//	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	//	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"

	//	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	//	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sesv2_tenant_resource_association", name="Tenant Resource Association", tags=false)
// @Testing(importStateIdAttribute="tenant_name")
func newResourceTenantResourceAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTenantResourceAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameTenantResourceAssociation = "Tenant Resource Association"
)

type resourceTenantResourceAssociation struct {
	framework.ResourceWithModel[resourceTenantResourceAssociationModel]
	framework.WithTimeouts
}

func (r *resourceTenantResourceAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"resource_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tenant_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceTenantResourceAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().SESV2Client(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceTenantResourceAssociationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	var input sesv2.CreateTenantResourceAssociationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 4. Call the AWS Create function
	out, err := conn.CreateTenantResourceAssociation(ctx, &input)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.TenantName.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.TenantName.String())
		return
	}

	plan.ID = types.StringValue(createID(plan.TenantName.ValueString(), plan.ResourceArn.ValueString()))
	//fmt.Printf("DEBUG :::: ID of the resource is %v\n", plan.ID)

	// TIP: -- 5. Using the output from the create function, set attributes
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Use a waiter to wait for create to complete
	//	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	//	_, err = waitTenantResourceAssociationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	//	if err != nil {
	//		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
	//		return
	//	}

	// TIP: -- 7. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceTenantResourceAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().SESV2Client(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceTenantResourceAssociationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findTenantResourceAssociationByID(ctx, conn, state.ID.ValueString())
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
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Set the state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

//func (r *resourceTenantResourceAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
//	// TIP: ==== RESOURCE UPDATE ====
//	// Not all resources have Update functions. There are a few reasons:
//	// a. The AWS API does not support changing a resource
//	// b. All arguments have RequiresReplace() plan modifiers
//	// c. The AWS API uses a create call to modify an existing resource
//	//
//	// In the cases of a. and b., the resource will not have an update method
//	// defined. In the case of c., Update and Create can be refactored to call
//	// the same underlying function.
//	//
//	// The rest of the time, there should be an Update function and it should
//	// do the following things. Make sure there is a good reason if you don't
//	// do one of these.
//	//
//	// 1. Get a client connection to the relevant service
//	// 2. Fetch the plan and state
//	// 3. Populate a modify input structure and check for changes
//	// 4. Call the AWS modify/update function
//	// 5. Use a waiter to wait for update to complete
//	// 6. Save the request plan to response state
//	// TIP: -- 1. Get a client connection to the relevant service
//	conn := r.Meta().SESV2Client(ctx)
//
//	// TIP: -- 2. Fetch the plan
//	var plan, state resourceTenantResourceAssociationModel
//	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
//	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	// TIP: -- 3. Get the difference between the plan and state, if any
//	diff, d := flex.Diff(ctx, plan, state)
//	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	if diff.HasChanges() {
//		var input sesv2.UpdateTenantResourceAssociationInput
//		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test")))
//		if resp.Diagnostics.HasError() {
//			return
//		}
//
//		// TIP: -- 4. Call the AWS modify/update function
//		out, err := conn.UpdateTenantResourceAssociation(ctx, &input)
//		if err != nil {
//			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
//			return
//		}
//		if out == nil || out.TenantResourceAssociation == nil {
//			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
//			return
//		}
//
//		// TIP: Using the output from the update function, re-set any computed attributes
//		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
//		if resp.Diagnostics.HasError() {
//			return
//		}
//	}
//
//	// TIP: -- 5. Use a waiter to wait for update to complete
//	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
//	_, err := waitTenantResourceAssociationUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
//	if err != nil {
//		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
//		return
//	}
//
//	// TIP: -- 6. Save the request plan to response state
//	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
//}

func (r *resourceTenantResourceAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().SESV2Client(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceTenantResourceAssociationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := sesv2.DeleteTenantResourceAssociationInput{
		ResourceArn: state.ResourceArn.ValueStringPointer(),
		TenantName:  state.TenantName.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteTenantResourceAssociation(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	//	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	//	_, err = waitTenantResourceAssociationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	//	if err != nil {
	//		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
	//		return
	//	}
}

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceTenantResourceAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	customID := req.ID
	//fmt.Printf("DEBUG :::: Custom ID %v\n", customID)
	parts := strings.Split(customID, "|")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid ID Format",
			fmt.Sprintf("Expected ID in the format of tenant_name|resource_arn, got: %s", customID),
		)
	}
	tenantName := parts[0]
	resourceARN := parts[1]
	resp.State.SetAttribute(ctx, path.Root("id"), customID)
	resp.State.SetAttribute(ctx, path.Root("tenant_name"), tenantName)
	resp.State.SetAttribute(ctx, path.Root("resource_arn"), resourceARN)
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
//const (
//	statusChangePending = "Pending"
//	statusDeleting      = "Deleting"
//	statusNormal        = "Normal"
//	statusUpdated       = "Updated"
//)

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
//func waitTenantResourceAssociationCreated(ctx context.Context, conn *sesv2.Client, id string, timeout time.Duration) (*awstypes.TenantResource, error) {
//	stateConf := &retry.StateChangeConf{
//		Pending:                   []string{},
//		Target:                    []string{statusNormal},
//		Refresh:                   statusTenantResourceAssociation(conn, id),
//		Timeout:                   timeout,
//		NotFoundChecks:            20,
//		ContinuousTargetOccurence: 2,
//	}
//
//	outputRaw, err := stateConf.WaitForStateContext(ctx)
//	if out, ok := outputRaw.(*awstypes.TenantResource); ok {
//		return out, smarterr.NewError(err)
//	}
//
//	return nil, smarterr.NewError(err)
//}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
//func waitTenantResourceAssociationUpdated(ctx context.Context, conn *sesv2.Client, id string, timeout time.Duration) (*awstypes.TenantResource, error) {
//	stateConf := &retry.StateChangeConf{
//		Pending:                   []string{statusChangePending},
//		Target:                    []string{statusUpdated},
//		Refresh:                   statusTenantResourceAssociation(conn, id),
//		Timeout:                   timeout,
//		NotFoundChecks:            20,
//		ContinuousTargetOccurence: 2,
//	}
//
//	outputRaw, err := stateConf.WaitForStateContext(ctx)
//	if out, ok := outputRaw.(*awstypes.TenantResource); ok {
//		return out, smarterr.NewError(err)
//	}
//
//	return nil, smarterr.NewError(err)
//}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
//func waitTenantResourceAssociationDeleted(ctx context.Context, conn *sesv2.Client, id string, timeout time.Duration) (*awstypes.TenantResource, error) {
//	stateConf := &retry.StateChangeConf{
//		Pending: []string{statusDeleting, statusNormal},
//		Target:  []string{},
//		Refresh: statusTenantResourceAssociation(conn, id),
//		Timeout: timeout,
//	}
//
//	outputRaw, err := stateConf.WaitForStateContext(ctx)
//	if out, ok := outputRaw.(*awstypes.TenantResource); ok {
//		return out, smarterr.NewError(err)
//	}
//
//	return nil, smarterr.NewError(err)
//}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
//func statusTenantResourceAssociation(conn *sesv2.Client, id string) retry.StateRefreshFunc {
//	return func(ctx context.Context) (any, string, error) {
//		out, err := findTenantResourceAssociationByID(ctx, conn, id)
//		if retry.NotFound(err) {
//			return nil, "", nil
//		}
//
//		if err != nil {
//			return nil, "", smarterr.NewError(err)
//		}
//
//		return out, aws.ToString(out.Status), nil
//	}
//}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findTenantResourceAssociationByID(
	ctx context.Context,
	conn *sesv2.Client,
	resourceID string,
) (*awstypes.TenantResource, error) {

	parts := strings.SplitN(resourceID, "|", 2)
	if len(parts) != 2 {
		return nil, smarterr.NewError(
			tfresource.NewEmptyResultError(resourceID),
		)
	}

	//fmt.Printf("DEBUG :::: %v\n", resourceID)
	tenantName := parts[0]
	resourceARN := parts[1]
	//	fmt.Printf("DEBUG :::: Tenant Name => %v\n", tenantName)
	//	fmt.Printf("DEBUG :::: Resource ARN => %v\n", resourceARN)

	input := &sesv2.ListTenantResourcesInput{
		TenantName: aws.String(tenantName),
	}

	p := sesv2.NewListTenantResourcesPaginator(conn, input)

	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, tfresource.ErrEmptyResult
		}

		for _, tenantResource := range out.TenantResources {
			if aws.ToString(tenantResource.ResourceArn) == resourceARN {
				return &tenantResource, nil
			}
		}
	}

	return nil, tfresource.ErrEmptyResult
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
type resourceTenantResourceAssociationModel struct {
	framework.WithRegionModel
	ResourceArn types.String `tfsdk:"resource_arn"`
	ID          types.String `tfsdk:"id"`
	TenantName  types.String `tfsdk:"tenant_name"`
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
//	awsv2.Register("aws_sesv2_tenant_resource_association", sweepTenantResourceAssociations)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepTenantResourceAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := sesv2.ListResourceTenantsInput{}
	conn := client.SESV2Client(ctx)
	var sweepResources []sweep.Sweepable

	pages := sesv2.NewListResourceTenantsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ResourceTenants {
			sweepResources = append(
				sweepResources,
				sweepfw.NewSweepResource(
					newResourceTenantResourceAssociation,
					client,
					sweepfw.NewAttribute(
						names.AttrID,
						createID(
							aws.ToString(v.TenantName),
							aws.ToString(v.ResourceArn),
						),
					),
				),
			)
		}
	}

	return sweepResources, nil
}

func createID(tenantName, resourceARN string) string {
	return fmt.Sprintf("%s|%s", tenantName, resourceARN)
}
