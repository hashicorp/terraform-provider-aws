// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_quicksight_account_custom_permission")
func newAccountCustomPermissionResourceAsListResource() list.ListResourceWithConfigure {
	return &accountCustomPermissionListResource{}
}

var _ list.ListResource = &accountCustomPermissionListResource{}

type accountCustomPermissionListResource struct {
	accountCustomPermissionResource
	framework.WithList
}

func (l *accountCustomPermissionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().QuickSightClient(ctx)
	accountID := l.Meta().AccountID(ctx)

	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrAWSAccountID), accountID)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey(names.AttrAWSAccountID): accountID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		out, err := findAccountCustomPermissionByID(ctx, conn, accountID)
		if retry.NotFound(err) {
			return
		}
		if err != nil {
			yield(smerr.NewListResultError(ctx, err))
			return
		}

		result := request.NewListResult(ctx)

		var data accountCustomPermissionResourceModel
		l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
			data.AWSAccountID = flex.StringValueToFramework(ctx, accountID)

			smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, out, &data), smerr.ID, accountID)
			if result.Diagnostics.HasError() {
				return
			}

			result.DisplayName = aws.ToString(out.CustomPermissionsName)
		})

		yield(result)
	}
}
