// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

func expandServerScope(inputs []interface{}) []awstypes.ResourceServerScopeType {
	configs := make([]awstypes.ResourceServerScopeType, len(inputs))
	for i, input := range inputs {
		param := input.(map[string]interface{})
		config := awstypes.ResourceServerScopeType{}

		if v, ok := param["scope_description"]; ok {
			config.ScopeDescription = aws.String(v.(string))
		}

		if v, ok := param["scope_name"]; ok {
			config.ScopeName = aws.String(v.(string))
		}

		configs[i] = config
	}

	return configs
}

func flattenServerScope(inputs []awstypes.ResourceServerScopeType) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, input := range inputs {
		var value = map[string]interface{}{
			"scope_name":        aws.ToString(input.ScopeName),
			"scope_description": aws.ToString(input.ScopeDescription),
		}
		values = append(values, value)
	}
	return values
}
