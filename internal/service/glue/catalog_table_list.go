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
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_glue_catalog_table")
func newCatalogTableResourceAsListResource() inttypes.ListResourceForSDK {
	l := catalogTableListResource{}
	l.SetResourceSchema(resourceCatalogTable())
	return &l
}

var _ list.ListResource = &catalogTableListResource{}

type catalogTableListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listCatalogTableModel struct {
	framework.WithRegionModel
	CatalogID    fwtypes.String `tfsdk:"catalog_id"`
	DatabaseName fwtypes.String `tfsdk:"database_name"`
}

func (l *catalogTableListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
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

func (l *catalogTableListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.GlueClient(ctx)

	var query listCatalogTableModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	dbName := query.DatabaseName.ValueString()

	tflog.Info(ctx, "Listing Glue Catalog Tables", map[string]any{
		logging.ResourceAttributeKey(names.AttrDatabaseName): dbName,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := glue.GetTablesInput{
			DatabaseName: aws.String(dbName),
		}

		if !query.CatalogID.IsNull() && !query.CatalogID.IsUnknown() {
			input.CatalogId = query.CatalogID.ValueStringPointer()
		}

		for table, err := range listCatalogTables(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			catalogID := aws.ToString(table.CatalogId)
			name := aws.ToString(table.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)
			result := request.NewListResult(ctx)

			id := catalogTableCreateResourceID(catalogID, dbName, name)
			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set(names.AttrName, name)
			rd.Set(names.AttrDatabaseName, dbName)
			rd.Set(names.AttrCatalogID, catalogID)

			if request.IncludeResource {
				if err := resourceCatalogTableFlatten(ctx, awsClient, catalogID, dbName, name, &table, rd); err != nil {
					tflog.Error(ctx, "Flattening Glue Catalog Table", map[string]any{
						names.AttrName: name,
						"error":        err.Error(),
					})
					continue
				}
			}

			result.DisplayName = name

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
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
