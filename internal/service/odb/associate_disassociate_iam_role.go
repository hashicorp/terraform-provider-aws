// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var AssociateDisassociateIAMRole = newResourceAssociateDisassociateIAMRole

// @FrameworkResource("aws_odb_associate_disassociate_iam_role", name="Associate Disassociate IAM Role")
func newResourceAssociateDisassociateIAMRole(context.Context) (resource.ResourceWithConfigure, error) {
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
			"aws_integration": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[odbtypes.IamRoleStatus](),
				Computed:   true,
			},
			"status_reason": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"composite_arn": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceCompositeARNModel](ctx),
				Validators: []validator.List{
					// Only one combination of resource ARN and IAM role ARN is mandatory
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
					listvalidator.IsRequired(),
				},
				Description: "Combination of resource ARN and IAM role ARN is mandatory",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"iam_role_arn": schema.StringAttribute{
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
					},
				},
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
	var combinedARN resourceCompositeARNModel
	flex.Flatten(ctx, &combinedARN, plan.IAMRoleResourceCombinedARN)
	var input odb.AssociateIamRoleToResourceInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AssociateDisassociateIAMRole"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.ResourceArn = combinedARN.ResourceARN.ValueStringPointer()
	input.IamRoleArn = combinedARN.IAMRoleARN.ValueStringPointer()
	out, err := conn.AssociateIamRoleToResource(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAssociateDisassociateIAMRole, plan.IAMRoleResourceCombinedARN.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAssociateDisassociateIAMRole, plan.IAMRoleResourceCombinedARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitAssociateDisassociateIAMRoleCreated(ctx, conn, input.ResourceArn, input.IamRoleArn, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameAssociateDisassociateIAMRole, plan.IAMRoleResourceCombinedARN.String(), err),
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
	var combinedARN resourceCompositeARNModel
	flex.Flatten(ctx, &combinedARN, state.IAMRoleResourceCombinedARN)
	out, err := FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, combinedARN.ResourceARN.ValueStringPointer(), combinedARN.ResourceARN.ValueStringPointer())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameAssociateDisassociateIAMRole, state.IAMRoleResourceCombinedARN.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAssociateDisassociateIAMRole) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state resourceAssociateDisassociateIAMRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var combinedARN resourceCompositeARNModel
	flex.Flatten(ctx, &combinedARN, state.IAMRoleResourceCombinedARN)
	var input odb.DisassociateIamRoleFromResourceInput
	resp.Diagnostics.Append(flex.Expand(ctx, state, &input, flex.WithFieldNamePrefix("AssociateDisassociateIAMRole"))...)
	input.IamRoleArn = combinedARN.IAMRoleARN.ValueStringPointer()
	input.ResourceArn = combinedARN.ResourceARN.ValueStringPointer()
	output, err := conn.DisassociateIamRoleFromResource(ctx, &input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, state.IAMRoleResourceCombinedARN.String(), err),
			err.Error(),
		)
		return
	}
	if output == nil {
		err = errors.New("disassociate IAM role returning nil response  : " + state.IAMRoleResourceCombinedARN.String())
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, state.IAMRoleResourceCombinedARN.String(), err),
			err.Error(),
		)
		return
	}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAssociateDisassociateIAMRoleDeleted(ctx, conn, combinedARN.ResourceARN.ValueStringPointer(), combinedARN.IAMRoleARN.ValueStringPointer(), deleteTimeout)
	if err != nil {
		err = errors.New("disassociate IAM role returning nil response  : " + state.IAMRoleResourceCombinedARN.String())
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, state.IAMRoleResourceCombinedARN.String(), err),
			err.Error(),
		)
		return
	}
}
func (r *resourceAssociateDisassociateIAMRole) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitAssociateDisassociateIAMRoleCreated(ctx context.Context, conn *odb.Client, resourceARN *string, iamRoleARN *string, timeout time.Duration) (*odb.AssociateIamRoleToResourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(odbtypes.IamRoleStatusAssociating),
		Target:                    enum.Slice(odbtypes.IamRoleStatusFailed, odbtypes.IamRoleStatusConnected),
		Refresh:                   statusAssociateDisassociateIAMRole(ctx, conn, resourceARN, iamRoleARN),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odb.AssociateIamRoleToResourceOutput); ok {
		return out, err
	}

	return nil, err
}

func waitAssociateDisassociateIAMRoleDeleted(ctx context.Context, conn *odb.Client, resourceARN *string, iamEoleARN *string, timeout time.Duration) (*odb.DisassociateIamRoleFromResourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.IamRoleStatusDisassociating),
		Target:  []string{},
		Refresh: statusAssociateDisassociateIAMRole(ctx, conn, resourceARN, iamEoleARN),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odb.DisassociateIamRoleFromResourceOutput); ok {
		return out, err
	}
	return nil, err
}

func statusAssociateDisassociateIAMRole(ctx context.Context, conn *odb.Client, resourceARN *string, roleARN *string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, resourceARN, roleARN)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

func FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx context.Context, conn *odb.Client, resourceARN *string, roleARN *string) (*odbtypes.IamRole, error) {
	parsedResourceARN, err := arn.Parse(*resourceARN)
	if err != nil {
		return nil, err
	}
	resourceType := strings.Split(parsedResourceARN.Resource, "/")[0]
	resourceId := strings.Split(parsedResourceARN.Resource, "/")[1]
	switch resourceType {
	case "cloud-vm-cluster":
		input := odb.GetCloudVmClusterInput{
			CloudVmClusterId: &resourceId,
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
				return &element, nil
			}
		}
		err = errors.New("no IAM role found for the vm cluster : " + *resourceARN)
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}

	case "cloud-autonomous-vm-cluster":
		input := odb.GetCloudAutonomousVmClusterInput{
			CloudAutonomousVmClusterId: &resourceId,
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
				return &element, nil
			}
		}
		err = errors.New("no IAM role found for the cloud autonomous vm cluster : " + *resourceARN)
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}
	}
	return nil, errors.New("IAM role association / disassociation not supported : " + *resourceARN)
}

type resourceAssociateDisassociateIAMRoleResourceModel struct {
	framework.WithRegionModel
	IAMRoleResourceCombinedARN fwtypes.ListNestedObjectValueOf[resourceCompositeARNModel] `tfsdk:"composite_arn"`
	AWSIntegration             types.String                                               `tfsdk:"aws_integration"`
	Status                     fwtypes.StringEnum[odbtypes.IamRoleStatus]                 `tfsdk:"status"`
	StatusReason               types.String                                               `tfsdk:"status_reason"`
	Timeouts                   timeouts.Value                                             `tfsdk:"timeouts"`
}

// Composite ID for IAM role resource
type resourceCompositeARNModel struct {
	IAMRoleARN  types.String `tfsdk:"iam_role_arn"`
	ResourceARN types.String `tfsdk:"resource_arn"`
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
