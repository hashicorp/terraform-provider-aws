// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				"data_set_identifier": stringLenBetweenSchema(attrRequired, 1, 2048),
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
							names.AttrContentType: stringEnumSchema[awstypes.CustomContentType](attrOptional),
							"content_url":         stringLenBetweenSchema(attrOptional, 1, 2048),
							"image_scaling":       stringEnumSchema[awstypes.CustomContentImageScalingConfiguration](attrOptional),
						},
					},
				},
				"subtitle": visualSubtitleLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
				"title":    visualTitleLabelOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
			},
		},
	}
}

func expandCustomContentVisual(tfList []any) *awstypes.CustomContentVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomContentVisual{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		apiObject.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandCustomContentConfiguration(v)
	}
	if v, ok := tfMap["subtitle"].([]any); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandCustomContentConfiguration(tfList []any) *awstypes.CustomContentConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomContentConfiguration{}

	if v, ok := tfMap[names.AttrContentType].(string); ok && v != "" {
		apiObject.ContentType = awstypes.CustomContentType(v)
	}
	if v, ok := tfMap["content_url"].(string); ok && v != "" {
		apiObject.ContentUrl = aws.String(v)
	}
	if v, ok := tfMap["image_scaling"].(string); ok && v != "" {
		apiObject.ImageScaling = awstypes.CustomContentImageScalingConfiguration(v)
	}

	return apiObject
}

func flattenCustomContentVisual(apiObject *awstypes.CustomContentVisual) []any {
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
	if apiObject.ChartConfiguration != nil {
		tfMap["chart_configuration"] = flattenCustomContentConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []any{tfMap}
}

func flattenCustomContentConfiguration(apiObject *awstypes.CustomContentConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap[names.AttrContentType] = apiObject.ContentType
	if apiObject.ContentUrl != nil {
		tfMap["content_url"] = aws.ToString(apiObject.ContentUrl)
	}
	tfMap["image_scaling"] = apiObject.ImageScaling

	return []any{tfMap}
}
