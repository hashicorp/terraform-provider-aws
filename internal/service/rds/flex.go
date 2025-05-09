// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenManagedMasterUserSecret(apiObject *types.MasterUserSecret) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(v)
	}

	if v := apiObject.SecretArn; v != nil {
		tfMap["secret_arn"] = aws.ToString(v)
	}

	if v := apiObject.SecretStatus; v != nil {
		tfMap["secret_status"] = aws.ToString(v)
	}

	return tfMap
}

func expandParameters(tfList []any) []types.Parameter {
	var apiObjects []types.Parameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		if tfMap[names.AttrName].(string) == "" {
			continue
		}

		apiObject := types.Parameter{
			ParameterName:  aws.String(strings.ToLower(tfMap[names.AttrName].(string))),
			ParameterValue: aws.String(tfMap[names.AttrValue].(string)),
		}

		if v, ok := tfMap["apply_method"].(string); ok && v != "" {
			apiObject.ApplyMethod = types.ApplyMethod(strings.ToLower(v))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenParameters(apiObject []types.Parameter) []any {
	apiObjects := make([]any, 0)

	for _, apiObject := range apiObject {
		if apiObject.ParameterName == nil {
			continue
		}

		tfMap := make(map[string]any)

		tfMap["apply_method"] = strings.ToLower(string(apiObject.ApplyMethod))
		tfMap[names.AttrName] = strings.ToLower(aws.ToString(apiObject.ParameterName))

		// Default empty string, guard against nil parameter values.
		tfMap[names.AttrValue] = ""
		if apiObject.ParameterValue != nil {
			tfMap[names.AttrValue] = aws.ToString(apiObject.ParameterValue)
		}

		apiObjects = append(apiObjects, tfMap)
	}

	return apiObjects
}
