// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

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
	// using the services/iam/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_iam_openid_connect_provider_client_id", name="Open ID Connect Provider Client ID")
// @IdentityAttribute("openid_connect_provider_arn")
// @IdentityAttribute("client_id")
//
// TIP: ==== GENERATED ACCEPTANCE TESTS ====
// Resource Identity and tagging make use of automatically generated acceptance tests.
// For more information about automatically generated acceptance tests, see
// https://hashicorp.github.io/terraform-provider-aws/acc-test-generation/
//
// Some common annotations are included below:
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/iam;iam.DescribeOpenIDConnectProviderClientIDResponse")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importIgnore="...;...")
// @Testing(hasNoPreExistingResource=true)
func newOpenIDConnectProviderClientIDResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &openIDConnectProviderClientIDResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameOpenIDConnectProviderClientID = "Open ID Connect Provider Client ID"
)

type openIDConnectProviderClientIDResource struct {
	framework.ResourceWithModel[openIDConnectProviderClientIDResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *openIDConnectProviderClientIDResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Required: true,
			},
			"openid_connect_provider_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required: true,
			},
		},
	}
}

func (r *openIDConnectProviderClientIDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	conn := r.Meta().IAMClient(ctx)
	
	// TIP: -- 2. Fetch the plan
	var plan openIDConnectProviderClientIDResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	var input iam.CreateOpenIDConnectProviderClientIDInput
	// TIP: Using a field name prefix allows mapping fields such as `ID` to `OpenIDConnectProviderClientIDId`
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("OpenIDConnectProviderClientID")))
	if resp.Diagnostics.HasError() {
		return
	}
	

	// TIP: -- 4. Call the AWS Create function
	out, err := conn.CreateOpenIDConnectProviderClientID(ctx, &input)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.OpenIDConnectProviderClientID == nil {
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
	_, err = waitOpenIDConnectProviderClientIDCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	
	// TIP: -- 7. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *openIDConnectProviderClientIDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().IAMClient(ctx)
	
	// TIP: -- 2. Fetch the state
	var state openIDConnectProviderClientIDResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	
	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findOpenIDConnectProviderClientIDByID(ctx, conn, state.ID.ValueString())
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

func (r *openIDConnectProviderClientIDResource) flatten(ctx context.Context, openIDConnectProviderClientID *awstypes.OpenIDConnectProviderClientID, data *openIDConnectProviderClientIDResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, openIDConnectProviderClientID, data)...)
	return diags
}

func (r *openIDConnectProviderClientIDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	conn := r.Meta().IAMClient(ctx)
	
	// TIP: -- 2. Fetch the plan
	var plan, state openIDConnectProviderClientIDResourceModel
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
		var input iam.UpdateOpenIDConnectProviderClientIDInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test")))
		if resp.Diagnostics.HasError() {
			return
		}
		
		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateOpenIDConnectProviderClientID(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil || out.OpenIDConnectProviderClientID == nil {
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
	_, err := waitOpenIDConnectProviderClientIDUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	// TIP: -- 6. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *openIDConnectProviderClientIDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().IAMClient(ctx)
	
	// TIP: -- 2. Fetch the state
	var state openIDConnectProviderClientIDResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	
	// TIP: -- 3. Populate a delete input structure
	input := iam.DeleteOpenIDConnectProviderClientIDInput{
		OpenIDConnectProviderClientIDId: state.ID.ValueStringPointer(),
	}
	
	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteOpenIDConnectProviderClientID(ctx, &input)
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
	_, err = waitOpenIDConnectProviderClientIDDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
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
// func (r *openIDConnectProviderClientIDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
func waitOpenIDConnectProviderClientIDCreated(ctx context.Context, conn *iam.Client, id string, timeout time.Duration) (*awstypes.OpenIDConnectProviderClientID, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusOpenIDConnectProviderClientID(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.OpenIDConnectProviderClientID); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitOpenIDConnectProviderClientIDUpdated(ctx context.Context, conn *iam.Client, id string, timeout time.Duration) (*awstypes.OpenIDConnectProviderClientID, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusOpenIDConnectProviderClientID(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.OpenIDConnectProviderClientID); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitOpenIDConnectProviderClientIDDeleted(ctx context.Context, conn *iam.Client, id string, timeout time.Duration) (*awstypes.OpenIDConnectProviderClientID, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusOpenIDConnectProviderClientID(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.OpenIDConnectProviderClientID); ok {
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
func statusOpenIDConnectProviderClientID(conn *iam.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findOpenIDConnectProviderClientIDByID(ctx, conn, id)
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
func findOpenIDConnectProviderClientIDByID(ctx context.Context, conn *iam.Client, id string) (*awstypes.OpenIDConnectProviderClientID, error) {
	input := iam.GetOpenIDConnectProviderClientIDInput{
		Id: aws.String(id),
	}

	out, err := conn.GetOpenIDConnectProviderClientID(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.OpenIDConnectProviderClientID == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.OpenIDConnectProviderClientID, nil
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
type openIDConnectProviderClientIDResourceModel struct {
	framework.WithRegionModel
	ARN             types.String                                          `tfsdk:"arn"`
	ComplexArgument fwtypes.ListNestedObjectValueOf[complexArgumentModel] `tfsdk:"complex_argument"`
	Description     types.String                                          `tfsdk:"description"`
	ID              types.String                                          `tfsdk:"id"`
	Name            types.String                                          `tfsdk:"name"`
	Timeouts        timeouts.Value                                        `tfsdk:"timeouts"`
	Type            types.String                                          `tfsdk:"type"`
}

type complexArgumentModel struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
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
	_ inttypes.ImportIDParser = openIDConnectProviderClientIDImportID{}
)

type openIDConnectProviderClientIDImportID struct{}

func (openIDConnectProviderClientIDImportID) Parse(id string) (string, map[string]string, error) {
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
//  awsv2.Register("aws_iam_openid_connect_provider_client_id", sweepOpenIDConnectProviderClientIDs)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepOpenIDConnectProviderClientIDs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := iam.ListOpenIDConnectProviderClientIDsInput{}
	conn := client.IAMClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := iam.NewListOpenIDConnectProviderClientIDsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.OpenIDConnectProviderClientIDs {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newOpenIDConnectProviderClientIDResource, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.OpenIDConnectProviderClientIDId))),
			)
		}
	}

	return sweepResources, nil
}
