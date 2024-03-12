// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func newDataSourceDomainName(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDomainName{}, nil
}

const (
	DSNameDomainName = "Domain Name Data Source"
)

type dataSourceDomainName struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceDomainName) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_apigatewayv2_domain_name"
}

func (d *dataSourceDomainName) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_mapping_selection_expression": schema.StringAttribute{
				Computed: true,
			},
			"domain_name": schema.StringAttribute{
				Required: true,
			},
			"domain_name_configurations": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceDomainNameConfigurationData](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"certificate_arn":                        types.StringType,
						"certificate_name":                       types.StringType,
						"certificate_upload_date":                fwtypes.TimestampType,
						"domain_name_status":                     types.StringType,
						"domain_name_status_message":             types.StringType,
						"endpoint_type":                          types.StringType,
						"hosted_zone_id":                         types.StringType,
						"ownership_verification_certificate_arn": types.StringType,
						"security_policy":                        types.StringType,
					},
				},
			},
			"mutual_tls_authentication": schema.ObjectAttribute{
				Computed:   true,
				CustomType: fwtypes.NewObjectTypeOf[dataSourceMutualTlsAuthenticationData](ctx),
				AttributeTypes: map[string]attr.Type{
					"truststore_uri":      types.StringType,
					"truststore_version":  types.StringType,
					"truststore_warnings": fwtypes.ListOfStringType,
				},
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *dataSourceDomainName) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().APIGatewayV2Conn(ctx)

	var data dataSourceDomainNameData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindDomainName(ctx, conn, data.DomainName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGatewayV2, create.ErrActionReading, DSNameDomainName, data.DomainName.String(), err),
			err.Error(),
		)
		return
	}

	diags := flex.Flatten(ctx, out, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	ignoreTagsConfig := d.Meta().IgnoreTagsConfig
	tags := KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	data.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.Map())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceDomainNameData struct {
	ApiMappingSelectionExpression types.String                                                           `tfsdk:"api_mapping_selection_expression"`
	DomainName                    types.String                                                           `tfsdk:"domain_name"`
	DomainNameConfigurations      fwtypes.ListNestedObjectValueOf[dataSourceDomainNameConfigurationData] `tfsdk:"domain_name_configurations"`
	MutualTlsAuthentication       fwtypes.ObjectValueOf[dataSourceMutualTlsAuthenticationData]           `tfsdk:"mutual_tls_authentication"`
	Tags                          types.Map                                                              `tfsdk:"tags"`
}

type dataSourceDomainNameConfigurationData struct {
	CertificateArn                      types.String      `tfsdk:"certificate_arn"`
	CertificateName                     types.String      `tfsdk:"certificate_name"`
	CertificateUploadDate               fwtypes.Timestamp `tfsdk:"certificate_upload_date"`
	DomainNameStatus                    types.String      `tfsdk:"domain_name_status"`
	DomainNameStatusMessage             types.String      `tfsdk:"domain_name_status_message"`
	EndpointType                        types.String      `tfsdk:"endpoint_type"`
	HostedZoneId                        types.String      `tfsdk:"hosted_zone_id"`
	OwnershipVerificationCertificateArn types.String      `tfsdk:"ownership_verification_certificate_arn"`
	SecurityPolicy                      types.String      `tfsdk:"security_policy"`
}

type dataSourceMutualTlsAuthenticationData struct {
	TruststoreUri      types.String                      `tfsdk:"truststore_uri"`
	TruststoreVersion  types.String                      `tfsdk:"truststore_version"`
	TruststoreWarnings fwtypes.ListValueOf[types.String] `tfsdk:"truststore_warnings"`
}
