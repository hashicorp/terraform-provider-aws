package opensearchserverless

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
)

func expandSAMLOptions(l []interface{}) *types.SamlConfigOptions {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &types.SamlConfigOptions{}
	}

	m := l[0].(map[string]interface{})

	options := &types.SamlConfigOptions{
		Metadata: aws.String(m["metadata"].(string)),
	}

	if v, ok := m["group_attribute"].(string); ok && v != "" {
		options.GroupAttribute = aws.String(v)
	}

	if v, ok := m["session_timeout"].(int); ok && v != 0 {
		options.SessionTimeout = aws.Int32(int32(v))
	}

	if v, ok := m["user_attribute"].(string); ok && v != "" {
		options.UserAttribute = aws.String(v)
	}

	return options
}

func flattenSAMLOptions(samlOptions *types.SamlConfigOptions) []interface{} {
	if samlOptions == nil {
		return nil
	}

	m := map[string]interface{}{
		"group_attribute": aws.ToString(samlOptions.GroupAttribute),
		"metadata":        aws.ToString(samlOptions.Metadata),
		"session_timeout": int(aws.ToInt32(samlOptions.SessionTimeout)),
		"user_attribute":  aws.ToString(samlOptions.UserAttribute),
	}

	return []interface{}{m}
}
