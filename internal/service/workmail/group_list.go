// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workmail/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
)

// @FrameworkListResource("aws_workmail_group")
func newGroupResourceAsListResource() list.ListResourceWithConfigure {
	return &groupListResource{}
}

var _ list.ListResource = &groupListResource{}

type groupListResource struct {
	groupResource
	framework.WithList
}

func (l *groupListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"organization_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the WorkMail organization to list groups from.",
			},
		},
	}
}

func (l *groupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().WorkMailClient(ctx)

	var query listGroupModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	organizationID := query.OrganizationID.ValueString()

	tflog.Info(ctx, "Listing WorkMail Groups", map[string]any{
		"organization_id": organizationID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := workmail.ListGroupsInput{
			OrganizationId: aws.String(organizationID),
		}

		for item, err := range listGroups(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			if item.State == awstypes.EntityStateDeleted {
				continue
			}

			groupID := aws.ToString(item.Id)
			groupName := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("group_id"), groupID)

			result := request.NewListResult(ctx)

			var data groupResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.OrganizationId = flex.StringValueToFramework(ctx, organizationID)
				data.GroupId = flex.StringValueToFramework(ctx, groupID)

				if request.IncludeResource {
					out, err := findGroupByTwoPartKey(ctx, conn, organizationID, groupID)
					if err != nil {
						result.Diagnostics.Append(fwdiag.NewListResultErrorDiagnostic(err).Diagnostics...)
						return
					}

					result.Diagnostics.Append(l.flatten(ctx, out, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = groupName
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

type listGroupModel struct {
	framework.WithRegionModel
	OrganizationID types.String `tfsdk:"organization_id"`
}

func listGroups(ctx context.Context, conn *workmail.Client, input *workmail.ListGroupsInput) iter.Seq2[awstypes.Group, error] {
	return func(yield func(awstypes.Group, error) bool) {
		pages := workmail.NewListGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Group{}, fmt.Errorf("listing WorkMail Group resources: %w", err))
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
