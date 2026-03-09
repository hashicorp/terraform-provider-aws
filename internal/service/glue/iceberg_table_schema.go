// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func createIcebergTableInputSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrLocation: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 2056),
				},
				"partition_spec": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"fields": {
								Type:     schema.TypeList,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"field_id": {
											Type:     schema.TypeInt,
											Optional: true,
										},
										names.AttrName: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 1024),
										},
										"source_id": {
											Type:     schema.TypeInt,
											Required: true,
										},
										"transform": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"spec_id": {
								Type:     schema.TypeInt,
								Optional: true,
							},
						},
					},
				},
				names.AttrProperties: {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"schema": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"fields": {
								Type:     schema.TypeList,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"doc": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(0, 255),
										},
										names.AttrID: {
											Type:     schema.TypeInt,
											Required: true,
										},
										"initial_default": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateFunc:     validation.StringIsJSON,
											DiffSuppressFunc: flex.SuppressEquivalentJSONDiffs,
										},
										names.AttrName: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 1024),
										},
										names.AttrRequired: {
											Type:     schema.TypeBool,
											Required: true,
										},
										names.AttrType: {
											Type:             schema.TypeString,
											Required:         true,
											ValidateFunc:     validation.StringIsJSON,
											DiffSuppressFunc: flex.SuppressEquivalentJSONDiffs,
										},
										"write_default": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateFunc:     validation.StringIsJSON,
											DiffSuppressFunc: flex.SuppressEquivalentJSONDiffs,
										},
									},
								},
							},
							"identifier_field_ids": {
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeInt},
							},
							"schema_id": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							names.AttrType: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringInSlice([]string{"struct"}, false),
							},
						},
					},
				},
				"write_order": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"fields": {
								Type:     schema.TypeList,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"direction": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
										},
										"null_order": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice([]string{"nulls-first", "nulls-last"}, false),
										},
										"source_id": {
											Type:     schema.TypeInt,
											Required: true,
										},
										"transform": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"order_id": {
								Type:     schema.TypeInt,
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

func expandCreateIcebergTableInput(tfList []any) *awstypes.CreateIcebergTableInput {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.CreateIcebergTableInput{}

	if v, ok := tfMap[names.AttrLocation].(string); ok && v != "" {
		apiObject.Location = aws.String(v)
	}

	if v, ok := tfMap["partition_spec"].([]any); ok && len(v) > 0 {
		apiObject.PartitionSpec = expandIcebergPartitionSpec(v)
	}

	if v, ok := tfMap[names.AttrProperties].(map[string]any); ok && len(v) > 0 {
		apiObject.Properties = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["schema"].([]any); ok && len(v) > 0 {
		apiObject.Schema = expandIcebergSchema(v)
	}

	if v, ok := tfMap["write_order"].([]any); ok && len(v) > 0 {
		apiObject.WriteOrder = expandIcebergSortOrder(v)
	}

	return apiObject
}

func expandIcebergSchema(tfList []any) *awstypes.IcebergSchema {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.IcebergSchema{}

	if v, ok := tfMap["fields"].([]any); ok && len(v) > 0 {
		apiObject.Fields = expandIcebergStructFields(v)
	}

	if v, ok := tfMap["identifier_field_ids"].([]any); ok && len(v) > 0 {
		apiObject.IdentifierFieldIds = flex.ExpandInt32ValueList(v)
	}

	if v, ok := tfMap["schema_id"].(int); ok {
		apiObject.SchemaId = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandIcebergStructFields(tfList []any) []awstypes.IcebergStructField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.IcebergStructField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.IcebergStructField{}

		if v, ok := tfMap["doc"].(string); ok && v != "" {
			apiObject.Doc = aws.String(v)
		}

		if v, ok := tfMap[names.AttrID].(int); ok {
			apiObject.Id = aws.Int32(int32(v))
		}

		if v, ok := tfMap["initial_default"].(string); ok && v != "" {
			apiObject.InitialDefault = aws.String(v)
		}

		if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
			apiObject.Name = aws.String(v)
		}

		if v, ok := tfMap[names.AttrRequired].(bool); ok {
			apiObject.Required = aws.Bool(v)
		}

		if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
			apiObject.Type = aws.String(v)
		}

		if v, ok := tfMap["write_default"].(string); ok && v != "" {
			apiObject.WriteDefault = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandIcebergPartitionSpec(tfList []any) *awstypes.IcebergPartitionSpec {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.IcebergPartitionSpec{}

	if v, ok := tfMap["fields"].([]any); ok && len(v) > 0 {
		apiObject.Fields = expandIcebergPartitionFields(v)
	}

	if v, ok := tfMap["spec_id"].(int); ok {
		apiObject.SpecId = aws.Int32(int32(v))
	}

	return apiObject
}

func expandIcebergPartitionFields(tfList []any) []awstypes.IcebergPartitionField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.IcebergPartitionField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.IcebergPartitionField{}

		if v, ok := tfMap["field_id"].(int); ok {
			apiObject.FieldId = aws.Int32(int32(v))
		}

		if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
			apiObject.Name = aws.String(v)
		}

		if v, ok := tfMap["source_id"].(int); ok {
			apiObject.SourceId = aws.Int32(int32(v))
		}

		if v, ok := tfMap["transform"].(string); ok && v != "" {
			apiObject.Transform = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandIcebergSortOrder(tfList []any) *awstypes.IcebergSortOrder {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.IcebergSortOrder{}

	if v, ok := tfMap["fields"].([]any); ok && len(v) > 0 {
		apiObject.Fields = expandIcebergSortFields(v)
	}

	if v, ok := tfMap["order_id"].(int); ok {
		apiObject.OrderId = aws.Int32(int32(v))
	}

	return apiObject
}

func expandIcebergSortFields(tfList []any) []awstypes.IcebergSortField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.IcebergSortField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.IcebergSortField{}

		if v, ok := tfMap["direction"].(string); ok && v != "" {
			apiObject.Direction = aws.String(v)
		}

		if v, ok := tfMap["null_order"].(string); ok && v != "" {
			apiObject.NullOrder = aws.String(v)
		}

		if v, ok := tfMap["source_id"].(int); ok {
			apiObject.SourceId = aws.Int32(int32(v))
		}

		if v, ok := tfMap["transform"].(string); ok && v != "" {
			apiObject.Transform = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
