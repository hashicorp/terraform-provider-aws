// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package attrmap

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// AttributeMap represents a map of Terraform resource attribute name to AWS API attribute name.
// Useful for SQS Queue or SNS Topic attribute handling.
type attributeInfo struct {
	alwaysSendConfiguredValueOnCreate bool
	apiAttributeName                  string
	tfType                            schema.ValueType
	tfComputed                        bool
	tfOptional                        bool
	isIAMPolicy                       bool
	missingSetToNil                   bool
	skipUpdate                        bool
}

type AttributeMap map[string]attributeInfo

// New returns a new AttributeMap from the specified Terraform resource attribute name to AWS API attribute name map and resource schema.
func New(attrMap map[string]string, schemaMap map[string]*schema.Schema) AttributeMap {
	attributeMap := make(AttributeMap)

	for tfAttributeName, apiAttributeName := range attrMap {
		if s, ok := schemaMap[tfAttributeName]; ok {
			attributeInfo := attributeInfo{
				apiAttributeName: apiAttributeName,
				tfType:           s.Type,
			}

			attributeInfo.tfComputed = s.Computed
			attributeInfo.tfOptional = s.Optional

			attributeMap[tfAttributeName] = attributeInfo
		} else {
			log.Printf("[ERROR] Unknown attribute: %s", tfAttributeName)
		}
	}

	return attributeMap
}

// APIAttributesToResourceData sets Terraform ResourceData from a map of AWS API attributes.
func (m AttributeMap) APIAttributesToResourceData(apiAttributes map[string]string, d *schema.ResourceData) error {
	for tfAttributeName, attributeInfo := range m {
		if v, ok := apiAttributes[attributeInfo.apiAttributeName]; ok {
			var err error
			var tfAttributeValue interface{}

			switch t := attributeInfo.tfType; t {
			case schema.TypeBool:
				tfAttributeValue, err = strconv.ParseBool(v)

				if err != nil {
					return fmt.Errorf("parsing %s value (%s) into boolean: %w", tfAttributeName, v, err)
				}
			case schema.TypeInt:
				tfAttributeValue, err = strconv.Atoi(v)

				if err != nil {
					return fmt.Errorf("parsing %s value (%s) into integer: %w", tfAttributeName, v, err)
				}
			case schema.TypeString:
				tfAttributeValue = v

				if attributeInfo.isIAMPolicy {
					policy, err := verify.PolicyToSet(d.Get(tfAttributeName).(string), tfAttributeValue.(string))

					if err != nil {
						return err
					}

					tfAttributeValue = policy
				}
			default:
				return fmt.Errorf("attribute %s is of unsupported type: %d", tfAttributeName, t)
			}

			if err := d.Set(tfAttributeName, tfAttributeValue); err != nil {
				return fmt.Errorf("setting %s: %w", tfAttributeName, err)
			}
		} else if attributeInfo.missingSetToNil {
			d.Set(tfAttributeName, nil)
		}
	}

	return nil
}

// ResourceDataToAPIAttributesCreate returns a map of AWS API attributes from Terraform ResourceData.
// The API attributes map is suitable for resource create.
func (m AttributeMap) ResourceDataToAPIAttributesCreate(d *schema.ResourceData) (map[string]string, error) {
	apiAttributes := map[string]string{}

	for tfAttributeName, attributeInfo := range m {
		// Purely Computed values aren't specified on creation.
		if attributeInfo.tfComputed && !attributeInfo.tfOptional {
			continue
		}

		var apiAttributeValue string
		configuredValue := d.GetRawConfig().GetAttr(tfAttributeName)
		tfOptionalComputed := attributeInfo.tfComputed && attributeInfo.tfOptional

		switch v, t := d.Get(tfAttributeName), attributeInfo.tfType; t {
		case schema.TypeBool:
			if v := v.(bool); v || (attributeInfo.alwaysSendConfiguredValueOnCreate && !configuredValue.IsNull()) {
				apiAttributeValue = strconv.FormatBool(v)
			}
		case schema.TypeInt:
			// On creation don't specify any zero Optional/Computed attribute integer values.
			if v := v.(int); !tfOptionalComputed || v != 0 {
				apiAttributeValue = strconv.Itoa(v)
			}
		case schema.TypeString:
			apiAttributeValue = v.(string)

			if attributeInfo.isIAMPolicy && apiAttributeValue != "" {
				policy, err := structure.NormalizeJsonString(apiAttributeValue)
				if err != nil {
					return nil, fmt.Errorf("policy (%s) is invalid JSON: %w", apiAttributeValue, err)
				}

				apiAttributeValue = policy
			}
		default:
			return nil, fmt.Errorf("attribute %s is of unsupported type: %d", tfAttributeName, t)
		}

		if apiAttributeValue != "" {
			apiAttributes[attributeInfo.apiAttributeName] = apiAttributeValue
		}
	}

	return apiAttributes, nil
}

