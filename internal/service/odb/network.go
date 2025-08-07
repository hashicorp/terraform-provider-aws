//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
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
}

var OdbNetwork = newResourceNetwork
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
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "Display name for the network resource.",
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
				Description: " The CIDR range of the backup subnet for the ODB network. Changing this will force terraform to create new resource.\n" +
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
				Description: "pecifies the configuration for Zero-ETL access from the ODB network.",
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
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"domain_name":         types.StringType,
						"oci_dns_listener_ip": types.StringType,
					},
				},
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
			"percent_progress": schema.Float32Attribute{
				Computed:    true,
				Description: "The amount of progress made on the current operation on the ODB network, expressed as a percentage.",
			},
			"status": schema.StringAttribute{
				CustomType:  statusType,
				Computed:    true,
				Description: "The status of the network resource.",
			},
			"status_reason": schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the current status of the ODB network.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time when the ODB network was created.",
			},
			"managed_services": schema.ObjectAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewObjectTypeOf[odbNetworkManagedServicesResourceModel](ctx),
				Description: "The managed services configuration for the ODB network.",
				AttributeTypes: map[string]attr.Type{
					"service_network_arn":  types.StringType,
					"resource_gateway_arn": types.StringType,
					"managed_service_ipv4_cidrs": types.SetType{
						ElemType: types.StringType,
					},
					"service_network_endpoint": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"vpc_endpoint_id":   types.StringType,
							"vpc_endpoint_type": fwtypes.StringEnumType[odbtypes.VpcEndpointType](),
						},
					},
					"managed_s3_backup_access": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"status": fwtypes.StringEnumType[odbtypes.ManagedResourceStatus](),
							"ipv4_addresses": types.SetType{
								ElemType: types.StringType,
							},
						},
					},
					"zero_tl_access": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"status": fwtypes.StringEnumType[odbtypes.ManagedResourceStatus](),
							"cidr":   types.StringType,
						},
					},
					"s3_access": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"status": fwtypes.StringEnumType[odbtypes.ManagedResourceStatus](),
							"ipv4_addresses": types.SetType{
								ElemType: types.StringType,
							},
							"domain_name":        types.StringType,
							"s3_policy_document": types.StringType,
						},
					},
				},
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
		ClientToken: aws.String(id.UniqueId()),
		Tags:        getTagsIn(ctx),
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
	createdOdbNetwork, err := waitNetworkCreated(ctx, conn, *out.OdbNetworkId, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameNetwork, plan.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	//set zero etl access
	if plan.ZeroEtlAccess.ValueEnum() == odbtypes.AccessEnabled {
		_, err = waitManagedServiceEnabled(ctx, conn, *createdOdbNetwork.OdbNetworkId, managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
			return managedService.ZeroEtlAccess.Status
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		plan.ZeroEtlAccess = fwtypes.StringEnumValue(odbtypes.AccessEnabled)

	} else if plan.ZeroEtlAccess.ValueEnum() == odbtypes.AccessDisabled {
		_, err = waitManagedServiceDisabled(ctx, conn, *createdOdbNetwork.OdbNetworkId, managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
			return managedService.ZeroEtlAccess.Status
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		plan.ZeroEtlAccess = fwtypes.StringEnumValue(odbtypes.AccessDisabled)
	}

	//set s3 access
	if plan.S3Access.ValueEnum() == odbtypes.AccessEnabled {
		_, err = waitManagedServiceEnabled(ctx, conn, *createdOdbNetwork.OdbNetworkId, managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
			return managedService.S3Access.Status
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		plan.S3Access = fwtypes.StringEnumValue(odbtypes.AccessEnabled)

	} else if plan.S3Access.ValueEnum() == odbtypes.AccessDisabled {
		_, err = waitManagedServiceDisabled(ctx, conn, *createdOdbNetwork.OdbNetworkId, managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
			return managedService.S3Access.Status
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		plan.S3Access = fwtypes.StringEnumValue(odbtypes.AccessDisabled)
	}

	plan.S3PolicyDocument = types.StringPointerValue(createdOdbNetwork.ManagedServices.S3Access.S3PolicyDocument)
	plan.CreatedAt = types.StringValue(createdOdbNetwork.CreatedAt.Format(time.RFC3339))

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

	out, err := FindOdbNetworkByID(ctx, conn, state.OdbNetworkId.ValueString())
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
	if out.ManagedServices != nil {

		readS3AccessStatus, err := managedServiceStatusToAccessStatus(out.ManagedServices.S3Access.Status)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, state.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		state.S3Access = fwtypes.StringEnumValue(readS3AccessStatus)
		state.S3PolicyDocument = types.StringPointerValue(out.ManagedServices.S3Access.S3PolicyDocument)

		readZeroEtlAccessStatus, err := managedServiceStatusToAccessStatus(out.ManagedServices.ZeroEtlAccess.Status)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, state.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		state.ZeroEtlAccess = fwtypes.StringEnumValue(readZeroEtlAccessStatus)
	} else {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameNetwork, state.OdbNetworkId.String(), errors.New("odbNetwork managed service not found")),
			"Odb Network managed service cannot be nil",
		)
		return
	}
	state.CreatedAt = types.StringValue(out.CreatedAt.Format(time.RFC3339))
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

	isUpdateRequired := false
	var input odb.UpdateOdbNetworkInput

	if !plan.DisplayName.Equal(state.DisplayName) {
		isUpdateRequired = true
		input.DisplayName = plan.DisplayName.ValueStringPointer()
	}

	if !plan.S3Access.Equal(state.S3Access) {
		isUpdateRequired = true
		input.S3Access = plan.S3Access.ValueEnum()
	}

	if !plan.ZeroEtlAccess.Equal(state.ZeroEtlAccess) {
		isUpdateRequired = true
		input.ZeroEtlAccess = plan.ZeroEtlAccess.ValueEnum()
	}
	if !plan.S3PolicyDocument.Equal(state.S3PolicyDocument) {
		isUpdateRequired = true
		if !plan.S3PolicyDocument.IsNull() || !plan.S3PolicyDocument.IsUnknown() {
			input.S3PolicyDocument = plan.S3PolicyDocument.ValueStringPointer()
		}
	}
	if isUpdateRequired {
		input.OdbNetworkId = state.OdbNetworkId.ValueStringPointer()

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
	updatedOdbNwk, err := waitNetworkUpdated(ctx, conn, plan.OdbNetworkId.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}

	if plan.S3Access.ValueEnum() == odbtypes.AccessEnabled {
		_, err = waitManagedServiceEnabled(ctx, conn, plan.OdbNetworkId.ValueString(), managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
			return managedService.S3Access.Status
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		plan.S3Access = fwtypes.StringEnumValue(odbtypes.AccessEnabled)

	} else if plan.S3Access.ValueEnum() == odbtypes.AccessDisabled {
		_, err = waitManagedServiceDisabled(ctx, conn, plan.OdbNetworkId.ValueString(), managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
			return managedService.S3Access.Status
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		plan.S3Access = fwtypes.StringEnumValue(odbtypes.AccessDisabled)
	}

	if plan.ZeroEtlAccess.ValueEnum() == odbtypes.AccessEnabled {
		_, err = waitManagedServiceEnabled(ctx, conn, plan.OdbNetworkId.ValueString(), managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
			return managedService.ZeroEtlAccess.Status
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		plan.ZeroEtlAccess = fwtypes.StringEnumValue(odbtypes.AccessEnabled)

	} else if plan.ZeroEtlAccess.ValueEnum() == odbtypes.AccessDisabled {
		_, err = waitManagedServiceDisabled(ctx, conn, plan.OdbNetworkId.ValueString(), managedServiceTimeout, func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
			return managedService.ZeroEtlAccess.Status
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameNetwork, plan.OdbNetworkId.String(), err),
				err.Error(),
			)
			return
		}
		plan.ZeroEtlAccess = fwtypes.StringEnumValue(odbtypes.AccessDisabled)
	}
	plan.S3PolicyDocument = types.StringPointerValue(updatedOdbNwk.ManagedServices.S3Access.S3PolicyDocument)
	plan.CreatedAt = types.StringValue(updatedOdbNwk.CreatedAt.Format(time.RFC3339))

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

	deleteAssociatedResources := false
	input := odb.DeleteOdbNetworkInput{
		OdbNetworkId:              state.OdbNetworkId.ValueStringPointer(),
		DeleteAssociatedResources: &deleteAssociatedResources,
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

func (r *resourceNetwork) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitNetworkCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.OdbNetwork, error) {
	stateConf := &retry.StateChangeConf{
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

func managedServiceStatusToAccessStatus(mangedStatus odbtypes.ManagedResourceStatus) (odbtypes.Access, error) {
	if mangedStatus == odbtypes.ManagedResourceStatusDisabled {
		return odbtypes.AccessDisabled, nil
	}
	if mangedStatus == odbtypes.ManagedResourceStatusEnabled {
		return odbtypes.AccessEnabled, nil
	}
	return "", errors.New("can not convert managed status to access status")
}

func waitManagedServiceEnabled(ctx context.Context, conn *odb.Client, id string, timeout time.Duration, managedResourceStatus func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus) (*odbtypes.OdbNetwork, error) {
	stateConf := &retry.StateChangeConf{
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
}

func waitManagedServiceDisabled(ctx context.Context, conn *odb.Client, id string, timeout time.Duration, managedResourceStatus func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus) (*odbtypes.OdbNetwork, error) {
	stateConf := &retry.StateChangeConf{
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
}

func statusManagedService(ctx context.Context, conn *odb.Client, id string, managedResourceStatus func(managedService *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindOdbNetworkByID(ctx, conn, id)

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
	stateConf := &retry.StateChangeConf{
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
	stateConf := &retry.StateChangeConf{
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

func statusNetwork(ctx context.Context, conn *odb.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindOdbNetworkByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindOdbNetworkByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.OdbNetwork, error) {
	input := odb.GetOdbNetworkInput{
		OdbNetworkId: aws.String(id),
	}

	out, err := conn.GetOdbNetwork(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
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

func sweepNetworks(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := odb.ListOdbNetworksInput{}
	conn := client.ODBClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := odb.NewListOdbNetworksPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.OdbNetworks {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceNetwork, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.OdbNetworkId))),
			)
		}
	}

	return sweepResources, nil
}

type odbNetworkResourceModel struct {
	framework.WithRegionModel
	DisplayName             types.String                                                               `tfsdk:"display_name"`
	AvailabilityZoneId      types.String                                                               `tfsdk:"availability_zone_id"`
	ClientSubnetCidr        types.String                                                               `tfsdk:"client_subnet_cidr"`
	BackupSubnetCidr        types.String                                                               `tfsdk:"backup_subnet_cidr"`
	CustomDomainName        types.String                                                               `tfsdk:"custom_domain_name"`
	DefaultDnsPrefix        types.String                                                               `tfsdk:"default_dns_prefix"`
	S3Access                fwtypes.StringEnum[odbtypes.Access]                                        `tfsdk:"s3_access" autoflex:",noflatten"`
	ZeroEtlAccess           fwtypes.StringEnum[odbtypes.Access]                                        `tfsdk:"zero_etl_access" autoflex:",noflatten"`
	S3PolicyDocument        types.String                                                               `tfsdk:"s3_policy_document" autoflex:",noflatten"`
	OdbNetworkId            types.String                                                               `tfsdk:"id"`
	PeeredCidrs             fwtypes.SetValueOf[types.String]                                           `tfsdk:"peered_cidrs"`
	OciDnsForwardingConfigs fwtypes.ListNestedObjectValueOf[odbNwkOciDnsForwardingConfigResourceModel] `tfsdk:"oci_dns_forwarding_configs"`
	OciNetworkAnchorId      types.String                                                               `tfsdk:"oci_network_anchor_id"`
	OciNetworkAnchorUrl     types.String                                                               `tfsdk:"oci_network_anchor_url"`
	OciResourceAnchorName   types.String                                                               `tfsdk:"oci_resource_anchor_name"`
	OciVcnId                types.String                                                               `tfsdk:"oci_vcn_id"`
	OciVcnUrl               types.String                                                               `tfsdk:"oci_vcn_url"`
	OdbNetworkArn           types.String                                                               `tfsdk:"arn"`
	PercentProgress         types.Float32                                                              `tfsdk:"percent_progress"`
	Status                  fwtypes.StringEnum[odbtypes.ResourceStatus]                                `tfsdk:"status"`
	StatusReason            types.String                                                               `tfsdk:"status_reason"`
	Timeouts                timeouts.Value                                                             `tfsdk:"timeouts"`
	ManagedServices         fwtypes.ObjectValueOf[odbNetworkManagedServicesResourceModel]              `tfsdk:"managed_services"`
	CreatedAt               types.String                                                               `tfsdk:"created_at" autoflex:",noflatten"`
	Tags                    tftags.Map                                                                 `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                                 `tfsdk:"tags_all"`
}

type odbNwkOciDnsForwardingConfigResourceModel struct {
	DomainName       types.String `tfsdk:"domain_name"`
	OciDnsListenerIp types.String `tfsdk:"oci_dns_listener_ip"`
}
type odbNetworkManagedServicesResourceModel struct {
	ServiceNetworkArn        types.String                                                         `tfsdk:"service_network_arn"`
	ResourceGatewayArn       types.String                                                         `tfsdk:"resource_gateway_arn"`
	ManagedServicesIpv4Cidrs fwtypes.SetOfString                                                  `tfsdk:"managed_service_ipv4_cidrs"`
	ServiceNetworkEndpoint   fwtypes.ObjectValueOf[serviceNetworkEndpointOdbNetworkResourceModel] `tfsdk:"service_network_endpoint"`
	ManagedS3BackupAccess    fwtypes.ObjectValueOf[managedS3BackupAccessOdbNetworkResourceModel]  `tfsdk:"managed_s3_backup_access"`
	ZeroEtlAccess            fwtypes.ObjectValueOf[zeroEtlAccessOdbNetworkResourceModel]          `tfsdk:"zero_etl_access"`
	S3Access                 fwtypes.ObjectValueOf[s3AccessOdbNetworkResourceModel]               `tfsdk:"s3_access"`
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
