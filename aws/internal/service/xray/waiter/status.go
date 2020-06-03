package waiter

import (
	"github.com/aws/aws-sdk-go/service/xray"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// EncryptionConfigStatus fetches the Encryption Config and its Status
func EncryptionConfigStatus(conn *xray.XRay) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		output, _ := conn.GetEncryptionConfig(&xray.GetEncryptionConfigInput{})

		return output, aws.StringValue(output.EncryptionConfig.Status), nil
	}
}
