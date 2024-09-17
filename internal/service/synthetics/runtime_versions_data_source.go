// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Runtime Versions")
func newDataSourceRuntimeVersions(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceRuntimeVersions{}, nil
}

const (
	DSNameRuntimeVersions = "Runtime Versions Data Source"
)

type dataSourceRuntimeVersions struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceRuntimeVersions) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_synthetics_runtime_versions"
}

func (d *dataSourceRuntimeVersions) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"skip_deprecated": schema.BoolAttribute{
				Optional: true,
			},
			"version_names": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceRuntimeVersions) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SyntheticsClient(ctx)

	var data dataSourceRuntimeVersionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	skipDeprecated := data.SkipDeprecated.ValueBool()
	out, err := findRuntimeVersions(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Synthetics, create.ErrActionReading, DSNameRuntimeVersions, "", err),
			err.Error(),
		)
		return
	}

	var versionNames []string

	for _, v := range out {
		if !skipDeprecated || v.DeprecationDate == nil {
			versionNames = append(versionNames, aws.ToString(v.VersionName))
		}
	}

	data.ID = flex.StringToFramework(ctx, &d.Meta().Region)
	data.VersionNames = flex.FlattenFrameworkStringValueListOfString(ctx, versionNames)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceRuntimeVersionsModel struct {
	ID             types.String                      `tfsdk:"id"`
	SkipDeprecated types.Bool                        `tfsdk:"skip_deprecated"`
	VersionNames   fwtypes.ListValueOf[types.String] `tfsdk:"version_names"`
}
