// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func statusPermissions(ctx context.Context, conn *lakeformation.Client, input *lakeformation.ListPermissionsInput, tableType string, columnNames []string, excludedColumnNames []string, columnWildcard bool) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var permissions []awstypes.PrincipalResourcePermissions

		pages := lakeformation.NewListPermissionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if errs.IsA[*awstypes.EntityNotFoundException](err) {
				return nil, statusNotFound, err
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Invalid principal") {
				return nil, statusIAMDelay, nil
			}

			if err != nil {
				return nil, statusFailed, fmt.Errorf("listing permissions: %w", err)
			}

			for _, permission := range page.PrincipalResourcePermissions {
				if reflect.ValueOf(permission).IsZero() {
					continue
				}

				if aws.ToString(input.Principal.DataLakePrincipalIdentifier) != aws.ToString(permission.Principal.DataLakePrincipalIdentifier) {
					continue
				}

				permissions = append(permissions, permission)
			}
		}

		// clean permissions = filter out permissions that do not pertain to this specific resource
		cleanPermissions := FilterPermissions(input, tableType, columnNames, excludedColumnNames, columnWildcard, permissions)

		if len(cleanPermissions) == 0 {
			return nil, statusNotFound, nil
		}

		return permissions, statusAvailable, nil
	}
}
