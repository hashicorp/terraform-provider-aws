package ivs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindPlaybackKeyPairByID(ctx context.Context, conn *ivs.IVS, id string) (*ivs.PlaybackKeyPair, error) {
	in := &ivs.GetPlaybackKeyPairInput{
		Arn: aws.String(id),
	}
	out, err := conn.GetPlaybackKeyPairWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
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
