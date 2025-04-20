// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_vpc_route_server", name="VPC Route Server")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func newVPCRouteServerResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCRouteServer{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCRouteServer = "VPC Route Server"
)

type resourceVPCRouteServer struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceVPCRouteServer) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"amazon_side_asn": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"persist_routes": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.RouteServerPersistRoutesAction](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"persist_routes_duration": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 5),
				},
			},
			"persist_routes_state": schema.StringAttribute{
				Computed: true,
			},
			"sns_notifications_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"sns_topic_arn": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
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

func (r *resourceVPCRouteServer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCRouteServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.CreateRouteServerInput{
		ClientToken:       aws.String(sdkid.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeRouteServer),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateRouteServer(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServer, plan.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.RouteServer == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCRouteServer, plan.RouteServerId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.RouteServer, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitVPCRouteServerCreated(ctx, conn, plan.RouteServerId.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCRouteServer, plan.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVPCRouteServer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCRouteServerByID(ctx, conn, state.RouteServerId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, ResNameVPCRouteServer, state.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVPCRouteServer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan, state resourceVPCRouteServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input ec2.ModifyRouteServerInput

		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.ModifyRouteServer(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionUpdating, ResNameVPCRouteServer, plan.RouteServerId.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.RouteServer == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionUpdating, ResNameVPCRouteServer, plan.RouteServerId.String(), nil),
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
	_, err := waitVPCRouteServerUpdated(ctx, conn, plan.RouteServerId.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForUpdate, ResNameVPCRouteServer, plan.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVPCRouteServer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCRouteServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DeleteRouteServerInput{
		RouteServerId: state.RouteServerId.ValueStringPointer(),
	}

	_, err := conn.DeleteRouteServer(ctx, &input)
	if err != nil {
		if tfresource.NotFound(err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCRouteServer, state.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCRouteServerDeleted(ctx, conn, state.RouteServerId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameVPCRouteServer, state.RouteServerId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCRouteServer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitVPCRouteServerCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RouteServerStatePending),
		Target:                    enum.Slice(awstypes.RouteServerStateAvailable),
		Refresh:                   statusVPCRouteServer(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ec2.CreateRouteServerOutput); ok {
		return out.RouteServer, err
	}

	return nil, err
}

func waitVPCRouteServerUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RouteServerStateModifying),
		Target:                    enum.Slice(awstypes.RouteServerStateAvailable),
		Refresh:                   statusVPCRouteServer(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ec2.ModifyRouteServerOutput); ok {
		return out.RouteServer, err
	}

	return nil, err
}

func waitVPCRouteServerDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteServerStateDeleting),
		Target:  enum.Slice(awstypes.RouteServerStateDeleted),
		Refresh: statusVPCRouteServer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ec2.DeleteRouteServerOutput); ok {
		return out.RouteServer, err
	}

	return nil, err
}

func statusVPCRouteServer(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findVPCRouteServerByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.State)), nil
	}
}

func findVPCRouteServerByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.RouteServer, error) {
	input := ec2.DescribeRouteServersInput{
		RouteServerIds: []string{aws.ToString(aws.String(id))},
	}
	var routeServers []awstypes.RouteServer
	paginator := ec2.NewDescribeRouteServersPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if tfresource.NotFound(err) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: &input,
				}
			}
			return nil, err
		}
		if page != nil && len(page.RouteServers) > 0 {
			routeServers = append(routeServers, page.RouteServers...)
		}
	}
	return &routeServers[0], nil
}

// func findVPCRouteServerByName(ctx context.Context, conn *ec2.Client, name string) (*awstypes.RouteServer, error) {
// 	input := ec2.DescribeRouteServersInput{
// 		Filters: []awstypes.Filter{
// 			{
// 				Name:   aws.String("tag:Name"),
// 				Values: []string{name},
// 			},
// 		},
// 	}
// 	var routeServers []awstypes.RouteServer
// 	paginator := ec2.NewDescribeRouteServersPaginator(conn, &input)
// 	for paginator.HasMorePages() {
// 		page, err := paginator.NextPage(ctx)
// 		if err != nil {
// 			if tfresource.NotFound(err) {
// 				return nil, &retry.NotFoundError{
// 					LastError:   err,
// 					LastRequest: &input,
// 				}
// 			}
// 			return nil, err
// 		}
// 		if page != nil && len(page.RouteServers) > 0 {
// 			routeServers = append(routeServers, page.RouteServers...)
// 		}
// 	}
// 	return &routeServers[0], nil
// }

type resourceVPCRouteServerModel struct {
	AmazonSideAsn           types.Int64    `tfsdk:"amazon_side_asn"`
	RouteServerId           types.String   `tfsdk:"id"`
	PersistRoutes           types.String   `tfsdk:"persist_routes"`
	PersistRoutesDuration   types.Int64    `tfsdk:"persist_routes_duration"`
	PersistRoutesState      types.String   `tfsdk:"persist_routes_state"`
	SnsNotificationsEnabled types.Bool     `tfsdk:"sns_notifications_enabled"`
	SnsTopicArn             types.String   `tfsdk:"sns_topic_arn"`
	State                   types.String   `tfsdk:"state"`
	Tags                    tftags.Map     `tfsdk:"tags"`
	TagsAll                 tftags.Map     `tfsdk:"tags_all"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}
