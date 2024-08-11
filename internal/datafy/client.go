package datafy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-provider-aws/version"
)

type Client struct {
	config Config
}

func NewDatafyClient(config *Config) *Client {
	return &Client{
		config: *config,
	}
}

func (c *Client) sendRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.config.Url, endpoint), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", fmt.Sprintf("terraform-provider-datafyaws/%s (datafy.io)", version.ProviderVersion))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.Token))
	client := &http.Client{}
	return client.Do(req)
}

func (c *Client) GetVolume(volumeId string) (*Volume, error) {
	resp, err := c.sendRequest(http.MethodGet, fmt.Sprintf("api/v1/aws/volumes/%s", volumeId), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var out Volume
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, err
		}
		return &out, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFoundError
	}

	return nil, fmt.Errorf(resp.Status)
}
