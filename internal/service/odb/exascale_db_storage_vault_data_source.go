// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_odb_exascale_db_storage_vault", name="Exascale DB Storage Vault")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newExascaleDBStorageVaultDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &exascaleDBStorageVaultDataSource{}, nil
}

type exascaleDBStorageVaultDataSource struct {
	framework.DataSourceWithModel[exascaleDBStorageVaultDataSourceModel]
}

func (d *exascaleDBStorageVaultDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	highCapacityDatabaseStorageAttribute := framework.DataSourceComputedListOfObjectAttribute[exascaleDBStorageDetailsDataSourceModel](ctx)
	highCapacityDatabaseStorageAttribute.Description = "High-capacity database storage details for the Exascale DB storage vault."

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"additional_flash_cache_in_percent": schema.Int32Attribute{
				Computed:    true,
				Description: "Additional flash cache percentage for the Exascale DB storage vault.",
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"attached_shape_attributes": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringEnumType[awstypes.ShapeAttribute](),
				ElementType: types.StringType,
				Computed:    true,
				Description: "Shape attributes attached to the Exascale DB storage vault.",
			},
			"autoscale_limit_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "Autoscale limit, in GB, for the Exascale DB storage vault.",
			},
			names.AttrAvailabilityZone: schema.StringAttribute{
				Computed:    true,
				Description: "Availability Zone for the Exascale DB storage vault.",
			},
			"availability_zone_id": schema.StringAttribute{
				Computed:    true,
				Description: "Availability Zone ID for the Exascale DB storage vault.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Date and time when the Exascale DB storage vault was created.",
			},
			names.AttrDescription: schema.StringAttribute{
				Computed:    true,
				Description: "Description of the Exascale DB storage vault.",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Computed:    true,
				Description: "User-friendly name for the Exascale DB storage vault.",
			},
			"high_capacity_database_storage": highCapacityDatabaseStorageAttribute,
			names.AttrID: schema.StringAttribute{
				Required:    true,
				Description: "Unique identifier of the Exascale DB storage vault.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 2048),
				},
			},
			"is_autoscale_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether autoscaling is enabled for the Exascale DB storage vault.",
			},
			"ocid": schema.StringAttribute{
				Computed:    true,
				Description: "OCID of the Exascale DB storage vault.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the OCI resource anchor for the Exascale DB storage vault.",
			},
			"oci_url": schema.StringAttribute{
				Computed:    true,
				Description: "HTTPS URL for the Exascale DB storage vault in OCI.",
			},
			"percent_progress": schema.Float32Attribute{
				Computed:    true,
				Description: "Progress of the current operation on the Exascale DB storage vault, expressed as a percentage.",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.ResourceStatus](),
				Computed:    true,
				Description: "Current status of the Exascale DB storage vault.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the status of the Exascale DB storage vault.",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"time_zone": schema.StringAttribute{
				Computed:    true,
				Description: "Time zone of the Exascale DB storage vault.",
			},
			"vm_cluster_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "ARNs of the VM clusters associated with the Exascale DB storage vault.",
			},
			"vm_cluster_count": schema.Int32Attribute{
				Computed:    true,
				Description: "Number of VM clusters associated with the Exascale DB storage vault.",
			},
			"vm_cluster_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "Unique identifiers of the VM clusters associated with the Exascale DB storage vault.",
			},
		},
	}
}

func (d *exascaleDBStorageVaultDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)

	var data exascaleDBStorageVaultDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findExascaleDBStorageVaultByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("ExascaleDbStorageVault")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type exascaleDBStorageVaultDataSourceModel struct {
	framework.WithRegionModel
	AdditionalFlashCacheInPercent types.Int32                                                              `tfsdk:"additional_flash_cache_in_percent"`
	ARN                           types.String                                                             `tfsdk:"arn"`
	AttachedShapeAttributes       fwtypes.ListOfStringEnum[awstypes.ShapeAttribute]                        `tfsdk:"attached_shape_attributes"`
	AutoscaleLimitInGBs           types.Int32                                                              `tfsdk:"autoscale_limit_in_gbs"`
	AvailabilityZone              types.String                                                             `tfsdk:"availability_zone"`
	AvailabilityZoneID            types.String                                                             `tfsdk:"availability_zone_id"`
	CreatedAt                     timetypes.RFC3339                                                        `tfsdk:"created_at"`
	Description                   types.String                                                             `tfsdk:"description"`
	DisplayName                   types.String                                                             `tfsdk:"display_name"`
	HighCapacityDatabaseStorage   fwtypes.ListNestedObjectValueOf[exascaleDBStorageDetailsDataSourceModel] `tfsdk:"high_capacity_database_storage"`
	ID                            types.String                                                             `tfsdk:"id"`
	IsAutoscaleEnabled            types.Bool                                                               `tfsdk:"is_autoscale_enabled"`
	Ocid                          types.String                                                             `tfsdk:"ocid"`
	OciResourceAnchorName         types.String                                                             `tfsdk:"oci_resource_anchor_name"`
	OciUrl                        types.String                                                             `tfsdk:"oci_url"`
	PercentProgress               types.Float32                                                            `tfsdk:"percent_progress"`
	Status                        fwtypes.StringEnum[awstypes.ResourceStatus]                              `tfsdk:"status"`
	StatusReason                  types.String                                                             `tfsdk:"status_reason"`
	Tags                          tftags.Map                                                               `tfsdk:"tags"`
	TimeZone                      types.String                                                             `tfsdk:"time_zone"`
	VmClusterArns                 fwtypes.ListOfString                                                     `tfsdk:"vm_cluster_arns"`
	VmClusterCount                types.Int32                                                              `tfsdk:"vm_cluster_count"`
	VmClusterIds                  fwtypes.ListOfString                                                     `tfsdk:"vm_cluster_ids"`
}

type exascaleDBStorageDetailsDataSourceModel struct {
	AvailableSizeInGBs types.Int32 `tfsdk:"available_size_in_gbs"`
	TotalSizeInGBs     types.Int32 `tfsdk:"total_size_in_gbs"`
}
