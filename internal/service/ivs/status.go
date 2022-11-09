package ivs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	statusNormal = "Normal"

	statusActive   = ivs.RecordingConfigurationStateActive
	statusCreating = ivs.RecordingConfigurationStateCreating
)

func statusPlaybackKeyPair(ctx context.Context, conn *ivs.IVS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindPlaybackKeyPairByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func statusRecordingConfiguration(ctx context.Context, conn *ivs.IVS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindRecordingConfigurationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.State), nil
	}
}
