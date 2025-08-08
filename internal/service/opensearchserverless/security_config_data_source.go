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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_opensearchserverless_security_config", name="Security Config")
func newSecurityConfigDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &securityConfigDataSource{}, nil
}

const (
	DSNameSecurityConfig = "Security Config Data Source"
)

type securityConfigDataSource struct {
	framework.DataSourceWithModel[securityConfigDataSourceModel]
}

func (d *securityConfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config_version": schema.StringAttribute{
				Description: "The version of the security configuration.",
				Computed:    true,
			},
			names.AttrCreatedDate: schema.StringAttribute{
				Description: "The date the configuration was created.",
				Computed:    true,
			},
			names.AttrDescription: schema.StringAttribute{
				Description: "The description of the security configuration.",
				Computed:    true,
			},
			names.AttrID: schema.StringAttribute{
				Description: "The unique identifier of the security configuration.",
				Required:    true,
			},
			"last_modified_date": schema.StringAttribute{
				Description: "The date the configuration was last modified.",
				Computed:    true,
			},
			names.AttrType: schema.StringAttribute{
				Description: "The type of security configuration.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"saml_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[samlOptionsData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group_attribute": schema.StringAttribute{
							Description: "Group attribute for this SAML integration.",
							Computed:    true,
						},
						"metadata": schema.StringAttribute{
							Description: "The XML IdP metadata file generated from your identity provider.",
							Computed:    true,
						},
						"session_timeout": schema.Int64Attribute{
							Description: "Session timeout, in minutes. Minimum is 5 minutes and maximum is 720 minutes (12 hours). Default is 60 minutes.",
							Computed:    true,
						},
						"user_attribute": schema.StringAttribute{
							Description: "User attribute for this SAML integration.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *securityConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data securityConfigDataSourceModel
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

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data, fwflex.WithIgnoredFieldNames([]string{"CreatedDate", "LastModifiedDate"}))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Special handling for Unix time conversion
	data.CreatedDate = fwflex.StringValueToFramework(ctx, time.UnixMilli(aws.ToInt64(out.CreatedDate)).Format(time.RFC3339))
	data.LastModifiedDate = fwflex.StringValueToFramework(ctx, time.UnixMilli(aws.ToInt64(out.LastModifiedDate)).Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type securityConfigDataSourceModel struct {
	framework.WithRegionModel
	ConfigVersion    types.String                                     `tfsdk:"config_version"`
	CreatedDate      types.String                                     `tfsdk:"created_date"`
	Description      types.String                                     `tfsdk:"description"`
	ID               types.String                                     `tfsdk:"id"`
	LastModifiedDate types.String                                     `tfsdk:"last_modified_date"`
	SamlOptions      fwtypes.ListNestedObjectValueOf[samlOptionsData] `tfsdk:"saml_options"`
	Type             types.String                                     `tfsdk:"type"`
}
