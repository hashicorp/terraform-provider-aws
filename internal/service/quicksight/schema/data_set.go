// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSetColumnGroupsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 8,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"geo_spatial_column_group": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"columns": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 16,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringLenBetween(1, 128),
								},
							},
							"country_code": stringEnumSchema[awstypes.GeoSpatialCountryCode](attrRequired),
							names.AttrName: stringLenBetweenSchema(attrRequired, 1, 64),
						},
					},
				},
			},
		},
	}
}

func DataSetColumnGroupsSchemaDataSourceSchema() *schema.Schema {
	return sdkv2.DataSourcePropertyFromResourceProperty(DataSetColumnGroupsSchema())
}

func DataSetColumnLevelPermissionRulesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column_names": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"principals": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 100,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func DataSetColumnLevelPermissionRulesSchemaDataSourceSchema() *schema.Schema {
	return sdkv2.DataSourcePropertyFromResourceProperty(DataSetColumnLevelPermissionRulesSchema())
}

func DataSetUsageConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disable_use_as_direct_query_source": {
					Type:     schema.TypeBool,
					Computed: true,
					Optional: true,
				},
				"disable_use_as_imported_source": {
					Type:     schema.TypeBool,
					Computed: true,
					Optional: true,
				},
			},
		},
	}
}

func DataSetUsageConfigurationSchemaDataSourceSchema() *schema.Schema {
	return sdkv2.DataSourcePropertyFromResourceProperty(DataSetUsageConfigurationSchema())
}

func DataSetFieldFoldersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1000,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_folders_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				"columns": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 5000,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				names.AttrDescription: stringLenBetweenSchema(attrOptional, 0, 500),
			},
		},
	}
}

func DataSetFieldFoldersSchemaDataSourceSchema() *schema.Schema {
	return sdkv2.DataSourcePropertyFromResourceProperty(DataSetFieldFoldersSchema())
}

func DataSetLogicalTableMapSchema() *schema.Schema {
	logicalTableMapSchema := func() *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrAlias: stringLenBetweenSchema(attrRequired, 1, 64),
				"data_transforms": {
					Type:     schema.TypeList,
					Computed: true,
					Optional: true,
					MinItems: 1,
					MaxItems: 2048,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cast_column_type_operation": {
								Type:     schema.TypeList,
								Computed: true,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"column_name":     stringLenBetweenSchema(attrRequired, 1, 128),
										names.AttrFormat:  stringLenBetweenSchema(attrOptionalComputed, 0, 32),
										"new_column_type": stringEnumSchema[awstypes.ColumnDataType](attrRequired),
									},
								},
							},
							"create_columns_operation": {
								Type:     schema.TypeList,
								Computed: true,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"columns": {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 128,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"column_id":          stringLenBetweenSchema(attrRequired, 1, 64),
													"column_name":        stringLenBetweenSchema(attrRequired, 1, 128),
													names.AttrExpression: stringLenBetweenSchema(attrRequired, 1, 4096),
												},
											},
										},
									},
								},
							},
							"filter_operation": {
								Type:     schema.TypeList,
								Computed: true,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"condition_expression": stringLenBetweenSchema(attrRequired, 1, 4096),
									},
								},
							},
							"project_operation": {
								Type:     schema.TypeList,
								Computed: true,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"projected_columns": {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 2000,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
									},
								},
							},
							"rename_column_operation": {
								Type:     schema.TypeList,
								Computed: true,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"column_name":     stringLenBetweenSchema(attrRequired, 1, 128),
										"new_column_name": stringLenBetweenSchema(attrRequired, 1, 128),
									},
								},
							},
							"tag_column_operation": {
								Type:     schema.TypeList,
								Computed: true,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"column_name": stringLenBetweenSchema(attrRequired, 1, 128),
										names.AttrTags: {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 16,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"column_description": {
														Type:     schema.TypeList,
														Computed: true,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"text": stringLenBetweenSchema(attrOptionalComputed, 0, 500),
															},
														},
													},
													"column_geographic_role": stringEnumSchema[awstypes.GeoSpatialDataRole](attrOptionalComputed),
												},
											},
										},
									},
								},
							},
							"untag_column_operation": {
								Type:     schema.TypeList,
								Computed: true,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"column_name": stringLenBetweenSchema(attrRequired, 1, 128),
										"tag_names": {
											Type:     schema.TypeList,
											Required: true,
											Elem:     stringEnumSchema[awstypes.ColumnTagName](attrElem),
										},
									},
								},
							},
						},
					},
				},
				"logical_table_map_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrSource: {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_set_arn": {
								Type:     schema.TypeString,
								Computed: true,
								Optional: true,
							},
							"join_instruction": {
								Type:     schema.TypeList,
								Computed: true,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"left_join_key_properties": {
											Type:     schema.TypeList,
											Computed: true,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"unique_key": {
														Type:     schema.TypeBool,
														Computed: true,
														Optional: true,
													},
												},
											},
										},
										"left_operand": stringLenBetweenSchema(attrRequired, 1, 64),
										"on_clause":    stringLenBetweenSchema(attrRequired, 1, 512),
										"right_join_key_properties": {
											Type:     schema.TypeList,
											Computed: true,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"unique_key": {
														Type:     schema.TypeBool,
														Computed: true,
														Optional: true,
													},
												},
											},
										},
										"right_operand": stringLenBetweenSchema(attrRequired, 1, 64),
										names.AttrType:  stringEnumSchema[awstypes.JoinType](attrRequired),
									},
								},
							},
							"physical_table_id": stringLenBetweenSchema(attrOptionalComputed, 1, 64),
						},
					},
				},
			},
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		MaxItems: 64,
		Elem:     logicalTableMapSchema(),
	}
}

