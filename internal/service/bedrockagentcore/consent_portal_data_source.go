// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bedrockagentcore_consent_portal", name="Consent Portal")
// @Tags(identifierAttribute="consent_portal_arn")
// @Testing(preCheck="testAccPreCheckConsentPortals")
func newConsentPortalDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &consentPortalDataSource{}, nil
}

type consentPortalDataSource struct {
	framework.DataSourceWithModel[consentPortalDataSourceModel]
}

func (d *consentPortalDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"consent_portal_identifier": schema.StringAttribute{Required: true},
			"consent_portal_arn":        schema.StringAttribute{Computed: true},
			"consent_portal_id":         schema.StringAttribute{Computed: true},
			names.AttrDescription:       schema.StringAttribute{Computed: true},
			names.AttrExecutionRoleARN:  schema.StringAttribute{Computed: true, CustomType: fwtypes.ARNType},
			"idp_config":                framework.DataSourceComputedListOfObjectAttribute[consentPortalIDPConfigModel](ctx),
			names.AttrName:              schema.StringAttribute{Computed: true},
			"portal_url":                schema.StringAttribute{Computed: true},
			"sources":                   framework.DataSourceComputedListOfObjectAttribute[consentPortalSourceModel](ctx),
			names.AttrTags:              tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *consentPortalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data consentPortalDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}
	conn := d.Meta().BedrockAgentCoreClient(ctx)
	out, err := findConsentPortalByID(ctx, conn, data.ConsentPortalIdentifier.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ConsentPortalIdentifier.String())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type consentPortalDataSourceModel struct {
	framework.WithRegionModel
	ConsentPortalIdentifier types.String                                                 `tfsdk:"consent_portal_identifier"`
	ConsentPortalARN        types.String                                                 `tfsdk:"consent_portal_arn"`
	ConsentPortalID         types.String                                                 `tfsdk:"consent_portal_id"`
	Description             types.String                                                 `tfsdk:"description" autoflex:",omitempty"`
	ExecutionRoleARN        fwtypes.ARN                                                  `tfsdk:"execution_role_arn"`
	IDPConfig               fwtypes.ListNestedObjectValueOf[consentPortalIDPConfigModel] `tfsdk:"idp_config"`
	Name                    types.String                                                 `tfsdk:"name"`
	PortalURL               types.String                                                 `tfsdk:"portal_url"`
	Sources                 fwtypes.ListNestedObjectValueOf[consentPortalSourceModel]    `tfsdk:"sources"`
	Tags                    tftags.Map                                                   `tfsdk:"tags"`
}
