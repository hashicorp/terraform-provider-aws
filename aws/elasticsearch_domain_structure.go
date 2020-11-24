package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandAdvancedSecurityOptions(m []interface{}, create bool) *elasticsearch.AdvancedSecurityOptionsInput {
	config := elasticsearch.AdvancedSecurityOptionsInput{}
	group := m[0].(map[string]interface{})

	if advancedSecurityEnabled, ok := group["enabled"]; ok {
		config.Enabled = aws.Bool(advancedSecurityEnabled.(bool))

		if advancedSecurityEnabled.(bool) {
			if v, ok := group["internal_user_database_enabled"].(bool); ok {
				config.InternalUserDatabaseEnabled = aws.Bool(v)
			}

			if v, ok := group["master_user_options"].([]interface{}); ok {
				if len(v) > 0 && v[0] != nil {
					muo := elasticsearch.MasterUserOptions{}
					masterUserOptions := v[0].(map[string]interface{})

					if v, ok := masterUserOptions["master_user_arn"].(string); ok && v != "" {
						muo.MasterUserARN = aws.String(v)
					}

					if v, ok := masterUserOptions["master_user_name"].(string); ok && v != "" {
						muo.MasterUserName = aws.String(v)
					}

					if v, ok := masterUserOptions["master_user_password"].(string); ok && v != "" {
						muo.MasterUserPassword = aws.String(v)
					}

					config.SetMasterUserOptions(&muo)
				}
			}

			// You cannot specify SAML options during domain creation.
			if !create {
				if v, ok := group["saml_options"].([]interface{}); ok {
					if len(v) > 0 && v[0] != nil {
						options := elasticsearch.SAMLOptionsInput{}
						SAMLOptions := v[0].(map[string]interface{})

						if SAMLEnabled, ok := SAMLOptions["enabled"]; ok {
							options.Enabled = aws.Bool(SAMLEnabled.(bool))

							if SAMLEnabled.(bool) {
								options.Idp = expandSAMLOptionsIdp(SAMLOptions["idp"].([]interface{}))
								if v, ok := SAMLOptions["master_backend_role"].(string); ok && v != "" {
									options.MasterBackendRole = aws.String(v)
								}
								if v, ok := SAMLOptions["master_user_name"].(string); ok && v != "" {
									options.MasterUserName = aws.String(v)
								}
								if v, ok := SAMLOptions["roles_key"].(string); ok && v != "" {
									options.RolesKey = aws.String(v)
								}
								if v, ok := SAMLOptions["session_timeout_minutes"].(int); ok {
									options.SessionTimeoutMinutes = aws.Int64(int64(v))
								}
								if v, ok := SAMLOptions["subject_key"].(string); ok && v != "" {
									options.SubjectKey = aws.String(v)
								}
							}

							config.SetSAMLOptions(&options)
						}
					}
				}
			}
		}
	}

	return &config
}

func expandSAMLOptionsIdp(l []interface{}) *elasticsearch.SAMLIdp {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &elasticsearch.SAMLIdp{}
	}

	m := l[0].(map[string]interface{})

	return &elasticsearch.SAMLIdp{
		EntityId:        aws.String(m["entity_id"].(string)),
		MetadataContent: aws.String(m["metadata_content"].(string)),
	}
}

func flattenAdvancedSecurityOptions(advancedSecurityOptions *elasticsearch.AdvancedSecurityOptions) []map[string]interface{} {
	if advancedSecurityOptions == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}
	m["enabled"] = aws.BoolValue(advancedSecurityOptions.Enabled)
	if aws.BoolValue(advancedSecurityOptions.Enabled) {
		m["internal_user_database_enabled"] = aws.BoolValue(advancedSecurityOptions.InternalUserDatabaseEnabled)
	}

	return []map[string]interface{}{m}
}

func getMasterUserOptions(d *schema.ResourceData) []interface{} {
	if v, ok := d.GetOk("advanced_security_options"); ok {
		options := v.([]interface{})
		if len(options) > 0 && options[0] != nil {
			m := options[0].(map[string]interface{})
			if opts, ok := m["master_user_options"]; ok {
				return opts.([]interface{})
			}
		}
	}
	return []interface{}{}
}

func getUserDBEnabled(d *schema.ResourceData) bool {
	if v, ok := d.GetOk("advanced_security_options"); ok {
		options := v.([]interface{})
		if len(options) > 0 && options[0] != nil {
			m := options[0].(map[string]interface{})
			if enabled, ok := m["internal_user_database_enabled"]; ok {
				return enabled.(bool)
			}
		}
	}
	return false
}
