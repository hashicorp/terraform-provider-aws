package datafy

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/datafy-io/iac-gateway-lambda/iacgateway"
	"strings"
)

func apiNotFound(err error) bool {
	return strings.Contains(err.Error(), "404")
}

type Client struct {
	config Config

	client *iacgateway.AWSClient
}

func NewDatafyClient(config *Config) *Client {
	return &Client{
		config: *config,
		client: iacgateway.NewAwsClient(url, config.Token),
	}
}

func (c *Client) GetVolume(volumeId string) (*Volume, error) {
	volume, err := c.client.GetVolume(volumeId)
	if err != nil {
		if apiNotFound(err) {
			return nil, NotFoundError
		}
		return nil, err
	}

	return &Volume{
		Volume: &awstypes.Volume{
			VolumeId: aws.String(volume.VolumeId),
		},
		IsManaged:     volume.IsManaged,
		IsReplacement: volume.IsReplacement,
	}, nil
}
