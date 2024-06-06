// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="VPC Connection")
// @Tags(identifierAttribute="arn")
func newResourceVPCConnection(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCConnection{}
	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type resourceVPCConnection struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceVPCConnection) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_quicksight_vpc_connection"
}

const (
	ResNameVPCConnection = "VPC Connection"
	vpcConnectionIdRegex = "[\\w\\-]+"
	subnetIdRegex        = "^subnet-[0-9a-z]*$"
	securityGroupIdRegex = "^sg-[0-9a-z]*$"
)

func (r *resourceVPCConnection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAWSAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"vpc_connection_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.LengthAtMost(1000),
						stringvalidator.RegexMatches(regexache.MustCompile(vpcConnectionIdRegex), "VPC Connection ID must match regex: "+vpcConnectionIdRegex),
					),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(128),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(20, 2048),
				},
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 16),
					setvalidator.ValueStringsAre(
						stringvalidator.All(
							stringvalidator.LengthAtMost(255),
							stringvalidator.RegexMatches(regexache.MustCompile(securityGroupIdRegex), "Security group ID must match regex: "+securityGroupIdRegex),
						),
					),
				},
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(2, 15),
					setvalidator.ValueStringsAre(
						stringvalidator.All(
							stringvalidator.LengthAtMost(255),
							stringvalidator.RegexMatches(regexache.MustCompile(subnetIdRegex), "Subnet ID must match regex: "+subnetIdRegex),
						),
					),
				},
			},
			"dns_resolvers": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 15),
					setvalidator.ValueStringsAre(
						stringvalidator.All(
							stringvalidator.LengthBetween(7, 15),
						),
					),
				},
			},
			"availability_status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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

func (r *resourceVPCConnection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var plan resourceVPCConnectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID)
	}
	plan.ID = types.StringValue(createVPCConnectionID(plan.AWSAccountID.ValueString(), plan.VPCConnectionID.ValueString()))

	in := &quicksight.CreateVPCConnectionInput{
		AwsAccountId:     aws.String(plan.AWSAccountID.ValueString()),
		VPCConnectionId:  aws.String(plan.VPCConnectionID.ValueString()),
		Name:             aws.String(plan.Name.ValueString()),
		RoleArn:          aws.String(plan.RoleArn.ValueString()),
		SecurityGroupIds: flex.ExpandFrameworkStringSet(ctx, plan.SecurityGroupIds),
		SubnetIds:        flex.ExpandFrameworkStringSet(ctx, plan.SubnetIds),
		Tags:             getTagsIn(ctx),
	}

	if !plan.DnsResolvers.IsNull() {
		in.DnsResolvers = flex.ExpandFrameworkStringSet(ctx, plan.DnsResolvers)
	}

	// account for IAM propagation when attempting to assume role
	out, err := retryVPCConnectionCreate(ctx, conn, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameVPCConnection, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameVPCConnection, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	waitOut, err := waitVPCConnectionCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForCreation, ResNameVPCConnection, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, waitOut.Arn)
	plan.AvailabilityStatus = flex.StringToFramework(ctx, waitOut.AvailabilityStatus)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVPCConnection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceVPCConnectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindVPCConnectionByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, ResNameVPCConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	if aws.StringValue(out.Status) == quicksight.VPCConnectionResourceStatusDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	// To support import, parse the ID for the component keys and set
	// individual values in state
	awsAccountID, vpcConnectionID, err := ParseVPCConnectionID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, ResNameVPCConnection, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.VPCConnectionID = flex.StringValueToFramework(ctx, vpcConnectionID)
	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.RoleArn = flex.StringToFramework(ctx, out.RoleArn)
	state.SecurityGroupIds = flex.FlattenFrameworkStringSet(ctx, out.SecurityGroupIds)
	state.DnsResolvers = flex.FlattenFrameworkStringSet(ctx, out.DnsResolvers)
	state.AvailabilityStatus = flex.StringToFramework(ctx, out.AvailabilityStatus)
	var subnetIds []*string
	for _, iface := range out.NetworkInterfaces {
		subnetIds = append(subnetIds, iface.SubnetId)
	}
	state.SubnetIds = flex.FlattenFrameworkStringSet(ctx, subnetIds)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVPCConnection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var plan, state resourceVPCConnectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.DnsResolvers.Equal(state.DnsResolvers) ||
		!plan.RoleArn.Equal(state.RoleArn) ||
		!plan.SecurityGroupIds.Equal(state.SecurityGroupIds) ||
		!plan.SubnetIds.Equal(state.SubnetIds) {
		in := quicksight.UpdateVPCConnectionInput{
			AwsAccountId:     aws.String(plan.AWSAccountID.ValueString()),
			VPCConnectionId:  aws.String(plan.VPCConnectionID.ValueString()),
			Name:             aws.String(plan.Name.ValueString()),
			RoleArn:          aws.String(plan.RoleArn.ValueString()),
			SecurityGroupIds: flex.ExpandFrameworkStringSet(ctx, plan.SecurityGroupIds),
			SubnetIds:        flex.ExpandFrameworkStringSet(ctx, plan.SubnetIds),
		}

		if !plan.DnsResolvers.IsNull() {
			in.DnsResolvers = flex.ExpandFrameworkStringSet(ctx, plan.DnsResolvers)
		}

		out, err := conn.UpdateVPCConnectionWithContext(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameVPCConnection, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameVPCConnection, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitVPCConnectionUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForUpdate, ResNameVPCConnection, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	}

	// ensure tag only updates are copied into state
	if !plan.Tags.Equal(state.Tags) {
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	}
}

