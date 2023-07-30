package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Vpc Connection")
// @Tags(identifierAttribute="arn")
func newResourceVpcConnection(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVpcConnection{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVpcConnection = "Vpc Connection"
)

type resourceVpcConnection struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceVpcConnection) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_msk_vpc_connection"
}

func (r *resourceVpcConnection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),

			"authentication": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_subnets": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"security_groups": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"vpc_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_cluster_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *resourceVpcConnection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var data resourceVpcConnectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(createVpcConnectionID(data.TargetClusterArn.ValueString(), data.VpcId.ValueString()))

	in := &kafka.CreateVpcConnectionInput{
		Authentication:   aws.String(data.Authentication.ValueString()),
		ClientSubnets:    flex.ExpandFrameworkStringValueSet(ctx, data.ClientSubnets),
		SecurityGroups:   flex.ExpandFrameworkStringValueSet(ctx, data.SecurityGroups),
		TargetClusterArn: aws.String(data.TargetClusterArn.ValueString()),
		VpcId:            aws.String(data.VpcId.ValueString()),
	}

	out, err := conn.CreateVpcConnection(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionCreating, ResNameVpcConnection, data.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionCreating, ResNameVpcConnection, data.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	data.ARN = flex.StringToFramework(ctx, out.VpcConnectionArn)
	data.State = flex.StringValueToFramework(ctx, out.State)

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	_, err = waitVpcConnectionCreated(ctx, conn, data.ARN.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionWaitingForCreation, ResNameVpcConnection, data.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *resourceVpcConnection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var state resourceVpcConnectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindVpcConnectionByARN(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionSetting, ResNameVpcConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.VpcConnectionArn)
	state.Authentication = flex.StringToFramework(ctx, out.Authentication)
	state.SecurityGroups = flex.FlattenFrameworkStringValueSet(ctx, out.SecurityGroups)
	state.ClientSubnets = flex.FlattenFrameworkStringValueSet(ctx, out.Subnets)
	state.VpcId = flex.StringToFramework(ctx, out.VpcId)


	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVpcConnection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// conn := r.Meta().KafkaClient(ctx)

	var plan, state resourceVpcConnectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if !plan.ARN.Equal(state.ARN) ||
	// 	!plan.SecurityGroups.Equal(state.SecurityGroups) ||
	// 	!plan.ClientSubnets.Equal(state.ClientSubnets) ||
	// 	!plan.Authentication.Equal(state.Authentication){

	// 	in := &kafka.CreateVpcConnectionInput{
	// 		SecurityGroups:   aws.String(plan.SecurityGroups.ValueString()),
	// 		VpcConnectionName: aws.String(plan.Name.ValueString()),
	// 		VpcConnectionType: aws.String(plan.Type.ValueString()),
	// 	}

	// 	if !plan.Description.IsNull() {
	// 		// TIP: Optional fields should be set based on whether or not they are
	// 		// used.
	// 		in.Description = aws.String(plan.Description.ValueString())
	// 	}
	// 	if !plan.ComplexArgument.IsNull() {
	// 		// TIP: Use an expander to assign a complex argument. The elements must be
	// 		// deserialized into the appropriate struct before being passed to the expander.
	// 		var tfList []complexArgumentData
	// 		resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
	// 		if resp.Diagnostics.HasError() {
	// 			return
	// 		}

	// 		in.ComplexArgument = expandComplexArgument(tfList)
	// 	}

	// 	// TIP: -- 4. Call the AWS modify/update function
	// 	out, err := conn(ctx, in)
	// 	if err != nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.Kafka, create.ErrActionUpdating, ResNameVpcConnection, plan.ID.String(), err),
	// 			err.Error(),
	// 		)
	// 		return
	// 	}
	// 	if out == nil || out.VpcConnection == nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.Kafka, create.ErrActionUpdating, ResNameVpcConnection, plan.ID.String(), nil),
	// 			errors.New("empty output").Error(),
	// 		)
	// 		return
	// 	}

	// 	plan.ARN = flex.StringToFramework(ctx, out.VpcConnection.Arn)
	// 	plan.ID = flex.StringToFramework(ctx, out.VpcConnection.Arn)
	// }

	// updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	// _, err := waitVpcConnectionUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.Kafka, create.ErrActionWaitingForUpdate, ResNameVpcConnection, plan.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVpcConnection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var state resourceVpcConnectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &kafka.DeleteVpcConnectionInput{
		Arn: aws.String(state.ARN.ValueString()),
	}

	_, err := conn.DeleteVpcConnection(ctx, in)
	if err != nil {
		var nfe *awstypes.NotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionDeleting, ResNameVpcConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVpcConnectionDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Kafka, create.ErrActionWaitingForDeletion, ResNameVpcConnection, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVpcConnection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}


func waitVpcConnectionCreated(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*kafka.DescribeVpcConnectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VpcConnectionStateCreating),
		Target:                    enum.Slice(awstypes.VpcConnectionStateAvailable),
		Refresh:                   statusVpcConnection(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeVpcConnectionOutput); ok {
		return out, err
	}

	return nil, err
}

// func waitVpcConnectionUpdated(ctx context.Context, conn *kafka.Client, id string, timeout time.Duration) (*kafka.DescribeVpcConnectionOutput, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:    enum.Slice(awstypes.VpcConnectionStateCreating),
// 		Target:     enum.Slice(awstypes.VpcConnectionStateAvailable),
// 		Refresh:                   statusVpcConnection(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*kafka.DescribeVpcConnectionOutput); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

func waitVpcConnectionDeleted(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*kafka.DescribeVpcConnectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpcConnectionStateAvailable, awstypes.VpcConnectionStateInactive, awstypes.VpcConnectionStateDeactivating, awstypes.VpcConnectionStateDeleting),
		Target:  []string{},
		Refresh: statusVpcConnection(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeVpcConnectionOutput); ok {
		return out, err
	}

	return nil, err
}

func statusVpcConnection(ctx context.Context, conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindVpcConnectionByARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindVpcConnectionByARN(ctx context.Context, conn *kafka.Client, arn string) (*kafka.DescribeVpcConnectionOutput, error) {
	in := &kafka.DescribeVpcConnectionInput{
		Arn: aws.String(arn),
	}

	out, err := conn.DescribeVpcConnection(ctx, in)
	if err != nil {
		var nfe *awstypes.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func createVpcConnectionID(TargetClusterArn, VpcId string) string {
	return fmt.Sprintf("%s,%s", TargetClusterArn, VpcId)
}

type resourceVpcConnectionData struct {
	ARN              types.String   `tfsdk:"arn"`
	ID               types.String   `tfsdk:"id"`
	Authentication   types.String   `tfsdk:"authentication"`
	ClientSubnets    types.Set      `tfsdk:"client_subnets"`
	SecurityGroups   types.Set      `tfsdk:"security_groups"`
	TargetClusterArn types.String   `tfsdk:"target_cluster_arn"`
	Tags             types.Map      `tfsdk:"tags"`
	State            types.String   `tfsdk:"state"`
	TagsAll          types.Map      `tfsdk:"tags_all"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
	VpcId            types.String   `tfsdk:"vpc_id"`
}
