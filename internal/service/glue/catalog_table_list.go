// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_glue_catalog_table")
func newCatalogTableResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceCatalogTable(), framework.SDKListSpec[awstypes.Table, listCatalogTableModel]{
		ConfigSchema: listCatalogTableConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listCatalogTableModel) (iter.Seq2[awstypes.Table, error], diag.Diagnostics) {
			input := glue.GetTablesInput{
				DatabaseName: query.DatabaseName.ValueStringPointer(),
			}
			if !query.CatalogID.IsNull() && !query.CatalogID.IsUnknown() {
				input.CatalogId = query.CatalogID.ValueStringPointer()
			}

			return listCatalogTables(ctx, c.GlueClient(ctx), &input), nil
		},
		ID: func(t awstypes.Table, query listCatalogTableModel) string {
			return catalogTableCreateResourceID(aws.ToString(t.CatalogId), query.DatabaseName.ValueString(), aws.ToString(t.Name))
		},
		Prepare: func(_ context.Context, t awstypes.Table, query listCatalogTableModel, rd *schema.ResourceData) {
			// All three make up this resource's identity, and Flatten only runs
			// when resource state is asked for.
			rd.Set(names.AttrCatalogID, aws.ToString(t.CatalogId))
			rd.Set(names.AttrDatabaseName, query.DatabaseName.ValueString())
			rd.Set(names.AttrName, aws.ToString(t.Name))
		},
		Flatten: resourceCatalogTableFlatten,
		// Flatten reads the table's partition indexes, so a listing that only
		// wants identity should not pay for that call.
		FlattenWhenRequested: true,
		DisplayName:          func(t awstypes.Table) string { return aws.ToString(t.Name) },
	})
}

type listCatalogTableModel struct {
	framework.WithRegionModel
	CatalogID    fwtypes.String `tfsdk:"catalog_id"`
	DatabaseName fwtypes.String `tfsdk:"database_name"`
}

func listCatalogTables(ctx context.Context, conn *glue.Client, input *glue.GetTablesInput) iter.Seq2[awstypes.Table, error] {
	return func(yield func(awstypes.Table, error) bool) {
		pages := glue.NewGetTablesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Table{}, fmt.Errorf("listing Glue Catalog Tables: %w", err))
				return
			}
			for _, table := range page.TableList {
				if !yield(table, nil) {
					return
				}
			}
		}
	}
}

func listCatalogTableConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrCatalogID: listschema.StringAttribute{
				Optional:    true,
				Description: "ID of the Glue Catalog to list tables from. Defaults to the AWS account ID.",
			},
			names.AttrDatabaseName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the Glue Catalog Database to list tables from.",
			},
		},
	}
}
