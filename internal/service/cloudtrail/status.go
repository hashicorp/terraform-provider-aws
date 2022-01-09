package cloudtrail

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	eventDataStoreStatusCreating  = "created"
	eventDataStoreStatusAvailable = "enabled"
	eventDataStoreStatusDeleting  = "pending"
)

// statusEventDataStore fetches the Event Data Store and its status.
func statusEventDataStore(ctx context.Context, conn *cloudtrail.CloudTrail, eventDataStoreName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		eventDataStore, err := FindEventDataStoreByName(ctx, conn, eventDataStoreName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return eventDataStore, aws.StringValue(eventDataStore.Status), nil
	}
}
