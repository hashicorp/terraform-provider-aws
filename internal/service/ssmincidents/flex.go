package ssmincidents

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
)

func expandRegions(regions []interface{}) map[string]types.RegionMapInputValue {
	if len(regions) == 0 {
		return nil
	}

	regionMap := make(map[string]types.RegionMapInputValue)
	for _, region := range regions {
		regionData := region.(map[string]interface{})

		input := types.RegionMapInputValue{}

		if kmsKey := regionData["kms_key_arn"].(string); kmsKey != "DefaultKey" {
			input.SseKmsKeyId = aws.String(kmsKey)
		}

		regionMap[regionData["name"].(string)] = input
	}

	return regionMap
}

func flattenRegions(regions map[string]types.RegionInfo) []map[string]interface{} {
	if len(regions) == 0 {
		return nil
	}

	tfRegionData := make([]map[string]interface{}, 0)
	for regionName, regionData := range regions {
		region := make(map[string]interface{})

		region["name"] = regionName
		region["status"] = regionData.Status
		region["kms_key_arn"] = aws.ToString(regionData.SseKmsKeyId)

		if v := regionData.StatusMessage; v != nil {
			region["status_message"] = aws.ToString(v)
		}

		tfRegionData = append(tfRegionData, region)
	}

	return tfRegionData
}
