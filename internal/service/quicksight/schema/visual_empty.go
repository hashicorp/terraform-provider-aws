// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				"data_set_identifier": stringLenBetweenSchema(attrRequired, 1, 2048),
				"visual_id":           idSchema(),
				names.AttrActions:     visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
			},
		},
	}
}

func expandEmptyVisual(tfList []any) *awstypes.EmptyVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.EmptyVisual{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		apiObject.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}

	return apiObject
}

func flattenEmptyVisual(apiObject *awstypes.EmptyVisual) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"data_set_identifier": aws.ToString(apiObject.DataSetIdentifier),
		"visual_id":           aws.ToString(apiObject.VisualId),
	}

	if apiObject.Actions != nil {
		tfMap[names.AttrActions] = flattenVisualCustomAction(apiObject.Actions)
	}

	return []any{tfMap}
}
