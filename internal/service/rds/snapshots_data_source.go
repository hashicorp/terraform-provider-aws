// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_rds_snapshots", name="Snapshots")
func newSnapshotsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &snapshotsDataSource{}, nil
}

const (
	DSNameSnapshots = "Snapshots Data Source"
)

type snapshotsDataSource struct {
	framework.DataSourceWithModel[snapshotsDataSourceModel]
}

func (d *snapshotsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"db_instance_identifier": schema.StringAttribute{
				Optional: true,
			},
			"db_snapshot_identifier": schema.StringAttribute{
				Optional: true,
			},
			"include_public": schema.BoolAttribute{
				Optional: true,
			},
			"include_shared": schema.BoolAttribute{
				Optional: true,
			},
			"snapshot_type": schema.StringAttribute{
				Optional: true,
			},
			"snapshots": framework.DataSourceComputedListOfObjectAttribute[snapshotModel](ctx),
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[snapshotFilterModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrValues: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (d *snapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().RDSClient(ctx)
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)

	var data snapshotsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input rds.DescribeDBSnapshotsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDBSnapshots(ctx, conn, &input, tfslices.PredicateTrue[*rdstypes.DBSnapshot]())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.Snapshots))
	if resp.Diagnostics.HasError() {
		return
	}

	// Tags are ignored by AutoFlex; populate them manually.
	snapshots, diags := data.Snapshots.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}
	for i, snapshot := range snapshots {
		snapshot.Tags = tftags.NewMapFromMapValue(flex.FlattenFrameworkStringValueMapLegacy(ctx, keyValueTags(ctx, out[i].TagList).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()))
	}
	data.Snapshots = fwtypes.NewListNestedObjectValueOfSliceMust[snapshotModel](ctx, snapshots)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type snapshotsDataSourceModel struct {
	framework.WithRegionModel
	DBInstanceIdentifier types.String                                        `tfsdk:"db_instance_identifier"`
	DBSnapshotIdentifier types.String                                        `tfsdk:"db_snapshot_identifier"`
	Filters              fwtypes.SetNestedObjectValueOf[snapshotFilterModel] `tfsdk:"filter"`
	IncludePublic        types.Bool                                          `tfsdk:"include_public"`
	IncludeShared        types.Bool                                          `tfsdk:"include_shared"`
	SnapshotType         types.String                                        `tfsdk:"snapshot_type"`
	Snapshots            fwtypes.ListNestedObjectValueOf[snapshotModel]      `tfsdk:"snapshots"`
}

type snapshotFilterModel struct {
	Name   types.String        `tfsdk:"name"`
	Values fwtypes.SetOfString `tfsdk:"values"`
}

type snapshotModel struct {
	AllocatedStorage           types.Int64       `tfsdk:"allocated_storage"`
	AvailabilityZone           types.String      `tfsdk:"availability_zone"`
	DBInstanceIdentifier       types.String      `tfsdk:"db_instance_identifier"`
	DBSnapshotARN              types.String      `tfsdk:"db_snapshot_arn"`
	DBSnapshotIdentifier       types.String      `tfsdk:"db_snapshot_identifier"`
	Encrypted                  types.Bool        `tfsdk:"encrypted"`
	Engine                     types.String      `tfsdk:"engine"`
	EngineVersion              types.String      `tfsdk:"engine_version"`
	IOPS                       types.Int64       `tfsdk:"iops"`
	KMSKeyID                   types.String      `tfsdk:"kms_key_id"`
	LicenseModel               types.String      `tfsdk:"license_model"`
	OptionGroupName            types.String      `tfsdk:"option_group_name"`
	OriginalSnapshotCreateTime timetypes.RFC3339 `tfsdk:"original_snapshot_create_time"`
	Port                       types.Int64       `tfsdk:"port"`
	SnapshotCreateTime         timetypes.RFC3339 `tfsdk:"snapshot_create_time"`
	SnapshotType               types.String      `tfsdk:"snapshot_type"`
	SourceDBSnapshotIdentifier types.String      `tfsdk:"source_db_snapshot_identifier"`
	SourceRegion               types.String      `tfsdk:"source_region"`
	Status                     types.String      `tfsdk:"status"`
	StorageType                types.String      `tfsdk:"storage_type"`
	Tags                       tftags.Map        `tfsdk:"tags"`
	VPCID                      types.String      `tfsdk:"vpc_id"`
}
