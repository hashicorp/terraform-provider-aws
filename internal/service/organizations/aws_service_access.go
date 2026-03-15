// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

// @FrameworkResource("aws_organizations_aws_service_access", name="Aws Service Access")
// @IdentityAttribute("service_principal")
//
// TIP: ==== GENERATED ACCEPTANCE TESTS ====
// Resource Identity and tagging make use of automatically generated acceptance tests.
// For more information about automatically generated acceptance tests, see
// https://hashicorp.github.io/terraform-provider-aws/acc-test-generation/
//
// Some common annotations are included below:
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/organizations;organizations.DescribeAwsServiceAccessResponse")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importIgnore="...;...")
// @Testing(hasNoPreExistingResource=true)
func newAwsServiceAccessResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &awsServiceAccessResource{}

	return r, nil
}

const (
	ResNameAwsServiceAccess = "Aws Service Access"
)

type awsServiceAccessResource struct {
	framework.ResourceWithModel[awsServiceAccessResourceModel]
	framework.WithImportByIdentity
}

func (r *awsServiceAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_principal": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *awsServiceAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	conn := r.Meta().OrganizationsClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan awsServiceAccessResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	var input organizations.EnableAWSServiceAccessInput
	// TIP: Using a field name prefix allows mapping fields such as `ID` to `AwsServiceAccessId`
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AwsServiceAccess")))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 4. Call the AWS Create function
	out, err := conn.EnableAWSServiceAccess(ctx, &input)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ServicePrincipal.String())
		return
	}
	// if out == nil || out.AwsServiceAccess == nil {
	// 	smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
	// 	return
	// }

	// TIP: -- 5. Using the output from the create function, set attributes
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Use a waiter to wait for create to complete
	// createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	// _, err = waitAwsServiceAccessCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	// if err != nil {
	// 	smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
	// 	return
	// }

	// TIP: -- 7. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *awsServiceAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	conn := r.Meta().OrganizationsClient(ctx)

	// TIP: -- 2. Fetch the state
	var state awsServiceAccessResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findAwsServiceAccessByID(ctx, conn, state.ServicePrincipal.String())
	// TIP: -- 4. Remove resource from state if it is not found
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ServicePrincipal.String())
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

func (r *awsServiceAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().OrganizationsClient(ctx)

	// TIP: -- 2. Fetch the state
	var state awsServiceAccessResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := organizations.DisableAWSServiceAccessInput{
		ServicePrincipal: state.ServicePrincipal.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DisableAWSServiceAccess(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		// if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		// 	return
		// }

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ServicePrincipal.String())
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	// deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	// _, err = waitAwsServiceAccessDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	// if err != nil {
	// 	smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
	// 	return
	// }
	return
}

// TIP: ==== TERRAFORM IMPORTING ====
// The built-in import function, and Import ID Handler, if any, should handle populating the required
// attributes from the Import ID or Resource Identity.
// In some cases, additional attributes must be set when importing.
// Adding a custom ImportState function can handle those.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/add-resource-identity-support/
// func (r *awsServiceAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	r.WithImportByIdentity.ImportState(ctx, req, resp)
//
// 	// Set needed attribute values here
// }

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findAwsServiceAccessByID(ctx context.Context, conn *organizations.Client, servicePrincipal string) (*awstypes.EnabledServicePrincipal, error) {
	var enabledServices []awstypes.EnabledServicePrincipal

	pages := organizations.NewListAWSServiceAccessForOrganizationPaginator(conn, nil)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, sp := range page.EnabledServicePrincipals {
			if sp.ServicePrincipal == &servicePrincipal {
				enabledServices = append(enabledServices, sp)
			}
		}
	}
	// if err != nil {
	// 	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
	// 		return nil, smarterr.NewError(&retry.NotFoundError{
	// 			LastError:   err,
	// 		})
	// 	}

	// 	return nil, smarterr.NewError(err)
	// }

	return tfresource.AssertSingleValueResult(enabledServices)
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
type awsServiceAccessResourceModel struct {
	ServicePrincipal types.String `tfsdk:"service_principal"`
}
