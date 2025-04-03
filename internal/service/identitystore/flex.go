// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/document"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenAddress(apiObject *types.Address) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Country; v != nil {
		tfMap["country"] = aws.ToString(v)
	}

	if v := apiObject.Formatted; v != nil {
		tfMap["formatted"] = aws.ToString(v)
	}

	if v := apiObject.Locality; v != nil {
		tfMap["locality"] = aws.ToString(v)
	}

	if v := apiObject.PostalCode; v != nil {
		tfMap["postal_code"] = aws.ToString(v)
	}

	tfMap["primary"] = apiObject.Primary

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	if v := apiObject.StreetAddress; v != nil {
		tfMap["street_address"] = aws.ToString(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	return tfMap
}

func expandAddress(tfMap map[string]any) *types.Address {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Address{}

	if v, ok := tfMap["country"].(string); ok && v != "" {
		apiObject.Country = aws.String(v)
	}

	if v, ok := tfMap["formatted"].(string); ok && v != "" {
		apiObject.Formatted = aws.String(v)
	}

	if v, ok := tfMap["locality"].(string); ok && v != "" {
		apiObject.Locality = aws.String(v)
	}

	if v, ok := tfMap["postal_code"].(string); ok && v != "" {
		apiObject.PostalCode = aws.String(v)
	}

	apiObject.Primary = tfMap["primary"].(bool)

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	if v, ok := tfMap["street_address"].(string); ok && v != "" {
		apiObject.StreetAddress = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func flattenAddresses(apiObjects []types.Address) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenAddress(&apiObject))
	}

	return tfList
}

func expandAddresses(tfList []any) []types.Address {
	apiObjects := make([]types.Address, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandAddress(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandAlternateIdentifier(tfMap map[string]any) types.AlternateIdentifier {
	if tfMap == nil {
		return nil
	}

	if v, ok := tfMap[names.AttrExternalID]; ok && len(v.([]any)) > 0 {
		return &types.AlternateIdentifierMemberExternalId{
			Value: *expandExternalID(v.([]any)[0].(map[string]any)),
		}
	} else if v, ok := tfMap["unique_attribute"]; ok && len(v.([]any)) > 0 {
		return &types.AlternateIdentifierMemberUniqueAttribute{
			Value: *expandUniqueAttribute(v.([]any)[0].(map[string]any)),
		}
	}

	return nil
}

func flattenEmail(apiObject *types.Email) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["primary"] = apiObject.Primary

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func expandEmail(tfMap map[string]any) *types.Email {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Email{}

	apiObject.Primary = tfMap["primary"].(bool)

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func flattenEmails(apiObjects []types.Email) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenEmail(&apiObject))
	}

	return tfList
}

func expandEmails(tfList []any) []types.Email {
	apiObjects := make([]types.Email, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandEmail(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandExternalID(tfMap map[string]any) *types.ExternalId {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ExternalId{}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap[names.AttrIssuer].(string); ok && v != "" {
		apiObject.Issuer = aws.String(v)
	}

	return apiObject
}

func flattenExternalID(apiObject *types.ExternalId) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Id; v != nil {
		tfMap[names.AttrID] = aws.ToString(v)
	}

	if v := apiObject.Issuer; v != nil {
		tfMap[names.AttrIssuer] = aws.ToString(v)
	}

	return tfMap
}

func flattenExternalIDs(apiObjects []types.ExternalId) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenExternalID(&apiObject))
	}

	return tfList
}

func flattenName(apiObject *types.Name) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.FamilyName; v != nil {
		tfMap["family_name"] = aws.ToString(v)
	}

	if v := apiObject.Formatted; v != nil {
		tfMap["formatted"] = aws.ToString(v)
	}

	if v := apiObject.GivenName; v != nil {
		tfMap["given_name"] = aws.ToString(v)
	}

	if v := apiObject.HonorificPrefix; v != nil {
		tfMap["honorific_prefix"] = aws.ToString(v)
	}

	if v := apiObject.HonorificSuffix; v != nil {
		tfMap["honorific_suffix"] = aws.ToString(v)
	}

	if v := apiObject.MiddleName; v != nil {
		tfMap["middle_name"] = aws.ToString(v)
	}

	return tfMap
}

func expandName(tfMap map[string]any) *types.Name {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Name{}

	if v, ok := tfMap["family_name"].(string); ok && v != "" {
		apiObject.FamilyName = aws.String(v)
	}

	if v, ok := tfMap["formatted"].(string); ok && v != "" {
		apiObject.Formatted = aws.String(v)
	}

	if v, ok := tfMap["given_name"].(string); ok && v != "" {
		apiObject.GivenName = aws.String(v)
	}

	if v, ok := tfMap["honorific_prefix"].(string); ok && v != "" {
		apiObject.HonorificPrefix = aws.String(v)
	}

	if v, ok := tfMap["honorific_suffix"].(string); ok && v != "" {
		apiObject.HonorificSuffix = aws.String(v)
	}

	if v, ok := tfMap["middle_name"].(string); ok && v != "" {
		apiObject.MiddleName = aws.String(v)
	}

	return apiObject
}

func flattenPhoneNumber(apiObject *types.PhoneNumber) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["primary"] = apiObject.Primary

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func expandPhoneNumber(tfMap map[string]any) *types.PhoneNumber {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PhoneNumber{}

	apiObject.Primary = tfMap["primary"].(bool)

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func flattenPhoneNumbers(apiObjects []types.PhoneNumber) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPhoneNumber(&apiObject))
	}

	return tfList
}

func expandPhoneNumbers(tfList []any) []types.PhoneNumber {
	apiObjects := make([]types.PhoneNumber, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandPhoneNumber(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandUniqueAttribute(tfMap map[string]any) *types.UniqueAttribute {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UniqueAttribute{}

	if v, ok := tfMap["attribute_path"].(string); ok && v != "" {
		apiObject.AttributePath = aws.String(v)
	}

	if v, ok := tfMap["attribute_value"].(string); ok && v != "" {
		apiObject.AttributeValue = document.NewLazyDocument(v)
	}

	return apiObject
}
