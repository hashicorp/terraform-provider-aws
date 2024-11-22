// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Repositories")
func newRepositoriesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &repositoriesDataSource{}, nil
}

type repositoriesDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *repositoriesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_ecr_repositories"
}

func (d *repositoriesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrNames: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}
func (d *repositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data repositoriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECRClient(ctx)

	output, err := findRepositories(ctx, conn, &ecr.DescribeRepositoriesInput{})

	if err != nil {
		resp.Diagnostics.AddError("reading ECR Repositories", err.Error())

		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, d.Meta().Region)
	data.Names.SetValue = fwflex.FlattenFrameworkStringValueSet(ctx, tfslices.ApplyToAll(output, func(v awstypes.Repository) string {
		return aws.ToString(v.RepositoryName)
	}))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findRepositories(ctx context.Context, conn *ecr.Client, input *ecr.DescribeRepositoriesInput) ([]awstypes.Repository, error) {
	var output []awstypes.Repository

	pages := ecr.NewDescribeRepositoriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Repositories...)
	}

	return output, nil
}

type repositoriesDataSourceModel struct {
	ID    types.String                     `tfsdk:"id"`
	Names fwtypes.SetValueOf[types.String] `tfsdk:"names"`
}
