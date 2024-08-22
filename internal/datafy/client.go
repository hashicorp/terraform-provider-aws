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

func newVolume(volume iacgateway.AWSVolume) *Volume {
	return &Volume{
		Volume: &awstypes.Volume{
			VolumeId: aws.String(volume.VolumeId),
		},
		IsManaged:     volume.IsManaged,
		IsDatafied:    volume.IsDatafied,
		IsReplacement: volume.IsReplacement,
	}
}

type Client struct {
	config Config

	client *iacgateway.AWSClient
}

func NewDatafyClient(config *Config) *Client {
	return &Client{
		config: *config,
		client: iacgateway.NewAwsClient(config.Url, config.Token),
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

	return newVolume(*volume), nil
}
