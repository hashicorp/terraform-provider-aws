package xray

import (
	"time"

	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	encryptionConfigAvailableTimeout = 15 * time.Minute
)

// waitEncryptionConfigAvailable waits for a EncryptionConfig to return Available
func waitEncryptionConfigAvailable(conn *xray.XRay) (*xray.EncryptionConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{xray.EncryptionStatusUpdating},
		Target:  []string{xray.EncryptionStatusActive},
		Refresh: statusEncryptionConfig(conn),
		Timeout: encryptionConfigAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*xray.EncryptionConfig); ok {
		return v, err
	}

	return nil, err
}
