// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_odb_network", name="Network")
// @Tags(identifierAttribute="arn")
func newResourceNetwork(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceNetwork{}
	r.SetDefaultCreateTimeout(24 * time.Hour)
	r.SetDefaultUpdateTimeout(24 * time.Hour)
	r.SetDefaultDeleteTimeout(24 * time.Hour)

	return r, nil
}

const (
	ResNameNetwork = "Odb Network"
)

type resourceNetwork struct {
	framework.ResourceWithModel[odbNetworkResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

var OracleDBNetwork = newResourceNetwork
var managedServiceTimeout = 15 * time.Minute

func (r *resourceNetwork) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	statusType := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	stringLengthBetween1And255Validator := []validator.String{
		stringvalidator.LengthBetween(1, 255),
	}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrDisplayName: schema.StringAttribute{
				Required:    true,
				Description: "The user-friendly name for the odb network. Changing this will force terraform to create a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrAvailabilityZone: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The name of the Availability Zone (AZ) where the odb network is located. Changing this will force terraform to create new resource",
			},
			"availability_zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The AZ ID of the AZ where the ODB network is located. Changing this will force terraform to create new resource.",
			},
			"client_subnet_cidr": schema.StringAttribute{
				Required:   true,
				Validators: stringLengthBetween1And255Validator,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The CIDR notation for the network resource. Changing this will force terraform to create new resource.\n" +
					" Constraints:\n  " +
					"\t - Must not overlap with the CIDR range of the backup subnet.\n   " +
					"\t- Must not overlap with the CIDR ranges of the VPCs that are connected to the\n  " +
					" ODB network.\n  " +
					"\t- Must not use the following CIDR ranges that are reserved by OCI:\n  " +
					"\t - 100.106.0.0/16 and 100.107.0.0/16\n  " +
					"\t - 169.254.0.0/16\n   " +
					"\t- 224.0.0.0 - 239.255.255.255\n   " +
					"\t- 240.0.0.0 - 255.255.255.255",
			},
			"backup_subnet_cidr": schema.StringAttribute{
				Required:   true,
				Validators: stringLengthBetween1And255Validator,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The CIDR range of the backup subnet for the ODB network. Changing this will force terraform to create new resource.\n" +
					"\tConstraints:\n" +
					"\t   - Must not overlap with the CIDR range of the client subnet.\n" +
					"\t   - Must not overlap with the CIDR ranges of the VPCs that are connected to the\n" +
					"\t   ODB network.\n" +
					"\t   - Must not use the following CIDR ranges that are reserved by OCI:\n" +
					"\t   - 100.106.0.0/16 and 100.107.0.0/16\n" +
					"\t   - 169.254.0.0/16\n" +
					"\t   - 224.0.0.0 - 239.255.255.255\n" +
					"\t   - 240.0.0.0 - 255.255.255.255",
			},

			"custom_domain_name": schema.StringAttribute{
				Optional:   true,
				Validators: stringLengthBetween1And255Validator,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the custom domain that the network is located. custom_domain_name and default_dns_prefix both can't be given.",
			},
			"default_dns_prefix": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The default DNS prefix for the network resource. Changing this will force terraform to create new resource.",
			},
			"s3_access": schema.StringAttribute{
				Required:    true,
				CustomType:  fwtypes.StringEnumType[odbtypes.Access](),
				Description: "Specifies the configuration for Amazon S3 access from the ODB network.",
			},
			"zero_etl_access": schema.StringAttribute{
				Required:    true,
				CustomType:  fwtypes.StringEnumType[odbtypes.Access](),
				Description: "Specifies the configuration for Zero-ETL access from the ODB network.",
			},
			"s3_policy_document": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Specifies the endpoint policy for Amazon S3 access from the ODB network.",
			},
			"oci_dns_forwarding_configs": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[odbNwkOciDnsForwardingConfigResourceModel](ctx),
				Computed:    true,
				Description: "The DNS resolver endpoint in OCI for forwarding DNS queries for the ociPrivateZone domain.",
			},
			"peered_cidrs": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "The list of CIDR ranges from the peered VPC that are allowed access to the ODB network. Please refer odb network peering documentation.",
			},
			"oci_network_anchor_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the OCI network anchor for the ODB network.",
			},
			"oci_network_anchor_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the OCI network anchor for the ODB network.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the OCI resource anchor for the ODB network.",
			},
			"oci_vcn_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier  Oracle Cloud ID (OCID) of the OCI VCN for the ODB network.",
			},
			"oci_vcn_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the OCI VCN for the ODB network.",
			},
			"delete_associated_resources": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "If set to true deletes associated OCI resources. Default false.",
			},
			"percent_progress": schema.Float32Attribute{
				Computed:    true,
				Description: "The amount of progress made on the current operation on the ODB network, expressed as a percentage.",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType:  statusType,
				Computed:    true,
				Description: "The status of the network resource.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the current status of the ODB network.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "The date and time when the ODB network was created.",
			},
			"managed_services": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewListNestedObjectTypeOf[odbNetworkManagedServicesResourceModel](ctx),
				Description: "The managed services configuration for the ODB network.",
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

