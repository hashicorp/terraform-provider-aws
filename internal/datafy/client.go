package datafy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-provider-aws/version"
)

type tags struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type createFromSnapshotsSource struct {
	DatafySnapshotId string `json:"datafySnapshotId"`
}

type createFromSnapshotsVolumeProperties struct {
	AvailabilityZone string `json:"availabilityZone"`
	DiskSize         int64  `json:"diskSize"`
	VolumeIops       int32  `json:"volumeIops"`
	VolumeThroughput int32  `json:"volumeThroughput"`
	Tags             []tags `json:"tags"`
}

type createFromSnapshotsRequest struct {
	Source           createFromSnapshotsSource           `json:"source"`
	VolumeProperties createFromSnapshotsVolumeProperties `json:"volumeProperties"`
}

type attachVolumeRequest struct {
	InstanceId string `json:"instanceId"`
	DeviceName string `json:"deviceName"`
}

type detachVolumeRequest struct {
	InstanceId string `json:"instanceId"`
	Force      bool   `json:"force"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type Client struct {
	config Config
}

func NewDatafyClient(config *Config) *Client {
	return &Client{
		config: *config,
	}
}

func toError(response *http.Response) error {
	var errResp errorResponse
	if err := json.NewDecoder(response.Body).Decode(&errResp); err == nil && errResp.Message != "" {
		return fmt.Errorf(errResp.Message)
	}
	return fmt.Errorf(response.Status)
}

func (c *Client) sendRequest(method, endpoint string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.config.Url, endpoint), bodyReader)
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

func (c *Client) GetSnapshot(snapshotId string) (*Snapshot, error) {
	resp, err := c.sendRequest(http.MethodGet, fmt.Sprintf("api/v1/aws/snapshots/%s", snapshotId), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var out Snapshot
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

func (c *Client) CreateVolumeFromSnapshot(datafySnapshotId string, availabilityZone string, iops int32, throughput int32, tagz map[string]string) (*RestoredVolume, error) {
	tagsList := make([]tags, 0, len(tagz))
	for k, v := range tagz {
		tagsList = append(tagsList, tags{Key: k, Value: v})
	}
	request := createFromSnapshotsRequest{
		Source: createFromSnapshotsSource{
			DatafySnapshotId: datafySnapshotId,
		},
		VolumeProperties: createFromSnapshotsVolumeProperties{
			VolumeIops:       iops,
			VolumeThroughput: throughput,
			AvailabilityZone: availabilityZone,
			Tags:             tagsList,
		},
	}

	resp, err := c.sendRequest(http.MethodPost, "api/v1/aws/volumes/create-from-snapshots", request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var out RestoredVolume
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, err
		}
		return &out, nil
	}

	return nil, toError(resp)
}

func (c *Client) AttachVolume(instanceId string, volumeId string, deviceName string) error {
	request := attachVolumeRequest{
		InstanceId: instanceId,
		DeviceName: deviceName,
	}

	resp, err := c.sendRequest(http.MethodPost, fmt.Sprintf("api/v1/aws/volumes/%s/attach", volumeId), request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf(resp.Status)
}

func (c *Client) DetachVolume(instanceId string, volumeId string) error {
	request := detachVolumeRequest{
		InstanceId: instanceId,
	}

	resp, err := c.sendRequest(http.MethodPost, fmt.Sprintf("api/v1/aws/volumes/%s/detach", volumeId), request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf(resp.Status)
}
