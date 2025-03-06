// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_servicecatalogappregistry_attribute_group_associations", name="Attribute Group Associations")
func newDataSourceAttributeGroupAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAttributeGroupAssociations{}, nil
}

const (
	DSNameAttributeGroupAssociations = "Attribute Group Associations Data Source"
)

type dataSourceAttributeGroupAssociations struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceAttributeGroupAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"attribute_group_ids": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *dataSourceAttributeGroupAssociations) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrID),
			path.MatchRoot(names.AttrName),
		),
	}
}

func (d *dataSourceAttributeGroupAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ServiceCatalogAppRegistryClient(ctx)

	var data dataSourceAttributeGroupAssociationsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var id string
	if !data.ID.IsNull() {
		id = data.ID.ValueString()
	} else if !data.Name.IsNull() {
		id = data.Name.ValueString()
	}

	out, err := findAttributeGroupAssociationsByID(ctx, conn, id)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionReading, DSNameAttributeGroupAssociations, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	data.AttributeGroups = flex.FlattenFrameworkStringValueSet(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findAttributeGroupAssociationsByID(ctx context.Context, conn *servicecatalogappregistry.Client, id string) ([]string, error) {
	in := &servicecatalogappregistry.ListAssociatedAttributeGroupsInput{
		Application: aws.String(id),
	}

	var out []string
	paginator := servicecatalogappregistry.NewListAssociatedAttributeGroupsPaginator(conn, in)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: in,
				}
			}

			return nil, err
		}

		out = append(out, page.AttributeGroups...)
	}

	return out, nil
}

type dataSourceAttributeGroupAssociationsData struct {
	AttributeGroups types.Set    `tfsdk:"attribute_group_ids"`
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
}
