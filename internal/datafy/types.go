package datafy

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type RestoredVolume struct {
	VolumeId     string `json:"volumeId"`
	VolumeSizeGB int32  `json:"volumeSizeGB"`
}

type Volume struct {
	*types.Volume

	HasSource  bool
	IsManaged  bool
	IsDatafied bool
	ReplacedBy string
}

func (v *Volume) UnmarshalJSON(data []byte) error {
	iac := struct {
		VolumeId string `json:"volumeId"`

		HasSource  bool   `json:"hasSource"`
		IsManaged  bool   `json:"isManaged"`
		IsDatafied bool   `json:"isDatafied"`
		ReplacedBy string `json:"replacedBy"`
	}{}
	if err := json.Unmarshal(data, &iac); err != nil {
		return err
	}

	v.Volume = &types.Volume{
		VolumeId: aws.String(iac.VolumeId),
	}
	v.HasSource = iac.HasSource
	v.IsManaged = iac.IsManaged
	v.IsDatafied = iac.IsDatafied
	v.ReplacedBy = iac.ReplacedBy

	return nil
}

type Snapshot struct {
	*types.Snapshot

	Region               string
	ReadyToUse           bool
	SizeBytes            int64
	SnapshotCreationTime *time.Time
	DatafySnapshotIds    []string
}

func (v *Snapshot) UnmarshalJSON(data []byte) error {
	iac := struct {
		SnapshotId string `json:"snapshotId"`
		VolumeId   string `json:"volumeId"`
		Region     string `json:"region"`

		ReadyToUse           bool       `json:"readyToUse"`
		SizeBytes            int64      `json:"sizeBytes"`
		SnapshotCreationTime *time.Time `json:"snapshotCreationTime"`
		DatafySnapshotIds    []string   `json:"datafySnapshotIds,omitempty"`
	}{}
	if err := json.Unmarshal(data, &iac); err != nil {
		return err
	}

	v.Snapshot = &types.Snapshot{
		SnapshotId: aws.String(iac.SnapshotId),
		VolumeId:   aws.String(iac.VolumeId),
	}
	v.Region = iac.Region
	v.ReadyToUse = iac.ReadyToUse
	v.SizeBytes = iac.SizeBytes
	v.SnapshotCreationTime = iac.SnapshotCreationTime
	v.DatafySnapshotIds = iac.DatafySnapshotIds

	return nil
}
