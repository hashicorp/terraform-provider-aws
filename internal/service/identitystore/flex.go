package identitystore

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/document"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
)

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
