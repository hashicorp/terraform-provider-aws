package s3outposts

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	endpointStatusNotFound = "NotFound"
	endpointStatusUnknown  = "Unknown"
)

// statusEndpoint fetches the Endpoint and its Status
func statusEndpoint(conn *s3outposts.S3Outposts, endpointArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		endpoint, err := FindEndpoint(conn, endpointArn)

		if err != nil {
			return nil, endpointStatusUnknown, err
		}

		if endpoint == nil {
			return nil, endpointStatusNotFound, nil
		}

		return endpoint, aws.StringValue(endpoint.Status), nil
	}
}
