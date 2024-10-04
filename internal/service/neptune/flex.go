// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandParameters(configured []interface{}) []awstypes.Parameter {
	parameters := make([]awstypes.Parameter, 0, len(configured))

	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		p := awstypes.Parameter{
			ApplyMethod:    awstypes.ApplyMethod(data["apply_method"].(string)),
			ParameterName:  aws.String(data[names.AttrName].(string)),
			ParameterValue: aws.String(data[names.AttrValue].(string)),
		}

		parameters = append(parameters, p)
	}

	return parameters
}

func flattenParameters(list []awstypes.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))

	for _, i := range list {
		if i.ParameterValue != nil {
			result = append(result, map[string]interface{}{
				"apply_method":  string(i.ApplyMethod),
				names.AttrName:  aws.ToString(i.ParameterName),
				names.AttrValue: aws.ToString(i.ParameterValue),
			})
		}
	}

	return result
}
