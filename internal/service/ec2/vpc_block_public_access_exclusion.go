// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_block_public_access_exclusion", name="Block Public Access Exclusion")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=true)
func newResourceVPCBlockPublicAccessExclusion(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCBlockPublicAccessExclusion{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCBlockPublicAccessExclusion = "VPC Block Public Access Exclusion"
)

type resourceVPCBlockPublicAccessExclusion struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceVPCBlockPublicAccessExclusion) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_vpc_block_public_access_exclusion"
}

func (r *resourceVPCBlockPublicAccessExclusion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"creation_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"exclusion_id": schema.StringAttribute{
				Computed: true,
			},
			"internet_gateway_exclusion_mode": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.InternetGatewayExclusionMode](),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"last_update_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"reason": schema.StringAttribute{
				Computed: true,
			},
			names.AttrResourceARN: framework.ARNAttributeComputedOnly(),
			names.AttrSubnetID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *resourceVPCBlockPublicAccessExclusion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCBlockPublicAccessExclusionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &ec2.CreateVpcBlockPublicAccessExclusionInput{
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpcBlockPublicAccessExclusion),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)

	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateVpcBlockPublicAccessExclusion(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCBlockPublicAccessExclusion, plan.ExclusionID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.VpcBlockPublicAccessExclusion == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCBlockPublicAccessExclusion, plan.ExclusionID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.VpcBlockPublicAccessExclusion, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.setID()

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitVPCBlockPublicAccessExclusionCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCBlockPublicAccessExclusion, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TODO: Read again because the LastUpdateTimeStamp is not provided in the output of the Create Call. At the time of release, if this is changed, then remove this part and uncomment the part above  where the output of create is used to update the plan.
	desc_out, desc_err := FindVPCBlockPublicAccessExclusionByID(ctx, conn, plan.ID.ValueString())
	if tfresource.NotFound(desc_err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if desc_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameVPCBlockPublicAccessExclusion, plan.ID.String(), desc_err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, desc_out, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVPCBlockPublicAccessExclusion) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func (r *resourceVPCBlockPublicAccessExclusion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCBlockPublicAccessExclusionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := state.InitFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	out, err := FindVPCBlockPublicAccessExclusionByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameVPCBlockPublicAccessExclusion, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract VPC Id and Subnet Id to support import
	resource_arn := state.ResourceARN.ValueString()

	arn, err := arn.Parse(resource_arn)
	if err != nil {
		resp.Diagnostics.AddError("Parsing Resource ARN", err.Error())
		return
	}

	if strings.HasPrefix(arn.Resource, "vpc/") {
		vpc_id := strings.TrimPrefix(arn.Resource, "vpc/")
		state.VPCID = types.StringValue(vpc_id)
	} else if strings.HasPrefix(arn.Resource, "subnet/") {
		subnet_id := strings.TrimPrefix(arn.Resource, "subnet/")
		state.SubnetID = types.StringValue(subnet_id)
	} else {
		resp.Diagnostics.AddError("Parsing Resource_ARN", fmt.Sprintf("Unknown resource type: %s", arn.Resource))
		return
	}

	setTagsOut(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVPCBlockPublicAccessExclusion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan, state resourceVPCBlockPublicAccessExclusionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.InternetGatewayExclusionMode.Equal(state.InternetGatewayExclusionMode) {
		input := ec2.ModifyVpcBlockPublicAccessExclusionInput{
			ExclusionId:                  state.ID.ValueStringPointer(),
			InternetGatewayExclusionMode: awstypes.InternetGatewayExclusionMode(plan.InternetGatewayExclusionMode.ValueString()),
		}

		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.ModifyVpcBlockPublicAccessExclusion(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionUpdating, ResNameVPCBlockPublicAccessExclusion, plan.ExclusionID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.VpcBlockPublicAccessExclusion == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionUpdating, ResNameVPCBlockPublicAccessExclusion, plan.ExclusionID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitVPCBlockPublicAccessExclusionUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForUpdate, ResNameVPCBlockPublicAccessExclusion, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	out, err := FindVPCBlockPublicAccessExclusionByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameVPCBlockPublicAccessExclusion, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVPCBlockPublicAccessExclusion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCBlockPublicAccessExclusionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DeleteVpcBlockPublicAccessExclusionInput{
		ExclusionId: state.ExclusionID.ValueStringPointer(),
	}

	_, err := conn.DeleteVpcBlockPublicAccessExclusion(ctx, &input)
	if err != nil {
		if !tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "is in delete-complete state and cannot be deleted") {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCBlockPublicAccessExclusion, state.ID.String(), err),
				err.Error(),
			)
		}
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCBlockPublicAccessExclusionDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameVPCBlockPublicAccessExclusion, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCBlockPublicAccessExclusion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitVPCBlockPublicAccessExclusionCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcBlockPublicAccessExclusion, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VpcBlockPublicAccessExclusionStateCreateInProgress),
		Target:                    enum.Slice(awstypes.VpcBlockPublicAccessExclusionStateCreateComplete),
		Refresh:                   statusVPCBlockPublicAccessExclusion(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            2,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(awstypes.VpcBlockPublicAccessExclusion); ok {
		return &out, err
	}

	return nil, err
}

func waitVPCBlockPublicAccessExclusionUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcBlockPublicAccessExclusion, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpcBlockPublicAccessExclusionStateUpdateInProgress),
		Target: enum.Slice(awstypes.VpcBlockPublicAccessExclusionStateUpdateComplete,
			awstypes.VpcBlockPublicAccessExclusionStateCreateComplete),
		Refresh:                   statusVPCBlockPublicAccessExclusion(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(awstypes.VpcBlockPublicAccessExclusion); ok {
		return &out, err
	}

	return nil, err
}

func waitVPCBlockPublicAccessExclusionDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcBlockPublicAccessExclusion, error) {
	stateConf := &retry.StateChangeConf{
		// There might API inconsistencies where even after invoking delete, the Describe might come back with a CreateComplete or UpdateComplete status (the status before delete was invoked). To account for that, we are also adding those two statuses as valid statues to retry.
		Pending: enum.Slice(awstypes.VpcBlockPublicAccessExclusionStateUpdateComplete,
			awstypes.VpcBlockPublicAccessExclusionStateCreateComplete,
			awstypes.VpcBlockPublicAccessExclusionStateDeleteInProgress),
		Target:                    enum.Slice(awstypes.VpcBlockPublicAccessExclusionStateDeleteComplete),
		Refresh:                   statusVPCBlockPublicAccessExclusion(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            1,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(awstypes.VpcBlockPublicAccessExclusion); ok {
		return &out, err
	}

	return nil, err
}

func statusVPCBlockPublicAccessExclusion(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindVPCBlockPublicAccessExclusionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return &out, string(out.State), nil
	}
}

func FindVPCBlockPublicAccessExclusionByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcBlockPublicAccessExclusion, error) {
	in := &ec2.DescribeVpcBlockPublicAccessExclusionsInput{
		ExclusionIds: []string{id},
	}

	out, err := conn.DescribeVpcBlockPublicAccessExclusions(ctx, in)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCBlockPublicAccessExclusionID) {
		return nil, &retry.NotFoundError{
			Message:     "Exclusion Id:" + id + " Not Found",
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.VpcBlockPublicAccessExclusions == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &(out.VpcBlockPublicAccessExclusions[0]), nil
}

type resourceVPCBlockPublicAccessExclusionModel struct {
	CreationTimestamp            timetypes.RFC3339 `tfsdk:"creation_timestamp"`
	ExclusionID                  types.String      `tfsdk:"exclusion_id"`
	ID                           types.String      `tfsdk:"id"`
	InternetGatewayExclusionMode types.String      `tfsdk:"internet_gateway_exclusion_mode"`
	LastUpdateTimestamp          timetypes.RFC3339 `tfsdk:"last_update_timestamp"`
	Reason                       types.String      `tfsdk:"reason"`
	ResourceARN                  types.String      `tfsdk:"resource_arn"`
	SubnetID                     types.String      `tfsdk:"subnet_id"`
	Timeouts                     timeouts.Value    `tfsdk:"timeouts"`
	Tags                         tftags.Map        `tfsdk:"tags"`
	TagsAll                      tftags.Map        `tfsdk:"tags_all"`
	VPCID                        types.String      `tfsdk:"vpc_id"`
}

func (data *resourceVPCBlockPublicAccessExclusionModel) InitFromID() error {
	data.ExclusionID = data.ID
	return nil
}

func (data *resourceVPCBlockPublicAccessExclusionModel) setID() {
	data.ID = data.ExclusionID
}
