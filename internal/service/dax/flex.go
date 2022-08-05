package dax

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"
)

func expandEncryptAtRestOptions(m map[string]interface{}) *dax.SSESpecification {
	options := dax.SSESpecification{}

	if v, ok := m["enabled"]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}

	return &options
}

func expandParameterGroupParameterNameValue(config []interface{}) []*dax.ParameterNameValue {
	if len(config) == 0 {
		return nil
	}
	results := make([]*dax.ParameterNameValue, 0, len(config))
	for _, raw := range config {
		m := raw.(map[string]interface{})
		pnv := &dax.ParameterNameValue{
			ParameterName:  aws.String(m["name"].(string)),
			ParameterValue: aws.String(m["value"].(string)),
		}
		results = append(results, pnv)
	}
	return results
}

func flattenEncryptAtRestOptions(options *dax.SSEDescription) []map[string]interface{} {
	m := map[string]interface{}{
		"enabled": false,
	}

	if options == nil {
		return []map[string]interface{}{m}
	}

	m["enabled"] = (aws.StringValue(options.Status) == dax.SSEStatusEnabled)

	return []map[string]interface{}{m}
}

func flattenParameterGroupParameters(params []*dax.Parameter) []map[string]interface{} {
	if len(params) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, p := range params {
		m := map[string]interface{}{
			"name":  aws.StringValue(p.ParameterName),
			"value": aws.StringValue(p.ParameterValue),
		}
		results = append(results, m)
	}
	return results
}

func flattenSecurityGroupIDs(securityGroups []*dax.SecurityGroupMembership) []string {
	result := make([]string, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if sg.SecurityGroupIdentifier != nil {
			result = append(result, *sg.SecurityGroupIdentifier)
		}
	}
	return result
}