func (r *resourceNetwork) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)
	var plan odbNetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.CreateOdbNetworkInput{
		Tags: getTagsIn(ctx),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := conn.CreateOdbNetwork(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameNetwork, plan.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.OdbNetworkId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameNetwork, plan.DisplayName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitNetworkCreated(ctx, conn, *out.OdbNetworkId, createTimeout)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), aws.ToString(out.OdbNetworkId))...)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameNetwork, plan.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	//wait for zero etl access
	_, err = waitForManagedService(ctx, plan.ZeroEtlAccess.ValueEnum(), conn, *out.OdbNetworkId, managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
		return managedService.ZeroEtlAccess.Status
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}
	//wait for s3 access
	createdOdbNetwork, err := waitForManagedService(ctx, plan.S3Access.ValueEnum(), conn, *out.OdbNetworkId, managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
		return managedService.S3Access.Status
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}
	//since zero_etl_access, s3_access and s3_policy_document are not returned directly by underlying API we need to set it.
	readZeroEtlAccessStatus, err := mapManagedServiceStatusToAccessStatus(createdOdbNetwork.ManagedServices.ZeroEtlAccess.Status)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, plan.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	plan.ZeroEtlAccess = fwtypes.StringEnumValue(readZeroEtlAccessStatus)

	readS3AccessStatus, err := mapManagedServiceStatusToAccessStatus(createdOdbNetwork.ManagedServices.S3Access.Status)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, plan.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	plan.S3Access = fwtypes.StringEnumValue(readS3AccessStatus)
	plan.S3PolicyDocument = types.StringPointerValue(createdOdbNetwork.ManagedServices.S3Access.S3PolicyDocument)

	resp.Diagnostics.Append(flex.Flatten(ctx, createdOdbNetwork, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceNetwork) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state odbNetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindOracleDBNetworkResourceByID(ctx, conn, state.OdbNetworkId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, state.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}
	if out.ManagedServices == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, state.OdbNetworkId.String(), errors.New("odbNetwork managed service not found")),
			"Odb Network managed service cannot be nil",
		)
		return
	} else {
		readS3AccessStatus, err := mapManagedServiceStatusToAccessStatus(out.ManagedServices.S3Access.Status)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, state.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		state.S3Access = fwtypes.StringEnumValue(readS3AccessStatus)
		state.S3PolicyDocument = types.StringPointerValue(out.ManagedServices.S3Access.S3PolicyDocument)

		readZeroEtlAccessStatus, err := mapManagedServiceStatusToAccessStatus(out.ManagedServices.ZeroEtlAccess.Status)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, state.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		state.ZeroEtlAccess = fwtypes.StringEnumValue(readZeroEtlAccessStatus)
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNetwork) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ODBClient(ctx)
	var plan, state odbNetworkResourceModel
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
		var input odb.UpdateOdbNetworkInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		out, err := conn.UpdateOdbNetwork(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameNetwork, plan.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.OdbNetworkId == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameNetwork, plan.OdbNetworkId.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitNetworkUpdated(ctx, conn, plan.OdbNetworkId.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}

	//zero ETL access
	_, err = waitForManagedService(ctx, plan.ZeroEtlAccess.ValueEnum(), conn, plan.OdbNetworkId.ValueString(), managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
		return managedService.ZeroEtlAccess.Status
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}

	//s3 access
	updatedOdbNwk, err := waitForManagedService(ctx, plan.S3Access.ValueEnum(), conn, plan.OdbNetworkId.ValueString(), managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
		return managedService.S3Access.Status
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}

	readS3AccessStatus, err := mapManagedServiceStatusToAccessStatus(updatedOdbNwk.ManagedServices.S3Access.Status)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, state.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}
	plan.S3Access = fwtypes.StringEnumValue(readS3AccessStatus)
	plan.S3PolicyDocument = types.StringPointerValue(updatedOdbNwk.ManagedServices.S3Access.S3PolicyDocument)

	readZeroEtlAccessStatus, err := mapManagedServiceStatusToAccessStatus(updatedOdbNwk.ManagedServices.ZeroEtlAccess.Status)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, state.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}
	plan.ZeroEtlAccess = fwtypes.StringEnumValue(readZeroEtlAccessStatus)

	resp.Diagnostics.Append(flex.Flatten(ctx, updatedOdbNwk, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceNetwork) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)
	var state odbNetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.DeleteOdbNetworkInput{
		OdbNetworkId: state.OdbNetworkId.ValueStringPointer(),
	}

	input.DeleteAssociatedResources = aws.Bool(false)
	if !state.DeleteAssociatedResources.IsNull() || !state.DeleteAssociatedResources.IsUnknown() {
		input.DeleteAssociatedResources = state.DeleteAssociatedResources.ValueBoolPointer()
	}
	_, err := conn.DeleteOdbNetwork(ctx, &input)

	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameNetwork, state.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitNetworkDeleted(ctx, conn, state.OdbNetworkId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameNetwork, state.OdbNetworkArn.String(), err),
			err.Error(),
		)
		return
	}
}

func waitNetworkCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.OdbNetwork, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusProvisioning),
		Target:  enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh: statusNetwork(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.OdbNetwork); ok {
		return out, err
	}

	return nil, err
}

func waitForManagedService(ctx context.Context, targetStatus odbtypes.Access, conn *odb.Client, id string, timeout time.Duration, managedResourceStatus func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus) (*odbtypes.OdbNetwork, error) {
	switch targetStatus {
	case odbtypes.AccessEnabled:
		stateConf := &sdkretry.StateChangeConf{
			Pending: enum.Slice(odbtypes.ManagedResourceStatusEnabling),
			Target:  enum.Slice(odbtypes.ManagedResourceStatusEnabled),
			Refresh: statusManagedService(ctx, conn, id, managedResourceStatus),
			Timeout: timeout,
		}
		outputRaw, err := stateConf.WaitForStateContext(ctx)
		if out, ok := outputRaw.(*odbtypes.OdbNetwork); ok {
			return out, err
		}
		return nil, err
	case odbtypes.AccessDisabled:
		stateConf := &sdkretry.StateChangeConf{
			Pending: enum.Slice(odbtypes.ManagedResourceStatusDisabling),
			Target:  enum.Slice(odbtypes.ManagedResourceStatusDisabled),
			Refresh: statusManagedService(ctx, conn, id, managedResourceStatus),
			Timeout: timeout,
		}
		outputRaw, err := stateConf.WaitForStateContext(ctx)
		if out, ok := outputRaw.(*odbtypes.OdbNetwork); ok {
			return out, err
		}
		return nil, err
	default:
		return nil, errors.New("odb network invalid manged service status")
	}
}

func statusManagedService(ctx context.Context, conn *odb.Client, id string, managedResourceStatus func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindOracleDBNetworkResourceByID(ctx, conn, id)

		if err != nil {
			return nil, "", err
		}

		if out.ManagedServices == nil {
			return nil, "", nil
		}

		return out, string(managedResourceStatus(out.ManagedServices)), nil
	}
}

func waitNetworkUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.OdbNetwork, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusUpdating),
		Target:  enum.Slice(odbtypes.ResourceStatusAvailable, odbtypes.ResourceStatusFailed),
		Refresh: statusNetwork(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.OdbNetwork); ok {
		return out, err
	}

	return nil, err
}

func waitNetworkDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.OdbNetwork, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(odbtypes.ResourceStatusTerminating),
		Target:  []string{},
		Refresh: statusNetwork(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.OdbNetwork); ok {
		return out, err
	}

	return nil, err
}

