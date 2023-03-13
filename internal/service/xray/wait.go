package xray

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	encryptionConfigAvailableTimeout = 15 * time.Minute
)

// waitEncryptionConfigAvailable waits for a EncryptionConfig to return Available
func waitEncryptionConfigAvailable(ctx context.Context, conn *xray.XRay) (*xray.EncryptionConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{xray.EncryptionStatusUpdating},
		Target:  []string{xray.EncryptionStatusActive},
		Refresh: statusEncryptionConfig(ctx, conn),
		Timeout: encryptionConfigAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*xray.EncryptionConfig); ok {
		return v, err
	}

	return nil, err
}
