// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_iam_outbound_web_identity_federation", name="Outbound Web Identity Federation")
func newOutboundWebIdentityFederationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &outboundWebIdentityFederationDataSource{}, nil
}

const (
	DSNameOutboundWebIdentityFederation = "Outbound Web Identity Federation Data Source"
)

type outboundWebIdentityFederationDataSource struct {
	framework.DataSourceWithModel[outboundWebIdentityFederationDataSourceModel]
}

func (d *outboundWebIdentityFederationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrEnabled: schema.BoolAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			"issuer_identifier": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *outboundWebIdentityFederationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().IAMClient(ctx)

	var data outboundWebIdentityFederationDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findOutboundWebIdentityFederation(ctx, conn)
	if retry.NotFound(err) {
		// Feature is disabled
		data.Enabled = types.BoolValue(false)
		data.ID = types.StringValue(d.Meta().AccountID(ctx))
		data.IssuerIdentifier = types.StringNull()

		smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	data.Enabled = types.BoolValue(out.JwtVendingEnabled)
	data.ID = types.StringValue(d.Meta().AccountID(ctx))
	data.IssuerIdentifier = flex.StringToFramework(ctx, out.IssuerIdentifier)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type outboundWebIdentityFederationDataSourceModel struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	ID               types.String `tfsdk:"id"`
	IssuerIdentifier types.String `tfsdk:"issuer_identifier"`
}
