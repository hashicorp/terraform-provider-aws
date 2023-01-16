package ssmincidents

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
)

func ExpandRegions(regions []interface{}) map[string]types.RegionMapInputValue {
	if len(regions) == 0 {
		return nil
	}

	ret := make(map[string]types.RegionMapInputValue)
	for _, region := range regions {
		regionData := region.(map[string]interface{})

		input := types.RegionMapInputValue{}

		if kmsKey := regionData["kms_key_arn"].(string); kmsKey != "DefaultKey" {
			input.SseKmsKeyId = aws.String(kmsKey)
		}

		ret[regionData["name"].(string)] = input
	}

	return ret
}

func FlattenRegions(regions map[string]types.RegionInfo) []interface{} {
	if len(regions) == 0 {
		return nil
	}

	ret := make([]interface{}, 0)
	for name, v := range regions {
		region := make(map[string]interface{})

		region["name"] = name
		region["status"] = v.Status
		region["status_update_time"] = aws.ToTime(v.StatusUpdateDateTime).String()
		region["kms_key_arn"] = aws.ToString(v.SseKmsKeyId)
		region["status_message"] = aws.ToString(v.StatusMessage)

		ret = append(ret, region)
	}

	return ret
}
