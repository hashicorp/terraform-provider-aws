package odb

/*
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_odb_associate_disassociate_iam_role", name="Associate Disassociate IAM Role")
func newResourceAssociateDisassociateIAMRole(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAssociateDisassociateIAMRole{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAssociateDisassociateIAMRole = "Associate Disassociate IAM Role"
)

type resourceAssociateDisassociateIAMRole struct {
	framework.ResourceWithModel[resourceAssociateDisassociateIAMRoleResourceModel]
	framework.WithTimeouts
}

func (r *resourceAssociateDisassociateIAMRole) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"iam_role_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws_integration": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"status_reason": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceAssociateDisassociateIAMRole) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)
	var plan resourceAssociateDisassociateIAMRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var input odb.AssociateIamRoleToResourceInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AssociateDisassociateIAMRole"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := conn.AssociateIamRoleToResource(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAssociateDisassociateIAMRole, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAssociateDisassociateIAMRole, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitAssociateDisassociateIAMRoleCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameAssociateDisassociateIAMRole, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAssociateDisassociateIAMRole) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state resourceAssociateDisassociateIAMRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findAssociateDisassociateIAMRole(ctx, conn, state.ResourceARN.ValueString(), state.IAMRoleARN.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameAssociateDisassociateIAMRole, state.ID.String(), err),
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

func (r *resourceAssociateDisassociateIAMRole) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	conn := r.Meta().ODBClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceAssociateDisassociateIAMRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := odb.DeleteAssociateDisassociateIAMRoleInput{
		AssociateDisassociateIAMRoleId: state.ID.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteAssociateDisassociateIAMRole(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameAssociateDisassociateIAMRole, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAssociateDisassociateIAMRoleDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, state.ID.String(), err),
			err.Error(),
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
func (r *resourceAssociateDisassociateIAMRole) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
func waitAssociateDisassociateIAMRoleCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*awstypes.AssociateDisassociateIAMRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusAssociateDisassociateIAMRole(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odb.AssociateDisassociateIAMRole); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitAssociateDisassociateIAMRoleUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*awstypes.AssociateDisassociateIAMRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusAssociateDisassociateIAMRole(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odb.AssociateDisassociateIAMRole); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitAssociateDisassociateIAMRoleDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*awstypes.AssociateDisassociateIAMRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusAssociateDisassociateIAMRole(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odb.AssociateDisassociateIAMRole); ok {
		return out, err
	}

	return nil, err
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusAssociateDisassociateIAMRole(ctx context.Context, conn *odb.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findAssociateDisassociateIAMRoleByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

func findAssociateDisassociateIAMRole(ctx context.Context, conn *odb.Client, resourceARN *string, roleARN *string) (*iamRoleResourceInternal, error) {
	arn, err := arn.Parse(*resourceARN)
	if err != nil {
		return nil, err
	}
	switch arn.Resource {
	case "cloud_vm_cluster":
		input := odb.GetCloudVmClusterInput{
			CloudVmClusterId: resourceARN,
		}
		out, err := conn.GetCloudVmCluster(ctx, &input)
		if err != nil {
			return nil, err
		}
		iamRolesList := out.CloudVmCluster.IamRoles

		for _, element := range iamRolesList {
			if *element.IamRoleArn == *roleARN {
				//we found the correct role
				var iamRole iamRoleResourceInternal
				iamRole.iamRoleArn = element.IamRoleArn
				iamRole.awsIntegration = element.AwsIntegration
				iamRole.resourceARN = resourceARN
				iamRole.statusReason = element.StatusReason
				iamRole.status = element.Status
				return &iamRole, nil
			}
		}
		return nil, errors.New("no IAM role found for the vm cluster : " + *resourceARN)

	case "cloud_autonomous_vm_cluster":
		input := odb.GetCloudAutonomousVmClusterInput{
			CloudAutonomousVmClusterId: resourceARN,
		}
		out, err := conn.GetCloudAutonomousVmCluster(ctx, &input)
		if err != nil {
			return nil, err
		}
		for _, element := range out.CloudAutonomousVmCluster.IamRoles {
			if *element.IamRoleArn == *roleARN {
				//We found a match
				var iamRole iamRoleResourceInternal
				iamRole.iamRoleArn = element.IamRoleArn
				iamRole.awsIntegration = element.AwsIntegration
				iamRole.resourceARN = resourceARN
				iamRole.statusReason = element.StatusReason
				iamRole.status = element.Status
				return &iamRole, nil
			}
		}

		return nil, errors.New("no IAM role found for the cloud autonomous vm cluster : " + *resourceARN)
	}
	return nil, errors.New("IAM role association / disassociation not supported : " + *resourceARN)
}

type resourceAssociateDisassociateIAMRoleResourceModel struct {
	framework.WithRegionModel
	IAMRoleARN     types.String   `tfsdk:"iam_role_arn"`
	AWSIntegration types.String   `tfsdk:"aws_integration"`
	ResourceARN    types.String   `tfsdk:"resource_arn"`
	Status         types.String   `tfsdK:"status"`
	StatusReason   types.String   `tfsdk:"status_reason"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}
type iamRoleResourceInternal struct {

	// The Amazon Web Services integration configuration settings for the Amazon Web
	// Services Identity and Access Management (IAM) service role.
	awsIntegration odbtypes.SupportedAwsIntegration

	// The Amazon Resource Name (ARN) of the Amazon Web Services Identity and Access
	// Management (IAM) service role.
	iamRoleArn *string

	// The current status of the Amazon Web Services Identity and Access Management
	// (IAM) service role.
	status odbtypes.IamRoleStatus

	// Additional information about the current status of the Amazon Web Services
	// Identity and Access Management (IAM) service role, if applicable.
	statusReason *string
	//ARN of the resource for which the IAM role is configured.
	resourceARN *string
}
*/
