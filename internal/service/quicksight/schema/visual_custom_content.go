// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func customContentVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomContentVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_identifier": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
				"visual_id":           idSchema(),
				"actions":             visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomContentConfiguration.html
					Type:             schema.TypeList,
					Optional:         true,
					MinItems:         1,
					MaxItems:         1,
					DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"content_type":  stringSchema(false, enum.Validate[types.CustomContentType]()),
							"content_url":   stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
							"image_scaling": stringSchema(false, enum.Validate[types.CustomContentImageScalingConfiguration]()),
						},
					},
				},
				"subtitle": visualSubtitleLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
				"title":    visualTitleLabelOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
			},
		},
	}
}

func expandCustomContentVisual(tfList []interface{}) *types.CustomContentVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.CustomContentVisual{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		visual.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
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

func expandCustomContentConfiguration(tfList []interface{}) *types.CustomContentConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.CustomContentConfiguration{}

	if v, ok := tfMap["content_type"].(string); ok && v != "" {
		config.ContentType = types.CustomContentType(v)
	}
	if v, ok := tfMap["content_url"].(string); ok && v != "" {
		config.ContentUrl = aws.String(v)
	}
	if v, ok := tfMap["image_scaling"].(string); ok && v != "" {
		config.ImageScaling = types.CustomContentImageScalingConfiguration(v)
	}

	return config
}

func flattenCustomContentVisual(apiObject *types.CustomContentVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"data_set_identifier": aws.ToString(apiObject.DataSetIdentifier),
		"visual_id":           aws.ToString(apiObject.VisualId),
	}
	if apiObject.Actions != nil {
		tfMap["actions"] = flattenVisualCustomAction(apiObject.Actions)
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

func flattenCustomContentConfiguration(apiObject *types.CustomContentConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["content_type"] = types.CustomContentType(apiObject.ContentType)

	tfMap["content_url"] = aws.ToString(apiObject.ContentUrl)

	tfMap["image_scaling"] = types.CustomContentImageScalingConfiguration(apiObject.ImageScaling)

	return []interface{}{tfMap}
}
