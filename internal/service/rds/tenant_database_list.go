// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_rds_tenant_database")
func newTenantDatabaseResourceAsListResource() list.ListResourceWithConfigure {
	return &tenantDatabaseListResource{}
}

var _ list.ListResource = &tenantDatabaseListResource{}

type tenantDatabaseListResource struct {
	tenantDatabaseResource
	framework.WithList
}

func (l *tenantDatabaseListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"db_instance_identifier": listschema.StringAttribute{
				Optional:    true,
				Description: "Identifier of the DB instance whose tenant databases to list. Defaults to all DB instances.",
			},
		},
	}
}

func (l *tenantDatabaseListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query listTenantDatabaseModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = smerr.ListStreamEnrich(ctx, diags)
			return
		}
	}

	conn := l.Meta().RDSClient(ctx)

	input := rds.DescribeTenantDatabasesInput{
		DBInstanceIdentifier: fwflex.StringFromFramework(ctx, query.DBInstanceIdentifier),
	}

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("db_instance_identifier"): aws.ToString(input.DBInstanceIdentifier),
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listTenantDatabases(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			id := aws.ToString(item.TenantDatabaseResourceId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("tenant_database_resource_id"), id)

			result := request.NewListResult(ctx)

			var data tenantDatabaseResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data), smerr.ID, id)
				if result.Diagnostics.HasError() {
					return
				}

				// Tenant database names are only unique within a DB instance.
				result.DisplayName = aws.ToString(item.DBInstanceIdentifier) + "/" + aws.ToString(item.TenantDBName)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listTenantDatabaseModel struct {
	framework.WithRegionModel
	DBInstanceIdentifier types.String `tfsdk:"db_instance_identifier"`
}

func listTenantDatabases(ctx context.Context, conn *rds.Client, input *rds.DescribeTenantDatabasesInput) iter.Seq2[awstypes.TenantDatabase, error] {
	return func(yield func(awstypes.TenantDatabase, error) bool) {
		pages := rds.NewDescribeTenantDatabasesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.TenantDatabase{}, fmt.Errorf("listing RDS (Relational Database) Tenant Database resources: %w", err))
				return
			}

			for _, item := range page.TenantDatabases {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
