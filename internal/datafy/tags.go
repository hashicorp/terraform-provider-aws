package datafy

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	managedByTagKey    = "Managed-By"
	managedByTagValue  = "Datafy.io"
	sourceVolumeTagKey = "datafy:source-volume:id"
	tagsPrefix         = "datafy:"
)

func DescribeDatafiedVolumesInput(sourceVolumeId string) *ec2.DescribeVolumesInput {
	return &ec2.DescribeVolumesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", managedByTagKey)),
				Values: []string{managedByTagValue},
			},
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", sourceVolumeTagKey)),
				Values: []string{sourceVolumeId},
			},
		},
	}
}

func RemoveDatafyTags(tags []types.Tag) []types.Tag {
	return slices.DeleteFunc(tags, func(t types.Tag) bool {
		key := aws.ToString(t.Key)
		return strings.HasPrefix(key, tagsPrefix) || (key == managedByTagKey && aws.ToString(t.Value) == managedByTagValue)
	})
}
