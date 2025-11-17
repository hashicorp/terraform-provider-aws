// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_route_server", name="VPC Route Server")
// @Tags(identifierAttribute="route_server_id")
// @Testing(tagsTest=false)
func newVPCRouteServerResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcRouteServerResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type vpcRouteServerResource struct {
	framework.ResourceWithModel[vpcRouteServerResourceModel]
	framework.WithTimeouts
}

func (r *vpcRouteServerResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"amazon_side_asn": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 4294967295),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"persist_routes": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RouteServerPersistRoutesAction](),
				Computed:   true,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"persist_routes_duration": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 5),
					int64validator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("persist_routes"),
					),
				},
			},
			"route_server_id": framework.IDAttribute(),
			"sns_notifications_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrSNSTopicARN: schema.StringAttribute{
				Computed: true,
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

func (r *vpcRouteServerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcRouteServerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	var input ec2.CreateRouteServerInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeRouteServer)

	output, err := conn.CreateRouteServer(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating VPC Route Server", err.Error())

		return
	}

	// Set values for unknowns.
	rs := output.RouteServer
	id := aws.ToString(rs.RouteServerId)
	data.ARN = r.routeServerARN(ctx, id)
	data.PersistRoutes = fwtypes.StringEnumValue(routeServerPersistRoutesStateToRouteServerPersistRoutesAction(rs.PersistRoutesState))
	data.RouteServerID = fwflex.StringValueToFramework(ctx, id)
	data.SNSNotificationsEnabled = fwflex.BoolToFramework(ctx, rs.SnsNotificationsEnabled)
	data.SNSTopicARN = fwflex.StringToFramework(ctx, rs.SnsTopicArn)

	if _, err := waitRouteServerCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server (%s) create", id), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *vpcRouteServerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcRouteServerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.RouteServerID)
	rs, err := findRouteServerByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Route Server (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, rs, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ARN = r.routeServerARN(ctx, id)
	data.PersistRoutes = fwtypes.StringEnumValue(routeServerPersistRoutesStateToRouteServerPersistRoutesAction(rs.PersistRoutesState))
	setTagsOut(ctx, rs.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcRouteServerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old vpcRouteServerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	if !new.PersistRoutes.Equal(old.PersistRoutes) || !new.PersistRoutesDuration.Equal(old.PersistRoutesDuration) || !new.SNSNotificationsEnabled.Equal(old.SNSNotificationsEnabled) {
		id := fwflex.StringValueFromFramework(ctx, new.RouteServerID)
		var input ec2.ModifyRouteServerInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		output, err := conn.ModifyRouteServer(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating VPC Route Server (%s)", id), err.Error())

			return
		}

		// Set values for unknowns.
		rs := output.RouteServer
		new.SNSTopicARN = fwflex.StringToFramework(ctx, rs.SnsTopicArn)

		if _, err := waitRouteServerUpdated(ctx, conn, id, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server (%s) update", id), err.Error())

			return
		}
	} else {
		new.SNSTopicARN = old.SNSTopicARN
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *vpcRouteServerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpcRouteServerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.RouteServerID)
	input := ec2.DeleteRouteServerInput{
		RouteServerId: aws.String(id),
	}
	_, err := conn.DeleteRouteServer(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteServerIdNotFound, errCodeIncorrectState) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Route Server (%s)", id), err.Error())

		return
	}

	if _, err := waitRouteServerDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Route Server (%s) delete", id), err.Error())

		return
	}
}

func (r *vpcRouteServerResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("route_server_id"), request, response)
}

func (r *vpcRouteServerResource) routeServerARN(ctx context.Context, id string) types.String {
	return fwflex.StringValueToFramework(ctx, r.Meta().RegionalARN(ctx, names.EC2, "route-server/"+id))
}

func routeServerPersistRoutesStateToRouteServerPersistRoutesAction(state awstypes.RouteServerPersistRoutesState) awstypes.RouteServerPersistRoutesAction {
	if state == awstypes.RouteServerPersistRoutesStateEnabled {
		return awstypes.RouteServerPersistRoutesActionEnable
	}
	return awstypes.RouteServerPersistRoutesActionDisable
}

type vpcRouteServerResourceModel struct {
	framework.WithRegionModel
	AmazonSideASN           types.Int64                                                 `tfsdk:"amazon_side_asn"`
	ARN                     types.String                                                `tfsdk:"arn"`
	PersistRoutes           fwtypes.StringEnum[awstypes.RouteServerPersistRoutesAction] `tfsdk:"persist_routes"`
	PersistRoutesDuration   types.Int64                                                 `tfsdk:"persist_routes_duration"`
	RouteServerID           types.String                                                `tfsdk:"route_server_id"`
	SNSNotificationsEnabled types.Bool                                                  `tfsdk:"sns_notifications_enabled"`
	SNSTopicARN             types.String                                                `tfsdk:"sns_topic_arn"`
	Tags                    tftags.Map                                                  `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                  `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                              `tfsdk:"timeouts"`
}
