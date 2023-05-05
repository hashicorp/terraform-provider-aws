package xray

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	encryptionConfigStatusUnknown = "Unknown"
)

// statusEncryptionConfig fetches the Encryption Config and its Status
func statusEncryptionConfig(ctx context.Context, conn *xray.XRay) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _ := conn.GetEncryptionConfigWithContext(ctx, &xray.GetEncryptionConfigInput{})

		if output == nil || output.EncryptionConfig == nil {
			return output, encryptionConfigStatusUnknown, nil
		}

		return output, aws.StringValue(output.EncryptionConfig.Status), nil
	}
}
