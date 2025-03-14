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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_vpc_connection", name="VPC Connection")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.VPCConnection")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newVPCConnectionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcConnectionResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type vpcConnectionResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

const (
	resNameVPCConnection = "VPC Connection"
)

func (r *vpcConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					stringvalidator.LengthAtMost(1000),
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
							stringvalidator.RegexMatches(regexache.MustCompile(`^sg-[0-9a-z]*$`), ""),
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
							stringvalidator.RegexMatches(regexache.MustCompile(`^subnet-[0-9a-z]*$`), ""),
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

func (r *vpcConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan resourceVPCConnectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID(ctx))
	}
	awsAccountID, vpcConnectionID := flex.StringValueFromFramework(ctx, plan.AWSAccountID), flex.StringValueFromFramework(ctx, plan.VPCConnectionID)
	in := &quicksight.CreateVPCConnectionInput{
		AwsAccountId:     aws.String(awsAccountID),
		Name:             plan.Name.ValueStringPointer(),
		RoleArn:          plan.RoleArn.ValueStringPointer(),
		SecurityGroupIds: flex.ExpandFrameworkStringValueSet(ctx, plan.SecurityGroupIds),
		SubnetIds:        flex.ExpandFrameworkStringValueSet(ctx, plan.SubnetIds),
		Tags:             getTagsIn(ctx),
		VPCConnectionId:  aws.String(vpcConnectionID),
	}

	if !plan.DnsResolvers.IsNull() {
		in.DnsResolvers = flex.ExpandFrameworkStringValueSet(ctx, plan.DnsResolvers)
	}

	// account for IAM propagation when attempting to assume role
	out, err := retryVPCConnectionCreate(ctx, conn, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameVPCConnection, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameVPCConnection, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, vpcConnectionCreateResourceID(awsAccountID, vpcConnectionID))

	waitOut, err := waitVPCConnectionCreated(ctx, conn, awsAccountID, vpcConnectionID, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForCreation, resNameVPCConnection, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, waitOut.Arn)
	plan.AvailabilityStatus = flex.StringValueToFramework(ctx, waitOut.AvailabilityStatus)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *vpcConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceVPCConnectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, vpcConnectionID, err := vpcConnectionParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameVPCConnection, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	out, err := findVPCConnectionByTwoPartKey(ctx, conn, awsAccountID, vpcConnectionID)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameVPCConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.VPCConnectionID = flex.StringValueToFramework(ctx, vpcConnectionID)
	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.RoleArn = flex.StringToFramework(ctx, out.RoleArn)
	state.SecurityGroupIds = flex.FlattenFrameworkStringValueSet(ctx, out.SecurityGroupIds)
	state.DnsResolvers = flex.FlattenFrameworkStringValueSet(ctx, out.DnsResolvers)
	state.AvailabilityStatus = flex.StringValueToFramework(ctx, out.AvailabilityStatus)
	var subnetIds []*string
	for _, iface := range out.NetworkInterfaces {
		subnetIds = append(subnetIds, iface.SubnetId)
	}
	state.SubnetIds = flex.FlattenFrameworkStringSet(ctx, subnetIds)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vpcConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan, state resourceVPCConnectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, vpcConnectionID, err := vpcConnectionParseResourceID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameVPCConnection, plan.ID.String(), nil),
			err.Error(),
		)
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.DnsResolvers.Equal(state.DnsResolvers) ||
		!plan.RoleArn.Equal(state.RoleArn) ||
		!plan.SecurityGroupIds.Equal(state.SecurityGroupIds) ||
		!plan.SubnetIds.Equal(state.SubnetIds) {
		in := quicksight.UpdateVPCConnectionInput{
			AwsAccountId:     aws.String(awsAccountID),
			Name:             plan.Name.ValueStringPointer(),
			RoleArn:          plan.RoleArn.ValueStringPointer(),
			SecurityGroupIds: flex.ExpandFrameworkStringValueSet(ctx, plan.SecurityGroupIds),
			SubnetIds:        flex.ExpandFrameworkStringValueSet(ctx, plan.SubnetIds),
			VPCConnectionId:  aws.String(vpcConnectionID),
		}

		if !plan.DnsResolvers.IsNull() {
			in.DnsResolvers = flex.ExpandFrameworkStringValueSet(ctx, plan.DnsResolvers)
		}

		out, err := conn.UpdateVPCConnection(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameVPCConnection, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameVPCConnection, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		_, err = waitVPCConnectionUpdated(ctx, conn, awsAccountID, vpcConnectionID, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForUpdate, resNameVPCConnection, plan.ID.String(), err),
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

func (r *vpcConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceVPCConnectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, vpcConnectionID, err := vpcConnectionParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameVPCConnection, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteVPCConnection(ctx, &quicksight.DeleteVPCConnectionInput{
		AwsAccountId:    aws.String(awsAccountID),
		VPCConnectionId: aws.String(vpcConnectionID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "Cannot perform operation on deleted VPCConnection") {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameVPCConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	_, err = waitVPCConnectionDeleted(ctx, conn, awsAccountID, vpcConnectionID, r.DeleteTimeout(ctx, state.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForDeletion, resNameVPCConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findVPCConnectionByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, vpcConnectionID string) (*awstypes.VPCConnection, error) {
	input := &quicksight.DescribeVPCConnectionInput{
		AwsAccountId:    aws.String(awsAccountID),
		VPCConnectionId: aws.String(vpcConnectionID),
	}
	output, err := findVPCConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.VPCConnectionResourceStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCConnection(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeVPCConnectionInput) (*awstypes.VPCConnection, error) {
	output, err := conn.DescribeVPCConnection(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VPCConnection == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VPCConnection, nil
}

func retryVPCConnectionCreate(ctx context.Context, conn *quicksight.Client, in *quicksight.CreateVPCConnectionInput) (*quicksight.CreateVPCConnectionOutput, error) {
	outputRaw, err := tfresource.RetryWhen(ctx,
		iamPropagationTimeout,
		func() (any, error) {
			return conn.CreateVPCConnection(ctx, in)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.AccessDeniedException](err) {
				return true, err
			}

			return false, err
		},
	)

	output, _ := outputRaw.(*quicksight.CreateVPCConnectionOutput)
	return output, err
}

func waitVPCConnectionCreated(ctx context.Context, conn *quicksight.Client, awsAccountID, vpcConnectionID string, timeout time.Duration) (*awstypes.VPCConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VPCConnectionResourceStatusCreationInProgress),
		Target:     enum.Slice(awstypes.VPCConnectionResourceStatusCreationSuccessful),
		Refresh:    statusVPCConnection(ctx, conn, awsAccountID, vpcConnectionID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VPCConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPCConnectionUpdated(ctx context.Context, conn *quicksight.Client, awsAccountID, vpcConnectionID string, timeout time.Duration) (*awstypes.VPCConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VPCConnectionResourceStatusUpdateInProgress),
		Target:     enum.Slice(awstypes.VPCConnectionResourceStatusUpdateSuccessful),
		Refresh:    statusVPCConnection(ctx, conn, awsAccountID, vpcConnectionID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VPCConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPCConnectionDeleted(ctx context.Context, conn *quicksight.Client, awsAccountID, vpcConnectionID string, timeout time.Duration) (*awstypes.VPCConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VPCConnectionResourceStatusDeletionInProgress),
		Target:     []string{},
		Refresh:    statusVPCConnection(ctx, conn, awsAccountID, vpcConnectionID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VPCConnection); ok {
		return output, err
	}

	return nil, err
}

func statusVPCConnection(ctx context.Context, conn *quicksight.Client, awsAccountID, vpcConnectionID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findVPCConnectionByTwoPartKey(ctx, conn, awsAccountID, vpcConnectionID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

const vpcConnectionResourceIDSeparator = ","

func vpcConnectionCreateResourceID(awsAccountID, vpcConnectionID string) string {
	parts := []string{awsAccountID, vpcConnectionID}
	id := strings.Join(parts, vpcConnectionResourceIDSeparator)

	return id
}

func vpcConnectionParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, vpcConnectionResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sVPC_CONNECTION_ID", id, vpcConnectionResourceIDSeparator)
	}

	return parts[0], parts[1], nil
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
	Tags               tftags.Map     `tfsdk:"tags"`
	TagsAll            tftags.Map     `tfsdk:"tags_all"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}
