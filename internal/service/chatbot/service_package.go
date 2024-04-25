package chatbot

import (
	"context"
	"math/rand"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chatbot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*chatbot.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return chatbot.NewFromConfig(cfg, func(o *chatbot.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		} else if config["partition"].(string) == names.StandardPartitionID {
			// Chatbot endpoint is available only in the 4 regions us-east-2, us-west-2, eu-west-1, and ap-southeast-1.
			// So pick a region randomly so that even if one endpoint is down, a retry might lead to trying a different endpoint.
			available_regions := []string{
				names.USEast2RegionID,
				names.USWest2RegionID,
				names.EUWest1RegionID,
				names.APSoutheast1RegionID,
			}
			o.Region = available_regions[rand.Intn(len(available_regions))]
		}
	}), nil
}
