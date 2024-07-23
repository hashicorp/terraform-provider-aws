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

func customContentVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomContentVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_identifier": stringSchema(true, validation.StringLenBetween(1, 2048)),
				"visual_id":           idSchema(),
				names.AttrActions:     visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomContentConfiguration.html
					Type:             schema.TypeList,
					Optional:         true,
					MinItems:         1,
					MaxItems:         1,
					DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrContentType: stringSchema(false, validation.StringInSlice(quicksight.CustomContentType_Values(), false)),
							"content_url":         stringSchema(false, validation.StringLenBetween(1, 2048)),
							"image_scaling":       stringSchema(false, validation.StringInSlice(quicksight.CustomContentImageScalingConfiguration_Values(), false)),
						},
					},
				},
				"subtitle": visualSubtitleLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
				"title":    visualTitleLabelOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
			},
		},
	}
}

func expandCustomContentVisual(tfList []interface{}) *quicksight.CustomContentVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.CustomContentVisual{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		visual.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandCustomContentConfiguration(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		visual.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		visual.Title = expandVisualTitleLabelOptions(v)
	}

	return visual
}

func expandCustomContentConfiguration(tfList []interface{}) *quicksight.CustomContentConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.CustomContentConfiguration{}

	if v, ok := tfMap[names.AttrContentType].(string); ok && v != "" {
		config.ContentType = aws.String(v)
	}
	if v, ok := tfMap["content_url"].(string); ok && v != "" {
		config.ContentUrl = aws.String(v)
	}
	if v, ok := tfMap["image_scaling"].(string); ok && v != "" {
		config.ImageScaling = aws.String(v)
	}

	return config
}

func flattenCustomContentVisual(apiObject *quicksight.CustomContentVisual) []interface{} {
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
	if apiObject.ChartConfiguration != nil {
		tfMap["chart_configuration"] = flattenCustomContentConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []interface{}{tfMap}
}

func flattenCustomContentConfiguration(apiObject *quicksight.CustomContentConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ContentType != nil {
		tfMap[names.AttrContentType] = aws.StringValue(apiObject.ContentType)
	}
	if apiObject.ContentUrl != nil {
		tfMap["content_url"] = aws.StringValue(apiObject.ContentUrl)
	}
	if apiObject.ImageScaling != nil {
		tfMap["image_scaling"] = aws.StringValue(apiObject.ImageScaling)
	}

	return []interface{}{tfMap}
}
