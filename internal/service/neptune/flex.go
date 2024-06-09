// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandParameters(configured []interface{}) []*neptune.Parameter {
	parameters := make([]*neptune.Parameter, 0, len(configured))

	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		p := &neptune.Parameter{
			ApplyMethod:    aws.String(data["apply_method"].(string)),
			ParameterName:  aws.String(data[names.AttrName].(string)),
			ParameterValue: aws.String(data[names.AttrValue].(string)),
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
				"apply_method":  aws.StringValue(i.ApplyMethod),
				names.AttrName:  aws.StringValue(i.ParameterName),
				names.AttrValue: aws.StringValue(i.ParameterValue),
			})
		}
	}

	return result
}
