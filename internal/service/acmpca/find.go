// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindPolicyByARN(ctx context.Context, conn *acmpca.Client, arn string) (string, error) {
	input := &acmpca.GetPolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return "", &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.Policy == nil {
		return "", tfresource.NewEmptyResultError(input)
	}

	return aws.ToString(output.Policy), nil
}

func FindPermission(ctx context.Context, conn *acmpca.Client, certificateAuthorityARN, principal, sourceAccount string) (*awstypes.Permission, error) {
	input := &acmpca.ListPermissionsInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}

	var results []awstypes.Permission
	paginator := acmpca.NewListPermissionsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, permission := range page.Permissions {
			if aws.ToString(permission.Principal) == principal && (sourceAccount == "" || aws.ToString(permission.SourceAccount) == sourceAccount) {
				results = append(results, permission)
			}
		}
	}

	permission, err := tfresource.AssertSingleValueResult(results)
	if err != nil {
		return nil, err
	}
	return permission, nil
}