func statusNetwork(ctx context.Context, conn *odb.Client, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindOracleDBNetworkResourceByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func mapManagedServiceStatusToAccessStatus(mangedStatus odbtypes.ManagedResourceStatus) (odbtypes.Access, error) {
	if mangedStatus == odbtypes.ManagedResourceStatusDisabled {
		return odbtypes.AccessDisabled, nil
	}
	if mangedStatus == odbtypes.ManagedResourceStatusEnabled {
		return odbtypes.AccessEnabled, nil
	}
	return "", errors.New("can not convert managed status to access status")
}

func FindOracleDBNetworkResourceByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.OdbNetwork, error) {
	input := odb.GetOdbNetworkInput{
		OdbNetworkId: aws.String(id),
	}

	out, err := conn.GetOdbNetwork(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.OdbNetwork == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.OdbNetwork, nil
}

type odbNetworkResourceModel struct {
	framework.WithRegionModel
	DisplayName               types.String                                                               `tfsdk:"display_name"`
	AvailabilityZone          types.String                                                               `tfsdk:"availability_zone"`
	AvailabilityZoneId        types.String                                                               `tfsdk:"availability_zone_id"`
	ClientSubnetCidr          types.String                                                               `tfsdk:"client_subnet_cidr"`
	BackupSubnetCidr          types.String                                                               `tfsdk:"backup_subnet_cidr"`
	CustomDomainName          types.String                                                               `tfsdk:"custom_domain_name"`
	DefaultDnsPrefix          types.String                                                               `tfsdk:"default_dns_prefix"`
	S3Access                  fwtypes.StringEnum[odbtypes.Access]                                        `tfsdk:"s3_access" autoflex:",noflatten"`
	ZeroEtlAccess             fwtypes.StringEnum[odbtypes.Access]                                        `tfsdk:"zero_etl_access" autoflex:",noflatten"`
	S3PolicyDocument          types.String                                                               `tfsdk:"s3_policy_document" autoflex:",noflatten"`
	OdbNetworkId              types.String                                                               `tfsdk:"id"`
	PeeredCidrs               fwtypes.SetValueOf[types.String]                                           `tfsdk:"peered_cidrs"`
	OciDnsForwardingConfigs   fwtypes.ListNestedObjectValueOf[odbNwkOciDnsForwardingConfigResourceModel] `tfsdk:"oci_dns_forwarding_configs"`
	OciNetworkAnchorId        types.String                                                               `tfsdk:"oci_network_anchor_id"`
	OciNetworkAnchorUrl       types.String                                                               `tfsdk:"oci_network_anchor_url"`
	OciResourceAnchorName     types.String                                                               `tfsdk:"oci_resource_anchor_name"`
	OciVcnId                  types.String                                                               `tfsdk:"oci_vcn_id"`
	OciVcnUrl                 types.String                                                               `tfsdk:"oci_vcn_url"`
	OdbNetworkArn             types.String                                                               `tfsdk:"arn"`
	PercentProgress           types.Float32                                                              `tfsdk:"percent_progress"`
	Status                    fwtypes.StringEnum[odbtypes.ResourceStatus]                                `tfsdk:"status"`
	StatusReason              types.String                                                               `tfsdk:"status_reason"`
	Timeouts                  timeouts.Value                                                             `tfsdk:"timeouts"`
	ManagedServices           fwtypes.ListNestedObjectValueOf[odbNetworkManagedServicesResourceModel]    `tfsdk:"managed_services"`
	CreatedAt                 timetypes.RFC3339                                                          `tfsdk:"created_at"`
	DeleteAssociatedResources types.Bool                                                                 `tfsdk:"delete_associated_resources"`
	Tags                      tftags.Map                                                                 `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                                 `tfsdk:"tags_all"`
}

type odbNwkOciDnsForwardingConfigResourceModel struct {
	DomainName       types.String `tfsdk:"domain_name"`
	OciDnsListenerIp types.String `tfsdk:"oci_dns_listener_ip"`
}
type odbNetworkManagedServicesResourceModel struct {
	ServiceNetworkArn        types.String                                                                   `tfsdk:"service_network_arn"`
	ResourceGatewayArn       types.String                                                                   `tfsdk:"resource_gateway_arn"`
	ManagedServicesIpv4Cidrs fwtypes.SetOfString                                                            `tfsdk:"managed_service_ipv4_cidrs"`
	ServiceNetworkEndpoint   fwtypes.ListNestedObjectValueOf[serviceNetworkEndpointOdbNetworkResourceModel] `tfsdk:"service_network_endpoint"`
	ManagedS3BackupAccess    fwtypes.ListNestedObjectValueOf[managedS3BackupAccessOdbNetworkResourceModel]  `tfsdk:"managed_s3_backup_access"`
	ZeroEtlAccess            fwtypes.ListNestedObjectValueOf[zeroEtlAccessOdbNetworkResourceModel]          `tfsdk:"zero_etl_access"`
	S3Access                 fwtypes.ListNestedObjectValueOf[s3AccessOdbNetworkResourceModel]               `tfsdk:"s3_access"`
}

type serviceNetworkEndpointOdbNetworkResourceModel struct {
	VpcEndpointId   types.String                                 `tfsdk:"vpc_endpoint_id"`
	VpcEndpointType fwtypes.StringEnum[odbtypes.VpcEndpointType] `tfsdk:"vpc_endpoint_type"`
}

type managedS3BackupAccessOdbNetworkResourceModel struct {
	Status        fwtypes.StringEnum[odbtypes.ManagedResourceStatus] `tfsdk:"status"`
	Ipv4Addresses fwtypes.SetOfString                                `tfsdk:"ipv4_addresses"`
}

type zeroEtlAccessOdbNetworkResourceModel struct {
	Status fwtypes.StringEnum[odbtypes.ManagedResourceStatus] `tfsdk:"status"`
	Cidr   types.String                                       `tfsdk:"cidr"`
}

type s3AccessOdbNetworkResourceModel struct {
	Status           fwtypes.StringEnum[odbtypes.ManagedResourceStatus] `tfsdk:"status"`
	Ipv4Addresses    fwtypes.SetOfString                                `tfsdk:"ipv4_addresses"`
	DomainName       types.String                                       `tfsdk:"domain_name"`
	S3PolicyDocument types.String                                       `tfsdk:"s3_policy_document"`
}
