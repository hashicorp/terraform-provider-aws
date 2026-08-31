// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_odb_exadb_vm_cluster")
func newExaDBVMClusterResourceAsListResource() list.ListResourceWithConfigure {
	return &exaDBVMClusterListResource{}
}

var _ list.ListResource = &exaDBVMClusterListResource{}

type exaDBVMClusterListResource struct {
	exaDBVMClusterResource
	framework.WithList
}

func (l *exaDBVMClusterListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"exascale_db_storage_vault_id": listschema.StringAttribute{
				Optional:    true,
				Description: "Limits results to ExaDB VM Clusters associated with the specified Exascale DB Storage Vault ID.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 2048),
				},
			},
		},
	}
}

func (l *exaDBVMClusterListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ODBClient(ctx)

	var query listExaDBVMClusterModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input odb.ListExadbVmClustersInput
	if !query.ExascaleDBStorageVaultID.IsNull() {
		input.ExascaleDbStorageVaultId = query.ExascaleDBStorageVaultID.ValueStringPointer()
	}

	tflog.Info(ctx, "Listing Oracle Database@AWS ExaDB VM Clusters")

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listExaDBVMClusters(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			arn := aws.ToString(item.ExadbVmClusterArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			id := aws.ToString(item.ExadbVmClusterId)

			var output *odbtypes.ExadbVmCluster
			if request.IncludeResource {
				output, err = findExaDBVMClusterByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data exaDBVMClusterResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				} else {
					data.ExaDBVMClusterID = fwflex.StringValueToFramework(ctx, id)
				}

				result.DisplayName = aws.ToString(item.DisplayName)
			})

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

type listExaDBVMClusterModel struct {
	framework.WithRegionModel
	ExascaleDBStorageVaultID types.String `tfsdk:"exascale_db_storage_vault_id"`
}

func listExaDBVMClusters(ctx context.Context, conn *odb.Client, input *odb.ListExadbVmClustersInput) iter.Seq2[odbtypes.ExadbVmClusterSummary, error] {
	return func(yield func(odbtypes.ExadbVmClusterSummary, error) bool) {
		pages := odb.NewListExadbVmClustersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[odbtypes.ExadbVmClusterSummary](), fmt.Errorf("listing Oracle Database@AWS ExaDB VM Cluster resources: %w", err))
				return
			}

			for _, item := range page.ExadbVmClusters {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
