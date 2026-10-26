// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directoryservicedata

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservicedata"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservicedata/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// @FrameworkListResource("aws_directoryservicedata_group")
func newGroupResourceAsListResource() list.ListResourceWithConfigure {
	return &groupListResource{}
}

var _ list.ListResource = &groupListResource{}

type groupListResource struct {
	groupResource
	framework.WithList
}

func (l *groupListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"directory_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the Directory to list Groups from.",
			},
		},
	}
}

func (l *groupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().DirectoryServiceDataClient(ctx)

	var query listGroupModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	directoryID := query.DirectoryID.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("directory_id"): directoryID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := directoryservicedata.ListGroupsInput{
			DirectoryId: aws.String(directoryID),
		}
		for item, err := range listGroups(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			samAccountName := aws.ToString(item.SAMAccountName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("sam_account_name"), samAccountName)

			var out *directoryservicedata.DescribeGroupOutput
			if request.IncludeResource {
				var err error
				out, err = findGroupByTwoPartKey(ctx, conn, directoryID, samAccountName)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data groupResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.DirectoryID = types.StringValue(directoryID)
				data.SAMAccountName = types.StringValue(samAccountName)

				if request.IncludeResource {
					result.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}
				result.DisplayName = samAccountName
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listGroupModel struct {
	framework.WithRegionModel
	DirectoryID types.String `tfsdk:"directory_id"`
}

func listGroups(ctx context.Context, conn *directoryservicedata.Client, input *directoryservicedata.ListGroupsInput) iter.Seq2[awstypes.GroupSummary, error] {
	return func(yield func(awstypes.GroupSummary, error) bool) {
		pages := directoryservicedata.NewListGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.GroupSummary{}, fmt.Errorf("listing Directory Service Data Group resources: %w", err))
				return
			}

			for _, item := range page.Groups {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
