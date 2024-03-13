// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindResourceSharePrincipalAssociationByShareARNPrincipal(ctx context.Context, conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	input := &ram.GetResourceShareAssociationsInput{
		AssociationType:   aws.String(ram.ResourceShareAssociationTypePrincipal),
		Principal:         aws.String(principal),
		ResourceShareArns: aws.StringSlice([]string{resourceShareARN}),
	}

	return findResourceShareAssociation(ctx, conn, input)
}

func findResourceShareAssociation(ctx context.Context, conn *ram.RAM, input *ram.GetResourceShareAssociationsInput) (*ram.ResourceShareAssociation, error) {
	output, err := findResourceShareAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findResourceShareAssociations(ctx context.Context, conn *ram.RAM, input *ram.GetResourceShareAssociationsInput) ([]*ram.ResourceShareAssociation, error) {
	var output []*ram.ResourceShareAssociation

	err := conn.GetResourceShareAssociationsPagesWithContext(ctx, input, func(page *ram.GetResourceShareAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResourceShareAssociations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
