package conns

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	apigatewayv2_sdkv1 "github.com/aws/aws-sdk-go/service/apigatewayv2"
	mediaconvert_sdkv1 "github.com/aws/aws-sdk-go/service/mediaconvert"
	s3_sdkv1 "github.com/aws/aws-sdk-go/service/s3"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

type AWSClient struct {
	AccountID               string
	DefaultTagsConfig       *tftags.DefaultConfig
	DNSSuffix               string
	IgnoreTagsConfig        *tftags.IgnoreConfig
	MediaConvertAccountConn *mediaconvert_sdkv1.MediaConvert
	Partition               string
	Region                  string
	ReverseDNSPrefix        string
	ServicePackages         map[string]ServicePackage
	Session                 *session_sdkv1.Session
	TerraformVersion        string

	clients    map[string]any
	conns      map[string]any
	httpClient *http.Client
	lock       sync.Mutex
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

func (client *AWSClient) S3ConnURICleaningDisabled(ctx context.Context) *s3_sdkv1.S3 {
	config := client.S3Conn(ctx).Config
	config.DisableRestProtocolURICleaning = aws_sdkv1.Bool(true)

	return s3_sdkv1.New(client.Session.Copy(&config))
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

// APIGatewayInvokeURL returns the Amazon API Gateway (REST APIs) invoke URL for the configured AWS Region.
// See https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-call-api.html.
func (client *AWSClient) APIGatewayInvokeURL(restAPIID, stageName string) string {
	return fmt.Sprintf("https://%s/%s", client.RegionalHostname(fmt.Sprintf("%s.execute-api", restAPIID)), stageName)
}

// APIGatewayV2InvokeURL returns the Amazon API Gateway v2 (WebSocket & HTTP APIs) invoke URL for the configured AWS Region.
// See https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-publish.html and
// https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-set-up-websocket-deployment.html.
func (client *AWSClient) APIGatewayV2InvokeURL(protocolType, apiID, stageName string) string {
	if protocolType == apigatewayv2_sdkv1.ProtocolTypeWebsocket {
		return fmt.Sprintf("wss://%s/%s", client.RegionalHostname(fmt.Sprintf("%s.execute-api", apiID)), stageName)
	}

	if stageName == "$default" {
		return fmt.Sprintf("https://%s/", client.RegionalHostname(fmt.Sprintf("%s.execute-api", apiID)))
	}

	return fmt.Sprintf("https://%s/%s", client.RegionalHostname(fmt.Sprintf("%s.execute-api", apiID)), stageName)
}

// CloudFrontDistributionHostedZoneID returns the Route 53 hosted zone ID
// for Amazon CloudFront distributions in the configured AWS partition.
func (client *AWSClient) CloudFrontDistributionHostedZoneID() string {
	if client.Partition == endpoints_sdkv1.AwsCnPartitionID {
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

// GlobalAcceleratorHostedZoneID returns the Route 53 hosted zone ID
// for AWS Global Accelerator accelerators in the configured AWS partition.
func (client *AWSClient) GlobalAcceleratorHostedZoneID() string {
	return "Z2BJ6XQ5FK7U4H" // See https://docs.aws.amazon.com/general/latest/gr/global_accelerator.html#global_accelerator_region
}

// conn returns the AWS SDK for Go v1 API client for the specified service.
func conn[T any](ctx context.Context, c *AWSClient, servicePackageName string) (T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if raw, ok := c.conns[servicePackageName]; ok {
		if conn, ok := raw.(T); ok {
			return conn, nil
		} else {
			var zero T
			return zero, fmt.Errorf("AWS SDK v1 API client (%s): %T, want %T", servicePackageName, raw, zero)
		}
	}

	sp, ok := c.ServicePackages[servicePackageName]
	if !ok {
		var zero T
		return zero, fmt.Errorf("unknown service package: %s", servicePackageName)
	}

	v, ok := sp.(interface {
		NewConn(context.Context) (T, error)
	})
	if !ok {
		var zero T
		return zero, fmt.Errorf("no AWS SDK v1 API client factory: %s", servicePackageName)
	}

	conn, err := v.NewConn(ctx)
	if err != nil {
		var zero T
		return zero, err
	}

	if v, ok := sp.(interface {
		CustomizeConn(context.Context, T) (T, error)
	}); ok {
		conn, err = v.CustomizeConn(ctx, conn)
		if err != nil {
			var zero T
			return zero, err
		}
	}

	c.conns[servicePackageName] = conn

	return conn, nil
}

// client returns the AWS SDK for Go v2 API client for the specified service.
func client[T any](ctx context.Context, c *AWSClient, servicePackageName string) (T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if raw, ok := c.clients[servicePackageName]; ok {
		if client, ok := raw.(T); ok {
			return client, nil
		} else {
			var zero T
			return zero, fmt.Errorf("AWS SDK v2 API client (%s): %T, want %T", servicePackageName, raw, zero)
		}
	}

	sp, ok := c.ServicePackages[servicePackageName]
	if !ok {
		var zero T
		return zero, fmt.Errorf("unknown service package: %s", servicePackageName)
	}

	v, ok := sp.(interface {
		NewClient(context.Context) (T, error)
	})
	if !ok {
		var zero T
		return zero, fmt.Errorf("no AWS SDK v2 API client factory: %s", servicePackageName)
	}

	client, err := v.NewClient(ctx)
	if err != nil {
		var zero T
		return zero, err
	}

	if v, ok := sp.(interface {
		CustomizeClient(context.Context, T) (T, error)
	}); ok {
		client, err = v.CustomizeClient(ctx, client)
		if err != nil {
			var zero T
			return zero, err
		}
	}

	c.clients[servicePackageName] = client

	return client, nil
}
