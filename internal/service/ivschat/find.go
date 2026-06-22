// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ivschat

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivschat"
	"github.com/aws/aws-sdk-go-v2/service/ivschat/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findLoggingConfigurationByID(ctx context.Context, conn *ivschat.Client, id string) (*ivschat.GetLoggingConfigurationOutput, error) {
	in := &ivschat.GetLoggingConfigurationInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetLoggingConfiguration(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
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

func findRoomByID(ctx context.Context, conn *ivschat.Client, id string) (*ivschat.GetRoomOutput, error) {
	in := &ivschat.GetRoomInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetRoom(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
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
