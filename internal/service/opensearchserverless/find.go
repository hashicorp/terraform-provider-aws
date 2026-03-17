// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findAccessPolicyByNameAndType(ctx context.Context, conn *opensearchserverless.Client, id, policyType string) (*types.AccessPolicyDetail, error) {
	in := &opensearchserverless.GetAccessPolicyInput{
		Name: aws.String(id),
		Type: types.AccessPolicyType(policyType),
	}
	out, err := conn.GetAccessPolicy(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.AccessPolicyDetail == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.AccessPolicyDetail, nil
}

func findCollectionByID(ctx context.Context, conn *opensearchserverless.Client, id string) (*types.CollectionDetail, error) {
	in := &opensearchserverless.BatchGetCollectionInput{
		Ids: []string{id},
	}
	out, err := conn.BatchGetCollection(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(out.CollectionDetails)
}

func findCollectionByName(ctx context.Context, conn *opensearchserverless.Client, name string) (*types.CollectionDetail, error) {
	in := &opensearchserverless.BatchGetCollectionInput{
		Names: []string{name},
	}
	out, err := conn.BatchGetCollection(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(out.CollectionDetails)
}

func findSecurityConfigByID(ctx context.Context, conn *opensearchserverless.Client, id string) (*types.SecurityConfigDetail, error) {
	in := &opensearchserverless.GetSecurityConfigInput{
		Id: aws.String(id),
	}
	out, err := conn.GetSecurityConfig(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.SecurityConfigDetail == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.SecurityConfigDetail, nil
}

func findSecurityPolicyByNameAndType(ctx context.Context, conn *opensearchserverless.Client, name, policyType string) (*types.SecurityPolicyDetail, error) {
	in := &opensearchserverless.GetSecurityPolicyInput{
		Name: aws.String(name),
		Type: types.SecurityPolicyType(policyType),
	}
	out, err := conn.GetSecurityPolicy(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.SecurityPolicyDetail == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.SecurityPolicyDetail, nil
}

func findLifecyclePolicyByNameAndType(ctx context.Context, conn *opensearchserverless.Client, name, policyType string) (*types.LifecyclePolicyDetail, error) {
	in := &opensearchserverless.BatchGetLifecyclePolicyInput{
		Identifiers: []types.LifecyclePolicyIdentifier{
			{
				Name: aws.String(name),
				Type: types.LifecyclePolicyType(policyType),
			},
		},
	}

	out, err := conn.BatchGetLifecyclePolicy(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(out.LifecyclePolicyDetails)
}
