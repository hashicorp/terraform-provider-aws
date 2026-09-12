// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_capacity_reservation", name="Capacity Reservation")
func newCapacityReservationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &capacityReservationDataSource{}, nil
}

type capacityReservationDataSource struct {
	framework.DataSourceWithModel[capacityReservationDataSourceModel]
}

func (d *capacityReservationDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
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
			"commitment_info": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[commitmentInfoModel](ctx),
				Computed:   true,
			},
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
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
			"ephemeral_storage": schema.BoolAttribute{
				Computed: true,
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

func (d *capacityReservationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data capacityReservationDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)

	input := ec2.DescribeCapacityReservationsInput{
		Filters: newCustomFilterListFramework(ctx, data.Filters),
	}

	if !data.ID.IsNull() {
		input.CapacityReservationIds = []string{fwflex.StringValueFromFramework(ctx, data.ID)}
	}

	output, err := findCapacityReservation(ctx, conn, &input)
	if err != nil {
		response.Diagnostics.AddError(
			"reading EC2 Capacity Reservation",
			tfresource.SingularDataSourceFindError("EC2 Capacity Reservation", err).Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix("CapacityReservation"))...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Tags = tftags.FlattenStringValueMap(ctx, keyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type capacityReservationDataSourceModel struct {
	framework.WithRegionModel
	ARN                             types.String                                                     `tfsdk:"arn"`
	AvailabilityZone                types.String                                                     `tfsdk:"availability_zone"`
	AvailabilityZoneID              types.String                                                     `tfsdk:"availability_zone_id"`
	AvailableInstanceCount          types.Int64                                                      `tfsdk:"available_instance_count"`
	CommitmentInfo                  fwtypes.ObjectValueOf[commitmentInfoModel]                       `tfsdk:"commitment_info"`
	CreateDate                      timetypes.RFC3339                                                `tfsdk:"created_date"`
	EbsOptimized                    types.Bool                                                       `tfsdk:"ebs_optimized"`
	EndDate                         timetypes.RFC3339                                                `tfsdk:"end_date"`
	EndDateType                     fwtypes.StringEnum[awstypes.EndDateType]                         `tfsdk:"end_date_type"`
	EphemeralStorage                types.Bool                                                       `tfsdk:"ephemeral_storage"`
	Filters                         customFilters                                                    `tfsdk:"filter" autoflex:"-"`
	ID                              types.String                                                     `tfsdk:"id"`
	InstanceMatchCriteria           fwtypes.StringEnum[awstypes.InstanceMatchCriteria]               `tfsdk:"instance_match_criteria"`
	InstancePlatform                fwtypes.StringEnum[awstypes.CapacityReservationInstancePlatform] `tfsdk:"instance_platform"`
	InstanceType                    types.String                                                     `tfsdk:"instance_type"`
	InterruptibleCapacityAllocation fwtypes.ObjectValueOf[interruptibleCapacityAllocationModel]      `tfsdk:"interruptible_capacity_allocation"`
	InterruptionInfo                fwtypes.ObjectValueOf[interruptionInfoModel]                     `tfsdk:"interruption_info"`
	OutpostARN                      types.String                                                     `tfsdk:"outpost_arn"`
	OwnerID                         types.String                                                     `tfsdk:"owner_id"`
	PlacementGroupARN               types.String                                                     `tfsdk:"placement_group_arn"`
	ReservationType                 fwtypes.StringEnum[awstypes.CapacityReservationType]             `tfsdk:"reservation_type"`
	StartDate                       timetypes.RFC3339                                                `tfsdk:"start_date"`
	State                           fwtypes.StringEnum[awstypes.CapacityReservationState]            `tfsdk:"state"`
	Tags                            tftags.Map                                                       `tfsdk:"tags"`
	Tenancy                         fwtypes.StringEnum[awstypes.CapacityReservationTenancy]          `tfsdk:"tenancy"`
	TotalInstanceCount              types.Int64                                                      `tfsdk:"instance_count"`
}
