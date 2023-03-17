package lakeformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func statusPermissions(ctx context.Context, conn *lakeformation.LakeFormation, input *lakeformation.ListPermissionsInput, tableType string, columnNames []*string, excludedColumnNames []*string, columnWildcard bool) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var permissions []*lakeformation.PrincipalResourcePermissions

		err := conn.ListPermissionsPagesWithContext(ctx, input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				if aws.StringValue(input.Principal.DataLakePrincipalIdentifier) != aws.StringValue(permission.Principal.DataLakePrincipalIdentifier) {
					continue
				}

				permissions = append(permissions, permission)
			}
			return !lastPage
		})

		if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
			return nil, statusNotFound, err
		}

		if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
			return nil, statusIAMDelay, nil
		}

		if err != nil {
			return nil, statusFailed, fmt.Errorf("error listing permissions: %w", err)
		}

		// clean permissions = filter out permissions that do not pertain to this specific resource
		cleanPermissions := FilterPermissions(input, tableType, columnNames, excludedColumnNames, columnWildcard, permissions)

		if len(cleanPermissions) == 0 {
			return nil, statusNotFound, nil
		}

		return permissions, statusAvailable, nil
	}
}
