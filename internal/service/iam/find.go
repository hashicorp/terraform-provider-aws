// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindUsers(ctx context.Context, conn *iam.IAM, nameRegex, pathPrefix string) ([]*iam.User, error) {
	input := &iam.ListUsersInput{}

	if pathPrefix != "" {
		input.PathPrefix = aws.String(pathPrefix)
	}

	var results []*iam.User

	err := conn.ListUsersPagesWithContext(ctx, input, func(page *iam.ListUsersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, user := range page.Users {
			if user == nil {
				continue
			}

			if nameRegex != "" && !regexache.MustCompile(nameRegex).MatchString(aws.StringValue(user.UserName)) {
				continue
			}

			results = append(results, user)
		}

		return !lastPage
	})

	return results, err
}

func FindServiceSpecificCredential(ctx context.Context, conn *iam.IAM, serviceName, userName, credID string) (*iam.ServiceSpecificCredentialMetadata, error) {
	input := &iam.ListServiceSpecificCredentialsInput{
		ServiceName: aws.String(serviceName),
		UserName:    aws.String(userName),
	}

	output, err := conn.ListServiceSpecificCredentialsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output.ServiceSpecificCredentials) == 0 || output.ServiceSpecificCredentials[0] == nil {
		return nil, tfresource.NewEmptyResultError(output)
	}

	var cred *iam.ServiceSpecificCredentialMetadata

	for _, crd := range output.ServiceSpecificCredentials {
		if aws.StringValue(crd.ServiceName) == serviceName &&
			aws.StringValue(crd.UserName) == userName &&
			aws.StringValue(crd.ServiceSpecificCredentialId) == credID {
			cred = crd
			break
		}
	}

	if cred == nil {
		return nil, tfresource.NewEmptyResultError(cred)
	}

	return cred, nil
}

func FindSigningCertificate(ctx context.Context, conn *iam.IAM, userName, certId string) (*iam.SigningCertificate, error) {
	input := &iam.ListSigningCertificatesInput{
		UserName: aws.String(userName),
	}

	output, err := conn.ListSigningCertificatesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output.Certificates) == 0 || output.Certificates[0] == nil {
		return nil, tfresource.NewEmptyResultError(output)
	}

	var cert *iam.SigningCertificate

	for _, crt := range output.Certificates {
		if aws.StringValue(crt.UserName) == userName &&
			aws.StringValue(crt.CertificateId) == certId {
			cert = crt
			break
		}
	}

	if cert == nil {
		return nil, tfresource.NewEmptyResultError(cert)
	}

	return cert, nil
}

func FindAccessKey(ctx context.Context, conn *iam.IAM, username, id string) (*iam.AccessKeyMetadata, error) {
	accessKeys, err := FindAccessKeys(ctx, conn, username)
	if err != nil {
		return nil, err
	}

	for _, accessKey := range accessKeys {
		if aws.StringValue(accessKey.AccessKeyId) == id {
			return accessKey, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindAccessKeys(ctx context.Context, conn *iam.IAM, username string) ([]*iam.AccessKeyMetadata, error) {
	input := &iam.ListAccessKeysInput{
		UserName: aws.String(username),
	}
	var output []*iam.AccessKeyMetadata

	err := conn.ListAccessKeysPagesWithContext(ctx, input, func(page *iam.ListAccessKeysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.AccessKeyMetadata...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	return output, err
}
