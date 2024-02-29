package costoptimizationhub

import (
	"context"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	costoptimizationhub_sdkv2 "github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*costoptimizationhub_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return costoptimizationhub_sdkv2.NewFromConfig(cfg, func(o *costoptimizationhub_sdkv2.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)
		} else if config["partition"].(string) == endpoints_sdkv1.AwsPartitionID {
			// Cost Optimization Hub endpoint is available only in us-east-1 Region.
			o.Region = endpoints_sdkv1.UsEast1RegionID
		}
	}), nil
}
