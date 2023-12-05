// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkDataSource(name="Repositories")
func newRepositoriesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &repositoriesDataSource{}, nil
}

type repositoriesDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *repositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_ecr_repositories"
}

func (d *repositoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"names": schema.SetAttribute{
				ElementType: fwtypes.StringType,
				Computed:    true,
			},
		},
	}
}
func (d *repositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ECRClient(ctx)

	var data repositoriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var names []string

	pages := ecr.NewDescribeRepositoriesPaginator(conn, &ecr.DescribeRepositoriesInput{})
	for pages.HasMorePages() {
		output, err := pages.NextPage(ctx)

		if err != nil {
			resp.Diagnostics.AddError("reading ECR Repositories", err.Error())
			return
		}

		for _, v := range output.Repositories {
			names = append(names, aws.ToString(v.RepositoryName))
		}
	}

	data.ID = flex.StringValueToFramework(ctx, d.Meta().Region)
	data.Names = flex.FlattenFrameworkStringValueSet(ctx, names)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type repositoriesDataSourceModel struct {
	ID    fwtypes.String `tfsdk:"id"`
	Names fwtypes.Set    `tfsdk:"names"`
}
