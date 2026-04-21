// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package odb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_odb_network", name="Network")
// @Tags(identifierAttribute="arn")
func newDataSourceNetwork(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceNetwork{}, nil
}

const (
	DSNameNetwork = "Odb Network Data Source"
)

type dataSourceNetwork struct {
	framework.DataSourceWithModel[odbNetworkDataSourceModel]
}

func (d *dataSourceNetwork) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	statusType := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			names.AttrDisplayName: schema.StringAttribute{
				Computed:    true,
				Description: "Display name for the network resource.",
			},
			"availability_zone_id": schema.StringAttribute{
				Computed:    true,
				Description: "The AZ ID of the AZ where the ODB network is located.",
			},
			names.AttrAvailabilityZone: schema.StringAttribute{
				Computed:    true,
				Description: "The availability zone where the ODB network is located.",
			},
			"backup_subnet_cidr": schema.StringAttribute{
				Computed:    true,
				Description: " The CIDR range of the backup subnet for the ODB network.",
			},
			"client_subnet_cidr": schema.StringAttribute{
				Computed:    true,
				Description: "The CIDR notation for the network resource.",
			},
			"custom_domain_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the custom domain that the network is located.",
			},
			"default_dns_prefix": schema.StringAttribute{
				Computed:    true,
				Description: "The default DNS prefix for the network resource.",
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
			"percent_progress": schema.Float64Attribute{
				Computed:    true,
				Description: "The amount of progress made on the current operation on the ODB network, expressed as a percentage.",
			},
			"peered_cidrs": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "The list of CIDR ranges from the peered VPC that are allowed access to the ODB network. Please refer odb network peering documentation.",
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
				CustomType:  fwtypes.NewListNestedObjectTypeOf[odbNetworkManagedServicesDataSourceModel](ctx),
				Description: "The managed services configuration for the ODB network.",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"oci_dns_forwarding_configs": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewListNestedObjectTypeOf[odbNwkOciDnsForwardingConfigDataSourceModel](ctx),
				Description: "The DNS resolver endpoint in OCI for forwarding DNS queries for the ociPrivateZone domain.",
			},
		},
	}
}

