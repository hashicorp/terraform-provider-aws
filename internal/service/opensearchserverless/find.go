// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findAccessPolicyByNameAndType(ctx context.Context, conn *opensearchserverless.Client, id, policyType string) (*types.AccessPolicyDetail, error) {
	in := &opensearchserverless.GetAccessPolicyInput{
		Name: aws.String(id),
		Type: types.AccessPolicyType(policyType),
	}
	out, err := conn.GetAccessPolicy(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.AccessPolicyDetail == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.AccessPolicyDetail, nil
}

func findCollectionByID(ctx context.Context, conn *opensearchserverless.Client, id string) (*types.CollectionDetail, error) {
	in := &opensearchserverless.BatchGetCollectionInput{
		Ids: []string{id},
	}
	out, err := conn.BatchGetCollection(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.CollectionDetails == nil || len(out.CollectionDetails) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.CollectionDetails[0], nil
}

func findCollectionByName(ctx context.Context, conn *opensearchserverless.Client, name string) (*types.CollectionDetail, error) {
	in := &opensearchserverless.BatchGetCollectionInput{
		Names: []string{name},
	}
	out, err := conn.BatchGetCollection(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.CollectionDetails == nil || len(out.CollectionDetails) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if len(out.CollectionDetails) > 1 {
		return nil, tfresource.NewTooManyResultsError(len(out.CollectionDetails), in)
	}

	return &out.CollectionDetails[0], nil
}

func findSecurityConfigByID(ctx context.Context, conn *opensearchserverless.Client, id string) (*types.SecurityConfigDetail, error) {
	in := &opensearchserverless.GetSecurityConfigInput{
		Id: aws.String(id),
	}
	out, err := conn.GetSecurityConfig(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.SecurityConfigDetail == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.SecurityConfigDetail, nil
}

func findSecurityPolicyByNameAndType(ctx context.Context, conn *opensearchserverless.Client, name, policyType string) (*types.SecurityPolicyDetail, error) {
	in := &opensearchserverless.GetSecurityPolicyInput{
		Name: aws.String(name),
		Type: types.SecurityPolicyType(policyType),
	}
	out, err := conn.GetSecurityPolicy(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.SecurityPolicyDetail == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.SecurityPolicyDetail, nil
}

func findVPCEndpointByID(ctx context.Context, conn *opensearchserverless.Client, id string) (*types.VpcEndpointDetail, error) {
	in := &opensearchserverless.BatchGetVpcEndpointInput{
		Ids: []string{id},
	}
	out, err := conn.BatchGetVpcEndpoint(ctx, in)

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.VpcEndpointDetails == nil || len(out.VpcEndpointDetails) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	vpcEndpointDetail := &out.VpcEndpointDetails[0]

	// Ensure default values if nil
	if vpcEndpointDetail.FailureCode == nil {
		vpcEndpointDetail.FailureCode = aws.String("")
	}
	if vpcEndpointDetail.FailureMessage == nil {
		vpcEndpointDetail.FailureMessage = aws.String("")
	}

	return vpcEndpointDetail, nil
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
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.LifecyclePolicyDetails == nil || len(out.LifecyclePolicyDetails) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if len(out.LifecyclePolicyDetails) > 1 {
		return nil, tfresource.NewTooManyResultsError(len(out.LifecyclePolicyDetails), in)
	}

	return &out.LifecyclePolicyDetails[0], nil
}
