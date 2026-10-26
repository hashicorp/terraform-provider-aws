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
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_backup_selection")
func newSelectionResourceAsListResource() inttypes.ListResourceForSDK {
	l := selectionListResource{}
	l.SetResourceSchema(resourceSelection())
	return &l
}

var _ list.ListResource = &selectionListResource{}

type selectionListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *selectionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.BackupClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		planInput := &backup.ListBackupPlansInput{}
		for plan, err := range listPlans(ctx, conn, planInput) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			planID := aws.ToString(plan.BackupPlanId)
			input := &backup.ListBackupSelectionsInput{
				BackupPlanId: aws.String(planID),
			}

			for item, err := range listSelections(ctx, conn, input) {
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}

				selectionID := aws.ToString(item.SelectionId)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), selectionID)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.SetId(selectionID)
				rd.Set("plan_id", planID)

				if request.IncludeResource {
					selection, err := findSelectionByTwoPartKey(ctx, conn, planID, selectionID)
					if err != nil {
						tflog.Error(ctx, "Reading Backup Selection", map[string]any{
							"error": err.Error(),
						})
						continue
					}

					if err := resourceSelectionFlatten(planID, rd, selection); err != nil {
						tflog.Error(ctx, "Flattening Backup Selection", map[string]any{
							"error": err,
						})
						continue
					}
				}

				result.DisplayName = aws.ToString(item.SelectionName)

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
}

func listSelections(ctx context.Context, conn *backup.Client, input *backup.ListBackupSelectionsInput) iter.Seq2[awstypes.BackupSelectionsListMember, error] {
	return func(yield func(awstypes.BackupSelectionsListMember, error) bool) {
		pages := backup.NewListBackupSelectionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.BackupSelectionsListMember{}, fmt.Errorf("listing Backup Selections: %w", err))
				return
			}

			for _, item := range page.BackupSelectionsList {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
