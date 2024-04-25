// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAliasByName(ctx context.Context, conn *kms.Client, name string) (*awstypes.AliasListEntry, error) {
	input := &kms.ListAliasesInput{}
	pages := kms.NewListAliasesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, alias := range page.Aliases {
			if aws.ToString(alias.AliasName) == name {
				return &alias, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindCustomKeyStoreByID(ctx context.Context, conn *kms.Client, in *kms.DescribeCustomKeyStoresInput) (*awstypes.CustomKeyStoresListEntry, error) {
	out, err := conn.DescribeCustomKeyStores(ctx, in)

	if errs.IsA[*awstypes.CustomKeyStoreNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || len(out.CustomKeyStores) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.CustomKeyStores[0], nil
}

func FindKeyByID(ctx context.Context, conn *kms.Client, id string) (*awstypes.KeyMetadata, error) {
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(id),
	}

	output, err := conn.DescribeKey(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.KeyMetadata == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	keyMetadata := output.KeyMetadata

	// Once the CMK is in the pending (replica) deletion state Terraform considers it logically deleted.
	if state := keyMetadata.KeyState; state == awstypes.KeyStatePendingDeletion || state == awstypes.KeyStatePendingReplicaDeletion {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return keyMetadata, nil
}

func findDefaultKey(ctx context.Context, client *conns.AWSClient, service, region string) (string, error) {
	conn := client.KMSConnForRegion(ctx, region)

	k, err := FindKeyByID(ctx, conn, fmt.Sprintf("alias/aws/%s", service)) //default key
	if err != nil {
		return "", fmt.Errorf("finding default key: %s", err)
	}

	return aws.ToString(k.Arn), nil
}

func FindKeyPolicyByKeyIDAndPolicyName(ctx context.Context, conn *kms.Client, keyID, policyName string) (*string, error) {
	input := &kms.GetKeyPolicyInput{
		KeyId:      aws.String(keyID),
		PolicyName: aws.String(policyName),
	}

	output, err := conn.GetKeyPolicy(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policy, nil
}

func FindKeyRotationEnabledByKeyID(ctx context.Context, conn *kms.Client, keyID string) (*bool, error) {
	input := &kms.GetKeyRotationStatusInput{
		KeyId: aws.String(keyID),
	}

	output, err := conn.GetKeyRotationStatus(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return &output.KeyRotationEnabled, nil
}
