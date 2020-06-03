package waiter

import (
	"github.com/aws/aws-sdk-go/service/xray"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// EncryptionConfigAvailable waits for a EncryptionConfig to return Available
func EncryptionConfigAvailable(conn *xray.XRay) (*xray.EncryptionConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{xray.EncryptionStatusUpdating},
		Target:  []string{xray.EncryptionStatusActive},
		Refresh: EncryptionConfigStatus(conn),
		Timeout: 15 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*xray.EncryptionConfig); ok {
		return v, err
	}

	return nil, err
}
