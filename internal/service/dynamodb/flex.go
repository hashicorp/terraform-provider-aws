// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"bytes"
	"encoding/json"
	"strings"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

func expandTableItemAttributes(tfString string) (map[string]awstypes.AttributeValue, error) {
	var apiObject map[string]awstypes.AttributeValue

	err := json.NewDecoder(strings.NewReader(tfString)).Decode(&apiObject)
	if err != nil {
		return nil, err
	}

	return apiObject, nil
}

func flattenTableItemAttributes(apiObject map[string]awstypes.AttributeValue) (string, error) {
	buf := bytes.NewBufferString("")
	encoder := json.NewEncoder(buf)

	err := encoder.Encode(tfmaps.ApplyToAllValues(apiObject, func(v awstypes.AttributeValue) attributeValue {
		return attributeValue{v}
	}))

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type attributeValue struct {
	awstypes.AttributeValue
}

func (v attributeValue) MarshalJSON() ([]byte, error) {
	m := map[string]any{}

	switch v := v.AttributeValue.(type) {
	case *awstypes.AttributeValueMemberB:
		m["B"] = v.Value
	case *awstypes.AttributeValueMemberBOOL:
		m["BOOL"] = v.Value
	case *awstypes.AttributeValueMemberBS:
		m["BS"] = v.Value
	case *awstypes.AttributeValueMemberL:
		m["L"] = tfslices.ApplyToAll(v.Value, func(v awstypes.AttributeValue) attributeValue {
			return attributeValue{v}
		})
	case *awstypes.AttributeValueMemberM:
		m["M"] = tfmaps.ApplyToAllValues(v.Value, func(v awstypes.AttributeValue) attributeValue {
			return attributeValue{v}
		})
	case *awstypes.AttributeValueMemberN:
		m["N"] = v.Value
	case *awstypes.AttributeValueMemberNS:
		m["NS"] = v.Value
	case *awstypes.AttributeValueMemberNULL:
		m["NULL"] = v.Value
	case *awstypes.AttributeValueMemberS:
		m["S"] = v.Value
	case *awstypes.AttributeValueMemberSS:
		m["SS"] = v.Value
	}

	return json.Marshal(m)
}
