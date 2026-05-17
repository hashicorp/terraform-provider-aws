// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusPermissions(conn *lakeformation.Client, input *lakeformation.ListPermissionsInput, filter PermissionsFilter) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		var permissions []awstypes.PrincipalResourcePermissions

		pages := lakeformation.NewListPermissionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if errs.IsA[*awstypes.EntityNotFoundException](err) {
				return nil, "", nil
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Invalid principal") {
				return nil, statusIAMDelay, nil
			}

			if err != nil {
				return nil, "", fmt.Errorf("listing permissions: %w", err)
			}

			for _, permission := range page.PrincipalResourcePermissions {
				if filter(permission) {
					permissions = append(permissions, permission)
				}
			}
		}

		if len(permissions) == 0 {
			return nil, "", nil
		}

		return permissions, statusAvailable, nil
	}
}
