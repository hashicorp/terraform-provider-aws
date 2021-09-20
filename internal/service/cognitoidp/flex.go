package cognitoidp

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

func expandServerScope(inputs []interface{}) []*cognitoidentityprovider.ResourceServerScopeType {
	configs := make([]*cognitoidentityprovider.ResourceServerScopeType, len(inputs))
	for i, input := range inputs {
		param := input.(map[string]interface{})
		config := &cognitoidentityprovider.ResourceServerScopeType{}

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

func flattenServerScope(inputs []*cognitoidentityprovider.ResourceServerScopeType) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, input := range inputs {
		if input == nil {
			continue
		}
		var value = map[string]interface{}{
			"scope_name":        aws.StringValue(input.ScopeName),
			"scope_description": aws.StringValue(input.ScopeDescription),
		}
		values = append(values, value)
	}
	return values
}
