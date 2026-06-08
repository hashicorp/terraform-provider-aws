// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package odb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_odb_cloud_exadata_infrastructure", name="Cloud Exadata Infrastructure")
// @Tags(identifierAttribute="arn")
func newDataSourceCloudExadataInfrastructure(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCloudExadataInfrastructure{}, nil
}

const (
	DSNameCloudExadataInfrastructure = "Cloud Exadata Infrastructure Data Source"
)

type dataSourceCloudExadataInfrastructure struct {
	framework.DataSourceWithModel[cloudExadataInfrastructureDataSourceModel]
}

func (d *dataSourceCloudExadataInfrastructure) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	statusType := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	computeModelType := fwtypes.StringEnumType[odbtypes.ComputeModel]()
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"activated_storage_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of storage servers requested for the Exadata infrastructure.",
			},
			"additional_storage_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of storage servers requested for the Exadata infrastructure.",
			},
			"available_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of available storage, in gigabytes (GB), for the Exadata infrastructure.",
			},
			names.AttrAvailabilityZone: schema.StringAttribute{
				Computed:    true,
				Description: "The name of the Availability Zone (AZ) where the Exadata infrastructure is located.",
			},
			"availability_zone_id": schema.StringAttribute{
				Computed:    true,
				Description: "The AZ ID of the AZ where the Exadata infrastructure is located.",
			},
			names.AttrARN: schema.StringAttribute{
				Computed:    true,
				Description: "The Amazon Resource Name (ARN) for the Exadata infrastructure.",
			},
			names.AttrID: schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the Exadata infrastructure.",
			},
			"compute_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of database servers for the Exadata infrastructure.",
			},
			"cpu_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPU cores that are allocated to the Exadata infrastructure.",
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The size of the Exadata infrastructure's data disk group, in terabytes (TB).",
			},
			"db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
				Description: "The database server model type of the Exadata infrastructure. For the list of\n" +
					"valid model names, use the ListDbSystemShapes operation.",
			},
			"db_server_version": schema.StringAttribute{
				Computed:    true,
				Description: "The version of the Exadata infrastructure.",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the Exadata infrastructure.",
			},
			"last_maintenance_run_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Oracle Cloud Identifier (OCID) of the last maintenance run for the Exadata infrastructure.",
			},
			"max_cpu_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPU cores available on the Exadata infrastructure.",
			},
			"max_data_storage_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The total amount of data disk group storage, in terabytes (TB), that's available on the Exadata infrastructure.",
			},
			"max_db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total amount of local node storage, in gigabytes (GB), that's available on the Exadata infrastructure.",
			},
			"max_memory_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total amount of memory, in gigabytes (GB), that's available on the Exadata infrastructure.",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of memory, in gigabytes (GB), that's allocated on the Exadata infrastructure.",
			},
			"monthly_db_server_version": schema.StringAttribute{
				Computed:    true,
				Description: "The monthly software version of the database servers installed on the Exadata infrastructure.",
			},
			"monthly_storage_server_version": schema.StringAttribute{
				Computed:    true,
				Description: "The monthly software version of the storage servers installed on the Exadata infrastructure.",
			},
			"next_maintenance_run_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the next maintenance run for the Exadata infrastructure.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the OCI resource anchor for the Exadata infrastructure.",
			},
			"oci_url": schema.StringAttribute{
				Computed:    true,
				Description: "The HTTPS link to the Exadata infrastructure in OCI.",
			},
			"ocid": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the Exadata infrastructure in OCI.",
			},
			"percent_progress": schema.Float64Attribute{
				Computed:    true,
				Description: "The amount of progress made on the current operation on the Exadata infrastructure expressed as a percentage.",
			},
			"shape": schema.StringAttribute{
				Computed:    true,
				Description: "The model name of the Exadata infrastructure.",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType:  statusType,
				Computed:    true,
				Description: "The status of the Exadata infrastructure.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the status of the Exadata infrastructure.",
			},
			"storage_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of storage servers that are activated for the Exadata infrastructure.",
			},
			"storage_server_version": schema.StringAttribute{
				Computed:    true,
				Description: "The software version of the storage servers on the Exadata infrastructure.",
			},
			"total_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total amount of storage, in gigabytes (GB), on the the Exadata infrastructure.",
			},
			"compute_model": schema.StringAttribute{
				CustomType: computeModelType,
				Computed:   true,
				Description: "The OCI model compute model used when you create or clone an instance: ECPU or\n" +
					"OCPU. An ECPU is an abstracted measure of compute resources. ECPUs are based on\n" +
					"the number of cores elastically allocated from a pool of compute and storage\n" +
					"servers. An OCPU is a legacy physical measure of compute resources. OCPUs are\n" +
					"based on the physical core of a processor with hyper-threading enabled.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "The time when the Exadata infrastructure was created.",
			},
			"database_server_type": schema.StringAttribute{
				Computed:    true,
				Description: "The database server model type of the Exadata infrastructure. For the list of valid model names, use the ListDbSystemShapes operation.",
			},
			"storage_server_type": schema.StringAttribute{
				Computed:    true,
				Description: "The storage server model type of the Exadata infrastructure. For the list of valid model names, use the ListDbSystemShapes operation.",
			},
			"customer_contacts_to_send_to_oci": schema.SetAttribute{
				CustomType:  fwtypes.NewSetNestedObjectTypeOf[customerContactExaInfraDataSourceModel](ctx),
				Computed:    true,
				Description: "The email addresses of contacts to receive notification from Oracle about maintenance updates for the Exadata infrastructure.",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"maintenance_window": schema.ListAttribute{
				Computed:    true,
				Description: " The scheduling details for the maintenance window. Patching and system updates take place during the maintenance window ",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudExadataInfraMaintenanceWindowDataSourceModel](ctx),
			},
		},
	}
}

