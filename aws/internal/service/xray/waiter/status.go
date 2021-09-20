package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	encryptionConfigStatusUnknown = "Unknown"
)

// statusEncryptionConfig fetches the Encryption Config and its Status
func statusEncryptionConfig(conn *xray.XRay) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		output, _ := conn.GetEncryptionConfig(&xray.GetEncryptionConfigInput{})

		if output == nil || output.EncryptionConfig == nil {
			return output, encryptionConfigStatusUnknown, nil
		}

		return output, aws.StringValue(output.EncryptionConfig.Status), nil
	}
}
