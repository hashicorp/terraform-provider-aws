// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearchserverless_vpc_endpoint", name="VPC Endpoint")
func newVPCEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := vpcEndpointResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return &r, nil
}

const (
	resNameVPCEndpoint = "VPC Endpoint"
)

type vpcEndpointResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *vpcEndpointResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
				ElementType: types.StringType,
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 5),
				},
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				ElementType: types.StringType,
				CustomType:  fwtypes.SetOfStringType,
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

func (r *vpcEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data vpcEndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.CreateVpcEndpointInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(sdkid.UniqueId())

	output, err := conn.CreateVpcEndpoint(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionCreating, resNameVPCEndpoint, data.Name.String(), nil),
			err.Error(),
		)
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.CreateVpcEndpointDetail.Id)

	if _, err := waitVPCEndpointCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionWaitingForCreation, resNameVPCEndpoint, data.Name.String(), nil),
			err.Error(),
		)
		return
	}

	// Security Group IDs are not returned and must be retrieved from the EC2 API.
	vpce, err := tfec2.FindVPCEndpointByID(ctx, r.Meta().EC2Client(ctx), data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, resNameVPCEndpoint, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	var securityGroupIDs []*string
	for _, group := range vpce.Groups {
		securityGroupIDs = append(securityGroupIDs, group.GroupId)
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, securityGroupIDs, &data.SecurityGroupIDs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vpcEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data vpcEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	output, err := findVPCEndpointByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, resNameVPCEndpoint, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Security Group IDs are not returned and must be retrieved from the EC2 API.
	vpce, err := tfec2.FindVPCEndpointByID(ctx, r.Meta().EC2Client(ctx), data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, resNameVPCEndpoint, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	var securityGroupIDs []*string
	for _, group := range vpce.Groups {
		securityGroupIDs = append(securityGroupIDs, group.GroupId)
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, securityGroupIDs, &data.SecurityGroupIDs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vpcEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var new, old vpcEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.UpdateVpcEndpointInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Id:          fwflex.StringFromFramework(ctx, new.ID),
	}

	if !new.SecurityGroupIDs.Equal(old.SecurityGroupIDs) {
		newSGs := fwflex.ExpandFrameworkStringValueSet(ctx, new.SecurityGroupIDs)
		oldSGs := fwflex.ExpandFrameworkStringValueSet(ctx, old.SecurityGroupIDs)

		if add := newSGs.Difference(oldSGs); len(add) > 0 {
			input.AddSecurityGroupIds = add
		}

		if del := oldSGs.Difference(newSGs); len(del) > 0 {
			input.RemoveSecurityGroupIds = del
		}
	}

	if !new.SubnetIDs.Equal(old.SubnetIDs) {
		oldSubnets := fwflex.ExpandFrameworkStringValueSet(ctx, old.SubnetIDs)
		newSubnets := fwflex.ExpandFrameworkStringValueSet(ctx, new.SubnetIDs)

		if add := newSubnets.Difference(oldSubnets); len(add) > 0 {
			input.AddSubnetIds = add
		}

		if del := oldSubnets.Difference(newSubnets); len(del) > 0 {
			input.RemoveSubnetIds = del
		}
	}

	_, err := conn.UpdateVpcEndpoint(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionDeleting, resNameVPCEndpoint, new.Name.String(), nil),
			err.Error(),
		)
		return
	}

	if _, err := waitVPCEndpointUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionWaitingForUpdate, resNameVPCEndpoint, new.Name.String(), nil),
			err.Error(),
		)
		return
	}

	// Security Group IDs are not returned and must be retrieved from the EC2 API.
	vpce, err := tfec2.FindVPCEndpointByID(ctx, r.Meta().EC2Client(ctx), new.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, resNameVPCEndpoint, new.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	var securityGroupIDs []*string
	for _, group := range vpce.Groups {
		securityGroupIDs = append(securityGroupIDs, group.GroupId)
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, securityGroupIDs, &new.SecurityGroupIDs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *vpcEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vpcEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	_, err := conn.DeleteVpcEndpoint(ctx, &opensearchserverless.DeleteVpcEndpointInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Id:          fwflex.StringFromFramework(ctx, data.ID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionDeleting, resNameVPCEndpoint, data.Name.String(), nil),
			err.Error(),
		)
		return
	}

	if _, err := waitVPCEndpointDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionWaitingForDeletion, resNameVPCEndpoint, data.Name.String(), nil),
			err.Error(),
		)
		return
	}
}

type vpcEndpointResourceModel struct {
	ID               types.String                     `tfsdk:"id"`
	Name             types.String                     `tfsdk:"name"`
	SecurityGroupIDs fwtypes.SetValueOf[types.String] `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
	Timeouts         timeouts.Value                   `tfsdk:"timeouts"`
	VPCID            types.String                     `tfsdk:"vpc_id"`
}

func findVPCEndpointByID(ctx context.Context, conn *opensearchserverless.Client, id string) (*awstypes.VpcEndpointDetail, error) {
	input := &opensearchserverless.BatchGetVpcEndpointInput{
		Ids: []string{id},
	}

	return findVPCEndpoint(ctx, conn, input)
}

func findVPCEndpoint(ctx context.Context, conn *opensearchserverless.Client, input *opensearchserverless.BatchGetVpcEndpointInput) (*awstypes.VpcEndpointDetail, error) {
	output, err := findVPCEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpoints(ctx context.Context, conn *opensearchserverless.Client, input *opensearchserverless.BatchGetVpcEndpointInput) ([]awstypes.VpcEndpointDetail, error) {
	output, err := conn.BatchGetVpcEndpoint(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VpcEndpointDetails, nil
}

func statusVPCEndpoint(ctx context.Context, conn *opensearchserverless.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findVPCEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
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

	if output, ok := outputRaw.(*awstypes.VpcEndpointDetail); ok {
		if output.Status == awstypes.VpcEndpointStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCEndpointUpdated(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*awstypes.VpcEndpointDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VpcEndpointStatusPending),
		Target:                    enum.Slice(awstypes.VpcEndpointStatusActive),
		Refresh:                   statusVPCEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcEndpointDetail); ok {
		if output.Status == awstypes.VpcEndpointStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCEndpointDeleted(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*awstypes.VpcEndpointDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VpcEndpointStatusDeleting, awstypes.VpcEndpointStatusActive),
		Target:                    []string{},
		Refresh:                   statusVPCEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 5,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcEndpointDetail); ok {
		if output.Status == awstypes.VpcEndpointStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}
