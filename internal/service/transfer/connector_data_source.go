// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"fmt"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_transfer_connector", name="Connector")
// @Tags
func newConnectorDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &connectorDataSource{}, nil
}

type connectorDataSource struct {
	framework.DataSourceWithModel[connectorDataSourceModel]
}

func (d *connectorDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_role": schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"as2_config":    framework.DataSourceComputedListOfObjectAttribute[as2ConnectorConfigModel](ctx),
			"egress_config": framework.DataSourceComputedListOfObjectAttribute[describedConnectorEgressConfigModel](ctx),
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			"logging_role": schema.StringAttribute{
				Computed: true,
			},
			"security_policy_name": schema.StringAttribute{
				Computed: true,
			},
			"service_managed_egress_ip_addresses": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			"sftp_config":  framework.DataSourceComputedListOfObjectAttribute[sftpConnectorConfigModel](ctx),
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			names.AttrURL: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *connectorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().TransferClient(ctx)

	connectorID := fwflex.StringValueFromFramework(ctx, data.ConnectorID)
	output, err := findConnectorByID(ctx, conn, connectorID)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading Transfer Connector (%s)", connectorID), err.Error())

		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type connectorDataSourceModel struct {
	framework.WithRegionModel
	ARN                             types.String                                                         `tfsdk:"arn"`
	AccessRole                      types.String                                                         `tfsdk:"access_role"`
	AS2Config                       fwtypes.ListNestedObjectValueOf[as2ConnectorConfigModel]             `tfsdk:"as2_config"`
	ConnectorID                     types.String                                                         `tfsdk:"id"`
	EgressConfig                    fwtypes.ListNestedObjectValueOf[describedConnectorEgressConfigModel] `tfsdk:"egress_config"`
	LoggingRole                     types.String                                                         `tfsdk:"logging_role"`
	SecurityPolicyName              types.String                                                         `tfsdk:"security_policy_name"`
	ServiceManagedEgressIPAddresses fwtypes.ListOfString                                                 `tfsdk:"service_managed_egress_ip_addresses"`
	SFTPConfig                      fwtypes.ListNestedObjectValueOf[sftpConnectorConfigModel]            `tfsdk:"sftp_config"`
	Tags                            tftags.Map                                                           `tfsdk:"tags"`
	URL                             types.String                                                         `tfsdk:"url"`
}

type as2ConnectorConfigModel struct {
	BasicAuthSecretID   types.String `tfsdk:"basic_auth_secret_id"`
	Compression         types.String `tfsdk:"compression"`
	EncryptionAlgorithm types.String `tfsdk:"encryption_algorithm"`
	LocalProfileID      types.String `tfsdk:"local_profile_id"`
	MDNResponse         types.String `tfsdk:"mdn_response"`
	MDNSigningAlgorithm types.String `tfsdk:"mdn_signing_algorithm"`
	MessageSubject      types.String `tfsdk:"message_subject"`
	PartnerProfileID    types.String `tfsdk:"partner_profile_id"`
	SigningAlgorithm    types.String `tfsdk:"singing_algorithm"`
}

type sftpConnectorConfigModel struct {
	TrustedHostKeys fwtypes.ListOfString `tfsdk:"trusted_host_keys"`
	UserSecretID    types.String         `tfsdk:"user_secret_id"`
}

type describedConnectorEgressConfigModel struct {
	VPCLattice fwtypes.ListNestedObjectValueOf[describedConnectorVpcLatticeEgressConfigModel] `tfsdk:"vpc_lattice"`
}

type describedConnectorVpcLatticeEgressConfigModel struct {
	PortNumber               types.Int64  `tfsdk:"port_number"`
	ResourceConfigurationARN types.String `tfsdk:"resource_configuration_arn"`
}

var (
	_ fwflex.Flattener = &describedConnectorEgressConfigModel{}
)

func (m *describedConnectorEgressConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.DescribedConnectorEgressConfigMemberVpcLattice:
		var data describedConnectorVpcLatticeEgressConfigModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.VPCLattice = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("egress_config flatten: %T", v),
		)
	}
	return diags
}
