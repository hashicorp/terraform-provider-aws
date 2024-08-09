// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceVPCEndpoint(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := resourceVpcEndpoint{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return &r, nil
}

type resourceVpcEndpointData struct {
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	SecurityGroupIds types.Set      `tfsdk:"security_group_ids"`
	SubnetIds        types.Set      `tfsdk:"subnet_ids"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
	VpcId            types.String   `tfsdk:"vpc_id"`
}

const (
	ResNameVPCEndpoint = "VPC Endpoint"
)

type resourceVpcEndpoint struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceVpcEndpoint) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_opensearchserverless_vpc_endpoint"
}

func (r *resourceVpcEndpoint) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 5),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 6),
				},
			},
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
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

func (r *resourceVpcEndpoint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceVpcEndpointData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	in := &opensearchserverless.CreateVpcEndpointInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(plan.Name.ValueString()),
		SubnetIds:   flex.ExpandFrameworkStringValueSet(ctx, plan.SubnetIds),
		VpcId:       aws.String(plan.VpcId.ValueString()),
	}

	if !plan.SecurityGroupIds.IsNull() && !plan.SecurityGroupIds.IsUnknown() {
		in.SecurityGroupIds = flex.ExpandFrameworkStringValueSet(ctx, plan.SecurityGroupIds)
	}

	out, err := conn.CreateVpcEndpoint(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionCreating, ResNameVPCEndpoint, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	if _, err := waitVPCEndpointCreated(ctx, conn, *out.CreateVpcEndpointDetail.Id, createTimeout); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionWaitingForCreation, ResNameVPCEndpoint, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	// The create operation only returns the Id and Name so retrieve the newly
	// created VPC Endpoint so we can store the possibly computed
	// security_group_ids in state
	vpcEndpoint, err := findVPCEndpointByID(ctx, conn, aws.ToString(out.CreateVpcEndpointDetail.Id))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionChecking, ResNameVPCEndpoint, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	state.refreshFromOutput(ctx, vpcEndpoint)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVpcEndpoint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state resourceVpcEndpointData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCEndpointByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, ResNameVPCEndpoint, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.refreshFromOutput(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVpcEndpoint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	update := false

	var plan, state resourceVpcEndpointData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &opensearchserverless.UpdateVpcEndpointInput{
		ClientToken: aws.String(id.UniqueId()),
		Id:          aws.String(plan.ID.ValueString()),
	}

	if !plan.SecurityGroupIds.Equal(state.SecurityGroupIds) {
		newSGs := flex.ExpandFrameworkStringValueSet(ctx, plan.SecurityGroupIds)
		oldSGs := flex.ExpandFrameworkStringValueSet(ctx, state.SecurityGroupIds)

		if add := newSGs.Difference(oldSGs); len(add) > 0 {
			input.AddSecurityGroupIds = add
		}

		if del := oldSGs.Difference(newSGs); len(del) > 0 {
			input.RemoveSecurityGroupIds = del
		}

		update = true
	}

	if !plan.SubnetIds.Equal(state.SubnetIds) {
		old := flex.ExpandFrameworkStringValueSet(ctx, state.SubnetIds)
		new := flex.ExpandFrameworkStringValueSet(ctx, plan.SubnetIds)

		if add := new.Difference(old); len(add) > 0 {
			input.AddSubnetIds = add
		}

		if del := old.Difference(new); len(del) > 0 {
			input.RemoveSubnetIds = del
		}

		update = true
	}

	if !update {
		return
	}

	log.Printf("[DEBUG] Updating OpenSearchServerless VPC Endpoint (%s): %#v", plan.ID.ValueString(), input)
	out, err := conn.UpdateVpcEndpoint(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("updating VPC Endpoint (%s)", plan.ID.ValueString()), err.Error())
		return
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	if _, err := waitVPCEndpointUpdated(ctx, conn, *out.UpdateVpcEndpointDetail.Id, updateTimeout); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionWaitingForUpdate, ResNameVPCEndpoint, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	// The update operation only returns security_group_ids if those were
	// changed so retrieve the updated VPC Endpoint so we can store the
	// actual security_group_ids in state
	vpcEndpoint, err := findVPCEndpointByID(ctx, conn, *out.UpdateVpcEndpointDetail.Id)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionChecking, ResNameVPCEndpoint, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	plan.refreshFromOutput(ctx, vpcEndpoint)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVpcEndpoint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state resourceVpcEndpointData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteVpcEndpoint(ctx, &opensearchserverless.DeleteVpcEndpointInput{
		ClientToken: aws.String(id.UniqueId()),
		Id:          aws.String(state.ID.ValueString()),
	})

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionDeleting, ResNameVPCEndpoint, state.Name.String(), nil),
			err.Error(),
		)
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if _, err := waitVPCEndpointDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionWaitingForDeletion, ResNameVPCEndpoint, state.Name.String(), nil),
			err.Error(),
		)
		return
	}
}

func (r *resourceVpcEndpoint) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

// refreshFromOutput writes state data from an AWS response object
func (rd *resourceVpcEndpointData) refreshFromOutput(ctx context.Context, out *awstypes.VpcEndpointDetail) {
	if out == nil {
		return
	}

	rd.ID = flex.StringToFramework(ctx, out.Id)
	rd.Name = flex.StringToFramework(ctx, out.Name)
	rd.SecurityGroupIds = flex.FlattenFrameworkStringValueSet(ctx, out.SecurityGroupIds)
	rd.SubnetIds = flex.FlattenFrameworkStringValueSet(ctx, out.SubnetIds)
	rd.VpcId = flex.StringToFramework(ctx, out.VpcId)
}

func waitVPCEndpointCreated(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*awstypes.VpcEndpointDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VpcEndpointStatusPending),
		Target:                    enum.Slice(awstypes.VpcEndpointStatusActive),
		Refresh:                   statusVPCEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.VpcEndpointDetail); ok {
		return out, err
	}

	return nil, err
}

func waitVPCEndpointUpdated(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*awstypes.VpcEndpointDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VpcEndpointStatusPending),
		Target:                    enum.Slice(awstypes.VpcEndpointStatusActive),
		Refresh:                   statusVPCEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.VpcEndpointDetail); ok {
		return out, err
	}

	return nil, err
}

func waitVPCEndpointDeleted(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*awstypes.VpcEndpointDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpcEndpointStatusDeleting, awstypes.VpcEndpointStatusActive),
		Target:  []string{},
		Refresh: statusVPCEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.VpcEndpointDetail); ok {
		return out, err
	}

	return nil, err
}

func statusVPCEndpoint(ctx context.Context, conn *opensearchserverless.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findVPCEndpointByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}
