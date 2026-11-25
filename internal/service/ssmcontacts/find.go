// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findContactByID(ctx context.Context, conn *ssmcontacts.Client, id string) (*ssmcontacts.GetContactOutput, error) {
	in := &ssmcontacts.GetContactInput{
		ContactId: aws.String(id),
	}
	out, err := conn.GetContact(ctx, in)
	if err != nil {
		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func findContactChannelByID(ctx context.Context, conn *ssmcontacts.Client, id string) (*ssmcontacts.GetContactChannelOutput, error) {
	in := &ssmcontacts.GetContactChannelInput{
		ContactChannelId: aws.String(id),
	}
	out, err := conn.GetContactChannel(ctx, in)
	if err != nil {
		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}