func DataSetLogicalTableMapSchemaDataSourceSchema() *schema.Schema {
	return sdkv2.DataSourcePropertyFromResourceProperty(DataSetLogicalTableMapSchema())
}

func DataSetOutputColumnsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrDescription: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrType: {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func DataSetPhysicalTableMapSchema() *schema.Schema {
	physicalTableMapSchema := func() *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_sql": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"columns": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 2048,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: stringLenBetweenSchema(attrRequired, 1, 128),
										names.AttrType: stringEnumSchema[awstypes.InputColumnDataType](attrRequired),
									},
								},
							},
							"data_source_arn": arnStringSchema(attrRequired),
							names.AttrName:    stringLenBetweenSchema(attrRequired, 1, 64),
							"sql_query":       stringLenBetweenSchema(attrRequired, 1, 65536),
						},
					},
				},
				"physical_table_map_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				"relational_table": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"catalog":         stringLenBetweenSchema(attrOptional, 0, 256),
							"data_source_arn": arnStringSchema(attrRequired),
							"input_columns": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 2048,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: stringLenBetweenSchema(attrRequired, 1, 128),
										names.AttrType: stringEnumSchema[awstypes.InputColumnDataType](attrRequired),
									},
								},
							},
							names.AttrName: stringLenBetweenSchema(attrRequired, 1, 64),
							names.AttrSchema: {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"s3_source": {
					Type:     schema.TypeList,
					Computed: true,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_source_arn": arnStringSchema(attrRequired),
							"input_columns": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 2048,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: stringLenBetweenSchema(attrRequired, 1, 128),
										names.AttrType: stringEnumSchema[awstypes.InputColumnDataType](attrRequired),
									},
								},
							},
							"upload_settings": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"contains_header": {
											Type:     schema.TypeBool,
											Computed: true,
											Optional: true,
										},
										"delimiter":      stringLenBetweenSchema(attrOptionalComputed, 1, 1),
										names.AttrFormat: stringEnumSchema[awstypes.FileFormat](attrOptionalComputed),
										"start_from_row": {
											Type:         schema.TypeInt,
											Computed:     true,
											Optional:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
										"text_qualifier": stringEnumSchema[awstypes.TextQualifier](attrOptionalComputed),
									},
								},
							},
						},
					},
				},
			},
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 32,
		Elem:     physicalTableMapSchema(),
	}
}

func DataSetPhysicalTableMapSchemaDataSourceSchema() *schema.Schema {
	return sdkv2.DataSourcePropertyFromResourceProperty(DataSetPhysicalTableMapSchema())
}

func DataSetRowLevelPermissionDataSetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrARN:       arnStringSchema(attrRequired),
				"format_version":    stringEnumSchema[awstypes.RowLevelPermissionFormatVersion](attrOptional),
				names.AttrNamespace: stringLenBetweenSchema(attrOptional, 0, 64),
				"permission_policy": stringEnumSchema[awstypes.RowLevelPermissionPolicy](attrRequired),
				names.AttrStatus:    stringEnumSchema[awstypes.Status](attrOptional),
			},
		},
	}
}

func DataSetRowLevelPermissionDataSetSchemaDataSourceSchema() *schema.Schema {
	return sdkv2.DataSourcePropertyFromResourceProperty(DataSetRowLevelPermissionDataSetSchema())
}

func DataSetRowLevelPermissionTagConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrStatus: stringEnumSchema[awstypes.Status](attrOptional),
				"tag_rules": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 50,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"column_name": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"match_all_value":           stringLenBetweenSchema(attrOptional, 1, 256),
							"tag_key":                   stringLenBetweenSchema(attrRequired, 1, 128),
							"tag_multi_value_delimiter": stringLenBetweenSchema(attrOptional, 1, 10),
						},
					},
				},
			},
		},
	}
}

func DataSetRowLevelPermissionTagConfigurationSchemaDataSourceSchema() *schema.Schema {
	return sdkv2.DataSourcePropertyFromResourceProperty(DataSetRowLevelPermissionTagConfigurationSchema())
}

func DataSetRefreshPropertiesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"refresh_configuration": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"incremental_refresh": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"lookback_window": {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"column_name": {
														Type:     schema.TypeString,
														Required: true,
													},
													names.AttrSize: {
														Type:     schema.TypeInt,
														Required: true,
													},
													"size_unit": stringEnumSchema[awstypes.LookbackWindowSizeUnit](attrRequired),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func ExpandColumnGroups(tfList []any) []awstypes.ColumnGroup {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandColumnGroup(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandColumnGroup(tfMap map[string]any) *awstypes.ColumnGroup {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &awstypes.ColumnGroup{}

	if tfMapRaw, ok := tfMap["geo_spatial_column_group"].([]any); ok {
		apiObject.GeoSpatialColumnGroup = expandGeoSpatialColumnGroup(tfMapRaw[0].(map[string]any))
	}

	return apiObject
}

func expandGeoSpatialColumnGroup(tfMap map[string]any) *awstypes.GeoSpatialColumnGroup {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.GeoSpatialColumnGroup{}

	if v, ok := tfMap["columns"].([]any); ok {
		apiObject.Columns = flex.ExpandStringValueList(v)
	}
	if v, ok := tfMap["country_code"].(string); ok && v != "" {
		apiObject.CountryCode = awstypes.GeoSpatialCountryCode(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func ExpandColumnLevelPermissionRules(tfList []any) []awstypes.ColumnLevelPermissionRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnLevelPermissionRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.ColumnLevelPermissionRule{}

		if v, ok := tfMap["column_names"].([]any); ok {
			apiObject.ColumnNames = flex.ExpandStringValueList(v)
		}
		if v, ok := tfMap["principals"].([]any); ok {
			apiObject.Principals = flex.ExpandStringValueList(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func ExpandDataSetUsageConfiguration(tfList []any) *awstypes.DataSetUsageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataSetUsageConfiguration{}

	if v, ok := tfMap["disable_use_as_direct_query_source"].(bool); ok {
		apiObject.DisableUseAsDirectQuerySource = v
	}
	if v, ok := tfMap["disable_use_as_imported_source"].(bool); ok {
		apiObject.DisableUseAsImportedSource = v
	}

	return apiObject
}

func ExpandFieldFolders(tfList []any) map[string]awstypes.FieldFolder {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]awstypes.FieldFolder)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.FieldFolder{}

		if v, ok := tfMap["columns"].([]any); ok && len(v) > 0 {
			apiObject.Columns = flex.ExpandStringValueList(v)
		}
		if v, ok := tfMap[names.AttrDescription].(string); ok {
			apiObject.Description = aws.String(v)
		}

		apiObjects[tfMap["field_folders_id"].(string)] = apiObject
	}

	return apiObjects
}

func ExpandLogicalTableMap(tfList []any) map[string]awstypes.LogicalTable {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]awstypes.LogicalTable)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.LogicalTable{}

		if v, ok := tfMap[names.AttrAlias].(string); ok {
			apiObject.Alias = aws.String(v)
		}
		if v, ok := tfMap[names.AttrSource].([]any); ok {
			apiObject.Source = expandLogicalTableSource(v[0].(map[string]any))
		}
		if v, ok := tfMap["data_transforms"].([]any); ok {
			apiObject.DataTransforms = expandTransformOperations(v)
		}

		apiObjects[tfMap["logical_table_map_id"].(string)] = apiObject
	}

	return apiObjects
}

func expandLogicalTableSource(tfMap map[string]any) *awstypes.LogicalTableSource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LogicalTableSource{}

	if v, ok := tfMap["data_set_arn"].(string); ok && v != "" {
		apiObject.DataSetArn = aws.String(v)
	}
	if v, ok := tfMap["physical_table_id"].(string); ok && v != "" {
		apiObject.PhysicalTableId = aws.String(v)
	}
	if v, ok := tfMap["join_instruction"].([]any); ok && len(v) > 0 {
		apiObject.JoinInstruction = expandJoinInstruction(v[0].(map[string]any))
	}

	return apiObject
}