func (r *resourceVPCConnection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceVPCConnectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &quicksight.DeleteVPCConnectionInput{
		AwsAccountId:    aws.String(state.AWSAccountID.ValueString()),
		VPCConnectionId: aws.String(state.VPCConnectionID.ValueString()),
	}

	_, err := conn.DeleteVPCConnectionWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return
		}
		if tfawserr.ErrMessageContains(err, quicksight.ErrCodeConflictException, "Cannot perform operation on deleted VPCConnection") {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameVPCConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCConnectionDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForDeletion, ResNameVPCConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCConnection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceVPCConnection) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func FindVPCConnectionByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.VPCConnection, error) {
	awsAccountID, vpcConnectionId, err := ParseVPCConnectionID(id)
	if err != nil {
		return nil, err
	}

	in := &quicksight.DescribeVPCConnectionInput{
		AwsAccountId:    aws.String(awsAccountID),
		VPCConnectionId: aws.String(vpcConnectionId),
	}

	out, err := conn.DescribeVPCConnectionWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}
	if out == nil || out.VPCConnection == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.VPCConnection, nil
}

func retryVPCConnectionCreate(ctx context.Context, conn *quicksight.QuickSight, in *quicksight.CreateVPCConnectionInput) (*quicksight.CreateVPCConnectionOutput, error) {
	outputRaw, err := tfresource.RetryWhen(ctx,
		iamPropagationTimeout,
		func() (interface{}, error) {
			return conn.CreateVPCConnectionWithContext(ctx, in)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeAccessDeniedException) {
				return true, err
			}

			return false, err
		},
	)

	output, _ := outputRaw.(*quicksight.CreateVPCConnectionOutput)
	return output, err
}

func waitVPCConnectionCreated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.VPCConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			quicksight.VPCConnectionResourceStatusCreationInProgress,
		},
		Target: []string{
			quicksight.VPCConnectionResourceStatusCreationSuccessful,
		},
		Refresh:    statusVPCConnection(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*quicksight.VPCConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPCConnectionUpdated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.VPCConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			quicksight.VPCConnectionResourceStatusUpdateInProgress,
		},
		Target: []string{
			quicksight.VPCConnectionResourceStatusUpdateSuccessful,
		},
		Refresh:    statusVPCConnection(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*quicksight.VPCConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPCConnectionDeleted(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.VPCConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			quicksight.VPCConnectionResourceStatusDeletionInProgress,
		},
		Target: []string{
			quicksight.VPCConnectionResourceStatusDeleted,
		},
		Refresh:    statusVPCConnection(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*quicksight.VPCConnection); ok {
		return output, err
	}

	return nil, err
}

func statusVPCConnection(ctx context.Context, conn *quicksight.QuickSight, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func ParseVPCConnectionID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,VPC_CONNECTION_ID", id)
	}
	return parts[0], parts[1], nil
}

func createVPCConnectionID(awsAccountID, vpcConnectionID string) string {
	return strings.Join([]string{awsAccountID, vpcConnectionID}, ",")
}

type resourceVPCConnectionData struct {
	ID                 types.String   `tfsdk:"id"`
	ARN                types.String   `tfsdk:"arn"`
	AWSAccountID       types.String   `tfsdk:"aws_account_id"`
	VPCConnectionID    types.String   `tfsdk:"vpc_connection_id"`
	Name               types.String   `tfsdk:"name"`
	RoleArn            types.String   `tfsdk:"role_arn"`
	AvailabilityStatus types.String   `tfsdk:"availability_status"`
	SecurityGroupIds   types.Set      `tfsdk:"security_group_ids"`
	SubnetIds          types.Set      `tfsdk:"subnet_ids"`
	DnsResolvers       types.Set      `tfsdk:"dns_resolvers"`
	Tags               types.Map      `tfsdk:"tags"`
	TagsAll            types.Map      `tfsdk:"tags_all"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}
