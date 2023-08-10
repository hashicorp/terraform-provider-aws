// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAPICacheByID(ctx context.Context, conn *appsync.AppSync, id string) (*appsync.ApiCache, error) {
	input := &appsync.GetApiCacheInput{
		ApiId: aws.String(id),
	}
	out, err := conn.GetApiCacheWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out.ApiCache, nil
}

func FindDomainNameByID(ctx context.Context, conn *appsync.AppSync, id string) (*appsync.DomainNameConfig, error) {
	input := &appsync.GetDomainNameInput{
		DomainName: aws.String(id),
	}
	out, err := conn.GetDomainNameWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out.DomainNameConfig, nil
}

func FindDomainNameAPIAssociationByID(ctx context.Context, conn *appsync.AppSync, id string) (*appsync.ApiAssociation, error) {
	input := &appsync.GetApiAssociationInput{
		DomainName: aws.String(id),
	}
	out, err := conn.GetApiAssociationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out.ApiAssociation, nil
}

func FindTypeByThreePartKey(ctx context.Context, conn *appsync.AppSync, apiID, format, name string) (*appsync.Type, error) {
	input := &appsync.GetTypeInput{
		ApiId:    aws.String(apiID),
		Format:   aws.String(format),
		TypeName: aws.String(name),
	}

	output, err := conn.GetTypeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Type == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Type, nil
}
