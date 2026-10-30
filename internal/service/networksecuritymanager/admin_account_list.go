// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_networksecuritymanager_admin_account")
func newAdminAccountResourceAsListResource() list.ListResourceWithConfigure {
	return &adminAccountListResource{}
}

var _ list.ListResource = &adminAccountListResource{}

type adminAccountListResource struct {
	adminAccountResource
	framework.WithList
}

func (l *adminAccountListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().NetworkSecurityManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input networksecuritymanager.ListAdminAccountsInput
		for item, err := range listAdminAccounts(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}
			adminAccountID := aws.ToString(item.AccountId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("admin_account_id"), adminAccountID)

			// The summary carries no status, and a removed administrator account
			// stays listed as OFFBOARDED for a while, so every item is read to
			// skip the ones Read would treat as gone.
			output, err := findAdminAccountDetailsByID(ctx, conn, adminAccountID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			result := request.NewListResult(ctx)

			var data adminAccountResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.AdminAccountID = fwflex.StringValueToFramework(ctx, adminAccountID)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = adminAccountID
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

func listAdminAccounts(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListAdminAccountsInput, optFns ...func(*networksecuritymanager.Options)) iter.Seq2[awstypes.AdminAccountSummary, error] {
	return tfiter.ConcatValuesWithError(listAdminAccountPages(ctx, conn, input, optFns...))
}

func listAdminAccountPages(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListAdminAccountsInput, optFns ...func(*networksecuritymanager.Options)) iter.Seq2[[]awstypes.AdminAccountSummary, error] {
	return func(yield func([]awstypes.AdminAccountSummary, error) bool) {
		pages := networksecuritymanager.NewListAdminAccountsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Network Security Manager Admin Accounts: %w", err))
				return
			}

			if !yield(page.AdminAccounts, nil) {
				return
			}
		}
	}
}
