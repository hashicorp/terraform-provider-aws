// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func emptyVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_EmptyVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_identifier": stringSchema(true, validation.StringLenBetween(1, 2048)),
				"visual_id":           idSchema(),
				names.AttrActions:     visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
			},
		},
	}
}

func expandEmptyVisual(tfList []interface{}) *quicksight.EmptyVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.EmptyVisual{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		visual.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}

	return visual
}

func flattenEmptyVisual(apiObject *quicksight.EmptyVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"data_set_identifier": aws.StringValue(apiObject.DataSetIdentifier),
		"visual_id":           aws.StringValue(apiObject.VisualId),
	}
	if apiObject.Actions != nil {
		tfMap[names.AttrActions] = flattenVisualCustomAction(apiObject.Actions)
	}

	return []interface{}{tfMap}
}
