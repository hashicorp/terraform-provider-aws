// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
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

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_odb_exascale_db_storage_vault")
func newExascaleDBStorageVaultResourceAsListResource() list.ListResourceWithConfigure {
	return &exascaleDBStorageVaultListResource{}
}

var _ list.ListResource = &exascaleDBStorageVaultListResource{}

type exascaleDBStorageVaultListResource struct {
	exascaleDBStorageVaultResource
	framework.WithList
}

func (l *exascaleDBStorageVaultListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ODBClient(ctx)

	var query listExascaleDBStorageVaultModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Oracle Database@AWS Exascale DB Storage Vaults")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input odb.ListExascaleDbStorageVaultsInput
		for item, err := range listExascaleDBStorageVaults(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.ExascaleDbStorageVaultArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			id := aws.ToString(item.ExascaleDbStorageVaultId)

			var out *awstypes.ExascaleDbStorageVault
			if request.IncludeResource {
				var err error
				out, err = findExascaleDBStorageVaultByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data exascaleDBStorageVaultResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, out, &data))
					if result.Diagnostics.HasError() {
						return
					}
				} else {
					data.ID = fwflex.StringValueToFramework(ctx, id)
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

type listExascaleDBStorageVaultModel struct {
	framework.WithRegionModel
}

func listExascaleDBStorageVaults(ctx context.Context, conn *odb.Client, input *odb.ListExascaleDbStorageVaultsInput) iter.Seq2[awstypes.ExascaleDbStorageVaultSummary, error] {
	return func(yield func(awstypes.ExascaleDbStorageVaultSummary, error) bool) {
		pages := odb.NewListExascaleDbStorageVaultsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.ExascaleDbStorageVaultSummary](), fmt.Errorf("listing Oracle Database@AWS Exascale DB Storage Vault resources: %w", err))
				return
			}

			for _, item := range page.ExascaleDbStorageVaults {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
