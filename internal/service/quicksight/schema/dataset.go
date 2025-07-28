// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var dataSetIdentifierDeclarationsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetIdentifierDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 50,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_arn":       arnStringSchema(attrOptional),
				names.AttrIdentifier: stringLenBetweenSchema(attrOptional, 1, 2048),
			},
		},
	}
})

var dataSetReferencesSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetReference.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_arn": arnStringSchema(attrRequired),
				"data_set_placeholder": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
})

func expandDataSetIdentifierDeclarations(tfList []any) []awstypes.DataSetIdentifierDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DataSetIdentifierDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandDataSetIdentifierDeclaration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDataSetIdentifierDeclaration(tfMap map[string]any) *awstypes.DataSetIdentifierDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataSetIdentifierDeclaration{}

	if v, ok := tfMap["data_set_arn"].(string); ok && v != "" {
		apiObject.DataSetArn = aws.String(v)
	}
	if v, ok := tfMap[names.AttrIdentifier].(string); ok && v != "" {
		apiObject.Identifier = aws.String(v)
	}

	return apiObject
}

func flattenDataSetIdentifierDeclarations(apiObject []awstypes.DataSetIdentifierDeclaration) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObject {
		tfMap := map[string]any{}

		if apiObject.DataSetArn != nil {
			tfMap["data_set_arn"] = aws.ToString(apiObject.DataSetArn)
		}
		if apiObject.Identifier != nil {
			tfMap[names.AttrIdentifier] = aws.ToString(apiObject.Identifier)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
