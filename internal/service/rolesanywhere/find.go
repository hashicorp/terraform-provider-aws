// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rolesanywhere

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere"
	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindProfileByID(ctx context.Context, conn *rolesanywhere.Client, id string) (*types.ProfileDetail, error) {
	in := &rolesanywhere.GetProfileInput{
		ProfileId: aws.String(id),
	}

	out, err := conn.GetProfile(ctx, in)

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Profile == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Profile, nil
}

func FindTrustAnchorByID(ctx context.Context, conn *rolesanywhere.Client, id string) (*types.TrustAnchorDetail, error) {
	in := &rolesanywhere.GetTrustAnchorInput{
		TrustAnchorId: aws.String(id),
	}

	out, err := conn.GetTrustAnchor(ctx, in)

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.TrustAnchor == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.TrustAnchor, nil
}
