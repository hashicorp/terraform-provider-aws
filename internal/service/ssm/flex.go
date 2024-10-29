// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandTargets(tfList []interface{}) []awstypes.Target {
	apiObjects := make([]awstypes.Target, 0)

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := awstypes.Target{
			Key:    aws.String(tfMap[names.AttrKey].(string)),
			Values: flex.ExpandStringValueList(tfMap[names.AttrValues].([]interface{})),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenTargets(apiObjects []awstypes.Target) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]interface{}, 1)
		tfMap[names.AttrKey] = aws.ToString(apiObject.Key)
		tfMap[names.AttrValues] = apiObject.Values

		tfList = append(tfList, tfMap)
	}

	return tfList
}
