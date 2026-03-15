// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package organizations

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
	// using the services/organizations/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
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

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_organizations_aws_service_access", name="Aws Service Access")

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

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	// r.SetDefaultCreateTimeout(30 * time.Minute)
	// r.SetDefaultUpdateTimeout(30 * time.Minute)
	// r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAwsServiceAccess = "Aws Service Access"
)

type awsServiceAccessResource struct {
	framework.ResourceWithModel[awsServiceAccessResourceModel]
	// framework.WithTimeouts
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
func (r *awsServiceAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_principal": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// names.AttrARN: framework.ARNAttributeComputedOnly(),
			// names.AttrDescription: schema.StringAttribute{
			// 	Optional: true,
			// },
			// TIP: ==== "ID" ATTRIBUTE ====
			// When using the Terraform Plugin Framework, there is no required "id" attribute.
			// This is different from the Terraform Plugin SDK.
			//
			// Only include an "id" attribute if the AWS API has an "Id" field, such as "AwsServiceAccessId"
			// names.AttrID: framework.IDAttribute(),
			// names.AttrName: schema.StringAttribute{
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
//  awsv2.Register("aws_organizations_aws_service_access", sweepAwsServiceAccesss)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
