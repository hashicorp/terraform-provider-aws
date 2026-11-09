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

// @SDKListResource("aws_backup_plan")
func newPlanResourceAsListResource() inttypes.ListResourceForSDK {
	l := planListResource{}
	l.SetResourceSchema(resourcePlan())
	return &l
}

var _ list.ListResource = &planListResource{}

type planListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *planListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BackupClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := backup.ListBackupPlansInput{}
		for item, err := range listPlans(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.BackupPlanId)
			name := aws.ToString(item.BackupPlanName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				findOut, err := findPlanByID(ctx, conn, id)
				if err != nil {
					tflog.Error(ctx, "Reading Backup Plan", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourcePlanFlatten(ctx, findOut, rd); err != nil {
					tflog.Error(ctx, "Flattening Backup Plan", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = name

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

func listPlans(ctx context.Context, conn *backup.Client, input *backup.ListBackupPlansInput) iter.Seq2[awstypes.BackupPlansListMember, error] {
	return func(yield func(awstypes.BackupPlansListMember, error) bool) {
		pages := backup.NewListBackupPlansPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.BackupPlansListMember{}, fmt.Errorf("listing Backup Plans: %w", err))
				return
			}

			for _, item := range page.BackupPlansList {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