func expandJoinInstruction(tfMap map[string]any) *awstypes.JoinInstruction {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.JoinInstruction{}

	if v, ok := tfMap["left_operand"].(string); ok {
		apiObject.LeftOperand = aws.String(v)
	}
	if v, ok := tfMap["on_clause"].(string); ok {
		apiObject.OnClause = aws.String(v)
	}
	if v, ok := tfMap["right_operand"].(string); ok {
		apiObject.RightOperand = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok {
		apiObject.Type = awstypes.JoinType(v)
	}
	if v, ok := tfMap["left_join_key_properties"].(map[string]any); ok {
		apiObject.LeftJoinKeyProperties = expandJoinKeyProperties(v)
	}
	if v, ok := tfMap["right_join_key_properties"].(map[string]any); ok {
		apiObject.RightJoinKeyProperties = expandJoinKeyProperties(v)
	}

	return apiObject
}

func expandJoinKeyProperties(tfMap map[string]any) *awstypes.JoinKeyProperties {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.JoinKeyProperties{}

	if v, ok := tfMap["unique_key"].(bool); ok {
		apiObject.UniqueKey = aws.Bool(v)
	}

	return apiObject
}

func expandTransformOperations(tfList []any) []awstypes.TransformOperation {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.TransformOperation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandTransformOperation(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandTransformOperation(tfMap map[string]any) awstypes.TransformOperation {
	if tfMap == nil {
		return nil
	}

	var apiObject awstypes.TransformOperation

	if v, ok := tfMap["cast_column_type_operation"].([]any); ok && len(v) > 0 {
		if v := expandCastColumnTypeOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberCastColumnTypeOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["create_columns_operation"].([]any); ok && len(v) > 0 {
		if v := expandCreateColumnsOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberCreateColumnsOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["filter_operation"].([]any); ok && len(v) > 0 {
		if v := expandFilterOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberFilterOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["project_operation"].([]any); ok && len(v) > 0 {
		if v := expandProjectOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberProjectOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["rename_column_operation"].([]any); ok && len(v) > 0 {
		if v := expandRenameColumnOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberRenameColumnOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["tag_column_operation"].([]any); ok && len(v) > 0 {
		if v := expandTagColumnOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberTagColumnOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["untag_column_operation"].([]any); ok && len(v) > 0 {
		if v := expandUntagColumnOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberUntagColumnOperation{
				Value: *v,
			}
		}
	}

	return apiObject
}

func expandCastColumnTypeOperation(tfList []any) *awstypes.CastColumnTypeOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CastColumnTypeOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		apiObject.ColumnName = aws.String(v)
	}
	if v, ok := tfMap["new_column_type"].(string); ok {
		apiObject.NewColumnType = awstypes.ColumnDataType(v)
	}
	if v, ok := tfMap[names.AttrFormat].(string); ok {
		apiObject.Format = aws.String(v)
	}

	return apiObject
}

func expandCreateColumnsOperation(tfList []any) *awstypes.CreateColumnsOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CreateColumnsOperation{}

	if v, ok := tfMap["columns"].([]any); ok {
		apiObject.Columns = expandCalculatedColumns(v)
	}

	return apiObject
}

func expandCalculatedColumns(tfList []any) []awstypes.CalculatedColumn {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CalculatedColumn

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandCalculatedColumn(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCalculatedColumn(tfMap map[string]any) *awstypes.CalculatedColumn {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CalculatedColumn{}

	if v, ok := tfMap["column_id"].(string); ok {
		apiObject.ColumnId = aws.String(v)
	}
	if v, ok := tfMap["column_name"].(string); ok {
		apiObject.ColumnName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok {
		apiObject.Expression = aws.String(v)
	}

	return apiObject
}

func expandFilterOperation(tfList []any) *awstypes.FilterOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterOperation{}

	if v, ok := tfMap["condition_expression"].(string); ok {
		apiObject.ConditionExpression = aws.String(v)
	}

	return apiObject
}

func expandProjectOperation(tfList []any) *awstypes.ProjectOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ProjectOperation{}

	if v, ok := tfMap["projected_columns"].([]any); ok && len(v) > 0 {
		apiObject.ProjectedColumns = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandRenameColumnOperation(tfList []any) *awstypes.RenameColumnOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RenameColumnOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		apiObject.ColumnName = aws.String(v)
	}
	if v, ok := tfMap["new_column_name"].(string); ok {
		apiObject.NewColumnName = aws.String(v)
	}

	return apiObject
}

func expandTagColumnOperation(tfList []any) *awstypes.TagColumnOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TagColumnOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		apiObject.ColumnName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrTags].([]any); ok {
		apiObject.Tags = expandColumnTags(v)
	}

	return apiObject
}

func expandColumnTags(tfList []any) []awstypes.ColumnTag {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnTag

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandColumnTag(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandColumnTag(tfMap map[string]any) *awstypes.ColumnTag {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ColumnTag{}

	if v, ok := tfMap["column_description"].([]any); ok {
		apiObject.ColumnDescription = expandColumnDescription(v)
	}
	if v, ok := tfMap["column_geographic_role"].(string); ok {
		apiObject.ColumnGeographicRole = awstypes.GeoSpatialDataRole(v)
	}

	return apiObject
}

func expandColumnDescription(tfList []any) *awstypes.ColumnDescription {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ColumnDescription{}
	if v, ok := tfMap["text"].(string); ok {
		apiObject.Text = aws.String(v)
	}

	return apiObject
}

func expandUntagColumnOperation(tfList []any) *awstypes.UntagColumnOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.UntagColumnOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		apiObject.ColumnName = aws.String(v)
	}
	if v, ok := tfMap["tag_names"].([]any); ok {
		apiObject.TagNames = flex.ExpandStringyValueList[awstypes.ColumnTagName](v)
	}

	return apiObject
}

func ExpandPhysicalTableMap(tfList []any) map[string]awstypes.PhysicalTable {
	apiObjects := make(map[string]awstypes.PhysicalTable)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		var apiObject awstypes.PhysicalTable

		if v, ok := tfMap["custom_sql"].([]any); ok && len(v) > 0 && v[0] != nil {
			if v := expandCustomSQL(v[0].(map[string]any)); v != nil {
				apiObject = &awstypes.PhysicalTableMemberCustomSql{
					Value: *v,
				}
			}
		}
		if v, ok := tfMap["relational_table"].([]any); ok && len(v) > 0 && v[0] != nil {
			if v := expandRelationalTable(v[0].(map[string]any)); v != nil {
				apiObject = &awstypes.PhysicalTableMemberRelationalTable{
					Value: *v,
				}
			}
		}
		if v, ok := tfMap["s3_source"].([]any); ok && len(v) > 0 && v[0] != nil {
			if v := expandS3Source(v[0].(map[string]any)); v != nil {
				apiObject = &awstypes.PhysicalTableMemberS3Source{
					Value: *v,
				}
			}
		}

		apiObjects[tfMap["physical_table_map_id"].(string)] = apiObject
	}

	return apiObjects
}

func expandCustomSQL(tfMap map[string]any) *awstypes.CustomSql {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CustomSql{}

	if v, ok := tfMap["columns"].([]any); ok {
		apiObject.Columns = expandInputColumns(v)
	}
	if v, ok := tfMap["data_source_arn"].(string); ok {
		apiObject.DataSourceArn = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["sql_query"].(string); ok {
		apiObject.SqlQuery = aws.String(v)
	}

	return apiObject
}

func expandInputColumns(tfList []any) []awstypes.InputColumn {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.InputColumn

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandInputColumn(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandInputColumn(tfMap map[string]any) *awstypes.InputColumn {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InputColumn{}

	if v, ok := tfMap[names.AttrName].(string); ok {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok {
		apiObject.Type = awstypes.InputColumnDataType(v)
	}

	return apiObject
}

func expandRelationalTable(tfMap map[string]any) *awstypes.RelationalTable {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RelationalTable{}

	if v, ok := tfMap["input_columns"].([]any); ok {
		apiObject.InputColumns = expandInputColumns(v)
	}
	if v, ok := tfMap["catalog"].(string); ok {
		apiObject.Catalog = aws.String(v)
	}
	if v, ok := tfMap["data_source_arn"].(string); ok {
		apiObject.DataSourceArn = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrSchema].(string); ok {
		apiObject.Schema = aws.String(v)
	}

	return apiObject
}

func expandS3Source(tfMap map[string]any) *awstypes.S3Source {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3Source{}

	if v, ok := tfMap["input_columns"].([]any); ok {
		apiObject.InputColumns = expandInputColumns(v)
	}
	if v, ok := tfMap["upload_settings"].(map[string]any); ok {
		apiObject.UploadSettings = expandUploadSettings(v)
	}
	if v, ok := tfMap["data_source_arn"].(string); ok {
		apiObject.DataSourceArn = aws.String(v)
	}

	return apiObject
}

func expandUploadSettings(tfMap map[string]any) *awstypes.UploadSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.UploadSettings{}

	if v, ok := tfMap["contains_header"].(bool); ok {
		apiObject.ContainsHeader = aws.Bool(v)
	}
	if v, ok := tfMap["delimiter"].(string); ok {
		apiObject.Delimiter = aws.String(v)
	}
	if v, ok := tfMap[names.AttrFormat].(string); ok {
		apiObject.Format = awstypes.FileFormat(v)
	}
	if v, ok := tfMap["start_from_row"].(int); ok {
		apiObject.StartFromRow = aws.Int32(int32(v))
	}
	if v, ok := tfMap["text_qualifier"].(string); ok {
		apiObject.TextQualifier = awstypes.TextQualifier(v)
	}

	return apiObject
}

func ExpandRowLevelPermissionDataSet(tfList []any) *awstypes.RowLevelPermissionDataSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RowLevelPermissionDataSet{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}
	if v, ok := tfMap["format_version"].(string); ok && v != "" {
		apiObject.FormatVersion = awstypes.RowLevelPermissionFormatVersion(v)
	}
	if v, ok := tfMap[names.AttrNamespace].(string); ok && v != "" {
		apiObject.Namespace = aws.String(v)
	}
	if v, ok := tfMap["permission_policy"].(string); ok && v != "" {
		apiObject.PermissionPolicy = awstypes.RowLevelPermissionPolicy(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = awstypes.Status(v)
	}

	return apiObject
}

func ExpandRowLevelPermissionTagConfiguration(tfList []any) *awstypes.RowLevelPermissionTagConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RowLevelPermissionTagConfiguration{}

	if v, ok := tfMap["tag_rules"].([]any); ok {
		apiObject.TagRules = expandRowLevelPermissionTagRules(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok {
		apiObject.Status = awstypes.Status(v)
	}

	return apiObject
}

func ExpandDataSetRefreshProperties(tfList []any) *awstypes.DataSetRefreshProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataSetRefreshProperties{}

	if v, ok := tfMap["refresh_configuration"].([]any); ok {
		apiObject.RefreshConfiguration = expandRefreshConfiguration(v)
	}

	return apiObject
}

func expandRefreshConfiguration(tfList []any) *awstypes.RefreshConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RefreshConfiguration{}

	if v, ok := tfMap["incremental_refresh"].([]any); ok {
		apiObject.IncrementalRefresh = expandIncrementalRefresh(v)
	}

	return apiObject
}

func expandIncrementalRefresh(tfList []any) *awstypes.IncrementalRefresh {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.IncrementalRefresh{}

	if v, ok := tfMap["lookback_window"].([]any); ok {
		apiObject.LookbackWindow = expandLookbackWindow(v)
	}

	return apiObject
}

func expandLookbackWindow(tfList []any) *awstypes.LookbackWindow {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LookbackWindow{}

	if v, ok := tfMap["column_name"].(string); ok {
		apiObject.ColumnName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrSize].(int); ok {
		apiObject.Size = aws.Int64(int64(v))
	}
	if v, ok := tfMap["size_unit"].(string); ok {
		apiObject.SizeUnit = awstypes.LookbackWindowSizeUnit(v)
	}

	return apiObject
}

func expandRowLevelPermissionTagRules(tfList []any) []awstypes.RowLevelPermissionTagRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.RowLevelPermissionTagRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandRowLevelPermissionTagRule(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandRowLevelPermissionTagRule(tfMap map[string]any) *awstypes.RowLevelPermissionTagRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RowLevelPermissionTagRule{}

	if v, ok := tfMap["column_name"].(string); ok {
		apiObject.ColumnName = aws.String(v)
	}
	if v, ok := tfMap["tag_key"].(string); ok {
		apiObject.TagKey = aws.String(v)
	}
	if v, ok := tfMap["match_all_value"].(string); ok {
		apiObject.MatchAllValue = aws.String(v)
	}
	if v, ok := tfMap["tag_multi_value_delimiter"].(string); ok {
		apiObject.TagMultiValueDelimiter = aws.String(v)
	}

	return apiObject
}

func FlattenColumnGroups(apiObjects []awstypes.ColumnGroup) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.GeoSpatialColumnGroup != nil {
			tfMap["geo_spatial_column_group"] = flattenGeoSpatialColumnGroup(apiObject.GeoSpatialColumnGroup)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func FlattenOutputColumns(apiObjects []awstypes.OutputColumn) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Description != nil {
			tfMap[names.AttrDescription] = aws.ToString(apiObject.Description)
		}
		if apiObject.Name != nil {
			tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		}
		tfMap[names.AttrType] = apiObject.Type

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenGeoSpatialColumnGroup(apiObject *awstypes.GeoSpatialColumnGroup) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Columns != nil {
		tfMap["columns"] = apiObject.Columns
	}
	tfMap["country_code"] = apiObject.CountryCode
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}

	return []any{tfMap}
}

func FlattenColumnLevelPermissionRules(apiObjects []awstypes.ColumnLevelPermissionRule) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.ColumnNames != nil {
			tfMap["column_names"] = apiObject.ColumnNames
		}
		if apiObject.Principals != nil {
			tfMap["principals"] = apiObject.Principals
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func FlattenDataSetUsageConfiguration(apiObject *awstypes.DataSetUsageConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["disable_use_as_direct_query_source"] = apiObject.DisableUseAsDirectQuerySource
	tfMap["disable_use_as_imported_source"] = apiObject.DisableUseAsImportedSource

	return []any{tfMap}
}

func FlattenFieldFolders(apiObjects map[string]awstypes.FieldFolder) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for k, apiObject := range apiObjects {
		tfMap := map[string]any{
			"field_folders_id": k,
		}

		if len(apiObject.Columns) > 0 {
			tfMap["columns"] = apiObject.Columns
		}
		if apiObject.Description != nil {
			tfMap[names.AttrDescription] = aws.ToString(apiObject.Description)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func FlattenLogicalTableMap(apiObjects map[string]awstypes.LogicalTable) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for k, apiObject := range apiObjects {
		tfMap := map[string]any{
			"logical_table_map_id": k,
		}

		if apiObject.Alias != nil {
			tfMap[names.AttrAlias] = aws.ToString(apiObject.Alias)
		}
		if apiObject.DataTransforms != nil {
			tfMap["data_transforms"] = flattenTransformOperations(apiObject.DataTransforms)
		}
		if apiObject.Source != nil {
			tfMap[names.AttrSource] = flattenLogicalTableSource(apiObject.Source)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTransformOperations(apiObjects []awstypes.TransformOperation) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		switch v := apiObject.(type) {
		case *awstypes.TransformOperationMemberCastColumnTypeOperation:
			tfMap["cast_column_type_operation"] = flattenCastColumnTypeOperation(&v.Value)
		case *awstypes.TransformOperationMemberCreateColumnsOperation:
			tfMap["create_columns_operation"] = flattenCreateColumnsOperation(&v.Value)
		case *awstypes.TransformOperationMemberFilterOperation:
			tfMap["filter_operation"] = flattenFilterOperation(&v.Value)
		case *awstypes.TransformOperationMemberProjectOperation:
			tfMap["project_operation"] = flattenProjectOperation(&v.Value)
		case *awstypes.TransformOperationMemberRenameColumnOperation:
			tfMap["rename_column_operation"] = flattenRenameColumnOperation(&v.Value)
		case *awstypes.TransformOperationMemberTagColumnOperation:
			tfMap["tag_column_operation"] = flattenTagColumnOperation(&v.Value)
		case *awstypes.TransformOperationMemberUntagColumnOperation:
			tfMap["untag_column_operation"] = flattenUntagColumnOperation(&v.Value)
		default:
			continue
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCastColumnTypeOperation(apiObject *awstypes.CastColumnTypeOperation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.Format != nil {
		tfMap[names.AttrFormat] = aws.ToString(apiObject.Format)
	}
	tfMap["new_column_type"] = apiObject.NewColumnType

	return []any{tfMap}
}

func flattenCreateColumnsOperation(apiObject *awstypes.CreateColumnsOperation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Columns != nil {
		tfMap["columns"] = flattenCalculatedColumns(apiObject.Columns)
	}

	return []any{tfMap}
}

func flattenCalculatedColumns(apiObjects []awstypes.CalculatedColumn) any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.ColumnId != nil {
			tfMap["column_id"] = aws.ToString(apiObject.ColumnId)
		}
		if apiObject.ColumnName != nil {
			tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
		}
		if apiObject.Expression != nil {
			tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFilterOperation(apiObject *awstypes.FilterOperation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ConditionExpression != nil {
		tfMap["condition_expression"] = aws.ToString(apiObject.ConditionExpression)
	}

	return []any{tfMap}
}

func flattenProjectOperation(apiObject *awstypes.ProjectOperation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ProjectedColumns != nil {
		tfMap["projected_columns"] = apiObject.ProjectedColumns
	}

	return []any{tfMap}
}

func flattenRenameColumnOperation(apiObject *awstypes.RenameColumnOperation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.NewColumnName != nil {
		tfMap["new_column_name"] = aws.ToString(apiObject.NewColumnName)
	}

	return []any{tfMap}
}

func flattenTagColumnOperation(apiObject *awstypes.TagColumnOperation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.Tags != nil {
		tfMap[names.AttrTags] = flattenColumnTags(apiObject.Tags)
	}

	return []any{tfMap}
}

func flattenColumnTags(apiObjects []awstypes.ColumnTag) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.ColumnDescription != nil {
			tfMap["column_description"] = flattenColumnDescription(apiObject.ColumnDescription)
		}
		tfMap["column_geographic_role"] = apiObject.ColumnGeographicRole

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnDescription(apiObject *awstypes.ColumnDescription) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Text != nil {
		tfMap["text"] = aws.ToString(apiObject.Text)
	}

	return []any{tfMap}
}

func flattenUntagColumnOperation(apiObject *awstypes.UntagColumnOperation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.TagNames != nil {
		tfMap["tag_names"] = apiObject.TagNames
	}

	return []any{tfMap}
}

func flattenLogicalTableSource(apiObject *awstypes.LogicalTableSource) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DataSetArn != nil {
		tfMap["data_set_arn"] = aws.ToString(apiObject.DataSetArn)
	}
	if apiObject.JoinInstruction != nil {
		tfMap["join_instruction"] = flattenJoinInstruction(apiObject.JoinInstruction)
	}
	if apiObject.PhysicalTableId != nil {
		tfMap["physical_table_id"] = aws.ToString(apiObject.PhysicalTableId)
	}

	return []any{tfMap}
}

func flattenJoinInstruction(apiObject *awstypes.JoinInstruction) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.LeftJoinKeyProperties != nil {
		tfMap["left_join_key_properties"] = flattenJoinKeyProperties(apiObject.LeftJoinKeyProperties)
	}
	if apiObject.LeftOperand != nil {
		tfMap["left_operand"] = aws.ToString(apiObject.LeftOperand)
	}
	if apiObject.OnClause != nil {
		tfMap["on_clause"] = aws.ToString(apiObject.OnClause)
	}
	if apiObject.RightJoinKeyProperties != nil {
		tfMap["right_join_key_properties"] = flattenJoinKeyProperties(apiObject.RightJoinKeyProperties)
	}
	if apiObject.RightOperand != nil {
		tfMap["right_operand"] = aws.ToString(apiObject.RightOperand)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []any{tfMap}
}

func flattenJoinKeyProperties(apiObject *awstypes.JoinKeyProperties) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.UniqueKey != nil {
		tfMap["unique_key"] = aws.ToBool(apiObject.UniqueKey)
	}

	return tfMap
}

func FlattenPhysicalTableMap(apiObjects map[string]awstypes.PhysicalTable) []any {
	var tfList []any

	for k, apiObject := range apiObjects {
		tfMap := map[string]any{
			"physical_table_map_id": k,
		}

		switch v := apiObject.(type) {
		case *awstypes.PhysicalTableMemberCustomSql:
			tfMap["custom_sql"] = flattenCustomSQL(&v.Value)
		case *awstypes.PhysicalTableMemberRelationalTable:
			tfMap["relational_table"] = flattenRelationalTable(&v.Value)
		case *awstypes.PhysicalTableMemberS3Source:
			tfMap["s3_source"] = flattenS3Source(&v.Value)
		default:
			continue
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCustomSQL(apiObject *awstypes.CustomSql) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Columns != nil {
		tfMap["columns"] = flattenInputColumns(apiObject.Columns)
	}
	if apiObject.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.ToString(apiObject.DataSourceArn)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	if apiObject.SqlQuery != nil {
		tfMap["sql_query"] = aws.ToString(apiObject.SqlQuery)
	}

	return []any{tfMap}
}

func flattenInputColumns(apiObjects []awstypes.InputColumn) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Name != nil {
			tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		}
		tfMap[names.AttrType] = apiObject.Type

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRelationalTable(apiObject *awstypes.RelationalTable) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Catalog != nil {
		tfMap["catalog"] = aws.ToString(apiObject.Catalog)
	}
	if apiObject.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.ToString(apiObject.DataSourceArn)
	}
	if apiObject.InputColumns != nil {
		tfMap["input_columns"] = flattenInputColumns(apiObject.InputColumns)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	if apiObject.Schema != nil {
		tfMap[names.AttrSchema] = aws.ToString(apiObject.Schema)
	}

	return []any{tfMap}
}

func flattenS3Source(apiObject *awstypes.S3Source) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.ToString(apiObject.DataSourceArn)
	}
	if apiObject.InputColumns != nil {
		tfMap["input_columns"] = flattenInputColumns(apiObject.InputColumns)
	}
	if apiObject.UploadSettings != nil {
		tfMap["upload_settings"] = flattenUploadSettings(apiObject.UploadSettings)
	}

	return []any{tfMap}
}

func flattenUploadSettings(apiObject *awstypes.UploadSettings) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ContainsHeader != nil {
		tfMap["contains_header"] = aws.ToBool(apiObject.ContainsHeader)
	}
	if apiObject.Delimiter != nil {
		tfMap["delimiter"] = aws.ToString(apiObject.Delimiter)
	}
	tfMap[names.AttrFormat] = apiObject.Format
	if apiObject.StartFromRow != nil {
		tfMap["start_from_row"] = aws.ToInt32(apiObject.StartFromRow)
	}
	tfMap["text_qualifier"] = apiObject.TextQualifier

	return []any{tfMap}
}

func FlattenRowLevelPermissionDataSet(apiObject *awstypes.RowLevelPermissionDataSet) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Arn != nil {
		tfMap[names.AttrARN] = aws.ToString(apiObject.Arn)
	}
	tfMap["format_version"] = apiObject.FormatVersion
	if apiObject.Namespace != nil {
		tfMap[names.AttrNamespace] = aws.ToString(apiObject.Namespace)
	}
	tfMap["permission_policy"] = apiObject.PermissionPolicy
	tfMap[names.AttrStatus] = apiObject.Status

	return []any{tfMap}
}

func FlattenRowLevelPermissionTagConfiguration(apiObject *awstypes.RowLevelPermissionTagConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap[names.AttrStatus] = apiObject.Status
	if apiObject.TagRules != nil {
		tfMap["tag_rules"] = flattenRowLevelPermissionTagRules(apiObject.TagRules)
	}

	return []any{tfMap}
}

func FlattenDataSetRefreshProperties(apiObject *awstypes.DataSetRefreshProperties) any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.RefreshConfiguration != nil {
		tfMap["refresh_configuration"] = flattenRefreshConfiguration(apiObject.RefreshConfiguration)
	}

	return []any{tfMap}
}

func flattenRefreshConfiguration(apiObject *awstypes.RefreshConfiguration) any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.IncrementalRefresh != nil {
		tfMap["incremental_refresh"] = flattenIncrementalRefresh(apiObject.IncrementalRefresh)
	}

	return []any{tfMap}
}

func flattenIncrementalRefresh(apiObject *awstypes.IncrementalRefresh) any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.LookbackWindow != nil {
		tfMap["lookback_window"] = flattenLookbackWindow(apiObject.LookbackWindow)
	}

	return []any{tfMap}
}

func flattenLookbackWindow(apiObject *awstypes.LookbackWindow) any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.Size != nil {
		tfMap[names.AttrSize] = aws.ToInt64(apiObject.Size)
	}
	tfMap["size_unit"] = apiObject.SizeUnit

	return []any{tfMap}
}

func flattenRowLevelPermissionTagRules(apiObjects []awstypes.RowLevelPermissionTagRule) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.ColumnName != nil {
			tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
		}
		if apiObject.MatchAllValue != nil {
			tfMap["match_all_value"] = aws.ToString(apiObject.MatchAllValue)
		}
		if apiObject.TagKey != nil {
			tfMap["tag_key"] = aws.ToString(apiObject.TagKey)
		}
		if apiObject.TagMultiValueDelimiter != nil {
			tfMap["tag_multi_value_delimiter"] = aws.ToString(apiObject.TagMultiValueDelimiter)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
