// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandIdentityPoolRoleMappingsAttachment(rms []interface{}) map[string]awstypes.RoleMapping {
	values := make(map[string]awstypes.RoleMapping)

	if len(rms) == 0 {
		return values
	}

	for _, v := range rms {
		rm := v.(map[string]interface{})
		key := rm["identity_provider"].(string)

		roleMapping := awstypes.RoleMapping{
			Type: awstypes.RoleMappingType(rm[names.AttrType].(string)),
		}

		if sv, ok := rm["ambiguous_role_resolution"].(string); ok {
			roleMapping.AmbiguousRoleResolution = awstypes.AmbiguousRoleResolutionType(sv)
		}

		if mr, ok := rm["mapping_rule"].([]interface{}); ok && len(mr) > 0 {
			rct := &awstypes.RulesConfigurationType{}
			mappingRules := make([]awstypes.MappingRule, 0)

			for _, r := range mr {
				rule := r.(map[string]interface{})
				mr := awstypes.MappingRule{
					Claim:     aws.String(rule["claim"].(string)),
					MatchType: awstypes.MappingRuleMatchType(rule["match_type"].(string)),
					RoleARN:   aws.String(rule[names.AttrRoleARN].(string)),
					Value:     aws.String(rule[names.AttrValue].(string)),
				}

				mappingRules = append(mappingRules, mr)
			}

			rct.Rules = mappingRules
			roleMapping.RulesConfiguration = rct
		}

		values[key] = roleMapping
	}

	return values
}

func expandIdentityPoolRoles(config map[string]interface{}) map[string]string {
	m := map[string]string{}
	for k, v := range config {
		s := v.(string)
		m[k] = s
	}
	return m
}

func expandIdentityProviders(s *schema.Set) []awstypes.CognitoIdentityProvider {
	ips := make([]awstypes.CognitoIdentityProvider, 0)

	for _, v := range s.List() {
		s := v.(map[string]interface{})

		ip := awstypes.CognitoIdentityProvider{}

		if sv, ok := s[names.AttrClientID].(string); ok {
			ip.ClientId = aws.String(sv)
		}

		if sv, ok := s[names.AttrProviderName].(string); ok {
			ip.ProviderName = aws.String(sv)
		}

		if sv, ok := s["server_side_token_check"].(bool); ok {
			ip.ServerSideTokenCheck = aws.Bool(sv)
		}

		ips = append(ips, ip)
	}

	return ips
}

func expandSupportedLoginProviders(config map[string]interface{}) map[string]string {
	m := map[string]string{}
	for k, v := range config {
		s := v.(string)
		m[k] = s
	}
	return m
}

func flattenIdentityPoolRoleMappingsAttachment(rms map[string]awstypes.RoleMapping) []map[string]interface{} {
	roleMappings := make([]map[string]interface{}, 0)

	if rms == nil {
		return roleMappings
	}

	for k, v := range rms {
		m := make(map[string]interface{})

		if v.Type != "" {
			m[names.AttrType] = string(v.Type)
		}

		if v.AmbiguousRoleResolution != "" {
			m["ambiguous_role_resolution"] = string(v.AmbiguousRoleResolution)
		}

		if v.RulesConfiguration != nil && v.RulesConfiguration.Rules != nil {
			m["mapping_rule"] = flattenIdentityPoolRolesAttachmentMappingRules(v.RulesConfiguration.Rules)
		}

		m["identity_provider"] = k
		roleMappings = append(roleMappings, m)
	}

	return roleMappings
}

func flattenIdentityPoolRolesAttachmentMappingRules(d []awstypes.MappingRule) []interface{} {
	rules := make([]interface{}, 0)

	for _, rule := range d {
		r := make(map[string]interface{})
		r["claim"] = aws.ToString(rule.Claim)
		r["match_type"] = string(rule.MatchType)
		r[names.AttrRoleARN] = aws.ToString(rule.RoleARN)
		r[names.AttrValue] = aws.ToString(rule.Value)

		rules = append(rules, r)
	}

	return rules
}

func flattenIdentityProviders(ips []awstypes.CognitoIdentityProvider) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, v := range ips {
		ip := make(map[string]interface{})

		if v.ClientId != nil {
			ip[names.AttrClientID] = aws.ToString(v.ClientId)
		}

		if v.ProviderName != nil {
			ip[names.AttrProviderName] = aws.ToString(v.ProviderName)
		}

		if v.ServerSideTokenCheck != nil {
			ip["server_side_token_check"] = aws.ToBool(v.ServerSideTokenCheck)
		}

		values = append(values, ip)
	}

	return values
}
