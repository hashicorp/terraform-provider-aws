// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkDataSource("aws_iam_outbound_web_identity_federation", name="Outbound Web Identity Federation")
func newOutboundWebIdentityFederationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &outboundWebIdentityFederationDataSource{}

	return d, nil
}

type outboundWebIdentityFederationDataSource struct {
	framework.DataSourceWithModel[outboundWebIdentityFederationDataSourceModel]
}

func (d *outboundWebIdentityFederationDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"issuer_identifier": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *outboundWebIdentityFederationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data outboundWebIdentityFederationDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().IAMClient(ctx)

	output, err := findOutboundWebIdentityFederation(ctx, conn)
	if err != nil {
		response.Diagnostics.AddError("reading IAM Outbound Web Identity Federation", err.Error())
		return
	}

	data.IssuerIdentifier = fwflex.StringToFramework(ctx, output.IssuerIdentifier)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type outboundWebIdentityFederationDataSourceModel struct {
	IssuerIdentifier types.String `tfsdk:"issuer_identifier"`
}
