// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_s3files_mount_target", name="Mount Target")
func newMountTargetDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &mountTargetDataSource{}, nil
}

type mountTargetDataSource struct {
	framework.DataSourceWithModel[mountTargetDataSourceModel]
}

func (d *mountTargetDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone_id": schema.StringAttribute{
				Computed:    true,
				Description: "Availability Zone ID",
			},
			names.AttrFileSystemID: schema.StringAttribute{
				Computed:    true,
				Description: "File system ID",
			},
			names.AttrID: schema.StringAttribute{
				Required:    true,
				Description: "Mount target ID",
			},
			"ipv4_address": schema.StringAttribute{
				Computed:    true,
				Description: "IPv4 address",
			},
			"ipv6_address": schema.StringAttribute{
				Computed:    true,
				Description: "IPv6 address",
			},
			names.AttrNetworkInterfaceID: schema.StringAttribute{
				Computed:    true,
				Description: "Network interface ID",
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed:    true,
				Description: "AWS account ID of the owner",
			},
			names.AttrSecurityGroups: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "Security group IDs",
			},
			names.AttrStatus: schema.StringAttribute{
				Computed:    true,
				Description: "Mount target status",
			},
			names.AttrStatusMessage: schema.StringAttribute{
				Computed:    true,
				Description: "Status message",
			},
			names.AttrSubnetID: schema.StringAttribute{
				Computed:    true,
				Description: "Subnet ID",
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed:    true,
				Description: "VPC ID",
			},
		},
	}
}

func (d *mountTargetDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data mountTargetDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().S3FilesClient(ctx)

	output, err := findMountTargetByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringPointerValue(output.MountTargetId)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type mountTargetDataSourceModel struct {
	framework.WithRegionModel
	AvailabilityZoneID types.String                     `tfsdk:"availability_zone_id"`
	FileSystemID       types.String                     `tfsdk:"file_system_id"`
	ID                 types.String                     `tfsdk:"id"`
	Ipv4Address        types.String                     `tfsdk:"ipv4_address"`
	Ipv6Address        types.String                     `tfsdk:"ipv6_address"`
	NetworkInterfaceID types.String                     `tfsdk:"network_interface_id"`
	OwnerID            types.String                     `tfsdk:"owner_id"`
	SecurityGroups     fwtypes.SetValueOf[types.String] `tfsdk:"security_groups"`
	Status             types.String                     `tfsdk:"status"`
	StatusMessage      types.String                     `tfsdk:"status_message"`
	SubnetID           types.String                     `tfsdk:"subnet_id"`
	VPCID              types.String                     `tfsdk:"vpc_id"`
}
