// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_secretsmanager_secret_versions, name="Secret Versions")
func newDataSourceSecretVersions(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceSecretVersions{}, nil
}

const (
	DSNameSecretVersions = "Secret Versions Data Source"
)

type dataSourceSecretVersions struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceSecretVersions) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(20, 2048),
				},
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"include_deprecated": schema.BoolAttribute{
				Optional: true,
			},
			"secret_id": schema.StringAttribute{
				Required: true,
			},
			"versions": schema.ListAttribute{
				Computed:   true,
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsVersionsData](ctx),
			},
		},
	}
}

func (d *dataSourceSecretVersions) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SecretsManagerClient(ctx)

	var data dsSecretVersionsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	paginator := secretsmanager.NewListSecretVersionIdsPaginator(conn, &secretsmanager.ListSecretVersionIdsInput{
		SecretId: data.SecretID.ValueStringPointer(),
	})

	var out secretsmanager.ListSecretVersionIdsOutput
	commonFieldsSet := false
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecretsManager, create.ErrActionReading, DSNameSecretVersions, data.Arn.String(), err),
				err.Error(),
			)
			return
		}

		if page != nil && len(page.Versions) > 0 {
			if !commonFieldsSet {
				out.ARN = page.ARN
				out.Name = page.Name
				commonFieldsSet = true
			}
			out.Versions = append(out.Versions, page.Versions...)
		}
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dsSecretVersionsData struct {
	Arn               types.String                                    `tfsdk:"arn"`
	Name              types.String                                    `tfsdk:"name"`
	IncludeDeprecated types.Bool                                      `tfsdk:"include_deprecated"`
	SecretID          types.String                                    `tfsdk:"secret_id"`
	Versions          fwtypes.ListNestedObjectValueOf[dsVersionsData] `tfsdk:"versions"`
}

type dsVersionsData struct {
	CreatedDate      timetypes.RFC3339                 `tfsdk:"created_time"`
	LastAccessedDate types.String                      `tfsdk:"last_accessed_date"`
	VersionID        types.String                      `tfsdk:"version_id"`
	VersionStages    fwtypes.ListValueOf[types.String] `tfsdk:"version_stages"`
}
