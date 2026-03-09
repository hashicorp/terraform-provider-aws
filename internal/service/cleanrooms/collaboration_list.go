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
// @SDKListResource("aws_cleanrooms_collaboration")
func newCollaborationResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceCollaboration{}
	l.SetResourceSchema(resourceCollaboration())
	return &l
}

var _ list.ListResource = &listResourceCollaboration{}

type listResourceCollaboration struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceCollaboration) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CleanRoomsClient(ctx)

	var query listCollaborationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Clean Rooms Collaboration")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input cleanrooms.ListCollaborationsInput
		for item, err := range listCollaborations(ctx, conn, &input) {
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

			tflog.Info(ctx, "Reading Clean Rooms Collaboration")
			output, err := findCollaborationByID(ctx, conn, id)
			if err != nil {
				tflog.Error(ctx, "Reading Clean Rooms Collaboration", map[string]any{
					names.AttrID: id,
					"err":        err.Error(),
				})
				continue
			}

			diags := resourceCollaborationFlatten(ctx, conn, rd, output)
			if diags.HasError() {
				result.Diagnostics.Append(fwdiag.FromSDKDiagnostics(diags)...)
				yield(result)
				return
			}

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

type listCollaborationModel struct {
	framework.WithRegionModel
}

func listCollaborations(ctx context.Context, conn *cleanrooms.Client, input *cleanrooms.ListCollaborationsInput) iter.Seq2[awstypes.CollaborationSummary, error] {
	return func(yield func(awstypes.CollaborationSummary, error) bool) {
		pages := cleanrooms.NewListCollaborationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.CollaborationSummary{}, fmt.Errorf("listing Clean Rooms Collaboration resources: %w", err))
				return
			}

			for _, item := range page.CollaborationList {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
