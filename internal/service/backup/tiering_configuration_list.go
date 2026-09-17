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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_backup_tiering_configuration")
func newTieringConfigurationResourceAsListResource() list.ListResourceWithConfigure {
	return &tieringConfigurationListResource{}
}

var _ list.ListResource = &tieringConfigurationListResource{}

type tieringConfigurationListResource struct {
	tieringConfigurationResource
	framework.WithList
}

func (l *tieringConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BackupClient(ctx)

	var query listTieringConfigurationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input backup.ListTieringConfigurationsInput
		for item, err := range listTieringConfigurations(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.TieringConfigurationName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			// The list operation does not return the resource selection, so the
			// full resource is read only when it is requested.
			var tieringConfiguration *awstypes.TieringConfiguration
			if request.IncludeResource {
				tieringConfiguration, err = findTieringConfigurationByName(ctx, conn, name)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}
			}

			result := request.NewListResult(ctx)

			var data tieringConfigurationResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if tieringConfiguration != nil {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, tieringConfiguration, &data))
				} else {
					smerr.AddEnrich(ctx, &result.Diagnostics, fwflex.Flatten(ctx, &item, &data, fwflex.WithFieldNamePrefix("TieringConfiguration")))
				}
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = name
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listTieringConfigurationModel struct {
	framework.WithRegionModel
}

func listTieringConfigurations(ctx context.Context, conn *backup.Client, input *backup.ListTieringConfigurationsInput) iter.Seq2[awstypes.TieringConfigurationsListMember, error] {
	return func(yield func(awstypes.TieringConfigurationsListMember, error) bool) {
		pages := backup.NewListTieringConfigurationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.TieringConfigurationsListMember](), fmt.Errorf("listing Backup Tiering Configurations: %w", err))
				return
			}

			for _, item := range page.TieringConfigurations {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
