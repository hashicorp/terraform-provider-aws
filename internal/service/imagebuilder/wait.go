package imagebuilder

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// waitImageStatusAvailable waits for an Image to return Available
func waitImageStatusAvailable(ctx context.Context, conn *imagebuilder.Imagebuilder, imageBuildVersionArn string, timeout time.Duration) (*imagebuilder.Image, error) {
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
		Refresh: statusImage(ctx, conn, imageBuildVersionArn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*imagebuilder.Image); ok {
		return v, err
	}

	return nil, err
}
