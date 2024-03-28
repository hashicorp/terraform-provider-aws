// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAPICacheByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiCache, error) {
	input := &appsync.GetApiCacheInput{
		ApiId: aws.String(id),
	}
	out, err := conn.GetApiCache(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func FindDomainNameByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.DomainNameConfig, error) {
	input := &appsync.GetDomainNameInput{
		DomainName: aws.String(id),
	}
	out, err := conn.GetDomainName(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func FindDomainNameAPIAssociationByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiAssociation, error) {
	input := &appsync.GetApiAssociationInput{
		DomainName: aws.String(id),
	}
	out, err := conn.GetApiAssociation(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func FindTypeByThreePartKey(ctx context.Context, conn *appsync.Client, apiID, format, name string) (*awstypes.Type, error) {
	input := &appsync.GetTypeInput{
		ApiId:    aws.String(apiID),
		Format:   awstypes.TypeDefinitionFormat(format),
		TypeName: aws.String(name),
	}

	output, err := conn.GetType(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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
