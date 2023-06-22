// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Target Registration")
func newResourceTargetRegistration(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTargetRegistration{}, nil
}

const (
	ResNameTargetRegistration = "Target Registration"
)

type resourceTargetRegistration struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceTargetRegistration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lb_target_registration"
}

func (r *resourceTargetRegistration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"target_group_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"target": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"availability_zone": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"port": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"target_id": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceTargetRegistration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ELBV2Conn(ctx)

	var plan resourceTargetRegistrationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tfList []targetData
	resp.Diagnostics.Append(plan.Target.ElementsAs(ctx, &tfList, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(plan.TargetGroupARN.ValueString()),
		Targets:        expandTargets(tfList),
	}

	_, err := conn.RegisterTargetsWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ELBV2, create.ErrActionCreating, ResNameTargetRegistration, plan.TargetGroupARN.String(), err),
			err.Error(),
		)
		return
	}

	// Targets must be read after create to set computed attributes
	out, err := findTargetRegistrationByMultipartKey(ctx, conn, plan.TargetGroupARN.ValueString(), tfList)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ELBV2, create.ErrActionReading, ResNameTargetRegistration, plan.TargetGroupARN.String(), err),
			err.Error(),
		)
		return
	}
	targets, d := flattenTargets(ctx, out)
	resp.Diagnostics.Append(d...)
	plan.Target = targets

	plan.ID = types.StringValue(plan.TargetGroupARN.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTargetRegistration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ELBV2Conn(ctx)

	var state resourceTargetRegistrationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tfList []targetData
	resp.Diagnostics.Append(state.Target.ElementsAs(ctx, &tfList, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTargetRegistrationByMultipartKey(ctx, conn, state.TargetGroupARN.ValueString(), tfList)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ELBV2, create.ErrActionSetting, ResNameTargetRegistration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	targets, d := flattenTargets(ctx, out)
	resp.Diagnostics.Append(d...)
	state.Target = targets

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTargetRegistration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ELBV2Conn(ctx)

	var plan, state resourceTargetRegistrationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Target.Equal(state.Target) {
		var stateTargets []targetData
		var planTargets []targetData
		resp.Diagnostics.Append(state.Target.ElementsAs(ctx, &stateTargets, false)...)
		resp.Diagnostics.Append(plan.Target.ElementsAs(ctx, &planTargets, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		stateSet := flex.Set[targetData](stateTargets)
		planSet := flex.Set[targetData](planTargets)

		if dereg := stateSet.Difference(planSet); len(dereg) > 0 {
			in := &elbv2.DeregisterTargetsInput{
				TargetGroupArn: aws.String(state.TargetGroupARN.ValueString()),
				Targets:        expandTargets(dereg),
			}

			_, err := conn.DeregisterTargetsWithContext(ctx, in)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
					return
				}
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.ELBV2, create.ErrActionUpdating, ResNameTargetRegistration, state.ID.String(), err),
					err.Error(),
				)
				return
			}
		}

		if reg := planSet.Difference(stateSet); len(reg) > 0 {
			in := &elbv2.RegisterTargetsInput{
				TargetGroupArn: aws.String(plan.TargetGroupARN.ValueString()),
				Targets:        expandTargets(reg),
			}

			_, err := conn.RegisterTargetsWithContext(ctx, in)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.ELBV2, create.ErrActionUpdating, ResNameTargetRegistration, state.ID.String(), err),
					err.Error(),
				)
				return
			}
		}

		// Targets must be read after update to set computed attributes
		out, err := findTargetRegistrationByMultipartKey(ctx, conn, plan.TargetGroupARN.ValueString(), planTargets)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ELBV2, create.ErrActionReading, ResNameTargetRegistration, plan.TargetGroupARN.String(), err),
				err.Error(),
			)
			return
		}
		targets, d := flattenTargets(ctx, out)
		resp.Diagnostics.Append(d...)
		plan.Target = targets
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTargetRegistration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ELBV2Conn(ctx)

	var state resourceTargetRegistrationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tfList []targetData
	resp.Diagnostics.Append(state.Target.ElementsAs(ctx, &tfList, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(tfList) == 0 {
		// No active targets, nothing to do
		return
	}

	in := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(state.TargetGroupARN.ValueString()),
		Targets:        expandTargets(tfList),
	}

	_, err := conn.DeregisterTargetsWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ELBV2, create.ErrActionDeleting, ResNameTargetRegistration, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findTargetRegistrationByMultipartKey(ctx context.Context, conn *elbv2.ELBV2, targetGroupARN string, targets []targetData) ([]*elbv2.TargetHealthDescription, error) {
	in := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets:        expandTargets(targets),
	}

	out, err := conn.DescribeTargetHealthWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException, elbv2.ErrCodeInvalidTargetException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.TargetHealthDescriptions) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.TargetHealthDescriptions, nil
}

func flattenTargets(ctx context.Context, apiObjects []*elbv2.TargetHealthDescription) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: targetAttrTypes}

	if len(apiObjects) == 0 {
		return types.SetNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		if apiObject == nil || apiObject.Target == nil || apiObject.TargetHealth == nil {
			continue
		}

		// Unregistered targets, or targets in the process of deregistering should not be included
		reason := aws.StringValue(apiObject.TargetHealth.Reason)
		if reason == elbv2.TargetHealthReasonEnumTargetNotRegistered || reason == elbv2.TargetHealthReasonEnumTargetDeregistrationInProgress {
			continue
		}

		target := apiObject.Target
		obj := map[string]attr.Value{
			"availability_zone": flex.StringToFramework(ctx, target.AvailabilityZone),
			"port":              flex.Int64ToFramework(ctx, target.Port),
			"target_id":         flex.StringToFramework(ctx, target.Id),
		}
		objVal, d := types.ObjectValue(targetAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	// check resulting elem length in case none of the returned API objects are
	// actively registered
	if len(elems) == 0 {
		return types.SetNull(elemType), diags
	}

	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func expandTargets(tfList []targetData) []*elbv2.TargetDescription {
	if len(tfList) == 0 {
		return nil
	}

	var apiObject []*elbv2.TargetDescription

	for _, tfObj := range tfList {
		item := &elbv2.TargetDescription{
			Id: aws.String(tfObj.TargetID.ValueString()),
		}
		if !tfObj.AvailabilityZone.IsNull() && !tfObj.AvailabilityZone.IsUnknown() {
			item.AvailabilityZone = aws.String(tfObj.AvailabilityZone.ValueString())
		}
		if !tfObj.Port.IsNull() && !tfObj.Port.IsUnknown() {
			item.Port = aws.Int64(tfObj.Port.ValueInt64())
		}

		apiObject = append(apiObject, item)
	}

	return apiObject
}

type resourceTargetRegistrationData struct {
	ID             types.String `tfsdk:"id"`
	Target         types.Set    `tfsdk:"target"`
	TargetGroupARN types.String `tfsdk:"target_group_arn"`
}

type targetData struct {
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	Port             types.Int64  `tfsdk:"port"`
	TargetID         types.String `tfsdk:"target_id"`
}

var targetAttrTypes = map[string]attr.Type{
	"availability_zone": types.StringType,
	"port":              types.Int64Type,
	"target_id":         types.StringType,
}
