package conns

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/endpoints"
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

// CloudFrontDistributionHostedZoneIDForPartition returns for the Route 53 hosted zone ID
// for Amazon CloudFront distributions in the configured AWS partition.
func (client *AWSClient) CloudFrontDistributionHostedZoneID() string {
	if client.Partition == endpoints.AwsCnPartitionID {
		return "Z3RFFRIM2A3IF5" // See https://docs.amazonaws.cn/en_us/aws/latest/userguide/route53.html
	}
	return "Z2FDTNDATAQYW2" // See https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html#Route53-Type-AliasTarget-HostedZoneId
}

// DefaultKMSKeyPolicy returns the default policy for KMS keys in the configured AWS partition.
func (client *AWSClient) DefaultKMSKeyPolicy() string {
	return fmt.Sprintf(`
{
	"Id": "default",
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "Enable IAM User Permissions",
			"Effect": "Allow",
			"Principal": {
				"AWS": "arn:%[1]s:iam::%[2]s:root"
			},
			"Action": "kms:*",
			"Resource": "*"
		}
	]
}	
`, client.Partition, client.AccountID)
}
