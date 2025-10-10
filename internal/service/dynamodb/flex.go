// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"fmt"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func expandTableItemAttributes(jsonStream string) (map[string]awstypes.AttributeValue, error) {
	var m map[string]any

	err := tfjson.DecodeFromString(jsonStream, &m)
	if err != nil {
		return nil, err
	}

	return tfmaps.ApplyToAllValuesWithError(m, attributeFromRaw)
}

func flattenTableItemAttributes(apiObject map[string]awstypes.AttributeValue) (string, error) {
	m, err := tfmaps.ApplyToAllValuesWithError(apiObject, rawFromAttribute)
	if err != nil {
		return "", err
	}

	return tfjson.EncodeToString(m)
}

func attributeFromRaw(v any) (awstypes.AttributeValue, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected raw attribute type: %T", v)
	}

	if n := len(m); n != 1 {
		return nil, fmt.Errorf("invalid raw attribute map len: %d", n)
	}

	for k, v := range m {
		switch v := v.(type) {
		case bool:
			switch k {
			case dataTypeDescriptorBoolean:
				return &awstypes.AttributeValueMemberBOOL{Value: v}, nil
			case dataTypeDescriptorNull:
				return &awstypes.AttributeValueMemberNULL{Value: v}, nil
			}
		case string:
			switch k {
			case dataTypeDescriptorBinary:
				v, err := itypes.Base64Decode(v)
				if err != nil {
					return nil, err
				}
				return &awstypes.AttributeValueMemberB{Value: v}, nil
			case dataTypeDescriptorNumber:
				return &awstypes.AttributeValueMemberN{Value: v}, nil
			case dataTypeDescriptorString:
				return &awstypes.AttributeValueMemberS{Value: v}, nil
			}
		case []any:
			switch k {
			case dataTypeDescriptorBinarySet:
				v, err := tfslices.ApplyToAllWithError(v, func(v any) ([]byte, error) {
					switch v := v.(type) {
					case string:
						return itypes.Base64Decode(v)
					default:
						return nil, unexpectedRawAttributeElementTypeError(v, k)
					}
				})
				if err != nil {
					return nil, err
				}
				return &awstypes.AttributeValueMemberBS{Value: v}, nil
			case dataTypeDescriptorList:
				v, err := tfslices.ApplyToAllWithError(v, attributeFromRaw)
				if err != nil {
					return nil, err
				}
				return &awstypes.AttributeValueMemberL{Value: v}, nil
			case dataTypeDescriptorNumberSet, dataTypeDescriptorStringSet:
				v, err := tfslices.ApplyToAllWithError(v, func(v any) (string, error) {
					switch v := v.(type) {
					case string:
						return v, nil
					default:
						return "", unexpectedRawAttributeElementTypeError(v, k)
					}
				})
				if err != nil {
					return nil, err
				}
				if k == dataTypeDescriptorNumberSet {
					return &awstypes.AttributeValueMemberNS{Value: v}, nil
				}
				return &awstypes.AttributeValueMemberSS{Value: v}, nil
			}
		case map[string]any:
			switch k {
			case dataTypeDescriptorMap:
				v, err := tfmaps.ApplyToAllValuesWithError(v, attributeFromRaw)
				if err != nil {
					return nil, err
				}
				return &awstypes.AttributeValueMemberM{Value: v}, nil
			}
		}

		return nil, fmt.Errorf("unexpected raw attribute type (%T) for data type descriptor: %s", v, k)
	}

	panic("unreachable") //lintignore:R009
}

func rawFromAttribute(a awstypes.AttributeValue) (any, error) {
	m := map[string]any{}

	switch a := a.(type) {
	case *awstypes.AttributeValueMemberB:
		m[dataTypeDescriptorBinary] = itypes.Base64Encode(a.Value)
	case *awstypes.AttributeValueMemberBOOL:
		m[dataTypeDescriptorBoolean] = a.Value
	case *awstypes.AttributeValueMemberBS:
		m[dataTypeDescriptorBinarySet] = tfslices.ApplyToAll(a.Value, itypes.Base64Encode)
	case *awstypes.AttributeValueMemberL:
		v, err := tfslices.ApplyToAllWithError(a.Value, rawFromAttribute)
		if err != nil {
			return nil, err
		}
		m[dataTypeDescriptorList] = v
	case *awstypes.AttributeValueMemberM:
		v, err := tfmaps.ApplyToAllValuesWithError(a.Value, rawFromAttribute)
		if err != nil {
			return nil, err
		}
		m[dataTypeDescriptorMap] = v
	case *awstypes.AttributeValueMemberN:
		m[dataTypeDescriptorNumber] = a.Value
	case *awstypes.AttributeValueMemberNS:
		m[dataTypeDescriptorNumberSet] = a.Value
	case *awstypes.AttributeValueMemberNULL:
		m[dataTypeDescriptorNull] = a.Value
	case *awstypes.AttributeValueMemberS:
		m[dataTypeDescriptorString] = a.Value
	case *awstypes.AttributeValueMemberSS:
		m[dataTypeDescriptorStringSet] = a.Value
	default:
		return nil, fmt.Errorf("unexpected attribute type: %T", a)
	}

	return m, nil
}

// See https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.NamingRulesDataTypes.html#HowItWorks.DataTypes.
const (
	dataTypeDescriptorBinary    = "B"
	dataTypeDescriptorBinarySet = "BS"
	dataTypeDescriptorBoolean   = "BOOL"
	dataTypeDescriptorList      = "L"
	dataTypeDescriptorMap       = "M"
	dataTypeDescriptorNull      = "NULL"
	dataTypeDescriptorNumber    = "N"
	dataTypeDescriptorNumberSet = "NS"
	dataTypeDescriptorString    = "S"
	dataTypeDescriptorStringSet = "SS"
)

func unexpectedRawAttributeElementTypeError(v any, k string) error {
	return fmt.Errorf("unexpected raw attribute element type (%T) for data type descriptor: %s", v, k)
}
