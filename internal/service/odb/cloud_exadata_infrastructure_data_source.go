//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_odb_cloud_exadata_infrastructure", name="Cloud Exadata Infrastructure")
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
			"availability_zone": schema.StringAttribute{
				Computed:    true,
				Description: "he name of the Availability Zone (AZ) where the Exadata infrastructure is located.",
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
			"display_name": schema.StringAttribute{
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
			"status": schema.StringAttribute{
				CustomType:  statusType,
				Computed:    true,
				Description: "The status of the Exadata infrastructure.",
			},
			"status_reason": schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the status of the Exadata infrastructure.",
			},
			"storage_count": schema.Int32Attribute{
				Computed:    true,
				Description: "he number of storage servers that are activated for the Exadata infrastructure.",
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
			"created_at": schema.StringAttribute{
				Computed:    true,
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
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"maintenance_window": schema.ObjectAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewObjectTypeOf[cloudExadataInfraMaintenanceWindowDataSourceModel](ctx),
				Description: "The maintenance window for the Exadata infrastructure.",
				AttributeTypes: map[string]attr.Type{
					"custom_action_timeout_in_mins": types.Int32Type,
					"days_of_week": types.SetType{
						ElemType: fwtypes.StringEnumType[odbtypes.DayOfWeekName](),
					},
					"hours_of_day": types.SetType{
						ElemType: types.Int32Type,
					},
					"is_custom_action_timeout_enabled": types.BoolType,
					"lead_time_in_weeks":               types.Int32Type,
					"months": types.SetType{
						ElemType: fwtypes.StringEnumType[odbtypes.MonthName](),
					},
					"patching_mode": fwtypes.StringEnumType[odbtypes.PatchingModeType](),
					"preference":    fwtypes.StringEnumType[odbtypes.PreferenceType](),
					"weeks_of_month": types.SetType{
						ElemType: types.Int32Type,
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"customer_contacts_to_send_to_oci": schema.SetNestedBlock{
				Description: "Customer contact emails to send to OCI.",
				CustomType:  fwtypes.NewSetNestedObjectTypeOf[customerContactDataSourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							Computed: true,
						},
					},
				},
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

	out, err := FindOdbExaDataInfraForDataSourceByID(ctx, conn, data.CloudExadataInfrastructureId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameCloudExadataInfrastructure, data.CloudExadataInfrastructureId.String(), err),
			err.Error(),
		)
		return
	}
	tagsRead, err := listTags(ctx, conn, *out.CloudExadataInfrastructureArn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameCloudExadataInfrastructure, data.CloudExadataInfrastructureId.String(), err),
			err.Error(),
		)
		return
	}
	if tagsRead != nil {
		data.Tags = tftags.FlattenStringValueMap(ctx, tagsRead.Map())
	}
	data.CreatedAt = types.StringValue(out.CreatedAt.Format(time.RFC3339))
	data.MaintenanceWindow = d.flattenMaintenanceWindow(ctx, out.MaintenanceWindow)
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FindOdbExaDataInfraForDataSourceByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudExadataInfrastructure, error) {
	input := odb.GetCloudExadataInfrastructureInput{
		CloudExadataInfrastructureId: aws.String(id),
	}

	out, err := conn.GetCloudExadataInfrastructure(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.CloudExadataInfrastructure == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.CloudExadataInfrastructure, nil
}

func (d *dataSourceCloudExadataInfrastructure) flattenMaintenanceWindow(ctx context.Context, obdExaInfraMW *odbtypes.MaintenanceWindow) fwtypes.ObjectValueOf[cloudExadataInfraMaintenanceWindowDataSourceModel] {
	//days of week
	daysOfWeek := make([]attr.Value, 0, len(obdExaInfraMW.DaysOfWeek))
	for _, dayOfWeek := range obdExaInfraMW.DaysOfWeek {
		dayOfWeekStringValue := fwtypes.StringEnumValue(dayOfWeek.Name).StringValue
		daysOfWeek = append(daysOfWeek, dayOfWeekStringValue)
	}
	setValueOfDaysOfWeek, _ := basetypes.NewSetValue(types.StringType, daysOfWeek)
	daysOfWeekRead := fwtypes.SetValueOf[fwtypes.StringEnum[odbtypes.DayOfWeekName]]{
		SetValue: setValueOfDaysOfWeek,
	}
	//hours of the day
	hoursOfTheDay := make([]attr.Value, 0, len(obdExaInfraMW.HoursOfDay))
	for _, hourOfTheDay := range obdExaInfraMW.HoursOfDay {
		daysOfWeekInt32Value := types.Int32Value(hourOfTheDay)
		hoursOfTheDay = append(hoursOfTheDay, daysOfWeekInt32Value)
	}
	setValuesOfHoursOfTheDay, _ := basetypes.NewSetValue(types.Int32Type, hoursOfTheDay)
	hoursOfTheDayRead := fwtypes.SetValueOf[types.Int64]{
		SetValue: setValuesOfHoursOfTheDay,
	}
	//months
	months := make([]attr.Value, 0, len(obdExaInfraMW.Months))
	for _, month := range obdExaInfraMW.Months {
		monthStringValue := fwtypes.StringEnumValue(month.Name).StringValue
		months = append(months, monthStringValue)
	}
	setValuesOfMonth, _ := basetypes.NewSetValue(types.StringType, months)
	monthsRead := fwtypes.SetValueOf[fwtypes.StringEnum[odbtypes.MonthName]]{
		SetValue: setValuesOfMonth,
	}
	//weeks of month
	weeksOfMonth := make([]attr.Value, 0, len(obdExaInfraMW.WeeksOfMonth))
	for _, weekOfMonth := range obdExaInfraMW.WeeksOfMonth {
		weeksOfMonthInt32Value := types.Int32Value(weekOfMonth)
		weeksOfMonth = append(weeksOfMonth, weeksOfMonthInt32Value)
	}
	setValuesOfWeekOfMonth, _ := basetypes.NewSetValue(types.Int32Type, weeksOfMonth)
	weeksOfMonthRead := fwtypes.SetValueOf[types.Int64]{
		SetValue: setValuesOfWeekOfMonth,
	}

	flattenMW := cloudExadataInfraMaintenanceWindowDataSourceModel{
		CustomActionTimeoutInMins:    types.Int32PointerValue(obdExaInfraMW.CustomActionTimeoutInMins),
		DaysOfWeek:                   daysOfWeekRead,
		HoursOfDay:                   hoursOfTheDayRead,
		IsCustomActionTimeoutEnabled: types.BoolPointerValue(obdExaInfraMW.IsCustomActionTimeoutEnabled),
		LeadTimeInWeeks:              types.Int32PointerValue(obdExaInfraMW.LeadTimeInWeeks),
		Months:                       monthsRead,
		PatchingMode:                 fwtypes.StringEnumValue(obdExaInfraMW.PatchingMode),
		Preference:                   fwtypes.StringEnumValue(obdExaInfraMW.Preference),
		WeeksOfMonth:                 weeksOfMonthRead,
	}
	if obdExaInfraMW.LeadTimeInWeeks == nil {
		flattenMW.LeadTimeInWeeks = types.Int32Value(0)
	}
	if obdExaInfraMW.CustomActionTimeoutInMins == nil {
		flattenMW.CustomActionTimeoutInMins = types.Int32Value(0)
	}
	if obdExaInfraMW.IsCustomActionTimeoutEnabled == nil {
		flattenMW.IsCustomActionTimeoutEnabled = types.BoolValue(false)
	}

	result, _ := fwtypes.NewObjectValueOf[cloudExadataInfraMaintenanceWindowDataSourceModel](ctx, &flattenMW)
	return result
}

type cloudExadataInfrastructureDataSourceModel struct {
	framework.WithRegionModel
	ActivatedStorageCount         types.Int32                                                              `tfsdk:"activated_storage_count"`
	AdditionalStorageCount        types.Int32                                                              `tfsdk:"additional_storage_count"`
	AvailabilityZone              types.String                                                             `tfsdk:"availability_zone"`
	AvailabilityZoneId            types.String                                                             `tfsdk:"availability_zone_id"`
	AvailableStorageSizeInGBs     types.Int32                                                              `tfsdk:"available_storage_size_in_gbs"`
	CloudExadataInfrastructureArn types.String                                                             `tfsdk:"arn"`
	CloudExadataInfrastructureId  types.String                                                             `tfsdk:"id"`
	ComputeCount                  types.Int32                                                              `tfsdk:"compute_count"`
	CpuCount                      types.Int32                                                              `tfsdk:"cpu_count"`
	DataStorageSizeInTBs          types.Float64                                                            `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs        types.Int32                                                              `tfsdk:"db_node_storage_size_in_gbs"`
	DbServerVersion               types.String                                                             `tfsdk:"db_server_version"`
	DisplayName                   types.String                                                             `tfsdk:"display_name"`
	LastMaintenanceRunId          types.String                                                             `tfsdk:"last_maintenance_run_id"`
	MaxCpuCount                   types.Int32                                                              `tfsdk:"max_cpu_count"`
	MaxDataStorageInTBs           types.Float64                                                            `tfsdk:"max_data_storage_in_tbs"`
	MaxDbNodeStorageSizeInGBs     types.Int32                                                              `tfsdk:"max_db_node_storage_size_in_gbs"`
	MaxMemoryInGBs                types.Int32                                                              `tfsdk:"max_memory_in_gbs"`
	MemorySizeInGBs               types.Int32                                                              `tfsdk:"memory_size_in_gbs"`
	MonthlyDbServerVersion        types.String                                                             `tfsdk:"monthly_db_server_version"`
	MonthlyStorageServerVersion   types.String                                                             `tfsdk:"monthly_storage_server_version"`
	NextMaintenanceRunId          types.String                                                             `tfsdk:"next_maintenance_run_id"`
	OciResourceAnchorName         types.String                                                             `tfsdk:"oci_resource_anchor_name"`
	OciUrl                        types.String                                                             `tfsdk:"oci_url"`
	Ocid                          types.String                                                             `tfsdk:"ocid"`
	PercentProgress               types.Float64                                                            `tfsdk:"percent_progress"`
	Shape                         types.String                                                             `tfsdk:"shape"`
	Status                        fwtypes.StringEnum[odbtypes.ResourceStatus]                              `tfsdk:"status"`
	StatusReason                  types.String                                                             `tfsdk:"status_reason"`
	StorageCount                  types.Int32                                                              `tfsdk:"storage_count"`
	StorageServerVersion          types.String                                                             `tfsdk:"storage_server_version"`
	TotalStorageSizeInGBs         types.Int32                                                              `tfsdk:"total_storage_size_in_gbs"`
	CustomerContactsToSendToOCI   fwtypes.SetNestedObjectValueOf[customerContactDataSourceModel]           `tfsdk:"customer_contacts_to_send_to_oci"`
	ComputeModel                  fwtypes.StringEnum[odbtypes.ComputeModel]                                `tfsdk:"compute_model"`
	CreatedAt                     types.String                                                             `tfsdk:"created_at" autoflex:",noflatten"`
	DatabaseServerType            types.String                                                             `tfsdk:"database_server_type"`
	StorageServerType             types.String                                                             `tfsdk:"storage_server_type"`
	MaintenanceWindow             fwtypes.ObjectValueOf[cloudExadataInfraMaintenanceWindowDataSourceModel] `tfsdk:"maintenance_window" autoflex:",noflatten"`
	Tags                          tftags.Map                                                               `tfsdk:"tags"`
}

type cloudExadataInfraMaintenanceWindowDataSourceModel struct {
	CustomActionTimeoutInMins    types.Int32                                                    `tfsdk:"custom_action_timeout_in_mins"`
	DaysOfWeek                   fwtypes.SetValueOf[fwtypes.StringEnum[odbtypes.DayOfWeekName]] `tfsdk:"days_of_week"`
	HoursOfDay                   fwtypes.SetValueOf[types.Int64]                                `tfsdk:"hours_of_day"`
	IsCustomActionTimeoutEnabled types.Bool                                                     `tfsdk:"is_custom_action_timeout_enabled"`
	LeadTimeInWeeks              types.Int32                                                    `tfsdk:"lead_time_in_weeks"`
	Months                       fwtypes.SetValueOf[fwtypes.StringEnum[odbtypes.MonthName]]     `tfsdk:"months"`
	PatchingMode                 fwtypes.StringEnum[odbtypes.PatchingModeType]                  `tfsdk:"patching_mode"`
	Preference                   fwtypes.StringEnum[odbtypes.PreferenceType]                    `tfsdk:"preference"`
	WeeksOfMonth                 fwtypes.SetValueOf[types.Int64]                                `tfsdk:"weeks_of_month"`
}
type customerContactDataSourceModel struct {
	Email types.String `tfsdk:"email"`
}
