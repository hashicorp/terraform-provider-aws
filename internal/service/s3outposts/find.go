package s3outposts

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3outposts"
)

// FindEndpoint returns matching FindEndpoint by ARN.
func FindEndpoint(ctx context.Context, conn *s3outposts.S3Outposts, endpointArn string) (*s3outposts.Endpoint, error) {
	input := &s3outposts.ListEndpointsInput{}
	var result *s3outposts.Endpoint

	err := conn.ListEndpointsPagesWithContext(ctx, input, func(page *s3outposts.ListEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, endpoint := range page.Endpoints {
			if aws.StringValue(endpoint.EndpointArn) == endpointArn {
				result = endpoint
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
