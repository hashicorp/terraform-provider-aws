// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_servicecatalogappregistry_application_attribute_group_associations", name="Application Attribute Group Associations")
func newDataSourceApplicationAttributeGroupAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceApplicationAttributeGroupAssociations{}, nil
}

const (
	DSNameApplicationAttributeGroupAssociations = "Application Attribute Group Associations Data Source"
)

type dataSourceApplicationAttributeGroupAssociations struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceApplicationAttributeGroupAssociations) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_servicecatalogappregistry_application_attribute_group_associations"
}

func (d *dataSourceApplicationAttributeGroupAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			"attribute_group_ids": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}
func (d *dataSourceApplicationAttributeGroupAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ServiceCatalogAppRegistryClient(ctx)

	var data dataSourceApplicationAttributeGroupAssociationsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationAttributeGroupAssociationsByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceCatalogAppRegistry, create.ErrActionReading, DSNameApplicationAttributeGroupAssociations, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	data.AttributeGroups = flex.FlattenFrameworkStringValueSet(ctx, out.AttributeGroups)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findApplicationAttributeGroupAssociationsByID(ctx context.Context, conn *servicecatalogappregistry.Client, id string) (*servicecatalogappregistry.ListAssociatedAttributeGroupsOutput, error) {
	in := &servicecatalogappregistry.ListAssociatedAttributeGroupsInput{
		Application: aws.String(id),
	}

	out, err := conn.ListAssociatedAttributeGroups(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type dataSourceApplicationAttributeGroupAssociationsData struct {
	ID              types.String `tfsdk:"id"`
	AttributeGroups types.Set    `tfsdk:"attribute_group_ids"`
}