func (d *dataSourceNetwork) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data odbNetworkDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.GetOdbNetworkInput{
		OdbNetworkId: data.OdbNetworkId.ValueStringPointer(),
	}

	out, err := conn.GetOdbNetwork(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameNetwork, data.OdbNetworkId.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.OdbNetwork, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type odbNetworkDataSourceModel struct {
	framework.WithRegionModel
	AvailabilityZone        types.String                                                                 `tfsdk:"availability_zone"`
	AvailabilityZoneId      types.String                                                                 `tfsdk:"availability_zone_id"`
	BackupSubnetCidr        types.String                                                                 `tfsdk:"backup_subnet_cidr"`
	ClientSubnetCidr        types.String                                                                 `tfsdk:"client_subnet_cidr"`
	CustomDomainName        types.String                                                                 `tfsdk:"custom_domain_name"`
	DefaultDnsPrefix        types.String                                                                 `tfsdk:"default_dns_prefix"`
	DisplayName             types.String                                                                 `tfsdk:"display_name"`
	OciDnsForwardingConfigs fwtypes.ListNestedObjectValueOf[odbNwkOciDnsForwardingConfigDataSourceModel] `tfsdk:"oci_dns_forwarding_configs"`
	OciNetworkAnchorId      types.String                                                                 `tfsdk:"oci_network_anchor_id"`
	OciNetworkAnchorUrl     types.String                                                                 `tfsdk:"oci_network_anchor_url"`
	OciResourceAnchorName   types.String                                                                 `tfsdk:"oci_resource_anchor_name"`
	OciVcnId                types.String                                                                 `tfsdk:"oci_vcn_id"`
	OciVcnUrl               types.String                                                                 `tfsdk:"oci_vcn_url"`
	OdbNetworkArn           types.String                                                                 `tfsdk:"arn"`
	OdbNetworkId            types.String                                                                 `tfsdk:"id"`
	PeeredCidrs             fwtypes.SetValueOf[types.String]                                             `tfsdk:"peered_cidrs"`
	PercentProgress         types.Float64                                                                `tfsdk:"percent_progress"`
	Status                  fwtypes.StringEnum[odbtypes.ResourceStatus]                                  `tfsdk:"status"`
	StatusReason            types.String                                                                 `tfsdk:"status_reason"`
	CreatedAt               timetypes.RFC3339                                                            `tfsdk:"created_at"`
	ManagedServices         fwtypes.ListNestedObjectValueOf[odbNetworkManagedServicesDataSourceModel]    `tfsdk:"managed_services"`
	Tags                    tftags.Map                                                                   `tfsdk:"tags"`
}

type odbNwkOciDnsForwardingConfigDataSourceModel struct {
	DomainName       types.String `tfsdk:"domain_name"`
	OciDnsListenerIp types.String `tfsdk:"oci_dns_listener_ip"`
}

type odbNetworkManagedServicesDataSourceModel struct {
	ServiceNetworkArn                 types.String                                                                                `tfsdk:"service_network_arn"`
	ResourceGatewayArn                types.String                                                                                `tfsdk:"resource_gateway_arn"`
	ManagedServicesIpv4Cidrs          fwtypes.ListOfString                                                                        `tfsdk:"managed_service_ipv4_cidrs"`
	ServiceNetworkEndpoint            fwtypes.ListNestedObjectValueOf[serviceNetworkEndpointOdbNetworkDataSourceModel]            `tfsdk:"service_network_endpoint"`
	ManagedS3BackupAccess             fwtypes.ListNestedObjectValueOf[managedS3BackupAccessOdbNetworkDataSourceModel]             `tfsdk:"managed_s3_backup_access"`
	ZeroEtlAccess                     fwtypes.ListNestedObjectValueOf[zeroEtlAccessOdbNetworkDataSourceModel]                     `tfsdk:"zero_tl_access"`
	S3Access                          fwtypes.ListNestedObjectValueOf[s3AccessOdbNetworkDataSourceModel]                          `tfsdk:"s3_access"`
	StsAccess                         fwtypes.ListNestedObjectValueOf[stsAccessOdbNetworkDataSourceModel]                         `tfsdk:"sts_access"`
	KmsAccess                         fwtypes.ListNestedObjectValueOf[kmsAccessOdbNetworkDataSourceModel]                         `tfsdk:"kms_access"`
	CrossRegionS3RestoreSourcesAccess fwtypes.ListNestedObjectValueOf[crossRegionS3RestoreSourcesAccessOdbNetworkDataSourceModel] `tfsdk:"cross_region_s3_restore_sources_access"`
}

type serviceNetworkEndpointOdbNetworkDataSourceModel struct {
	VpcEndpointId   types.String                                 `tfsdk:"vpc_endpoint_id"`
	VpcEndpointType fwtypes.StringEnum[odbtypes.VpcEndpointType] `tfsdk:"vpc_endpoint_type"`
}

type managedS3BackupAccessOdbNetworkDataSourceModel struct {
	Status        fwtypes.StringEnum[odbtypes.ManagedResourceStatus] `tfsdk:"status"`
	Ipv4Addresses fwtypes.ListOfString                               `tfsdk:"ipv4_addresses"`
}

type zeroEtlAccessOdbNetworkDataSourceModel struct {
	Status fwtypes.StringEnum[odbtypes.ManagedResourceStatus] `tfsdk:"status"`
	Cidr   types.String                                       `tfsdk:"cidr"`
}

type s3AccessOdbNetworkDataSourceModel struct {
	Status           fwtypes.StringEnum[odbtypes.ManagedResourceStatus] `tfsdk:"status"`
	Ipv4Addresses    fwtypes.ListOfString                               `tfsdk:"ipv4_addresses"`
	DomainName       types.String                                       `tfsdk:"domain_name"`
	S3PolicyDocument types.String                                       `tfsdk:"s3_policy_document"`
}

type stsAccessOdbNetworkDataSourceModel struct {
	Status            fwtypes.StringEnum[odbtypes.ManagedResourceStatus] `tfsdk:"status"`
	Ipv4Addresses     fwtypes.ListOfString                               `tfsdk:"ipv4_addresses"`
	DomainName        types.String                                       `tfsdk:"domain_name"`
	StsPolicyDocument types.String                                       `tfsdk:"sts_policy_document"`
}

type kmsAccessOdbNetworkDataSourceModel struct {
	Status            fwtypes.StringEnum[odbtypes.ManagedResourceStatus] `tfsdk:"status"`
	Ipv4Addresses     fwtypes.ListOfString                               `tfsdk:"ipv4_addresses"`
	DomainName        types.String                                       `tfsdk:"domain_name"`
	KmsPolicyDocument types.String                                       `tfsdk:"kms_policy_document"`
}

type crossRegionS3RestoreSourcesAccessOdbNetworkDataSourceModel struct {
	Ipv4Addresses fwtypes.ListOfString                               `tfsdk:"ipv4_addresses"`
	Region        types.String                                       `tfsdk:"region"`
	Status        fwtypes.StringEnum[odbtypes.ManagedResourceStatus] `tfsdk:"status"`
}
