package cloudtrail

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindEventDataStoreByARN(ctx context.Context, conn *cloudtrail.CloudTrail, eventDataStoreArn string) (*cloudtrail.GetEventDataStoreOutput, error) {
	input := cloudtrail.GetEventDataStoreInput{
		EventDataStore: aws.String(eventDataStoreArn),
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

	if status := aws.StringValue(output.Status); status == cloudtrail.EventDataStoreStatusPendingDeletion {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}
