// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Security Config")
func newDataSourceSecurityConfig(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceSecurityConfig{}, nil
}

const (
	DSNameSecurityConfig = "Security Config Data Source"
)

type dataSourceSecurityConfig struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceSecurityConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_opensearchserverless_security_config"
}

func (d *dataSourceSecurityConfig) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreatedDate: schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			"last_modified_date": schema.StringAttribute{
				Computed: true,
			},
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"saml_options": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"group_attribute": schema.StringAttribute{
						Computed: true,
					},
					"metadata": schema.StringAttribute{
						Computed: true,
					},
					"session_timeout": schema.Int64Attribute{
						Computed: true,
					},
					"user_attribute": schema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func (d *dataSourceSecurityConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data dataSourceSecurityConfigData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSecurityConfigByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, DSNameSecurityConfig, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	createdDate := time.UnixMilli(aws.ToInt64(out.CreatedDate))
	data.CreatedDate = flex.StringValueToFramework(ctx, createdDate.Format(time.RFC3339))

	data.ConfigVersion = flex.StringToFramework(ctx, out.ConfigVersion)
	data.Description = flex.StringToFramework(ctx, out.Description)
	data.ID = flex.StringToFramework(ctx, out.Id)

	lastModifiedDate := time.UnixMilli(aws.ToInt64(out.LastModifiedDate))
	data.LastModifiedDate = flex.StringValueToFramework(ctx, lastModifiedDate.Format(time.RFC3339))

	data.Type = flex.StringValueToFramework(ctx, out.Type)

	samlOptions := flattenSAMLOptions(ctx, out.SamlOptions)
	data.SamlOptions = samlOptions

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceSecurityConfigData struct {
	ConfigVersion    types.String `tfsdk:"config_version"`
	CreatedDate      types.String `tfsdk:"created_date"`
	Description      types.String `tfsdk:"description"`
	ID               types.String `tfsdk:"id"`
	LastModifiedDate types.String `tfsdk:"last_modified_date"`
	SamlOptions      types.Object `tfsdk:"saml_options"`
	Type             types.String `tfsdk:"type"`
}
