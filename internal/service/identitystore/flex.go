package identitystore

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/document"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
)

func expandAlternateIdentifier(tfMap map[string]interface{}) types.AlternateIdentifier {
	if tfMap == nil {
		return nil
	}

	if v, ok := tfMap["external_id"]; ok && len(v.([]interface{})) > 0 {
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

func expandExternalId(tfMap map[string]interface{}) *types.ExternalId {
	if tfMap == nil {
		return nil
	}

	a := &types.ExternalId{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		a.Id = aws.String(v)
	}

	if v, ok := tfMap["issuer"].(string); ok && v != "" {
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
		m["id"] = aws.ToString(v)
	}

	if v := apiObject.Issuer; v != nil {
		m["issuer"] = aws.ToString(v)
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
