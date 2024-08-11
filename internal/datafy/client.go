package datafy

import (
	"github.com/datafy-io/iac-gateway-lambda/iacgateway"
)

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
