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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// @FrameworkResource("aws_odb_associate_disassociate_iam_role", name="Associate Disassociate IAM Role")
func newResourceAssociateDisassociateIAMRole(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAssociateDisassociateIAMRole{}

	r.SetDefaultCreateTimeout(36 * time.Hour)
	r.SetDefaultUpdateTimeout(36 * time.Hour)
	r.SetDefaultDeleteTimeout(36 * time.Hour)

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
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
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
	var combinedARNs []resourceCompositeARNModel
	plan.CompositeARN.ElementsAs(ctx, &combinedARNs, false)
	var input odb.AssociateIamRoleToResourceInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AssociateDisassociateIAMRole"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.ResourceArn = combinedARNs[0].ResourceARN.ValueStringPointer()
	input.IamRoleArn = combinedARNs[0].IAMRoleARN.ValueStringPointer()
	out, err := conn.AssociateIamRoleToResource(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAssociateDisassociateIAMRole, plan.CompositeARN.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAssociateDisassociateIAMRole, plan.CompositeARN.String(), nil),
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
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameAssociateDisassociateIAMRole, plan.CompositeARN.String(), err),
			err.Error(),
		)
		return
	}
	iamRoleOut, err := FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, *input.ResourceArn, *input.IamRoleArn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameAssociateDisassociateIAMRole, plan.CompositeARN.String(), err),
			err.Error(),
		)
		return
	}
	plan.Status = fwtypes.StringEnumValue(iamRoleOut.Status)
	plan.StatusReason = types.StringPointerValue(iamRoleOut.StatusReason)
	plan.ComputedId = types.StringValue(plan.CompositeARN.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAssociateDisassociateIAMRole) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state resourceAssociateDisassociateIAMRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var combinedARNs []resourceCompositeARNModel
	resp.Diagnostics.Append(state.CompositeARN.ElementsAs(ctx, &combinedARNs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(combinedARNs) == 0 {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameAssociateDisassociateIAMRole, state.CompositeARN.String(), errors.New("missing composite_arn")),
			"Expected at least one composite_arn entry in state.",
		)
		return
	}

	out, err := FindAssociatedDisassociatedIAMRoleOracleDBResource(ctx, conn, combinedARNs[0].ResourceARN.ValueString(), combinedARNs[0].IAMRoleARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameAssociateDisassociateIAMRole, state.CompositeARN.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compositeARN, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []resourceCompositeARNModel{
		{
			IAMRoleARN:  types.StringPointerValue(out.IamRoleArn),
			ResourceARN: combinedARNs[0].ResourceARN,
		},
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.CompositeARN = compositeARN
	state.ComputedId = types.StringValue(state.CompositeARN.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAssociateDisassociateIAMRole) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state resourceAssociateDisassociateIAMRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var combinedARNs []resourceCompositeARNModel
	resp.Diagnostics.Append(state.CompositeARN.ElementsAs(ctx, &combinedARNs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(combinedARNs) == 0 {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, state.CompositeARN.String(), errors.New("missing composite_arn")),
			"Expected at least one composite_arn entry in state.",
		)
		return
	}

	var input odb.DisassociateIamRoleFromResourceInput
	resp.Diagnostics.Append(flex.Expand(ctx, state, &input, flex.WithFieldNamePrefix("AssociateDisassociateIAMRole"))...)
	input.IamRoleArn = combinedARNs[0].IAMRoleARN.ValueStringPointer()
	input.ResourceArn = combinedARNs[0].ResourceARN.ValueStringPointer()
	output, err := conn.DisassociateIamRoleFromResource(ctx, &input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, state.CompositeARN.String(), err),
			err.Error(),
		)
		return
	}
	if output == nil {
		err = errors.New("disassociate IAM role returning nil response  : " + state.CompositeARN.String())
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, state.CompositeARN.String(), err),
			err.Error(),
		)
		return
	}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAssociateDisassociateIAMRoleDeleted(ctx, conn, combinedARNs[0].ResourceARN.ValueStringPointer(), combinedARNs[0].IAMRoleARN.ValueStringPointer(), deleteTimeout)
	if err != nil {
		err = errors.New("disassociate IAM role returning nil response  : " + state.CompositeARN.String())
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAssociateDisassociateIAMRole, state.CompositeARN.String(), err),
			err.Error(),
		)
		return
	}
}
func (r *resourceAssociateDisassociateIAMRole) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	const (
		keyIAMRoleARN  = "iam_role_arn"
		keyResourceARN = "resource_arn"
	)

	kvPairs := strings.Split(req.ID, ",")
	if len(kvPairs) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Expected import identifier in the format iam_role_arn=<value>,resource_arn=<value>.",
		)
		return
	}

	values := map[string]string{}
	for _, kvPair := range kvPairs {
		parts := strings.SplitN(strings.TrimSpace(kvPair), "=", 2)
		if len(parts) != 2 {
			resp.Diagnostics.AddError(
				"Invalid import identifier",
				"Expected key-value pairs in the format key=value.",
			)
			return
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case keyIAMRoleARN, keyResourceARN:
			if value == "" {
				resp.Diagnostics.AddError(
					"Invalid import identifier",
					"Import value for key "+key+" must be non-empty.",
				)
				return
			}
			values[key] = value
		default:
			resp.Diagnostics.AddError(
				"Invalid import identifier",
				"Unsupported key "+key+". Supported keys are iam_role_arn and resource_arn.",
			)
			return
		}
	}

	iamRoleARN, hasIAMRoleARN := values[keyIAMRoleARN]
	resourceARN, hasResourceARN := values[keyResourceARN]
	if !hasIAMRoleARN || !hasResourceARN {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Both iam_role_arn and resource_arn must be specified in the import identifier.",
		)
		return
	}

	compositeARN, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []resourceCompositeARNModel{
		{
			IAMRoleARN:  types.StringValue(iamRoleARN),
			ResourceARN: types.StringValue(resourceARN),
		},
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("composite_arn"), compositeARN)...)
}

func waitAssociateDisassociateIAMRoleCreated(ctx context.Context, conn *odb.Client, resourceARN *string, iamRoleARN *string, timeout time.Duration) (*odb.AssociateIamRoleToResourceOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(odbtypes.IamRoleStatusAssociating),
		Target:  enum.Slice(odbtypes.IamRoleStatusFailed, odbtypes.IamRoleStatusConnected),
		Refresh: statusAssociateDisassociateIAMRole(ctx, conn, resourceARN, iamRoleARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odb.AssociateIamRoleToResourceOutput); ok {
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
	CompositeARN   fwtypes.ListNestedObjectValueOf[resourceCompositeARNModel] `tfsdk:"composite_arn" noexpand:"true" noflatten:"true"`
	AWSIntegration types.String                                               `tfsdk:"aws_integration"`
	Status         fwtypes.StringEnum[odbtypes.IamRoleStatus]                 `tfsdk:"status"`
	StatusReason   types.String                                               `tfsdk:"status_reason"`
	ComputedId     types.String                                               `tfsdk:"id" noflatten:"true"`
	Timeouts       timeouts.Value                                             `tfsdk:"timeouts"`
}

// Composite ID for IAM role resource
type resourceCompositeARNModel struct {
	IAMRoleARN  types.String `tfsdk:"iam_role_arn"`
	ResourceARN types.String `tfsdk:"resource_arn"`
}
