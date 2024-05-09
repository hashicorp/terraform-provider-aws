// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandSourceS3Path(tfList []interface{}) *types.S3Path {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.S3Path{}

	if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		result.Key = aws.String(v)
	}

	return result
}

func flattenSourceS3Path(apiObject *types.S3Path) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		m[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Key; v != nil {
		m[names.AttrKey] = aws.ToString(v)
	}

	return []interface{}{m}
}
