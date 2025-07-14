// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_odb_cloud_exadata_infrastructure", name="Cloud Exadata Infrastructure")
// @Tags(identifierAttribute="arn")
func newResourceCloudExadataInfrastructure(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCloudExadataInfrastructure{}

	r.SetDefaultCreateTimeout(24 * time.Hour)
	r.SetDefaultUpdateTimeout(24 * time.Hour)
	r.SetDefaultDeleteTimeout(24 * time.Hour)

	return r, nil
}

const (
	ResNameCloudExadataInfrastructure     = "Cloud Exadata Infrastructure"
	ExaInfraStorageServerTypeNotAvailable = "Storage_Server_Type_NA"
	ExaInfraDBServerTypeNotAvailable      = "DB_Server_Type_NA"
)

type resourceCloudExadataInfrastructure struct {
	framework.ResourceWithModel[cloudExadataInfrastructureResourceModel]
	framework.WithTimeouts
}

// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (r *resourceCloudExadataInfrastructure) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	statusType := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	computeModelType := fwtypes.StringEnumType[odbtypes.ComputeModel]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"activated_storage_count": schema.Int32Attribute{
				Computed: true,
			},
			"additional_storage_count": schema.Int32Attribute{
				Computed: true,
			},
			"database_server_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"storage_server_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"available_storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"availability_zone": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"compute_count": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"cpu_count": schema.Int32Attribute{
				Computed: true,
			},
			"customer_contacts_to_send_to_oci": schema.SetAttribute{
				ElementType: types.StringType,
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Computed: true,
			},
			"db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"db_server_version": schema.StringAttribute{
				Computed: true,
			},
			"display_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_maintenance_run_id": schema.StringAttribute{
				Computed: true,
			},
			"max_cpu_count": schema.Int32Attribute{
				Computed: true,
			},
			"max_data_storage_in_tbs": schema.Float64Attribute{
				Computed: true,
			},
			"max_db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"max_memory_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"monthly_db_server_version": schema.StringAttribute{
				Computed: true,
			},
			"monthly_storage_server_version": schema.StringAttribute{
				Computed: true,
			},
			"next_maintenance_run_id": schema.StringAttribute{
				Computed: true,
			},
			"ocid": schema.StringAttribute{
				Computed: true,
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed: true,
			},
			"oci_url": schema.StringAttribute{
				Computed: true,
			},
			"percent_progress": schema.Float64Attribute{
				Computed: true,
			},
			"shape": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				CustomType: statusType,
				Computed:   true,
			},
			"status_reason": schema.StringAttribute{
				Computed: true,
			},
			"storage_count": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"storage_server_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"total_storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"compute_model": schema.StringAttribute{
				CustomType: computeModelType,
				Computed:   true,
			},
			"maintenance_window": schema.ObjectAttribute{
				Required:   true,
				CustomType: fwtypes.NewObjectTypeOf[cloudExadataInfraMaintenanceWindowResourceModel](ctx),

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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceCloudExadataInfrastructure) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan cloudExadataInfrastructureResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.CreateCloudExadataInfrastructureInput{
		Tags: getTagsIn(ctx),
	}
	maintenanceWindowInput, err := r.expandMaintenanceWindow(ctx, plan.MaintenanceWindow)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudExadataInfrastructure, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
	}
	input.MaintenanceWindow = maintenanceWindowInput

	if !plan.CustomerContactsToSendToOCI.IsNull() && !plan.CustomerContactsToSendToOCI.IsUnknown() {
		input.CustomerContactsToSendToOCI = r.expandCustomerContacts(ctx, plan.CustomerContactsToSendToOCI)
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateCloudExadataInfrastructure(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudExadataInfrastructure, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.CloudExadataInfrastructureId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameCloudExadataInfrastructure, plan.DisplayName.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	createdExaInfra, err := waitCloudExadataInfrastructureCreated(ctx, conn, *out.CloudExadataInfrastructureId, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameCloudExadataInfrastructure, plan.DisplayName.ValueString(), err),
			err.Error(),
		)
		return
	}

	plan.CustomerContactsToSendToOCI = r.flattenCustomerContacts(createdExaInfra.CustomerContactsToSendToOCI)
	plan.MaintenanceWindow = r.flattenMaintenanceWindow(ctx, createdExaInfra.MaintenanceWindow)

	plan.CreatedAt = types.StringValue(createdExaInfra.CreatedAt.Format(time.RFC3339))

	if createdExaInfra.DatabaseServerType == nil {
		plan.DatabaseServerType = types.StringValue(ExaInfraDBServerTypeNotAvailable)
	} else {
		plan.DatabaseServerType = types.StringValue(*createdExaInfra.DatabaseServerType)
	}
	if createdExaInfra.StorageServerType == nil {
		plan.StorageServerType = types.StringValue(ExaInfraStorageServerTypeNotAvailable)
	} else {
		plan.StorageServerType = types.StringValue(*createdExaInfra.StorageServerType)
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, createdExaInfra, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceCloudExadataInfrastructure) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state cloudExadataInfrastructureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindOdbExadataInfraResourceByID(ctx, conn, state.CloudExadataInfrastructureId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.String(), err),
			err.Error(),
		)
		return
	}

	state.CustomerContactsToSendToOCI = r.flattenCustomerContacts(out.CustomerContactsToSendToOCI)
	state.CreatedAt = types.StringValue(out.CreatedAt.Format(time.RFC3339))

	state.MaintenanceWindow = r.flattenMaintenanceWindow(ctx, out.MaintenanceWindow)

	if out.DatabaseServerType == nil {
		state.DatabaseServerType = types.StringValue("ExaInfraDBServerTypeNotAvailable")
	} else {
		state.DatabaseServerType = types.StringValue(*out.DatabaseServerType)
	}
	if out.StorageServerType == nil {
		state.StorageServerType = types.StringValue("ExaInfraStorageServerTypeNotAvailable")
	} else {
		state.StorageServerType = types.StringValue(*out.StorageServerType)
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCloudExadataInfrastructure) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state cloudExadataInfrastructureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().ODBClient(ctx)

	if !state.MaintenanceWindow.Equal(plan.MaintenanceWindow) {

		//we need to call update maintenance window

		updatedMW := odb.UpdateCloudExadataInfrastructureInput{
			CloudExadataInfrastructureId: plan.CloudExadataInfrastructureId.ValueStringPointer(),
		}
		mwInput, err := r.expandMaintenanceWindow(ctx, plan.MaintenanceWindow)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.ValueString(), err),
				err.Error(),
			)
			return
		}
		updatedMW.MaintenanceWindow = mwInput

		/*if updatedMW.MaintenanceWindow != nil {
			fmt.Println("MW preference : ", updatedMW.MaintenanceWindow.Preference)
			fmt.Println("MW hours : ", updatedMW.MaintenanceWindow.HoursOfDay)
			fmt.Println("MW weeks : ", updatedMW.MaintenanceWindow.WeeksOfMonth)
		} else {
			fmt.Println("Nil MW")
		}*/
		out, err := conn.UpdateCloudExadataInfrastructure(ctx, &updatedMW)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.ValueString(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.ValueString(), err),
				err.Error(),
			)
			return
		}

	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	updatedExaInfra, err := waitCloudExadataInfrastructureUpdated(ctx, conn, state.CloudExadataInfrastructureId.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.ValueString(), err),
			err.Error(),
		)
		return
	}
	plan.CustomerContactsToSendToOCI = r.flattenCustomerContacts(updatedExaInfra.CustomerContactsToSendToOCI)
	plan.CreatedAt = types.StringValue(updatedExaInfra.CreatedAt.Format(time.RFC3339))
	plan.MaintenanceWindow = r.flattenMaintenanceWindow(ctx, updatedExaInfra.MaintenanceWindow)
	if updatedExaInfra.DatabaseServerType == nil {
		plan.DatabaseServerType = types.StringValue(ExaInfraDBServerTypeNotAvailable)
	} else {
		plan.DatabaseServerType = types.StringValue(*updatedExaInfra.DatabaseServerType)
	}
	if updatedExaInfra.StorageServerType == nil {
		plan.StorageServerType = types.StringValue(ExaInfraStorageServerTypeNotAvailable)
	} else {
		plan.StorageServerType = types.StringValue(*updatedExaInfra.StorageServerType)
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, updatedExaInfra, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

func (r *resourceCloudExadataInfrastructure) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state cloudExadataInfrastructureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.DeleteCloudExadataInfrastructureInput{
		CloudExadataInfrastructureId: state.CloudExadataInfrastructureId.ValueStringPointer(),
	}

	_, err := conn.DeleteCloudExadataInfrastructure(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitCloudExadataInfrastructureDeleted(ctx, conn, state.CloudExadataInfrastructureId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameCloudExadataInfrastructure, state.CloudExadataInfrastructureId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCloudExadataInfrastructure) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitCloudExadataInfrastructureCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudExadataInfrastructure, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(odbtypes.ResourceStatusProvisioning),
		Target:       enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh:      statusCloudExadataInfrastructure(ctx, conn, id),
		PollInterval: 1 * time.Minute,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudExadataInfrastructure); ok {
		return out, err
	}
	return nil, err
}

func waitCloudExadataInfrastructureUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudExadataInfrastructure, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(odbtypes.ResourceStatusUpdating),
		Target:       enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh:      statusCloudExadataInfrastructure(ctx, conn, id),
		PollInterval: 1 * time.Minute,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudExadataInfrastructure); ok {
		return out, err
	}

	return nil, err
}

func waitCloudExadataInfrastructureDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.CloudExadataInfrastructure, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusTerminating),
		Target:  []string{},
		Refresh: statusCloudExadataInfrastructure(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.CloudExadataInfrastructure); ok {
		return out, err
	}

	return nil, err
}

func statusCloudExadataInfrastructure(ctx context.Context, conn *odb.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindOdbExadataInfraResourceByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}
func (r *resourceCloudExadataInfrastructure) expandCustomerContacts(ctx context.Context, contactsList fwtypes.SetValueOf[types.String]) []odbtypes.CustomerContact {
	if contactsList.IsNull() || contactsList.IsUnknown() {
		return nil
	}

	var contacts []types.String

	contactsList.ElementsAs(ctx, &contacts, false)

	result := make([]odbtypes.CustomerContact, 0, len(contacts))
	for _, element := range contacts {
		result = append(result, odbtypes.CustomerContact{
			Email: element.ValueStringPointer(),
		})
	}

	return result
}