func (d *dataSourceCloudExadataInfrastructure) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)

	var data cloudExadataInfrastructureDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindExaDataInfraForDataSourceByID(ctx, conn, data.CloudExadataInfrastructureId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameCloudExadataInfrastructure, data.CloudExadataInfrastructureId.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FindExaDataInfraForDataSourceByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudExadataInfrastructure, error) {
	input := odb.GetCloudExadataInfrastructureInput{
		CloudExadataInfrastructureId: aws.String(id),
	}

	out, err := conn.GetCloudExadataInfrastructure(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.CloudExadataInfrastructure == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.CloudExadataInfrastructure, nil
}

type cloudExadataInfrastructureDataSourceModel struct {
	framework.WithRegionModel
	ActivatedStorageCount         types.Int32                                                                        `tfsdk:"activated_storage_count"`
	AdditionalStorageCount        types.Int32                                                                        `tfsdk:"additional_storage_count"`
	AvailabilityZone              types.String                                                                       `tfsdk:"availability_zone"`
	AvailabilityZoneId            types.String                                                                       `tfsdk:"availability_zone_id"`
	AvailableStorageSizeInGBs     types.Int32                                                                        `tfsdk:"available_storage_size_in_gbs"`
	CloudExadataInfrastructureArn types.String                                                                       `tfsdk:"arn"`
	CloudExadataInfrastructureId  types.String                                                                       `tfsdk:"id"`
	ComputeCount                  types.Int32                                                                        `tfsdk:"compute_count"`
	CpuCount                      types.Int32                                                                        `tfsdk:"cpu_count"`
	DataStorageSizeInTBs          types.Float64                                                                      `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs        types.Int32                                                                        `tfsdk:"db_node_storage_size_in_gbs"`
	DbServerVersion               types.String                                                                       `tfsdk:"db_server_version"`
	DisplayName                   types.String                                                                       `tfsdk:"display_name"`
	LastMaintenanceRunId          types.String                                                                       `tfsdk:"last_maintenance_run_id"`
	MaxCpuCount                   types.Int32                                                                        `tfsdk:"max_cpu_count"`
	MaxDataStorageInTBs           types.Float64                                                                      `tfsdk:"max_data_storage_in_tbs"`
	MaxDbNodeStorageSizeInGBs     types.Int32                                                                        `tfsdk:"max_db_node_storage_size_in_gbs"`
	MaxMemoryInGBs                types.Int32                                                                        `tfsdk:"max_memory_in_gbs"`
	MemorySizeInGBs               types.Int32                                                                        `tfsdk:"memory_size_in_gbs"`
	MonthlyDbServerVersion        types.String                                                                       `tfsdk:"monthly_db_server_version"`
	MonthlyStorageServerVersion   types.String                                                                       `tfsdk:"monthly_storage_server_version"`
	NextMaintenanceRunId          types.String                                                                       `tfsdk:"next_maintenance_run_id"`
	OciResourceAnchorName         types.String                                                                       `tfsdk:"oci_resource_anchor_name"`
	OciUrl                        types.String                                                                       `tfsdk:"oci_url"`
	Ocid                          types.String                                                                       `tfsdk:"ocid"`
	PercentProgress               types.Float64                                                                      `tfsdk:"percent_progress"`
	Shape                         types.String                                                                       `tfsdk:"shape"`
	Status                        fwtypes.StringEnum[odbtypes.ResourceStatus]                                        `tfsdk:"status"`
	StatusReason                  types.String                                                                       `tfsdk:"status_reason"`
	StorageCount                  types.Int32                                                                        `tfsdk:"storage_count"`
	StorageServerVersion          types.String                                                                       `tfsdk:"storage_server_version"`
	TotalStorageSizeInGBs         types.Int32                                                                        `tfsdk:"total_storage_size_in_gbs"`
	CustomerContactsToSendToOCI   fwtypes.SetNestedObjectValueOf[customerContactExaInfraDataSourceModel]             `tfsdk:"customer_contacts_to_send_to_oci"`
	ComputeModel                  fwtypes.StringEnum[odbtypes.ComputeModel]                                          `tfsdk:"compute_model"`
	CreatedAt                     timetypes.RFC3339                                                                  `tfsdk:"created_at" `
	DatabaseServerType            types.String                                                                       `tfsdk:"database_server_type"`
	StorageServerType             types.String                                                                       `tfsdk:"storage_server_type"`
	MaintenanceWindow             fwtypes.ListNestedObjectValueOf[cloudExadataInfraMaintenanceWindowDataSourceModel] `tfsdk:"maintenance_window" `
	Tags                          tftags.Map                                                                         `tfsdk:"tags"`
}

type cloudExadataInfraMaintenanceWindowDataSourceModel struct {
	CustomActionTimeoutInMins    types.Int32                                                                       `tfsdk:"custom_action_timeout_in_mins"`
	DaysOfWeek                   fwtypes.SetNestedObjectValueOf[dayOfWeekExaInfraMaintenanceWindowDataSourceModel] `tfsdk:"days_of_week" `
	HoursOfDay                   fwtypes.SetValueOf[types.Int64]                                                   `tfsdk:"hours_of_day"`
	IsCustomActionTimeoutEnabled types.Bool                                                                        `tfsdk:"is_custom_action_timeout_enabled"`
	LeadTimeInWeeks              types.Int32                                                                       `tfsdk:"lead_time_in_weeks"`
	Months                       fwtypes.SetNestedObjectValueOf[monthExaInfraMaintenanceWindowDataSourceModel]     `tfsdk:"months" `
	PatchingMode                 fwtypes.StringEnum[odbtypes.PatchingModeType]                                     `tfsdk:"patching_mode"`
	Preference                   fwtypes.StringEnum[odbtypes.PreferenceType]                                       `tfsdk:"preference"`
	WeeksOfMonth                 fwtypes.SetValueOf[types.Int64]                                                   `tfsdk:"weeks_of_month"`
}

type dayOfWeekExaInfraMaintenanceWindowDataSourceModel struct {
	Name fwtypes.StringEnum[odbtypes.DayOfWeekName] `tfsdk:"name"`
}

type monthExaInfraMaintenanceWindowDataSourceModel struct {
	Name fwtypes.StringEnum[odbtypes.MonthName] `tfsdk:"name"`
}

type customerContactExaInfraDataSourceModel struct {
	Email types.String `tfsdk:"email"`
}
