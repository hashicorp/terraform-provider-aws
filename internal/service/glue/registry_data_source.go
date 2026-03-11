// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_glue_registry", name="Registry")
func newRegistryDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &registryDataSource{}, nil
}

const (
	DSNameRegistry = "Registry Data Source"
)

type registryDataSource struct {
	framework.DataSourceWithModel[registryDataSourceModel]
}

func (d *registryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *registryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().GlueClient(ctx)

	var data registryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRegistryByName(ctx, conn, data.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Glue, create.ErrActionReading, DSNameRegistry, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("Registry"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findRegistryByName(ctx context.Context, conn *glue.Client, name string) (*glue.GetRegistryOutput, error) {
	input := &glue.GetRegistryInput{
		RegistryId: &awstypes.RegistryId{
			RegistryName: aws.String(name),
		},
	}

	output, err := conn.GetRegistry(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

type registryDataSourceModel struct {
	framework.WithRegionModel
	ARN         types.String `tfsdk:"arn"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
}