func (r *resourceCloudExadataInfrastructure) flattenCustomerContacts(contacts []odbtypes.CustomerContact) fwtypes.SetValueOf[types.String] {
	if len(contacts) == 0 {
		return fwtypes.SetValueOf[types.String]{
			SetValue: basetypes.NewSetNull(types.StringType),
		}
	}

	elements := make([]attr.Value, 0, len(contacts))
	for _, contact := range contacts {
		if contact.Email != nil {
			stringValue := types.StringValue(*contact.Email)
			elements = append(elements, stringValue)
		}
	}

	list, _ := basetypes.NewSetValue(types.StringType, elements)

	return fwtypes.SetValueOf[types.String]{
		SetValue: list,
	}
}

func FindOdbExadataInfraResourceByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudExadataInfrastructure, error) {
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
func (r *resourceCloudExadataInfrastructure) expandMaintenanceWindow(ctx context.Context, exaInfraMWResourceObj fwtypes.ObjectValueOf[cloudExadataInfraMaintenanceWindowResourceModel]) (*odbtypes.MaintenanceWindow, error) {

	hasError := false
	var err error
	var exaInfraMWResource cloudExadataInfraMaintenanceWindowResourceModel

	exaInfraMWResourceObj.As(ctx, &exaInfraMWResource, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})

	if exaInfraMWResource.DaysOfWeek.IsNull() || exaInfraMWResource.DaysOfWeek.IsUnknown() {
		hasError = true
		err = errors.Join(err, errors.New("daysOfWeeks can not be null or unknown"))
	}
	if exaInfraMWResource.HoursOfDay.IsNull() || exaInfraMWResource.HoursOfDay.IsUnknown() {
		hasError = true
		err = errors.Join(err, errors.New("hoursOfDay can not be null or unknown"))
	}
	if exaInfraMWResource.WeeksOfMonth.IsNull() || exaInfraMWResource.WeeksOfMonth.IsUnknown() {
		hasError = true
		err = errors.Join(err, errors.New("weeksOfMonth can not be null or unknown"))
	}
	if exaInfraMWResource.Months.IsNull() || exaInfraMWResource.Months.IsUnknown() {
		hasError = true
		err = errors.Join(err, errors.New("months can not be null or unknown"))
	}
	if exaInfraMWResource.IsCustomActionTimeoutEnabled.IsNull() || exaInfraMWResource.IsCustomActionTimeoutEnabled.IsUnknown() {
		hasError = true
		err = errors.Join(err, errors.New("isCustomActionTimeoutEnabled can not be null or unknown"))
	}
	if exaInfraMWResource.CustomActionTimeoutInMins.IsNull() || exaInfraMWResource.CustomActionTimeoutInMins.IsUnknown() {
		hasError = true
		err = errors.Join(err, errors.New("customActionTimeoutInMins can not be null or unknown"))
	}
	if exaInfraMWResource.WeeksOfMonth.IsNull() || exaInfraMWResource.WeeksOfMonth.IsUnknown() {
		hasError = true
		err = errors.Join(err, errors.New("weeksOfMonth can not be null or unknown"))
	}
	if exaInfraMWResource.LeadTimeInWeeks.IsNull() || exaInfraMWResource.LeadTimeInWeeks.IsUnknown() {
		hasError = true
		err = errors.Join(err, errors.New("leadTimeInWeeks can not be null or unknown"))
	}

	var daysOfWeekNames []odbtypes.DayOfWeekName
	exaInfraMWResource.DaysOfWeek.ElementsAs(ctx, &daysOfWeekNames, false)
	daysOfWeek := make([]odbtypes.DayOfWeek, 0, len(daysOfWeekNames))

	for _, dayOfWeek := range daysOfWeekNames {
		daysOfWeek = append(daysOfWeek, odbtypes.DayOfWeek{
			Name: dayOfWeek,
		})
	}

	var hoursOfTheDay []int32
	exaInfraMWResource.HoursOfDay.ElementsAs(ctx, &hoursOfTheDay, false)

	var monthNames []odbtypes.MonthName
	exaInfraMWResource.Months.ElementsAs(ctx, &monthNames, false)
	months := make([]odbtypes.Month, 0, len(monthNames))
	for _, month := range monthNames {
		months = append(months, odbtypes.Month{
			Name: month,
		})
	}

	var weeksOfMonth []int32
	exaInfraMWResource.WeeksOfMonth.ElementsAs(ctx, &weeksOfMonth, false)
	odbTypeMW := odbtypes.MaintenanceWindow{
		CustomActionTimeoutInMins:    exaInfraMWResource.CustomActionTimeoutInMins.ValueInt32Pointer(),
		DaysOfWeek:                   daysOfWeek,
		HoursOfDay:                   hoursOfTheDay,
		IsCustomActionTimeoutEnabled: exaInfraMWResource.IsCustomActionTimeoutEnabled.ValueBoolPointer(),
		LeadTimeInWeeks:              exaInfraMWResource.LeadTimeInWeeks.ValueInt32Pointer(),
		Months:                       months,
		PatchingMode:                 exaInfraMWResource.PatchingMode.ValueEnum(),
		Preference:                   exaInfraMWResource.Preference.ValueEnum(),
		WeeksOfMonth:                 weeksOfMonth,
	}
	if odbTypeMW.Preference == odbtypes.PreferenceTypeNoPreference {
		if len(odbTypeMW.DaysOfWeek) != 0 {
			hasError = true
			err = errors.Join(err, errors.New("default maintenance window can't have daysOfWeek with values"))
		}
		if len(odbTypeMW.HoursOfDay) != 0 {
			hasError = true
			err = errors.Join(err, errors.New("default maintenance window can't have hoursOfDay with values"))
		}
		if len(odbTypeMW.WeeksOfMonth) != 0 {
			hasError = true
			err = errors.Join(err, errors.New("default maintenance window can't have weeksOfMonth with values"))
		}
		if len(odbTypeMW.Months) != 0 {
			hasError = true
			err = errors.Join(err, errors.New("default maintenance window can't have months with values"))
		}
	}

	if hasError {
		return nil, err
	}

	return &odbTypeMW, nil
}

