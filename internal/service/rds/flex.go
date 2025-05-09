// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"slices"
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
	var characterSetsAndCollation, others []types.Parameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		parameterName := strings.ToLower(tfMap[names.AttrName].(string))
		if parameterName == "" {
			continue
		}

		apiObject := types.Parameter{
			ParameterName:  aws.String(parameterName),
			ParameterValue: aws.String(tfMap[names.AttrValue].(string)),
		}

		if v, ok := tfMap["apply_method"].(string); ok && v != "" {
			apiObject.ApplyMethod = types.ApplyMethod(strings.ToLower(v))
		}

		// Keep character set and collation parameters together.
		if strings.HasPrefix(parameterName, "character_set_") || strings.HasPrefix(parameterName, "collation_") {
			characterSetsAndCollation = append(characterSetsAndCollation, apiObject)
		} else {
			others = append(others, apiObject)
		}
	}

	return slices.Concat(characterSetsAndCollation, others)
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