// ResourceDataToAPIAttributesUpdate returns a map of AWS API attributes from Terraform ResourceData.
// The API attributes map is suitable for resource update.
func (m AttributeMap) ResourceDataToAPIAttributesUpdate(d *schema.ResourceData) (map[string]string, error) {
	apiAttributes := map[string]string{}

	for tfAttributeName, attributeInfo := range m {
		if attributeInfo.skipUpdate {
			continue
		}

		// Purely Computed values aren't specified on update.
		if attributeInfo.tfComputed && !attributeInfo.tfOptional {
			continue
		}

		if d.HasChange(tfAttributeName) {
			v := d.Get(tfAttributeName)

			var apiAttributeValue string

			switch t := attributeInfo.tfType; t {
			case schema.TypeBool:
				apiAttributeValue = strconv.FormatBool(v.(bool))
			case schema.TypeInt:
				apiAttributeValue = strconv.Itoa(v.(int))
			case schema.TypeString:
				apiAttributeValue = v.(string)

				if attributeInfo.isIAMPolicy {
					policy, err := structure.NormalizeJsonString(apiAttributeValue)

					if err != nil {
						return nil, fmt.Errorf("policy (%s) is invalid JSON: %w", apiAttributeValue, err)
					}

					apiAttributeValue = policy
				}
			default:
				return nil, fmt.Errorf("attribute %s is of unsupported type: %d", tfAttributeName, t)
			}

			apiAttributes[attributeInfo.apiAttributeName] = apiAttributeValue
		}
	}

	return apiAttributes, nil
}

// APIAttributeNames returns the AWS API attribute names.
func (m AttributeMap) APIAttributeNames() []string {
	apiAttributeNames := []string{}

	for _, attributeInfo := range m {
		apiAttributeNames = append(apiAttributeNames, attributeInfo.apiAttributeName)
	}

	return apiAttributeNames
}

// WithAlwaysSendConfiguredBooleanValueOnCreate marks the specified Terraform Boolean attribute as always having any configured value sent on resource create.
// By default a Boolean value is only sent to the API on resource create if its configured value is true.
// This method is intended to be chained with other similar helper methods in a builder pattern.
func (m AttributeMap) WithAlwaysSendConfiguredBooleanValueOnCreate(tfAttributeName string) AttributeMap {
	if attributeInfo, ok := m[tfAttributeName]; ok && attributeInfo.tfType == schema.TypeBool {
		attributeInfo.alwaysSendConfiguredValueOnCreate = true
		m[tfAttributeName] = attributeInfo
	}

	return m
}

// WithIAMPolicyAttribute marks the specified Terraform attribute as holding an AWS IAM policy.
// AWS IAM policies get special handling.
// This method is intended to be chained with other similar helper methods in a builder pattern.
func (m AttributeMap) WithIAMPolicyAttribute(tfAttributeName string) AttributeMap {
	if attributeInfo, ok := m[tfAttributeName]; ok {
		attributeInfo.isIAMPolicy = true
		m[tfAttributeName] = attributeInfo
	}

	return m
}

// WithMissingSetToNil marks the specified Terraform attribute as being set to nil if it's missing after reading the API.
// An attribute name of "*" means all attributes get marked.
// This method is intended to be chained with other similar helper methods in a builder pattern.
func (m AttributeMap) WithMissingSetToNil(tfAttributeName string) AttributeMap {
	if tfAttributeName == "*" {
		for k, attributeInfo := range m {
			attributeInfo.missingSetToNil = true
			m[k] = attributeInfo
		}
	} else if attributeInfo, ok := m[tfAttributeName]; ok {
		attributeInfo.missingSetToNil = true
		m[tfAttributeName] = attributeInfo
	}

	return m
}

// WithSkipUpdate marks the specified Terraform attribute as skipping update handling.
// This method is intended to be chained with other similar helper methods in a builder pattern.
func (m AttributeMap) WithSkipUpdate(tfAttributeName string) AttributeMap {
	if attributeInfo, ok := m[tfAttributeName]; ok {
		attributeInfo.skipUpdate = true
		m[tfAttributeName] = attributeInfo
	}

	return m
}
