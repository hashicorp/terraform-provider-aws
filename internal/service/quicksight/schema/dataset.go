// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func dataSetIdentifierDeclarationsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetIdentifierDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 50,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_arn":       stringSchema(false, verify.ValidARN),
				names.AttrIdentifier: stringSchema(false, validation.StringLenBetween(1, 2048)),
			},
		},
	}
}

func dataSetReferencesSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetReference.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_arn": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"data_set_placeholder": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func expandDataSetIdentifierDeclarations(tfList []interface{}) []*quicksight.DataSetIdentifierDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var identifiers []*quicksight.DataSetIdentifierDeclaration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		identifier := expandDataSetIdentifierDeclaration(tfMap)
		if identifier == nil {
			continue
		}

		identifiers = append(identifiers, identifier)
	}

	return identifiers
}

func expandDataSetIdentifierDeclaration(tfMap map[string]interface{}) *quicksight.DataSetIdentifierDeclaration {
	if tfMap == nil {
		return nil
	}

	identifier := &quicksight.DataSetIdentifierDeclaration{}

	if v, ok := tfMap["data_set_arn"].(string); ok && v != "" {
		identifier.DataSetArn = aws.String(v)
	}
	if v, ok := tfMap[names.AttrIdentifier].(string); ok && v != "" {
		identifier.Identifier = aws.String(v)
	}

	return identifier
}

func flattenDataSetIdentifierDeclarations(apiObject []*quicksight.DataSetIdentifierDeclaration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, identifier := range apiObject {
		if identifier == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if identifier.DataSetArn != nil {
			tfMap["data_set_arn"] = aws.StringValue(identifier.DataSetArn)
		}
		if identifier.Identifier != nil {
			tfMap[names.AttrIdentifier] = aws.StringValue(identifier.Identifier)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}
