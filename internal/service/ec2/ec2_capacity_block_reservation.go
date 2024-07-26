// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_capacity_block_reservation",name="Capacity Block Reservation")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func newCapacityBlockReservationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &capacityBlockReservationResource{}
	r.SetDefaultCreateTimeout(40 * time.Minute)

	return r, nil
}

type capacityBlockReservationResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
	framework.WithNoOpUpdate[capacityBlockReservationReservationModel]
	framework.WithNoOpDelete
}

func (*capacityBlockReservationResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_ec2_capacity_block_reservation"
}

func (r *capacityBlockReservationResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
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

func (r *capacityBlockReservationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data capacityBlockReservationReservationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.PurchaseCapacityBlockInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeCapacityReservation)

	output, err := conn.PurchaseCapacityBlock(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("purchasing EC2 Capacity Block Reservation", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.CapacityReservation.CapacityReservationId)

	cr, err := waitCapacityBlockReservationActive(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Capacity Block Reservation (%s) active", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, cr, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *capacityBlockReservationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data capacityBlockReservationReservationModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	cr, err := findCapacityReservationByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Capacity Block Reservation (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, cr, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *capacityBlockReservationResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type capacityBlockReservationReservationModel struct {
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
