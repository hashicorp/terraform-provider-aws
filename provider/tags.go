package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/keyvaluetags"
)

func expandProviderDefaultTags(l []interface{}) *keyvaluetags.DefaultConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	defaultConfig := &keyvaluetags.DefaultConfig{}
	m := l[0].(map[string]interface{})

	if v, ok := m["tags"].(map[string]interface{}); ok {
		defaultConfig.Tags = keyvaluetags.New(v)
	}
	return defaultConfig
}

func expandProviderIgnoreTags(l []interface{}) *keyvaluetags.IgnoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	ignoreConfig := &keyvaluetags.IgnoreConfig{}
	m := l[0].(map[string]interface{})

	if v, ok := m["keys"].(*schema.Set); ok {
		ignoreConfig.Keys = keyvaluetags.New(v.List())
	}

	if v, ok := m["key_prefixes"].(*schema.Set); ok {
		ignoreConfig.KeyPrefixes = keyvaluetags.New(v.List())
	}

	return ignoreConfig
}
