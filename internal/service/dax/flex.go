// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dax

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dax/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandEncryptAtRestOptions(m map[string]interface{}) *awstypes.SSESpecification {
	options := awstypes.SSESpecification{}

	if v, ok := m[names.AttrEnabled]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}

	return &options
}

func expandParameterGroupParameterNameValue(config []interface{}) []awstypes.ParameterNameValue {
	if len(config) == 0 {
		return nil
	}
	results := make([]awstypes.ParameterNameValue, 0, len(config))
	for _, raw := range config {
		m := raw.(map[string]interface{})
		pnv := awstypes.ParameterNameValue{
			ParameterName:  aws.String(m[names.AttrName].(string)),
			ParameterValue: aws.String(m[names.AttrValue].(string)),
		}
		results = append(results, pnv)
	}
	return results
}

func flattenEncryptAtRestOptions(options *awstypes.SSEDescription) []map[string]interface{} {
	m := map[string]interface{}{
		names.AttrEnabled: false,
	}

	if options == nil {
		return []map[string]interface{}{m}
	}

	m[names.AttrEnabled] = options.Status == awstypes.SSEStatusEnabled

	return []map[string]interface{}{m}
}

func flattenParameterGroupParameters(params []awstypes.Parameter) []map[string]interface{} {
	if len(params) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, p := range params {
		m := map[string]interface{}{
			names.AttrName:  aws.ToString(p.ParameterName),
			names.AttrValue: aws.ToString(p.ParameterValue),
		}
		results = append(results, m)
	}
	return results
}

func flattenSecurityGroupIDs(securityGroups []awstypes.SecurityGroupMembership) []string {
	result := make([]string, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if sg.SecurityGroupIdentifier != nil {
			result = append(result, *sg.SecurityGroupIdentifier)
		}
	}
	return result
}
