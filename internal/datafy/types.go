package datafy

import "github.com/aws/aws-sdk-go-v2/service/ec2/types"

type Volume struct {
	*types.Volume

	IsManaged     bool
	IsDatafied    bool
	IsReplacement bool
}
