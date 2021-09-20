package cognitoidentity

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandIdentityPoolRoleMappingsAttachment(rms []interface{}) map[string]*cognitoidentity.RoleMapping {
	values := make(map[string]*cognitoidentity.RoleMapping)

	if len(rms) == 0 {
		return values
	}

	for _, v := range rms {
		rm := v.(map[string]interface{})
		key := rm["identity_provider"].(string)

		roleMapping := &cognitoidentity.RoleMapping{
			Type: aws.String(rm["type"].(string)),
		}

		if sv, ok := rm["ambiguous_role_resolution"].(string); ok {
			roleMapping.AmbiguousRoleResolution = aws.String(sv)
		}

		if mr, ok := rm["mapping_rule"].([]interface{}); ok && len(mr) > 0 {
			rct := &cognitoidentity.RulesConfigurationType{}
			mappingRules := make([]*cognitoidentity.MappingRule, 0)

			for _, r := range mr {
				rule := r.(map[string]interface{})
				mr := &cognitoidentity.MappingRule{
					Claim:     aws.String(rule["claim"].(string)),
					MatchType: aws.String(rule["match_type"].(string)),
					RoleARN:   aws.String(rule["role_arn"].(string)),
					Value:     aws.String(rule["value"].(string)),
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

func expandIdentityPoolRoles(config map[string]interface{}) map[string]*string {
	m := map[string]*string{}
	for k, v := range config {
		s := v.(string)
		m[k] = &s
	}
	return m
}

func expandIdentityProviders(s *schema.Set) []*cognitoidentity.Provider {
	ips := make([]*cognitoidentity.Provider, 0)

	for _, v := range s.List() {
		s := v.(map[string]interface{})

		ip := &cognitoidentity.Provider{}

		if sv, ok := s["client_id"].(string); ok {
			ip.ClientId = aws.String(sv)
		}

		if sv, ok := s["provider_name"].(string); ok {
			ip.ProviderName = aws.String(sv)
		}

		if sv, ok := s["server_side_token_check"].(bool); ok {
			ip.ServerSideTokenCheck = aws.Bool(sv)
		}

		ips = append(ips, ip)
	}

	return ips
}

func expandSupportedLoginProviders(config map[string]interface{}) map[string]*string {
	m := map[string]*string{}
	for k, v := range config {
		s := v.(string)
		m[k] = &s
	}
	return m
}

func flattenIdentityPoolRoleMappingsAttachment(rms map[string]*cognitoidentity.RoleMapping) []map[string]interface{} {
	roleMappings := make([]map[string]interface{}, 0)

	if rms == nil {
		return roleMappings
	}

	for k, v := range rms {
		m := make(map[string]interface{})

		if v == nil {
			return nil
		}

		if v.Type != nil {
			m["type"] = *v.Type
		}

		if v.AmbiguousRoleResolution != nil {
			m["ambiguous_role_resolution"] = *v.AmbiguousRoleResolution
		}

		if v.RulesConfiguration != nil && v.RulesConfiguration.Rules != nil {
			m["mapping_rule"] = flattenIdentityPoolRolesAttachmentMappingRules(v.RulesConfiguration.Rules)
		}

		m["identity_provider"] = k
		roleMappings = append(roleMappings, m)
	}

	return roleMappings
}

func flattenIdentityPoolRoles(config map[string]*string) map[string]string {
	m := map[string]string{}
	for k, v := range config {
		m[k] = *v
	}
	return m
}

func flattenIdentityPoolRolesAttachmentMappingRules(d []*cognitoidentity.MappingRule) []interface{} {
	rules := make([]interface{}, 0)

	for _, rule := range d {
		r := make(map[string]interface{})
		r["claim"] = *rule.Claim
		r["match_type"] = *rule.MatchType
		r["role_arn"] = *rule.RoleARN
		r["value"] = *rule.Value

		rules = append(rules, r)
	}

	return rules
}

func flattenIdentityProviders(ips []*cognitoidentity.Provider) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, v := range ips {
		ip := make(map[string]interface{})

		if v == nil {
			return nil
		}

		if v.ClientId != nil {
			ip["client_id"] = *v.ClientId
		}

		if v.ProviderName != nil {
			ip["provider_name"] = *v.ProviderName
		}

		if v.ServerSideTokenCheck != nil {
			ip["server_side_token_check"] = *v.ServerSideTokenCheck
		}

		values = append(values, ip)
	}

	return values
}

func flattenSupportedLoginProviders(config map[string]*string) map[string]string {
	m := map[string]string{}
	for k, v := range config {
		m[k] = *v
	}
	return m
}
