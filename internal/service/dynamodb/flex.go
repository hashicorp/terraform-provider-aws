// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func expandTableItemAttributes(jsonStream string) (map[string]awstypes.AttributeValue, error) {
	var apiObject map[string]attributeValue
	dec := json.NewDecoder(strings.NewReader(jsonStream))

	for {
		if err := dec.Decode(&apiObject); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}

	return tfmaps.ApplyToAllValues(apiObject, attributeValueStructToInterface), nil
}

func flattenTableItemAttributes(apiObject map[string]awstypes.AttributeValue) (string, error) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)

	if err := enc.Encode(tfmaps.ApplyToAllValues(apiObject, attributeValueInterfaceToStruct)); err != nil {
		return "", err
	}

	return b.String(), nil
}

func attributeValueInterfaceToStruct(v awstypes.AttributeValue) attributeValue {
	return attributeValue{v}
}

func attributeValueStructToInterface(v attributeValue) awstypes.AttributeValue {
	return v.AttributeValue
}

type attributeValue struct {
	awstypes.AttributeValue
}

func (a attributeValue) MarshalJSON() ([]byte, error) {
	m := map[string]any{}

	switch v := a.AttributeValue.(type) {
	case *awstypes.AttributeValueMemberB:
		m["B"] = v.Value
	case *awstypes.AttributeValueMemberBOOL:
		m["BOOL"] = v.Value
	case *awstypes.AttributeValueMemberBS:
		m["BS"] = v.Value
	case *awstypes.AttributeValueMemberL:
		m["L"] = tfslices.ApplyToAll(v.Value, attributeValueInterfaceToStruct)
	case *awstypes.AttributeValueMemberM:
		m["M"] = tfmaps.ApplyToAllValues(v.Value, attributeValueInterfaceToStruct)
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

func (a *attributeValue) UnmarshalJSON(b []byte) error {
	m := map[string]any{}

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	if n := len(m); n != 1 {
		return fmt.Errorf("invalid map len: %d", n)
	}

	var intf awstypes.AttributeValue
	for k, v := range m {
		switch k {
		case "B":
			v, err := itypes.Base64Decode(v.(string))
			if err != nil {
				return err
			}
			intf = &awstypes.AttributeValueMemberB{Value: v}
		case "BOOL":
			intf = &awstypes.AttributeValueMemberBOOL{Value: v.(bool)}
		case "BS":
			v, err := tfslices.ApplyToAllWithError(v.([]any), func(v any) ([]byte, error) {
				return itypes.Base64Decode(v.(string))
			})
			if err != nil {
				return err
			}
			intf = &awstypes.AttributeValueMemberBS{Value: v}
		// case "L":
		// 	intf = &awstypes.AttributeValueMemberL{Value: tfslices.ApplyToAll(v.([]any), attributeValueInterfaceToStruct)}
		// case "M":
		// 	intf = &awstypes.AttributeValueMemberM{Value: tfmaps.ApplyToAllValues(v.(map[string]any), attributeValueInterfaceToStruct)}
		case "N":
			intf = &awstypes.AttributeValueMemberN{Value: v.(string)}
		case "NS":
			intf = &awstypes.AttributeValueMemberNS{Value: tfslices.ApplyToAll(v.([]any), func(v any) string {
				return v.(string)
			})}
		case "NULL":
			intf = &awstypes.AttributeValueMemberNULL{Value: v.(bool)}
		case "S":
			intf = &awstypes.AttributeValueMemberS{Value: v.(string)}
		case "SS":
			intf = &awstypes.AttributeValueMemberSS{Value: tfslices.ApplyToAll(v.([]any), func(v any) string {
				return v.(string)
			})}
		}
	}

	*a = attributeValueInterfaceToStruct(intf)

	return nil
}
