// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	smithy "github.com/aws/smithy-go"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_dms_migration_project")
func newMigrationProjectResourceAsListResource() list.ListResourceWithConfigure {
	return &migrationProjectListResource{}
}

var _ list.ListResource = &migrationProjectListResource{}

type migrationProjectListResource struct {
	migrationProjectResource
	framework.WithList
}

func (l *migrationProjectListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().DMSClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input databasemigrationservice.DescribeMigrationProjectsInput
		for item, err := range listMigrationProjects(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			arn := aws.ToString(item.MigrationProjectArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var tags tftags.KeyValueTags
			if request.IncludeResource {
				var err error
				tags, err = listTags(ctx, conn, arn)
				if errs.IsA[*awstypes.ResourceNotFoundFault](err) || errs.IsAErrorMessageContains[smithy.APIError](err, "Unable to find") {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data migrationProjectResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data))
				if result.Diagnostics.HasError() {
					return
				}
				result.DisplayName = aws.ToString(item.MigrationProjectName)
				if request.IncludeResource {
					setTagsOut(ctx, svcTags(tags))
				}
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}
			if !yield(result) {
				return
			}
		}
	}
}

func listMigrationProjects(ctx context.Context, conn *databasemigrationservice.Client, input *databasemigrationservice.DescribeMigrationProjectsInput) iter.Seq2[awstypes.MigrationProject, error] {
	return func(yield func(awstypes.MigrationProject, error) bool) {
		pages := databasemigrationservice.NewDescribeMigrationProjectsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.MigrationProject{}, smarterr.NewError(err))
				return
			}
			for _, item := range page.MigrationProjects {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
