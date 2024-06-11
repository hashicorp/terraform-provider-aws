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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_capacity_block_reservation",name="Capacity BLock Reservation")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func newResourceCapacityBlockReservation(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCapacityBlockReservation{}
	r.SetDefaultCreateTimeout(40 * time.Minute)

	return r, nil
}

type resourceCapacityBlockReservation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
	framework.WithNoOpUpdate[resourceCapacityBlockReservationData]
	framework.WithNoOpDelete
}

func (r *resourceCapacityBlockReservation) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_ec2_capacity_block_reservation"
}

func (r *resourceCapacityBlockReservation) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrAvailabilityZone: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"capacity_block_offering_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ebs_optimized": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"end_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"end_date_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EndDateType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrInstanceCount: schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"instance_platform": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapacityReservationInstancePlatform](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrInstanceType: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"outpost_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"placement_group_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reservation_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapacityReservationType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"start_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"tenancy": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}

	response.Schema = s
}

const (
	ResNameCapacityBlockReservation = "Capacity Block Reservation"
)

func (r *resourceCapacityBlockReservation) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)
	var plan resourceCapacityBlockReservationData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := &ec2.PurchaseCapacityBlockInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)

	if response.Diagnostics.HasError() {
		return
	}

	input.TagSpecifications = getTagSpecificationsInV2(ctx, awstypes.ResourceTypeCapacityReservation)

	output, err := conn.PurchaseCapacityBlock(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameCapacityBlockReservation, plan.CapacityBlockOfferingID.String(), err),
			err.Error(),
		)
		return
	}

	if output == nil || output.CapacityReservation == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameCapacityBlockReservation, plan.CapacityBlockOfferingID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	cp := output.CapacityReservation
	state := plan
	state.ID = fwflex.StringToFramework(ctx, cp.CapacityReservationId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitCapacityBlockReservationActive(ctx, conn, createTimeout, state.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameCapacityBlockReservation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceCapacityBlockReservation) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)
	var data resourceCapacityBlockReservationData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findCapacityBlockReservationByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceCapacityBlockReservation) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceCapacityBlockReservationData struct {
	ARN                     types.String                                                     `tfsdk:"arn"`
	AvailabilityZone        types.String                                                     `tfsdk:"availability_zone"`
	CapacityBlockOfferingID types.String                                                     `tfsdk:"capacity_block_offering_id"`
	EbsOptimized            types.Bool                                                       `tfsdk:"ebs_optimized"`
	EndDate                 timetypes.RFC3339                                                `tfsdk:"end_date"`
	EndDateType             fwtypes.StringEnum[awstypes.EndDateType]                         `tfsdk:"end_date_type"`
	ID                      types.String                                                     `tfsdk:"id"`
	InstanceCount           types.Int64                                                      `tfsdk:"instance_count"`
	InstancePlatform        fwtypes.StringEnum[awstypes.CapacityReservationInstancePlatform] `tfsdk:"instance_platform"`
	InstanceType            types.String                                                     `tfsdk:"instance_type"`
	OutpostARN              types.String                                                     `tfsdk:"outpost_arn"`
	PlacementGroupARN       types.String                                                     `tfsdk:"placement_group_arn"`
	ReservationType         fwtypes.StringEnum[awstypes.CapacityReservationType]             `tfsdk:"reservation_type"`
	StartDate               timetypes.RFC3339                                                `tfsdk:"start_date"`
	Tags                    types.Map                                                        `tfsdk:"tags"`
	TagsAll                 types.Map                                                        `tfsdk:"tags_all"`
	Tenancy                 types.String                                                     `tfsdk:"tenancy"`
	Timeouts                timeouts.Value                                                   `tfsdk:"timeouts"`
}

func findCapacityBlockReservationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CapacityReservation, error) {
	input := &ec2.DescribeCapacityReservationsInput{
		CapacityReservationIds: []string{id},
	}

	output, err := conn.DescribeCapacityReservations(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidReservationNotFound, errCodeInvalidCapacityReservationIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CapacityReservations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	reservation, err := tfresource.AssertSingleValueResult(output.CapacityReservations)

	if err != nil {
		return nil, err
	}

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/capacity-reservations-using.html#capacity-reservations-view.
	if state := reservation.State; state == awstypes.CapacityReservationStateCancelled || state == awstypes.CapacityReservationStateExpired {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(reservation.CapacityReservationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return reservation, nil
}

func waitCapacityBlockReservationActive(ctx context.Context, conn *ec2.Client, timeout time.Duration, id string) (*awstypes.CapacityReservation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.CapacityReservationStatePaymentPending),
		Target:     enum.Slice(awstypes.CapacityReservationStateActive, awstypes.CapacityReservationStateScheduled),
		Refresh:    statusCapacityBlockReservation(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CapacityReservation); ok {
		return output, err
	}

	return nil, err
}

func statusCapacityBlockReservation(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCapacityBlockReservationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}
