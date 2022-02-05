package cloudtrail

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusEventDataStore fetches the CloudTrail Event Data Store its status.
func statusEventDataStore(ctx context.Context, conn *cloudtrail.CloudTrail, eventDataStoreArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		eventDataStore, err := FindEventDataStoreByArn(ctx, conn, eventDataStoreArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return eventDataStore, aws.StringValue(eventDataStore.Status), nil
	}
}
