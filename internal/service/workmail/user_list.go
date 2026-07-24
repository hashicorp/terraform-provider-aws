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

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_workmail_user")
func newUserResourceAsListResource() list.ListResourceWithConfigure {
	return &userListResource{}
}

var _ list.ListResource = &userListResource{}

type userListResource struct {
	userResource
	framework.WithList
}

func (l *userListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"organization_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the WorkMail organization to list users from.",
			},
		},
	}
}

func (l *userListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().WorkMailClient(ctx)

	var query listUserModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	organizationID := query.OrganizationID.ValueString()

	tflog.Info(ctx, "Listing WorkMail Users", map[string]any{
		"organization_id": organizationID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := workmail.ListUsersInput{
			OrganizationId: aws.String(organizationID),
		}

		for item, err := range listUsers(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			if item.State == awstypes.EntityStateDeleted {
				continue
			}

			userID := aws.ToString(item.Id)
			userName := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("user_id"), userID)

			result := request.NewListResult(ctx)

			var data userResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.OrganizationId = flex.StringValueToFramework(ctx, organizationID)
				data.UserId = flex.StringValueToFramework(ctx, userID)

				if request.IncludeResource {
					out, err := findUserByTwoPartKey(ctx, conn, organizationID, userID)
					if err != nil {
						result.Diagnostics.Append(fwdiag.NewListResultErrorDiagnostic(err).Diagnostics...)
						return
					}

					result.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = userName
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

type listUserModel struct {
	framework.WithRegionModel
	OrganizationID types.String `tfsdk:"organization_id"`
}

func listUsers(ctx context.Context, conn *workmail.Client, input *workmail.ListUsersInput) iter.Seq2[awstypes.User, error] {
	return func(yield func(awstypes.User, error) bool) {
		pages := workmail.NewListUsersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.User{}, fmt.Errorf("listing WorkMail User resources: %w", err))
				return
			}

			for _, item := range page.Users {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
