// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cleanrooms

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cleanrooms/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_cleanrooms_configured_table")
func newConfiguredTableResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceConfiguredTable{}
	l.SetResourceSchema(resourceConfiguredTable())
	return &l
}

var _ list.ListResource = &listResourceConfiguredTable{}

type listResourceConfiguredTable struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceConfiguredTable) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CleanRoomsClient(ctx)

	var query listConfiguredTableModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Clean Rooms Configured Table")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input cleanrooms.ListConfiguredTablesInput
		for item, err := range listConfiguredTables(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.Id)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)

			tflog.Info(ctx, "Reading Clean Rooms Configured Table")
			output, err := findConfiguredTableByID(ctx, conn, id)
			if err != nil {
				tflog.Error(ctx, "Reading Clean Rooms Configured Table", map[string]any{
					names.AttrID: id,
					"err":        err.Error(),
				})
				continue
			}

			resourceConfiguredTableFlatten(ctx, rd, output)

			result.DisplayName = aws.ToString(item.Name)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
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

type listConfiguredTableModel struct {
	framework.WithRegionModel
}

func listConfiguredTables(ctx context.Context, conn *cleanrooms.Client, input *cleanrooms.ListConfiguredTablesInput) iter.Seq2[awstypes.ConfiguredTableSummary, error] {
	return func(yield func(awstypes.ConfiguredTableSummary, error) bool) {
		pages := cleanrooms.NewListConfiguredTablesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ConfiguredTableSummary{}, fmt.Errorf("listing Clean Rooms Configured Table resources: %w", err))
				return
			}

			for _, item := range page.ConfiguredTableSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
