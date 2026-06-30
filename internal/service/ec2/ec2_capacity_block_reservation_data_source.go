// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_capacity_block_reservation", name="Capacity Block Reservation")
// @Tags
// @Testing(tagsTest=false)
func newCapacityBlockReservationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &capacityBlockReservationDataSource{}, nil
}

type capacityBlockReservationDataSource struct {
	framework.DataSourceWithModel[capacityBlockReservationDataSourceModel]
}

func (d *capacityBlockReservationDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrAvailabilityZone: schema.StringAttribute{
				Computed: true,
			},
			"availability_zone_id": schema.StringAttribute{
				Computed: true,
			},
			"available_instance_count": schema.Int64Attribute{
				Computed: true,
			},
			"capacity_block_id": schema.StringAttribute{
				Computed: true,
			},
			"commitment_info": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[commitmentInfoModel](ctx),
				Computed:   true,
			},
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"delivery_preference": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapacityReservationDeliveryPreference](),
				Computed:   true,
			},
			"ebs_optimized": schema.BoolAttribute{
				Computed: true,
			},
			"end_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"end_date_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EndDateType](),
				Computed:   true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrInstanceCount: schema.Int64Attribute{
				Computed: true,
			},
			"instance_match_criteria": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InstanceMatchCriteria](),
				Computed:   true,
			},
			"instance_platform": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapacityReservationInstancePlatform](),
				Computed:   true,
			},
			names.AttrInstanceType: schema.StringAttribute{
				Computed: true,
			},
			"interruptible_capacity_allocation": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[interruptibleCapacityAllocationModel](ctx),
				Computed:   true,
			},
			"interruption_info": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[interruptionInfoModel](ctx),
				Computed:   true,
			},
			names.AttrOutpostARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
			},
			"placement_group_arn": schema.StringAttribute{
				Computed: true,
			},
			"reservation_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapacityReservationType](),
				Computed:   true,
			},
			"start_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapacityReservationState](),
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"tenancy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapacityReservationTenancy](),
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(ctx),
		},
	}
}

func (d *capacityBlockReservationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data capacityBlockReservationDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := ec2.DescribeCapacityReservationsInput{
		Filters: newCustomFilterListFramework(ctx, data.Filters),
	}

	if !data.ID.IsNull() || !data.ID.IsUnknown() {
		input.CapacityReservationIds = []string{data.ID.ValueString()}
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := findCapacityReservation(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix("CapacityReservation")))
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type capacityBlockReservationDataSourceModel struct {
	framework.WithRegionModel
	ARN                             types.String                                                       `tfsdk:"arn"`
	AvailabilityZone                types.String                                                       `tfsdk:"availability_zone"`
	AvailabilityZoneID              types.String                                                       `tfsdk:"availability_zone_id"`
	AvailableInstanceCount          types.Int64                                                        `tfsdk:"available_instance_count"`
	CapacityBlockID                 types.String                                                       `tfsdk:"capacity_block_id"`
	CommitmentInfo                  fwtypes.ObjectValueOf[commitmentInfoModel]                         `tfsdk:"commitment_info"`
	CreateDate                      timetypes.RFC3339                                                  `tfsdk:"created_date"`
	DeliveryPreference              fwtypes.StringEnum[awstypes.CapacityReservationDeliveryPreference] `tfsdk:"delivery_preference"`
	EbsOptimized                    types.Bool                                                         `tfsdk:"ebs_optimized"`
	EndDate                         timetypes.RFC3339                                                  `tfsdk:"end_date"`
	EndDateType                     fwtypes.StringEnum[awstypes.EndDateType]                           `tfsdk:"end_date_type"`
	Filters                         customFilters                                                      `tfsdk:"filter" autoflex:"-"`
	ID                              types.String                                                       `tfsdk:"id"`
	InstanceMatchCriteria           fwtypes.StringEnum[awstypes.InstanceMatchCriteria]                 `tfsdk:"instance_match_criteria"`
	InstancePlatform                fwtypes.StringEnum[awstypes.CapacityReservationInstancePlatform]   `tfsdk:"instance_platform"`
	InstanceType                    types.String                                                       `tfsdk:"instance_type"`
	InterruptibleCapacityAllocation fwtypes.ObjectValueOf[interruptibleCapacityAllocationModel]        `tfsdk:"interruptible_capacity_allocation"`
	InterruptionInfo                fwtypes.ObjectValueOf[interruptionInfoModel]                       `tfsdk:"interruption_info"`
	OutpostARN                      types.String                                                       `tfsdk:"outpost_arn"`
	OwnerID                         types.String                                                       `tfsdk:"owner_id"`
	PlacementGroupARN               types.String                                                       `tfsdk:"placement_group_arn"`
	ReservationType                 fwtypes.StringEnum[awstypes.CapacityReservationType]               `tfsdk:"reservation_type"`
	StartDate                       timetypes.RFC3339                                                  `tfsdk:"start_date"`
	State                           fwtypes.StringEnum[awstypes.CapacityReservationState]              `tfsdk:"state"`
	Tags                            tftags.Map                                                         `tfsdk:"tags"`
	Tenancy                         fwtypes.StringEnum[awstypes.CapacityReservationTenancy]            `tfsdk:"tenancy"`
	TotalInstanceCount              types.Int64                                                        `tfsdk:"instance_count"`
}

type commitmentInfoModel struct {
	CommitmentEndDate      timetypes.RFC3339 `tfsdk:"commitment_end_date"`
	CommittedInstanceCount types.Int64       `tfsdk:"committed_instance_count"`
}

type interruptibleCapacityAllocationModel struct {
	InstanceCount                      types.Int64                                                                   `tfsdk:"instance_count"`
	InterruptibleCapacityReservationID types.String                                                                  `tfsdk:"interruptible_capacity_reservation_id"`
	InterruptionType                   fwtypes.StringEnum[awstypes.InterruptionType]                                 `tfsdk:"interruption_type"`
	Status                             fwtypes.StringEnum[awstypes.InterruptibleCapacityReservationAllocationStatus] `tfsdk:"status"`
	TargetInstanceCount                types.Int64                                                                   `tfsdk:"target_instance_count"`
}

type interruptionInfoModel struct {
	InterruptionType            fwtypes.StringEnum[awstypes.InterruptionType] `tfsdk:"interruption_type"`
	SourceCapacityReservationID types.String                                  `tfsdk:"source_capacity_reservation_id"`
}
