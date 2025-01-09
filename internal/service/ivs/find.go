// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ivs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindPlaybackKeyPairByID(ctx context.Context, conn *ivs.Client, id string) (*awstypes.PlaybackKeyPair, error) {
	in := &ivs.GetPlaybackKeyPairInput{
		Arn: aws.String(id),
	}
	out, err := conn.GetPlaybackKeyPair(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func FindRecordingConfigurationByID(ctx context.Context, conn *ivs.Client, id string) (*awstypes.RecordingConfiguration, error) {
	in := &ivs.GetRecordingConfigurationInput{
		Arn: aws.String(id),
	}
	out, err := conn.GetRecordingConfiguration(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func FindChannelByID(ctx context.Context, conn *ivs.Client, arn string) (*awstypes.Channel, error) {
	in := &ivs.GetChannelInput{
		Arn: aws.String(arn),
	}
	out, err := conn.GetChannel(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func FindStreamKeyByChannelID(ctx context.Context, conn *ivs.Client, channelArn string) (*awstypes.StreamKey, error) {
	in := &ivs.ListStreamKeysInput{
		ChannelArn: aws.String(channelArn),
	}
	out, err := conn.ListStreamKeys(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findStreamKeyByID(ctx context.Context, conn *ivs.Client, id string) (*awstypes.StreamKey, error) {
	in := &ivs.GetStreamKeyInput{
		Arn: aws.String(id),
	}
	out, err := conn.GetStreamKey(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
