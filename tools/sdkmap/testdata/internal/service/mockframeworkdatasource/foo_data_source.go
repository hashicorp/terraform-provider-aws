// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockframeworkdatasource

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockframeworkdatasource"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockframeworkdatasource/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_mockframeworkdatasource_foo", name="Foo")
// @FrameworkDataSource("aws_mockframeworkdatasource_legacy_foo", name="Foo")
func newFooDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &fooDataSource{}, nil
}

type fooDataSource struct {
	framework.DataSourceWithModel[fooDataSourceModel]
}

func (d *fooDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data fooDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().MockFrameworkDataSourceClient(ctx)

	name := data.Name.ValueString()
	foo, err := findFooByName(ctx, conn, name)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrName, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, foo, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func findFooByName(ctx context.Context, conn *mockframeworkdatasource.Client, name string) (*awstypes.Foo, error) {
	var foos []awstypes.Foo

	var input mockframeworkdatasource.ListFoosInput
	pages := mockframeworkdatasource.NewListFoosPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, foo := range page.Foos {
			if aws.ToString(foo.Name) == name {
				foos = append(foos, foo)
			}
		}
	}

	return tfslices.SingleElement(foos)
}

type fooDataSourceModel struct {
	framework.WithRegionModel
	ARN  types.String `tfsdk:"arn"`
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
