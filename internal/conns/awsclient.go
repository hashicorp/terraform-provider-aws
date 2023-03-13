package conns

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/service/s3"
)

// InitContext creates context.
func (client *AWSClient) InitContext(ctx context.Context) context.Context {
	return ctx
}

// PartitionHostname returns a hostname with the provider domain suffix for the partition
// e.g. PREFIX.amazonaws.com
// The prefix should not contain a trailing period.
func (client *AWSClient) PartitionHostname(prefix string) string {
	return fmt.Sprintf("%s.%s", prefix, client.DNSSuffix)
}

// RegionalHostname returns a hostname with the provider domain suffix for the region and partition
// e.g. PREFIX.us-west-2.amazonaws.com
// The prefix should not contain a trailing period.
func (client *AWSClient) RegionalHostname(prefix string) string {
	return fmt.Sprintf("%s.%s.%s", prefix, client.Region, client.DNSSuffix)
}

func (client *AWSClient) S3ConnURICleaningDisabled() *s3.S3 {
	return client.s3ConnURICleaningDisabled
}

// SetHTTPClient sets the http.Client used for AWS API calls.
// To have effect it must be called before the AWS SDK v1 Session is created.
func (client *AWSClient) SetHTTPClient(httpClient *http.Client) {
	if client.Session == nil {
		client.httpClient = httpClient
	}
}

// HTTPClient returns the http.Client used for AWS API calls.
func (client *AWSClient) HTTPClient() *http.Client {
	return client.httpClient
}
