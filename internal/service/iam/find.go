// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"reflect"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindUsers(ctx context.Context, conn *iam.Client, nameRegex, pathPrefix string) ([]awstypes.User, error) {
	input := &iam.ListUsersInput{}

	if pathPrefix != "" {
		input.PathPrefix = aws.String(pathPrefix)
	}

	var results []awstypes.User

	pages := iam.NewListUsersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, user := range page.Users {
			if nameRegex != "" && !regexache.MustCompile(nameRegex).MatchString(aws.ToString(user.UserName)) {
				continue
			}

			results = append(results, user)
		}
	}

	return results, nil
}

func FindServiceSpecificCredential(ctx context.Context, conn *iam.Client, serviceName, userName, credID string) (*awstypes.ServiceSpecificCredentialMetadata, error) {
	input := &iam.ListServiceSpecificCredentialsInput{
		ServiceName: aws.String(serviceName),
		UserName:    aws.String(userName),
	}

	output, err := conn.ListServiceSpecificCredentials(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output.ServiceSpecificCredentials) == 0 {
		return nil, tfresource.NewEmptyResultError(output)
	}

	var cred awstypes.ServiceSpecificCredentialMetadata

	for _, crd := range output.ServiceSpecificCredentials {
		if aws.ToString(crd.ServiceName) == serviceName &&
			aws.ToString(crd.UserName) == userName &&
			aws.ToString(crd.ServiceSpecificCredentialId) == credID {
			cred = crd
			break
		}
	}

	if reflect.ValueOf(cred).IsZero() {
		return nil, tfresource.NewEmptyResultError(cred)
	}

	return &cred, nil
}

func FindSigningCertificate(ctx context.Context, conn *iam.Client, userName, certId string) (*awstypes.SigningCertificate, error) {
	input := &iam.ListSigningCertificatesInput{
		UserName: aws.String(userName),
	}

	output, err := conn.ListSigningCertificates(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output.Certificates) == 0 {
		return nil, tfresource.NewEmptyResultError(output)
	}

	var cert awstypes.SigningCertificate

	for _, crt := range output.Certificates {
		if aws.ToString(crt.UserName) == userName &&
			aws.ToString(crt.CertificateId) == certId {
			cert = crt
			break
		}
	}

	if reflect.ValueOf(cert).IsZero() {
		return nil, tfresource.NewEmptyResultError(cert)
	}

	return &cert, nil
}
