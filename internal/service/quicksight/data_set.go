package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsQuickSightDataSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsQuickSightDataSetCreate,
		ReadWithoutTimeout:   resourceAwsQuickSightDataSetRead,
		UpdateWithoutTimeout: resourceAwsQuickSightDataSetUpdate,
		DeleteWithoutTimeout: resourceAwsQuickSightDataSetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},

			"data_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"import_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(quicksight.DataSetImportMode_Values(), false),
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},

			"physical_table_map": {
				Type:     schema.TypeMap,
				Required: true,
				MinItems: 0,
				MaxItems: 32,
				// how do i validate key length?
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cusom_sql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_source_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
									"sql_query": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 65536),
									},
									"columns": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 2048,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 128),
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(quicksight.InputColumnDataType_Values(), false),
												},
											},
										},
									},
								},
							},
						},
						"relational_table": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_source_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"input_columns": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 2048,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 128),
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(quicksight.InputColumnDataType_Values(), false),
												},
											},
										},
									},
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"catalog": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"schema": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"s3_source": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_source_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"input_columns": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 2048,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 128),
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(quicksight.InputColumnDataType_Values(), false),
												},
											},
										},
									},
									"upload_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"contains_header": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"delimiter": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 1),
												},
												"format": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(quicksight.FileFormat_Values(), false),
												},
												"start_from_row": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"text_qualifier": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(quicksight.TextQualifier_Values(), false),
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

			"column_groups": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 8,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"geo_spacial_column_group": {
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
									"country_code": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(quicksight.GeoSpatialCountryCode_Values(), false),
									},
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
								},
							},
						},
					},
				},
			},

			"column_level_permission_rules": {
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
			},

			"data_set_usage_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disable_use_as_direct_query_source": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"disable_use_as_imported_source": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"field_folders": {
				Type:     schema.TypeMap,
				Optional: true,
				MinItems: 1,
				MaxItems: 1000,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"columns": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 5000,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 500),
						},
					},
				},
			},

			"logical_table_map": {
				Type:     schema.TypeMap,
				Optional: true,
				MaxItems: 64,
				// key length constraint 1 to 64
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alias": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
						"source": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_set_arn": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"join_instruction": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"left_operand": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 64),
												},
												"on_clause": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 512),
												},
												"right_operand": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 64),
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(quicksight.JoinType_Values(), false),
												},
												"left_join_key_properties": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unique_key": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"right_join_key_properties": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unique_key": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"physical_table_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
								},
							},
						},
						"data_transforms": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 2048,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cast_column_type_operation": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"column_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 128),
												},
												"new_column_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(quicksight.ColumnDataType_Values(), false),
												},
												"format": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 32),
												},
											},
										},
									},
									"create_columns_operation": {
										Type:     schema.TypeList,
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
															"column_id": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 64),
															},
															"column_name": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 128),
															},
															"expression": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 4096),
															},
														},
													},
												},
											},
										},
									},
									"filter_operation": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"condition_expression": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 4096),
												},
											},
										},
									},
									"project_operation": {
										Type:     schema.TypeList,
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
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"column_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 128),
												},
												"new_column_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 128),
												},
											},
										},
									},
									"tag_column_operation": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"column_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 128),
												},
												"tags": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 16,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"column_description": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"text": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 500),
																		},
																	},
																},
															},
															"column_geographic_role": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringInSlice(quicksight.GeoSpatialDataRole_Values(), false),
															},
														},
													},
												},
											},
										},
									},
									"untag_column_operation": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"column_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 128),
												},
												"tag_names": {
													Type:         schema.TypeList,
													Required:     true,
													Elem:         &schema.Schema{Type: schema.TypeString},
													ValidateFunc: validation.StringInSlice(quicksight.ColumnTagName_Values(), false),
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

			"permissions": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 64,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 16,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"principal": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
					},
				},
			},

			"row_level_permission_data_set": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"permission_policy": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(quicksight.RowLevelPermissionPolicy_Values(), false),
						},
						"format_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(quicksight.RowLevelPermissionFormatVersion_Values(), false),
						},
						"namespace": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
						},
					},
				},
			},

			"row_level_permission_tag_configurations": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
									"tag_key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									"match_all_value": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"tag_multi_value_delimiter": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 10),
									},
								},
							},
						},
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
						},
					},
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsQuickSightDataSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).quicksightconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	awsAccountId := meta.(*AWSClient).accountid
	id := d.Get("data_set_id").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountId = v.(string)
	}

	params := &quicksight.CreateDataSetInput{
		AwsAccountId:     aws.String(awsAccountId),
		DataSetId:        aws.String(id),
		ImportMode:       aws.String(d.Get("import_mode").(string)),
		PhysicalTableMap: expandQuickSightDataSetPhysicalTableMap(d.Get("physical_table_map").([]interface{})),
		Name:             aws.String(d.Get("name").(string)),
	}

	if len(tags) > 0 {
		params.Tags = tags.IgnoreAws().QuicksightTags()
	}

	if v, ok := d.GetOk("column_groups"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.ColumnGroups = expandQuickSightDataSetColumnGroups(v.([]interface{}))
	}

	if v, ok := d.GetOk("column_level_permission_rules"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.ColumnLevelPermissionRules = expandQuickSightDataSetColumnLevelPermissionRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("data_set_usage_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.DataSetUsageConfiguration = expandQuickSightDataSetUsageConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("field_folders"); ok && len(v.(map[string]interface{})) != 0 {
		params.FieldFolders = expandQuickSightDataSourceFieldFolders(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("logical_table_map"); ok && len(v.(map[string]interface{})) != 0 {
		params.LogicalTableMap = expandQuickSightDataSourceLogicalTableMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("permissions"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.Permissions = expandQuickSightDataSetPermissions(v.([]interface{}))
	}

	if v, ok := d.GetOk("row_level_permission_data_set"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.RowLevelPermissionDataSet = expandQuickSightDataSetRowLevelPermissionDataSet(v.([]interface{}))
	}

	if v, ok := d.GetOk("row_level_permission_tag_configurations"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.RowLevelPermissionTagConfiguration = expandQuickSightDataSetRowLevelPermissionTagConfigurations(v.([]interface{}))
	}

	_, err := conn.CreateDataSetWithContext(ctx, params)
	if err != nil {
		return diag.Errorf("error creating QuickSight Data Set: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", awsAccountId, id))

	// confirm dataset has been created? having troubles due to a lack of output status and error handling.

	return resourceAwsQuickSightDataSetRead(ctx, d, meta)
}

func resourceAwsQuickSightDataSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceAwsQuickSightDataSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceAwsQuickSightDataSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func expandQuickSightDataSetPhysicalTableMap(tfList []interface{}) map[string]*quicksight.PhysicalTable {
	return nil
}

func expandQuickSightDataSetColumnGroups(tfList []interface{}) []*quicksight.ColumnGroup {
	return nil
}

func expandQuickSightDataSetColumnLevelPermissionRules(tfList []interface{}) []*quicksight.ColumnLevelPermissionRule {
	return nil
}

func expandQuickSightDataSetUsageConfiguration(tfList []interface{}) *quicksight.DataSetUsageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	usageConfiguration := &quicksight.DataSetUsageConfiguration{}

	if v, ok := tfMap["disable_use_as_direct_query_source"].(bool); ok {
		usageConfiguration.DisableUseAsDirectQuerySource = aws.Bool(v)
	}

	if v, ok := tfMap["disable_use_as_imported_source"].(bool); ok {
		usageConfiguration.DisableUseAsImportedSource = aws.Bool(v)
	}

	return usageConfiguration
}

func expandQuickSightDataSourceFieldFolders(tfMap map[string]interface{}) map[string]*quicksight.FieldFolder {
	return nil
}

func expandQuickSightDataSourceLogicalTableMap(tfMap map[string]interface{}) map[string]*quicksight.LogicalTable {
	return nil
}

func expandQuickSightDataSetPermissions(tfList []interface{}) []*quicksight.ResourcePermission {
	return nil
}

func expandQuickSightDataSetRowLevelPermissionDataSet(tfList []interface{}) *quicksight.RowLevelPermissionDataSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	rowLevelPermission := &quicksight.RowLevelPermissionDataSet{}

	if v, ok := tfMap["arn"].(string); ok {
		rowLevelPermission.Arn = aws.String(v)
	}

	if v, ok := tfMap["permission_policy"].(string); ok {
		rowLevelPermission.PermissionPolicy = aws.String(v)
	}

	if v, ok := tfMap["format_version"].(string); ok {
		rowLevelPermission.FormatVersion = aws.String(v)
	}

	if v, ok := tfMap["namespace"].(string); ok {
		rowLevelPermission.Namespace = aws.String(v)
	}

	if v, ok := tfMap["status"].(string); ok {
		rowLevelPermission.Status = aws.String(v)
	}

	return rowLevelPermission
}

func expandQuickSightDataSetRowLevelPermissionTagConfigurations(tfList []interface{}) *quicksight.RowLevelPermissionTagConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	rowLevelPermissionTagConfiguration := &quicksight.RowLevelPermissionTagConfiguration{}

	if v, ok := tfMap["tag_rules"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			tagRules := &quicksight.RowLevelPermissionTagRule{}

			if v, ok := m["column_name"].(string); ok {
				tagRules.ColumnName = aws.String(v)
			}

			if v, ok := m["tag_key"].(string); ok {
				tagRules.TagKey = aws.String(v)
			}

			if v, ok := m["match_all_value"].(string); ok {
				tagRules.MatchAllValue = aws.String(v)
			}

			if v, ok := m["tag_multi_value_delimiter"].(string); ok {
				tagRules.TagMultiValueDelimiter = aws.String(v)
			}
		}
	}

	if v, ok := tfMap["status"].(string); ok {
		rowLevelPermissionTagConfiguration.Status = aws.String(v)
	}

	return rowLevelPermissionTagConfiguration
}
