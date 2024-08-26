// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/document"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenAddress(apiObject *types.Address) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Country; v != nil {
		m["country"] = aws.ToString(v)
	}

	if v := apiObject.Formatted; v != nil {
		m["formatted"] = aws.ToString(v)
	}

	if v := apiObject.Locality; v != nil {
		m["locality"] = aws.ToString(v)
	}

	if v := apiObject.PostalCode; v != nil {
		m["postal_code"] = aws.ToString(v)
	}

	m["primary"] = apiObject.Primary

	if v := apiObject.Region; v != nil {
		m[names.AttrRegion] = aws.ToString(v)
	}

	if v := apiObject.StreetAddress; v != nil {
		m["street_address"] = aws.ToString(v)
	}

	if v := apiObject.Type; v != nil {
		m[names.AttrType] = aws.ToString(v)
	}

	return m
}

func expandAddress(tfMap map[string]interface{}) *types.Address {
	if tfMap == nil {
		return nil
	}

	a := &types.Address{}

	if v, ok := tfMap["country"].(string); ok && v != "" {
		a.Country = aws.String(v)
	}

	if v, ok := tfMap["formatted"].(string); ok && v != "" {
		a.Formatted = aws.String(v)
	}

	if v, ok := tfMap["locality"].(string); ok && v != "" {
		a.Locality = aws.String(v)
	}

	if v, ok := tfMap["postal_code"].(string); ok && v != "" {
		a.PostalCode = aws.String(v)
	}

	a.Primary = tfMap["primary"].(bool)

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		a.Region = aws.String(v)
	}

	if v, ok := tfMap["street_address"].(string); ok && v != "" {
		a.StreetAddress = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		a.Type = aws.String(v)
	}

	return a
}

func flattenAddresses(apiObjects []types.Address) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		apiObject := apiObject
		l = append(l, flattenAddress(&apiObject))
	}

	return l
}

func expandAddresses(tfList []interface{}) []types.Address {
	s := make([]types.Address, 0, len(tfList))

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandAddress(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandAlternateIdentifier(tfMap map[string]interface{}) types.AlternateIdentifier {
	if tfMap == nil {
		return nil
	}

	if v, ok := tfMap[names.AttrExternalID]; ok && len(v.([]interface{})) > 0 {
		return &types.AlternateIdentifierMemberExternalId{
			Value: *expandExternalId(v.([]interface{})[0].(map[string]interface{})),
		}
	} else if v, ok := tfMap["unique_attribute"]; ok && len(v.([]interface{})) > 0 {
		return &types.AlternateIdentifierMemberUniqueAttribute{
			Value: *expandUniqueAttribute(v.([]interface{})[0].(map[string]interface{})),
		}
	} else {
		return nil
	}
}

func flattenEmail(apiObject *types.Email) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["primary"] = apiObject.Primary

	if v := apiObject.Type; v != nil {
		m[names.AttrType] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		m[names.AttrValue] = aws.ToString(v)
	}

	return m
}

func expandEmail(tfMap map[string]interface{}) *types.Email {
	if tfMap == nil {
		return nil
	}

	a := &types.Email{}

	a.Primary = tfMap["primary"].(bool)

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		a.Type = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		a.Value = aws.String(v)
	}

	return a
}

func flattenEmails(apiObjects []types.Email) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		apiObject := apiObject
		l = append(l, flattenEmail(&apiObject))
	}

	return l
}

func expandEmails(tfList []interface{}) []types.Email {
	s := make([]types.Email, 0, len(tfList))

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandEmail(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandExternalId(tfMap map[string]interface{}) *types.ExternalId {
	if tfMap == nil {
		return nil
	}

	a := &types.ExternalId{}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		a.Id = aws.String(v)
	}

	if v, ok := tfMap[names.AttrIssuer].(string); ok && v != "" {
		a.Issuer = aws.String(v)
	}

	return a
}

func flattenExternalId(apiObject *types.ExternalId) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Id; v != nil {
		m[names.AttrID] = aws.ToString(v)
	}

	if v := apiObject.Issuer; v != nil {
		m[names.AttrIssuer] = aws.ToString(v)
	}

	return m
}

func flattenExternalIds(apiObjects []types.ExternalId) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		apiObject := apiObject
		l = append(l, flattenExternalId(&apiObject))
	}

	return l
}

func flattenName(apiObject *types.Name) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.FamilyName; v != nil {
		m["family_name"] = aws.ToString(v)
	}

	if v := apiObject.Formatted; v != nil {
		m["formatted"] = aws.ToString(v)
	}

	if v := apiObject.GivenName; v != nil {
		m["given_name"] = aws.ToString(v)
	}

	if v := apiObject.HonorificPrefix; v != nil {
		m["honorific_prefix"] = aws.ToString(v)
	}

	if v := apiObject.HonorificSuffix; v != nil {
		m["honorific_suffix"] = aws.ToString(v)
	}

	if v := apiObject.MiddleName; v != nil {
		m["middle_name"] = aws.ToString(v)
	}

	return m
}

func expandName(tfMap map[string]interface{}) *types.Name {
	if tfMap == nil {
		return nil
	}

	a := &types.Name{}

	if v, ok := tfMap["family_name"].(string); ok && v != "" {
		a.FamilyName = aws.String(v)
	}

	if v, ok := tfMap["formatted"].(string); ok && v != "" {
		a.Formatted = aws.String(v)
	}

	if v, ok := tfMap["given_name"].(string); ok && v != "" {
		a.GivenName = aws.String(v)
	}

	if v, ok := tfMap["honorific_prefix"].(string); ok && v != "" {
		a.HonorificPrefix = aws.String(v)
	}

	if v, ok := tfMap["honorific_suffix"].(string); ok && v != "" {
		a.HonorificSuffix = aws.String(v)
	}

	if v, ok := tfMap["middle_name"].(string); ok && v != "" {
		a.MiddleName = aws.String(v)
	}

	return a
}

func flattenPhoneNumber(apiObject *types.PhoneNumber) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["primary"] = apiObject.Primary

	if v := apiObject.Type; v != nil {
		m[names.AttrType] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		m[names.AttrValue] = aws.ToString(v)
	}

	return m
}

func expandPhoneNumber(tfMap map[string]interface{}) *types.PhoneNumber {
	if tfMap == nil {
		return nil
	}

	a := &types.PhoneNumber{}

	a.Primary = tfMap["primary"].(bool)

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		a.Type = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		a.Value = aws.String(v)
	}

	return a
}

func flattenPhoneNumbers(apiObjects []types.PhoneNumber) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		apiObject := apiObject
		l = append(l, flattenPhoneNumber(&apiObject))
	}

	return l
}

func expandPhoneNumbers(tfList []interface{}) []types.PhoneNumber {
	s := make([]types.PhoneNumber, 0, len(tfList))

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandPhoneNumber(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandUniqueAttribute(tfMap map[string]interface{}) *types.UniqueAttribute {
	if tfMap == nil {
		return nil
	}

	a := &types.UniqueAttribute{}

	if v, ok := tfMap["attribute_path"].(string); ok && v != "" {
		a.AttributePath = aws.String(v)
	}

	if v, ok := tfMap["attribute_value"].(string); ok && v != "" {
		a.AttributeValue = document.NewLazyDocument(v)
	}

	return a
}
