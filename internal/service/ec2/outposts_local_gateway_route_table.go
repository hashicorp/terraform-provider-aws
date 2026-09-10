// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_local_gateway_route_table", name="Local Gateway Route Table")
// @Tags(identifierAttribute="local_gateway_route_table_id")
// @Testing(tagsTest=false)
// @Testing(preCheck="acctest.PreCheckOutpostsOutposts")
// @Testing(generator=false)
func newLocalGatewayRouteTableResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &localGatewayRouteTableResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type localGatewayRouteTableResource struct {
	framework.ResourceWithModel[localGatewayRouteTableResourceModel]
	framework.WithTimeouts
}

func (r *localGatewayRouteTableResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"local_gateway_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"local_gateway_route_table_id": framework.IDAttribute(),
			names.AttrMode: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.LocalGatewayRouteTableMode](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrOutpostARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *localGatewayRouteTableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data localGatewayRouteTableResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.CreateLocalGatewayRouteTableInput{
		LocalGatewayId:    fwflex.StringFromFramework(ctx, data.LocalGatewayID),
		Mode:              awstypes.LocalGatewayRouteTableMode(data.Mode.ValueString()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeLocalGatewayRouteTable),
	}

	output, err := conn.CreateLocalGatewayRouteTable(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 Local Gateway Route Table", err.Error())
		return
	}

	rt := output.LocalGatewayRouteTable
	id := aws.ToString(rt.LocalGatewayRouteTableId)
	data.LocalGatewayRouteTableID = types.StringValue(id)
	data.ARN = fwflex.StringToFramework(ctx, rt.LocalGatewayRouteTableArn)
	data.OutpostARN = fwflex.StringToFramework(ctx, rt.OutpostArn)
	data.OwnerID = fwflex.StringToFramework(ctx, rt.OwnerId)

	waitOutput, err := waitLocalGatewayRouteTableAvailable(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Local Gateway Route Table (%s) create", id), err.Error())
		return
	}

	data.State = fwflex.StringToFramework(ctx, waitOutput.State)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *localGatewayRouteTableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data localGatewayRouteTableResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := data.LocalGatewayRouteTableID.ValueString()
	rt, err := findLocalGatewayRouteTableByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Local Gateway Route Table (%s)", id), err.Error())
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, rt.LocalGatewayRouteTableArn)
	data.LocalGatewayID = fwflex.StringToFramework(ctx, rt.LocalGatewayId)
	data.Mode = fwtypes.StringEnumValue(rt.Mode)
	data.OutpostARN = fwflex.StringToFramework(ctx, rt.OutpostArn)
	data.OwnerID = fwflex.StringToFramework(ctx, rt.OwnerId)
	data.State = fwflex.StringToFramework(ctx, rt.State)

	setTagsOut(ctx, rt.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *localGatewayRouteTableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data localGatewayRouteTableResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Tags only.

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *localGatewayRouteTableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data localGatewayRouteTableResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := data.LocalGatewayRouteTableID.ValueString()
	input := ec2.DeleteLocalGatewayRouteTableInput{
		LocalGatewayRouteTableId: aws.String(id),
	}

	_, err := conn.DeleteLocalGatewayRouteTable(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidLocalGatewayRouteTableIDNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Local Gateway Route Table (%s)", id), err.Error())
		return
	}

	if _, err := waitLocalGatewayRouteTableDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Local Gateway Route Table (%s) delete", id), err.Error())
		return
	}
}

func (r *localGatewayRouteTableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("local_gateway_route_table_id"), request, response)
}

func findLocalGatewayRouteTableByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.LocalGatewayRouteTable, error) {
	input := ec2.DescribeLocalGatewayRouteTablesInput{
		LocalGatewayRouteTableIds: []string{id},
	}

	output, err := findLocalGatewayRouteTable(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if aws.ToString(output.State) == "deleted" {
		return nil, &retry.NotFoundError{
			Message: "deleted",
		}
	}

	return output, nil
}

func statusLocalGatewayRouteTable(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findLocalGatewayRouteTableByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}

func waitLocalGatewayRouteTableAvailable(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.LocalGatewayRouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"available"},
		Refresh: statusLocalGatewayRouteTable(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.LocalGatewayRouteTable); ok {
		return output, err
	}

	return nil, err
}

func waitLocalGatewayRouteTableDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.LocalGatewayRouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"available", "deleting"},
		Target:  []string{},
		Refresh: statusLocalGatewayRouteTable(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.LocalGatewayRouteTable); ok {
		return output, err
	}

	return nil, err
}

type localGatewayRouteTableResourceModel struct {
	framework.WithRegionModel
	ARN                      types.String                                            `tfsdk:"arn"`
	LocalGatewayID           types.String                                            `tfsdk:"local_gateway_id"`
	LocalGatewayRouteTableID types.String                                            `tfsdk:"local_gateway_route_table_id"`
	Mode                     fwtypes.StringEnum[awstypes.LocalGatewayRouteTableMode] `tfsdk:"mode"`
	OutpostARN               types.String                                            `tfsdk:"outpost_arn"`
	OwnerID                  types.String                                            `tfsdk:"owner_id"`
	State                    types.String                                            `tfsdk:"state"`
	Tags                     tftags.Map                                              `tfsdk:"tags"`
	TagsAll                  tftags.Map                                              `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                          `tfsdk:"timeouts"`
}
