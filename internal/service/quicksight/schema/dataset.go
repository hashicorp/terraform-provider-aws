// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func dataSetIdentifierDeclarationsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetIdentifierDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 50,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_arn": stringSchema(false, validation.ToDiagFunc(verify.ValidARN)),
				"identifier":   stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
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
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
				},
				"data_set_placeholder": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func expandDataSetIdentifierDeclarations(tfList []interface{}) []types.DataSetIdentifierDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var identifiers []types.DataSetIdentifierDeclaration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		identifier := expandDataSetIdentifierDeclaration(tfMap)
		if identifier == nil {
			continue
		}

		identifiers = append(identifiers, *identifier)
	}

	return identifiers
}

func expandDataSetIdentifierDeclaration(tfMap map[string]interface{}) *types.DataSetIdentifierDeclaration {
	if tfMap == nil {
		return nil
	}

	identifier := &types.DataSetIdentifierDeclaration{}

	if v, ok := tfMap["data_set_arn"].(string); ok && v != "" {
		identifier.DataSetArn = aws.String(v)
	}
	if v, ok := tfMap["identifier"].(string); ok && v != "" {
		identifier.Identifier = aws.String(v)
	}

	return identifier
}

func flattenDataSetIdentifierDeclarations(apiObject []types.DataSetIdentifierDeclaration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, identifier := range apiObject {

		tfMap := map[string]interface{}{}
		if identifier.DataSetArn != nil {
			tfMap["data_set_arn"] = aws.ToString(identifier.DataSetArn)
		}
		if identifier.Identifier != nil {
			tfMap["identifier"] = aws.ToString(identifier.Identifier)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}
