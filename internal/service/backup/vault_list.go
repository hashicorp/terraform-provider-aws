// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_backup_vault")
func vaultResourceAsListResource() inttypes.ListResourceForSDK {
	l := vaultListResource{}
	l.SetResourceSchema(resourceVault())
	return &l
}

type vaultListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type vaultListResourceModel struct {
	framework.WithRegionModel
}

func (l *vaultListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query vaultListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := l.Meta()
	conn := awsClient.BackupClient(ctx)

	tflog.Info(ctx, "Listing Backup Vaults")
	stream.Results = func(yield func(list.ListResult) bool) {
		input := backup.ListBackupVaultsInput{
			ByVaultType: awstypes.VaultTypeBackupVault,
		}
		for item, err := range listVaults(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.BackupVaultName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				output, err := findBackupVaultByName(ctx, conn, name)
				if retry.NotFound(err) {
					tflog.Info(ctx, "Backup Vault not found, skipping", map[string]any{
						names.AttrName: name,
					})
					continue
				}
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}

				resourceVaultFlatten(rd, output)
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

func listVaults(ctx context.Context, conn *backup.Client, input *backup.ListBackupVaultsInput) iter.Seq2[awstypes.BackupVaultListMember, error] {
	return func(yield func(awstypes.BackupVaultListMember, error) bool) {
		pages := backup.NewListBackupVaultsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.BackupVaultListMember{}, fmt.Errorf("listing Backup Vaults: %w", err))
				return
			}

			for _, item := range page.BackupVaultList {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
