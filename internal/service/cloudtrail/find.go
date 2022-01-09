package cloudtrail

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindEventDataStoreByName(ctx context.Context, conn *cloudtrail.CloudTrail, name string) (*cloudtrail.GetEventDataStoreOutput, error) {
	input := cloudtrail.GetEventDataStoreInput{
		EventDataStore: aws.String(name),
	}

	output, err := conn.GetEventDataStoreWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, cloudtrail.ErrCodeEventDataStoreNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
