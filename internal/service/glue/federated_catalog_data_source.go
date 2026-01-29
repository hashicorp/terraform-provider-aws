// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_glue_federated_catalog", name="Federated Catalog")
func newDataSourceFederatedCatalog(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceFederatedCatalog{}, nil
}

const (
	DSNameFederatedCatalog = "Federated Catalog Data Source"
)

type dataSourceFederatedCatalog struct {
	framework.DataSourceWithModel[dataSourceFederatedCatalogModel]
}

func (d *dataSourceFederatedCatalog) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCatalogID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"federated_catalog": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[federatedCatalogModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"connection_name": schema.StringAttribute{
							Computed: true,
						},
						names.AttrIdentifier: schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceFederatedCatalog) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().GlueClient(ctx)
	var data dataSourceFederatedCatalogModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogId := data.CatalogId.ValueString()
	catalogName := data.Name.ValueString()

	if catalogId == "" {
		catalogId = d.Meta().AccountID(ctx)

		if catalogName == s3TablesCatalogName {
			catalogId = catalogName
		}
	}

	id := fmt.Sprintf("%s,%s", catalogId, catalogName)
	out, err := findFederatedCatalogByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.Name.ValueString())
		return
	}

	data.ID = types.StringValue(id)
	data.CatalogId = types.StringValue(catalogId)

	if out.ResourceArn != nil {
		data.ARN = types.StringValue(aws.ToString(out.ResourceArn))
	} else {
		partition := d.Meta().Partition(ctx)
		region := d.Meta().Region(ctx)
		accountID := d.Meta().AccountID(ctx)
		if catalogName == s3TablesCatalogName {
			data.ARN = types.StringValue(fmt.Sprintf("arn:%s:glue:%s:%s:catalog/%s", partition, region, accountID, catalogName))
		} else {
			data.ARN = types.StringValue(fmt.Sprintf("arn:%s:glue:%s:%s:catalog", partition, region, accountID))
		}
	}

	if out.Name != nil {
		data.Name = types.StringValue(aws.ToString(out.Name))
	}

	if out.Description != nil {
		data.Description = types.StringValue(aws.ToString(out.Description))
	}
	if out.FederatedCatalog != nil {
		fedCatalogModel := federatedCatalogModel{
			ConnectionName: types.StringValue(aws.ToString(out.FederatedCatalog.ConnectionName)),
			Identifier:     types.StringValue(aws.ToString(out.FederatedCatalog.Identifier)),
		}

		fedCatalogList, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &fedCatalogModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.FederatedCatalog = fedCatalogList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceFederatedCatalogModel struct {
	framework.WithRegionModel
	ARN              types.String                                           `tfsdk:"arn"`
	CatalogId        types.String                                           `tfsdk:"catalog_id"`
	Description      types.String                                           `tfsdk:"description"`
	FederatedCatalog fwtypes.ListNestedObjectValueOf[federatedCatalogModel] `tfsdk:"federated_catalog"`
	ID               types.String                                           `tfsdk:"id"`
	Name             types.String                                           `tfsdk:"name"`
}

// federatedCatalogModel is defined in federated_catalog.go to avoid duplication
