package datafy

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	ManagedByTagKey    = "Managed-By"
	ManagedByTagValue  = "Datafy.io"
	SourceVolumeTagKey = "datafy:source-volume:id"
)

type Volume struct {
	*types.Volume

	IsManaged  bool
	IsDatafied bool
	ReplacedBy string
}

func (v *Volume) UnmarshalJSON(data []byte) error {
	iac := struct {
		VolumeId string `json:"volumeId"`

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
	v.IsManaged = iac.IsManaged
	v.IsDatafied = iac.IsDatafied
	v.ReplacedBy = iac.ReplacedBy

	return nil
}
