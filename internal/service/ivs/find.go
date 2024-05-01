// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindPlaybackKeyPairByID(ctx context.Context, conn *ivs.IVS, id string) (*ivs.PlaybackKeyPair, error) {
	in := &ivs.GetPlaybackKeyPairInput{
		Arn: aws.String(id),
	}
	out, err := conn.GetPlaybackKeyPairWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.KeyPair == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.KeyPair, nil
}

func FindRecordingConfigurationByID(ctx context.Context, conn *ivs.IVS, id string) (*ivs.RecordingConfiguration, error) {
	in := &ivs.GetRecordingConfigurationInput{
		Arn: aws.String(id),
	}
	out, err := conn.GetRecordingConfigurationWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.RecordingConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.RecordingConfiguration, nil
}

func FindChannelByID(ctx context.Context, conn *ivs.IVS, arn string) (*ivs.Channel, error) {
	in := &ivs.GetChannelInput{
		Arn: aws.String(arn),
	}
	out, err := conn.GetChannelWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Channel == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Channel, nil
}

func FindStreamKeyByChannelID(ctx context.Context, conn *ivs.IVS, channelArn string) (*ivs.StreamKey, error) {
	in := &ivs.ListStreamKeysInput{
		ChannelArn: aws.String(channelArn),
	}
	out, err := conn.ListStreamKeysWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(out.StreamKeys) < 1 {
		return nil, &retry.NotFoundError{
			LastRequest: in,
		}
	}

	streamKeyArn := out.StreamKeys[0].Arn

	return findStreamKeyByID(ctx, conn, *streamKeyArn)
}

func findStreamKeyByID(ctx context.Context, conn *ivs.IVS, id string) (*ivs.StreamKey, error) {
	in := &ivs.GetStreamKeyInput{
		Arn: aws.String(id),
	}
	out, err := conn.GetStreamKeyWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	return out.StreamKey, nil
}