func (r *resourceCloudExadataInfrastructure) flattenMaintenanceWindow(ctx context.Context, obdExaInfraMW *odbtypes.MaintenanceWindow) fwtypes.ObjectValueOf[cloudExadataInfraMaintenanceWindowResourceModel] {
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

	flattenMW := cloudExadataInfraMaintenanceWindowResourceModel{
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
	result, _ := fwtypes.NewObjectValueOf[cloudExadataInfraMaintenanceWindowResourceModel](ctx, &flattenMW)
	return result
}

// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type cloudExadataInfrastructureResourceModel struct {
	framework.WithRegionModel
	ActivatedStorageCount         types.Int32                                                            `tfsdk:"activated_storage_count"`
	AdditionalStorageCount        types.Int32                                                            `tfsdk:"additional_storage_count"`
	DatabaseServerType            types.String                                                           `tfsdk:"database_server_type"`
	StorageServerType             types.String                                                           `tfsdk:"storage_server_type"`
	AvailabilityZone              types.String                                                           `tfsdk:"availability_zone"`
	AvailabilityZoneId            types.String                                                           `tfsdk:"availability_zone_id"`
	AvailableStorageSizeInGBs     types.Int32                                                            `tfsdk:"available_storage_size_in_gbs"`
	CloudExadataInfrastructureArn types.String                                                           `tfsdk:"arn"`
	CloudExadataInfrastructureId  types.String                                                           `tfsdk:"id"`
	ComputeCount                  types.Int32                                                            `tfsdk:"compute_count"`
	CpuCount                      types.Int32                                                            `tfsdk:"cpu_count"`
	CustomerContactsToSendToOCI   fwtypes.SetValueOf[types.String]                                       `tfsdk:"customer_contacts_to_send_to_oci" autoflex:"-"`
	DataStorageSizeInTBs          types.Float64                                                          `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs        types.Int32                                                            `tfsdk:"db_node_storage_size_in_gbs"`
	DbServerVersion               types.String                                                           `tfsdk:"db_server_version"`
	DisplayName                   types.String                                                           `tfsdk:"display_name"`
	LastMaintenanceRunId          types.String                                                           `tfsdk:"last_maintenance_run_id"`
	MaxCpuCount                   types.Int32                                                            `tfsdk:"max_cpu_count"`
	MaxDataStorageInTBs           types.Float64                                                          `tfsdk:"max_data_storage_in_tbs"`
	MaxDbNodeStorageSizeInGBs     types.Int32                                                            `tfsdk:"max_db_node_storage_size_in_gbs"`
	MaxMemoryInGBs                types.Int32                                                            `tfsdk:"max_memory_in_gbs"`
	MemorySizeInGBs               types.Int32                                                            `tfsdk:"memory_size_in_gbs"`
	MonthlyDbServerVersion        types.String                                                           `tfsdk:"monthly_db_server_version"`
	MonthlyStorageServerVersion   types.String                                                           `tfsdk:"monthly_storage_server_version"`
	NextMaintenanceRunId          types.String                                                           `tfsdk:"next_maintenance_run_id"`
	Ocid                          types.String                                                           `tfsdk:"ocid"`
	OciResourceAnchorName         types.String                                                           `tfsdk:"oci_resource_anchor_name"`
	OciUrl                        types.String                                                           `tfsdk:"oci_url"`
	PercentProgress               types.Float64                                                          `tfsdk:"percent_progress"`
	Shape                         types.String                                                           `tfsdk:"shape"`
	Status                        fwtypes.StringEnum[odbtypes.ResourceStatus]                            `tfsdk:"status"`
	StatusReason                  types.String                                                           `tfsdk:"status_reason"`
	StorageCount                  types.Int32                                                            `tfsdk:"storage_count"`
	StorageServerVersion          types.String                                                           `tfsdk:"storage_server_version"`
	TotalStorageSizeInGBs         types.Int32                                                            `tfsdk:"total_storage_size_in_gbs"`
	Timeouts                      timeouts.Value                                                         `tfsdk:"timeouts"`
	CreatedAt                     types.String                                                           `tfsdk:"created_at" autoflex:",noflatten"`
	ComputeModel                  fwtypes.StringEnum[odbtypes.ComputeModel]                              `tfsdk:"compute_model"`
	MaintenanceWindow             fwtypes.ObjectValueOf[cloudExadataInfraMaintenanceWindowResourceModel] `tfsdk:"maintenance_window" autoflex:"-"`
	Tags                          tftags.Map                                                             `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                             `tfsdk:"tags_all"`
}

type cloudExadataInfraMaintenanceWindowResourceModel struct {
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

// TIP: ==== SWEEPERS ====
// When acceptance testing resources, interrupted or failed tests may
// leave behind orphaned resources in an account. To facilitate cleaning
// up lingering resources, each resource implementation should include
// a corresponding "sweeper" function.
//
// The sweeper function lists all resources of a given type and sets the
// appropriate identifers required to delete the resource via the Delete
// method implemented above.
//
// Once the sweeper function is implemented, register it in sweep.go
// as follows:
//
//	awsv2.Register("aws_odb_cloud_exadata_infrastructure", sweepCloudExadataInfrastructures)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepCloudExadataInfrastructures(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := odb.ListCloudExadataInfrastructuresInput{}
	conn := client.ODBClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := odb.NewListCloudExadataInfrastructuresPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.CloudExadataInfrastructures {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceCloudExadataInfrastructure, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.CloudExadataInfrastructureId))),
			)
		}
	}

	return sweepResources, nil
}
