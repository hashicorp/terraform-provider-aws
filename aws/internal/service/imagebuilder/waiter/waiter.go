package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// waitImageStatusAvailable waits for an Image to return Available
func waitImageStatusAvailable(conn *imagebuilder.Imagebuilder, imageBuildVersionArn string, timeout time.Duration) (*imagebuilder.Image, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			imagebuilder.ImageStatusBuilding,
			imagebuilder.ImageStatusCreating,
			imagebuilder.ImageStatusDistributing,
			imagebuilder.ImageStatusIntegrating,
			imagebuilder.ImageStatusPending,
			imagebuilder.ImageStatusTesting,
		},
		Target:  []string{imagebuilder.ImageStatusAvailable},
		Refresh: statusImage(conn, imageBuildVersionArn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*imagebuilder.Image); ok {
		return v, err
	}

	return nil, err
}
