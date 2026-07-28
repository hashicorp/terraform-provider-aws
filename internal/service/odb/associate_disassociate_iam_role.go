// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var AssociateDisassociateIAMRole = newResourceAssociateDisassociateIAMRole

// @FrameworkResource("aws_odb_iam_role_association", name="IAM Role Association")
func newResourceAssociateDisassociateIAMRole(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAssociateDisassociateIAMRole{}

	r.SetDefaultCreateTimeout(1 * time.Hour)
	r.SetDefaultUpdateTimeout(1 * time.Hour)
	r.SetDefaultDeleteTimeout(1 * time.Hour)

	return r, nil
}

const (
	ResNameAssociateDisassociateIAMRole = "IAM Role Association"
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
			names.AttrIAMRoleARN: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrResourceARN: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[odbtypes.IamRoleStatus](),
				Computed:   true,
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
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
	associationDescription := iamRoleAssociationDescription(plan.IAMRoleARN, plan.ResourceARN)
	var input odb.AssociateIamRoleToResourceInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AssociateDisassociateIAMRole"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := conn.AssociateIamRoleToResource(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAssociateDisassociateIAMRole, associationDescription, err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAssociateDisassociateIAMRole, associationDescription, nil),
			errors.New("empty output").Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	iamRoleOut, err := waitAssociateDisassociateIAMRoleCreated(ctx, conn, input.ResourceArn, input.IamRoleArn, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameAssociateDisassociateIAMRole, associationDescription, err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, iamRoleOut, &plan)...)
	if resp.Diagnostics.HasError() {
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
	associationDescription := iamRoleAssociationDescription(state.IAMRoleARN, state.ResourceARN)
	out, err := FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, state.ResourceARN.ValueString(), state.IAMRoleARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameAssociateDisassociateIAMRole, associationDescription, err),
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
	associationDescription := iamRoleAssociationDescription(state.IAMRoleARN, state.ResourceARN)
	var input odb.DisassociateIamRoleFromResourceInput
	resp.Diagnostics.Append(flex.Expand(ctx, state, &input, flex.WithFieldNamePrefix("AssociateDisassociateIAMRole"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	output, err := conn.DisassociateIamRoleFromResource(ctx, &input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, associationDescription, err),
			err.Error(),
		)
		return
	}
	if output == nil {
		err = errors.New("disassociate IAM role returning nil response")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, associationDescription, err),
			err.Error(),
		)
		return
	}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAssociateDisassociateIAMRoleDeleted(ctx, conn, state.ResourceARN.ValueStringPointer(), state.IAMRoleARN.ValueStringPointer(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, associationDescription, err),
			err.Error(),
		)
		return
	}
}
func (r *resourceAssociateDisassociateIAMRole) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ODB resource ARNs cannot contain commas, but IAM role ARNs can.
	separatorIndex := strings.LastIndex(req.ID, ",")
	if separatorIndex == -1 {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Expected import identifier in the format <iam_role_arn>,<resource_arn>.",
		)
		return
	}

	iamRoleARN := strings.TrimSpace(req.ID[:separatorIndex])
	resourceARN := strings.TrimSpace(req.ID[separatorIndex+1:])
	if iamRoleARN == "" || resourceARN == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Both IAM role ARN and resource ARN must be non-empty.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrIAMRoleARN), iamRoleARN)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrResourceARN), resourceARN)...)
}

func waitAssociateDisassociateIAMRoleCreated(ctx context.Context, conn *odb.Client, resourceARN *string, iamRoleARN *string, timeout time.Duration) (*odbtypes.IamRole, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(odbtypes.IamRoleStatusAssociating),
		Target:  enum.Slice(odbtypes.IamRoleStatusFailed, odbtypes.IamRoleStatusConnected),
		Refresh: statusAssociateDisassociateIAMRole(ctx, conn, resourceARN, iamRoleARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.IamRole); ok {
		return out, err
	}

	return nil, err
}

func waitAssociateDisassociateIAMRoleDeleted(ctx context.Context, conn *odb.Client, resourceARN *string, iamEoleARN *string, timeout time.Duration) (*odb.DisassociateIamRoleFromResourceOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
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

func statusAssociateDisassociateIAMRole(ctx context.Context, conn *odb.Client, resourceARN *string, roleARN *string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, *resourceARN, *roleARN)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

func FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx context.Context, conn *odb.Client, resourceARN string, roleARN string) (*odbtypes.IamRole, error) {
	parsedResourceARN, err := arn.Parse(resourceARN)
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
			if aws.ToString(element.IamRoleArn) == roleARN {
				//we found the correct role
				return &element, nil
			}
		}
		err = errors.New("no IAM role found for the vm cluster : " + resourceARN)
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
			if aws.ToString(element.IamRoleArn) == roleARN {
				//We found a match
				return &element, nil
			}
		}
		err = errors.New("no IAM role found for the cloud autonomous vm cluster : " + resourceARN)
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}
	}
	return nil, errors.New("IAM role association / disassociation not supported : " + resourceARN)
}

type resourceAssociateDisassociateIAMRoleResourceModel struct {
	framework.WithRegionModel
	AWSIntegration types.String                               `tfsdk:"aws_integration"`
	IAMRoleARN     types.String                               `tfsdk:"iam_role_arn"`
	ResourceARN    types.String                               `tfsdk:"resource_arn"`
	Status         fwtypes.StringEnum[odbtypes.IamRoleStatus] `tfsdk:"status"`
	StatusReason   types.String                               `tfsdk:"status_reason"`
	Timeouts       timeouts.Value                             `tfsdk:"timeouts"`
}

func iamRoleAssociationDescription(iamRoleARN, resourceARN types.String) string {
	return "IAM role " + iamRoleARN.ValueString() + " for resource " + resourceARN.ValueString()
}
