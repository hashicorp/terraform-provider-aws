// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
)

func expandParameters(configured []interface{}) []*neptune.Parameter {
	parameters := make([]*neptune.Parameter, 0, len(configured))

	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		p := &neptune.Parameter{
			ApplyMethod:    aws.String(data["apply_method"].(string)),
			ParameterName:  aws.String(data["name"].(string)),
			ParameterValue: aws.String(data["value"].(string)),
		}

		parameters = append(parameters, p)
	}

	return parameters
}

func flattenParameters(list []*neptune.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))

	for _, i := range list {
		if i.ParameterValue != nil {
			result = append(result, map[string]interface{}{
				"apply_method": aws.StringValue(i.ApplyMethod),
				"name":         aws.StringValue(i.ParameterName),
				"value":        aws.StringValue(i.ParameterValue),
			})
		}
	}

	return result
}
