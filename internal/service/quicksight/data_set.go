package quicksight

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDataSet() *schema.Resource {
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
				ValidateFunc: verify.ValidAccountID,
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

			"data_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

			"import_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(quicksight.DataSetImportMode_Values(), false),
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

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
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

			"physical_table_map": {
				Type:     schema.TypeMap,
				Required: true,
				MinItems: 0,
				MaxItems: 32,
				// how do i validate key length?
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_sql": {
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

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAwsQuickSightDataSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	awsAccountId := meta.(*conns.AWSClient).AccountID
	id := d.Get("data_set_id").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountId = v.(string)
	}

	params := &quicksight.CreateDataSetInput{
		AwsAccountId:     aws.String(awsAccountId),
		DataSetId:        aws.String(id),
		ImportMode:       aws.String(d.Get("import_mode").(string)),
		PhysicalTableMap: expandQuickSightDataSetPhysicalTableMap(d.Get("physical_table_map").(map[string]interface{})),
		Name:             aws.String(d.Get("name").(string)),
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("column_groups"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.ColumnGroups = expandQuickSightDataSetColumnGroups(v.([]interface{}))
	}

	if v, ok := d.GetOk("column_level_permission_rules"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.ColumnLevelPermissionRules = expandQuickSightDataSetColumnLevelPermissionRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("data_set_usage_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.DataSetUsageConfiguration = expandQuickSightDataSetUsageConfiguration(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("field_folders"); ok && len(v.(map[string]interface{})) != 0 {
		params.FieldFolders = expandQuickSightDataSetFieldFolders(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("logical_table_map"); ok && len(v.(map[string]interface{})) != 0 {
		params.LogicalTableMap = expandQuickSightDataSetLogicalTableMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("permissions"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.Permissions = expandQuickSightDataSetResourcePermissions(v.([]interface{}))
	}

	if v, ok := d.GetOk("row_level_permission_data_set"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.RowLevelPermissionDataSet = expandQuickSightDataSetRowLevelPermissionDataSet(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("row_level_permission_tag_configurations"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.RowLevelPermissionTagConfiguration = expandQuickSightDataSetRowLevelPermissionTagConfigurations(v.(map[string]interface{}))
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

func expandQuickSightDataSetColumnGroups(tfList []interface{}) []*quicksight.ColumnGroup {
	if len(tfList) == 0 {
		return nil
	}

	var columnGroups []*quicksight.ColumnGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		columnGroup := expandQuickSightDataSetColumnGroup(tfMap)

		if columnGroup == nil {
			continue
		}

		columnGroups = append(columnGroups, columnGroup)
	}

	return columnGroups
}

func expandQuickSightDataSetColumnGroup(tfMap map[string]interface{}) *quicksight.ColumnGroup {
	if len(tfMap) == 0 {
		return nil
	}

	columnGroup := &quicksight.ColumnGroup{}

	if tfMapRaw, ok := tfMap["geo_spatial_column_group"].(map[string]interface{}); ok {
		columnGroup.GeoSpatialColumnGroup = expandQuickSightDataSetGeoSpatialColumnGroup(tfMapRaw)
	}

	return columnGroup
}

func expandQuickSightDataSetGeoSpatialColumnGroup(tfMap map[string]interface{}) *quicksight.GeoSpatialColumnGroup {
	if tfMap == nil {
		return nil
	}

	geoSpatialColumnGroup := &quicksight.GeoSpatialColumnGroup{}

	if v, ok := tfMap["columns"].([]string); ok {
		geoSpatialColumnGroup.Columns = aws.StringSlice(v)
	}

	if v, ok := tfMap["country_code"].(string); ok {
		geoSpatialColumnGroup.CountryCode = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok {
		geoSpatialColumnGroup.Name = aws.String(v)
	}

	return geoSpatialColumnGroup
}

func expandQuickSightDataSetColumnLevelPermissionRules(tfList []interface{}) []*quicksight.ColumnLevelPermissionRule {
	if len(tfList) == 0 {
		return nil
	}

	var columnLevelPermissionRules []*quicksight.ColumnLevelPermissionRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		columnLevelPermissionRule := expandQuickSightDataSetColumnLevelPermissionRule(tfMap)
		if columnLevelPermissionRule == nil {
			continue
		}

		columnLevelPermissionRules = append(columnLevelPermissionRules, columnLevelPermissionRule)
	}

	return columnLevelPermissionRules
}

func expandQuickSightDataSetColumnLevelPermissionRule(tfMap map[string]interface{}) *quicksight.ColumnLevelPermissionRule {
	if len(tfMap) == 0 {
		return nil
	}

	columnLevelPermissionRule := &quicksight.ColumnLevelPermissionRule{}

	if v, ok := tfMap["column_name"].([]string); ok {
		columnLevelPermissionRule.ColumnNames = aws.StringSlice(v)
	}

	if v, ok := tfMap["permission_policy"].([]string); ok {
		columnLevelPermissionRule.Principals = aws.StringSlice(v)
	}

	return columnLevelPermissionRule
}

func expandQuickSightDataSetUsageConfiguration(tfMap map[string]interface{}) *quicksight.DataSetUsageConfiguration {
	if len(tfMap) == 0 {
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

func expandQuickSightDataSetFieldFolders(tfMap map[string]interface{}) map[string]*quicksight.FieldFolder {
	if len(tfMap) == 0 {
		return nil
	}

	fieldFolderMap := make(map[string]*quicksight.FieldFolder)
	for k, v := range tfMap {

		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		fieldFolder := &quicksight.FieldFolder{}

		if v, ok := vMap["columns"].([]string); ok {
			fieldFolder.Columns = aws.StringSlice(v)
		}

		if v, ok := vMap["description"].(string); ok {
			fieldFolder.Description = aws.String(v)
		}

		fieldFolderMap[k] = fieldFolder
	}

	return fieldFolderMap
}

func expandQuickSightDataSetLogicalTableMap(tfMap map[string]interface{}) map[string]*quicksight.LogicalTable {
	if len(tfMap) == 0 {
		return nil
	}

	logicalTableMap := make(map[string]*quicksight.LogicalTable)
	for k, v := range tfMap {

		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		logicalTable := &quicksight.LogicalTable{}

		if v, ok := vMap["alias"].(string); ok {
			logicalTable.Alias = aws.String(v)
		}

		if v, ok := vMap["logical_table_source"].(map[string]interface{}); ok {
			logicalTable.Source = expandQuickSightDataSetLogicalTableSource(v)
		}

		if v, ok := vMap["data_transforms"].([]interface{}); ok {
			logicalTable.DataTransforms = expandQuickSightDataSetDataTransforms(v)
		}

		logicalTableMap[k] = logicalTable
	}

	return logicalTableMap
}

func expandQuickSightDataSetLogicalTableSource(tfMap map[string]interface{}) *quicksight.LogicalTableSource {
	if tfMap == nil {
		return nil
	}

	logicalTableSource := &quicksight.LogicalTableSource{}

	if v, ok := tfMap["data_set_arn"].(string); ok {
		logicalTableSource.DataSetArn = aws.String(v)
	}

	if v, ok := tfMap["physical_table_id"].(string); ok {
		logicalTableSource.PhysicalTableId = aws.String(v)
	}

	if v, ok := tfMap["join_instruction"].(map[string]interface{}); ok {
		logicalTableSource.JoinInstruction = expandQuickSightDataSetJoinInstruction(v)
	}

	return logicalTableSource
}

func expandQuickSightDataSetJoinInstruction(tfMap map[string]interface{}) *quicksight.JoinInstruction {
	if tfMap == nil {
		return nil
	}

	joinInstruction := &quicksight.JoinInstruction{}

	if v, ok := tfMap["left_operand"].(string); ok {
		joinInstruction.LeftOperand = aws.String(v)
	}

	if v, ok := tfMap["on_clause"].(string); ok {
		joinInstruction.OnClause = aws.String(v)
	}

	if v, ok := tfMap["right_operand"].(string); ok {
		joinInstruction.RightOperand = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok {
		joinInstruction.Type = aws.String(v)
	}

	if v, ok := tfMap["left_join_key_properties"].(map[string]interface{}); ok {
		joinInstruction.LeftJoinKeyProperties = expandQuickSightDataSetJoinKeyProperties(v)
	}

	if v, ok := tfMap["right_join_key_properties"].(map[string]interface{}); ok {
		joinInstruction.RightJoinKeyProperties = expandQuickSightDataSetJoinKeyProperties(v)
	}

	return joinInstruction
}

func expandQuickSightDataSetJoinKeyProperties(tfMap map[string]interface{}) *quicksight.JoinKeyProperties {
	if tfMap == nil {
		return nil
	}

	joinKeyProperties := &quicksight.JoinKeyProperties{}

	if v, ok := tfMap["unique_key"].(bool); ok {
		joinKeyProperties.UniqueKey = aws.Bool(v)
	}

	return joinKeyProperties
}

func expandQuickSightDataSetDataTransforms(tfList []interface{}) []*quicksight.TransformOperation {
	if len(tfList) == 0 {
		return nil
	}

	var transformOperations []*quicksight.TransformOperation

	for _, tfMapRaw := range tfList {

		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		transformOperation := expandQuickSightDataSetDataTransform(tfMap)

		if transformOperation == nil {
			continue
		}

		transformOperations = append(transformOperations, transformOperation)

	}

	return transformOperations
}

func expandQuickSightDataSetDataTransform(tfMap map[string]interface{}) *quicksight.TransformOperation {
	if tfMap == nil {
		return nil
	}

	transformOperation := &quicksight.TransformOperation{}

	if v, ok := tfMap["cast_column_type_operation"].(map[string]interface{}); ok {
		transformOperation.CastColumnTypeOperation = expandQuickSightDataSetCastColumnTypeOperation(v)
	}

	if v, ok := tfMap["create_columns_operation"].(map[string]interface{}); ok {
		transformOperation.CreateColumnsOperation = expandQuickSightDataSetCreateColumnsOperation(v)
	}

	if v, ok := tfMap["filter_operation"].(map[string]interface{}); ok {
		transformOperation.FilterOperation = expandQuickSightDataSetFilterOperation(v)
	}

	if v, ok := tfMap["project_operation"].(map[string]interface{}); ok {
		transformOperation.ProjectOperation = expandQuickSightDataSetProjectOperation(v)
	}

	if v, ok := tfMap["rename_column_operation"].(map[string]interface{}); ok {
		transformOperation.RenameColumnOperation = expandQuickSightDataSetRenameColumnOperation(v)
	}

	if v, ok := tfMap["tag_column_operation"].(map[string]interface{}); ok {
		transformOperation.TagColumnOperation = expandQuickSightDataSetTagColumnOperation(v)
	}

	if v, ok := tfMap["untag_column_operation"].(map[string]interface{}); ok {
		transformOperation.UntagColumnOperation = expandQuickSightDataSetUntagColumnOperation(v)
	}

	return transformOperation
}

func expandQuickSightDataSetCastColumnTypeOperation(tfMap map[string]interface{}) *quicksight.CastColumnTypeOperation {
	if tfMap == nil {
		return nil
	}

	castColumnTypeOperation := &quicksight.CastColumnTypeOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		castColumnTypeOperation.ColumnName = aws.String(v)
	}

	if v, ok := tfMap["new_column_type"].(string); ok {
		castColumnTypeOperation.NewColumnType = aws.String(v)
	}

	if v, ok := tfMap["format"].(string); ok {
		castColumnTypeOperation.Format = aws.String(v)
	}

	return castColumnTypeOperation
}

func expandQuickSightDataSetCreateColumnsOperation(tfMap map[string]interface{}) *quicksight.CreateColumnsOperation {
	if tfMap == nil {
		return nil
	}

	createColumnsOperation := &quicksight.CreateColumnsOperation{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
		createColumnsOperation.Columns = expandQuickSightDataSetCalculatedColumns(v)
	}

	return createColumnsOperation
}

func expandQuickSightDataSetCalculatedColumns(tfList []interface{}) []*quicksight.CalculatedColumn {
	if len(tfList) == 0 {
		return nil
	}

	var calculatedColumns []*quicksight.CalculatedColumn

	for _, tfMapRaw := range tfList {

		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		calculatedColumn := expandQuickSightDataSetCalculatedColumn(tfMap)
		if calculatedColumn == nil {
			continue
		}

		calculatedColumns = append(calculatedColumns, calculatedColumn)
	}

	return calculatedColumns
}

func expandQuickSightDataSetCalculatedColumn(tfMap map[string]interface{}) *quicksight.CalculatedColumn {
	if tfMap == nil {
		return nil
	}

	calculatedColumn := &quicksight.CalculatedColumn{}

	if v, ok := tfMap["column_id"].(string); ok {
		calculatedColumn.ColumnId = aws.String(v)
	}

	if v, ok := tfMap["column_name"].(string); ok {
		calculatedColumn.ColumnName = aws.String(v)
	}

	if v, ok := tfMap["expression"].(string); ok {
		calculatedColumn.Expression = aws.String(v)
	}

	return calculatedColumn
}

func expandQuickSightDataSetFilterOperation(tfMap map[string]interface{}) *quicksight.FilterOperation {
	if tfMap == nil {
		return nil
	}

	filterOperation := &quicksight.FilterOperation{}

	if v, ok := tfMap["condition_expression"].(string); ok {
		filterOperation.ConditionExpression = aws.String(v)
	}

	return filterOperation
}

func expandQuickSightDataSetProjectOperation(tfMap map[string]interface{}) *quicksight.ProjectOperation {
	if tfMap == nil {
		return nil
	}

	projectOperation := &quicksight.ProjectOperation{}

	if v, ok := tfMap["projected_columns"].([]string); ok && len(v) > 0 {
		projectOperation.ProjectedColumns = aws.StringSlice(v)
	}

	return projectOperation
}

func expandQuickSightDataSetRenameColumnOperation(tfMap map[string]interface{}) *quicksight.RenameColumnOperation {
	if tfMap == nil {
		return nil
	}

	renameColumnOperation := &quicksight.RenameColumnOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		renameColumnOperation.ColumnName = aws.String(v)
	}

	if v, ok := tfMap["new_column_name"].(string); ok {
		renameColumnOperation.NewColumnName = aws.String(v)
	}

	return renameColumnOperation
}

func expandQuickSightDataSetTagColumnOperation(tfMap map[string]interface{}) *quicksight.TagColumnOperation {
	if tfMap == nil {
		return nil
	}

	tagColumnOperation := &quicksight.TagColumnOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		tagColumnOperation.ColumnName = aws.String(v)
	}

	if v, ok := tfMap["tags"].([]interface{}); ok {
		tagColumnOperation.Tags = expandQuickSightDataSetTags(v)
	}

	return tagColumnOperation
}

func expandQuickSightDataSetTags(tfList []interface{}) []*quicksight.ColumnTag {
	if len(tfList) == 0 {
		return nil
	}

	var tags []*quicksight.ColumnTag

	for _, tfMapRaw := range tfList {

		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		tag := expandQuickSightDataSetTag(tfMap)
		if tag == nil {
			continue
		}

		tags = append(tags, tag)
	}

	return tags
}

func expandQuickSightDataSetTag(tfMap map[string]interface{}) *quicksight.ColumnTag {
	if tfMap == nil {
		return nil
	}

	tag := &quicksight.ColumnTag{}

	if v, ok := tfMap["column_description"].(map[string]interface{}); ok {
		tag.ColumnDescription = expandQuickSightDataSetColumnDescription(v)
	}

	if v, ok := tfMap["column_geographic_role"].(string); ok {
		tag.ColumnGeographicRole = aws.String(v)
	}

	return tag
}

func expandQuickSightDataSetColumnDescription(tfMap map[string]interface{}) *quicksight.ColumnDescription {
	if tfMap == nil {
		return nil
	}

	columnDescription := &quicksight.ColumnDescription{}

	if v, ok := tfMap["text"].(string); ok {
		columnDescription.Text = aws.String(v)
	}

	return columnDescription
}

func expandQuickSightDataSetUntagColumnOperation(tfMap map[string]interface{}) *quicksight.UntagColumnOperation {
	if tfMap == nil {
		return nil
	}

	untagColumnOperation := &quicksight.UntagColumnOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		untagColumnOperation.ColumnName = aws.String(v)
	}

	if v, ok := tfMap["tag_names"].([]string); ok {
		untagColumnOperation.TagNames = aws.StringSlice(v)
	}

	return untagColumnOperation
}

func expandQuickSightDataSetPhysicalTableMap(tfMap map[string]interface{}) map[string]*quicksight.PhysicalTable {
	if len(tfMap) == 0 {
		return nil
	}

	physicalTableMap := make(map[string]*quicksight.PhysicalTable)
	for k, v := range tfMap {

		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		physicalTable := &quicksight.PhysicalTable{}

		if v, ok := vMap["custom_sql"].(map[string]interface{}); ok {
			physicalTable.CustomSql = expandQuickSightDataSetCustomSql(v)
		}

		if v, ok := vMap["relational_table"].(map[string]interface{}); ok {
			physicalTable.RelationalTable = expandQuickSightDataSetRelationalTable(v)
		}

		if v, ok := vMap["s3_source"].(map[string]interface{}); ok {
			physicalTable.S3Source = expandQuickSightDataSetS3Source(v)
		}

		physicalTableMap[k] = physicalTable
	}

	return physicalTableMap
}

func expandQuickSightDataSetCustomSql(tfMap map[string]interface{}) *quicksight.CustomSql {
	if tfMap == nil {
		return nil
	}

	customSql := &quicksight.CustomSql{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
		customSql.Columns = expandQuickSightDataSetInputColumns(v)
	}

	if v, ok := tfMap["data_source_arn"].(string); ok {
		customSql.DataSourceArn = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok {
		customSql.Name = aws.String(v)
	}

	if v, ok := tfMap["sql_query"].(string); ok {
		customSql.SqlQuery = aws.String(v)
	}

	return customSql
}

func expandQuickSightDataSetInputColumns(tfList []interface{}) []*quicksight.InputColumn {
	if len(tfList) == 0 {
		return nil
	}

	var inputColumns []*quicksight.InputColumn

	for _, tfMapRaw := range tfList {

		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		inputColumn := expandQuickSightDataSetInputColumn(tfMap)
		if inputColumn == nil {
			continue
		}

		inputColumns = append(inputColumns, inputColumn)
	}
	return inputColumns
}

func expandQuickSightDataSetInputColumn(tfMap map[string]interface{}) *quicksight.InputColumn {
	if tfMap == nil {
		return nil
	}

	inputColumn := &quicksight.InputColumn{}

	if v, ok := tfMap["name"].(string); ok {
		inputColumn.Name = aws.String(v)
	}
	if v, ok := tfMap["type"].(string); ok {
		inputColumn.Type = aws.String(v)
	}

	return inputColumn
}

func expandQuickSightDataSetRelationalTable(tfMap map[string]interface{}) *quicksight.RelationalTable {
	if tfMap == nil {
		return nil
	}

	relationalTable := &quicksight.RelationalTable{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
		relationalTable.InputColumns = expandQuickSightDataSetInputColumns(v)
	}

	if v, ok := tfMap["catelog"].(string); ok {
		relationalTable.Catalog = aws.String(v)
	}

	if v, ok := tfMap["data_source_arn"].(string); ok {
		relationalTable.DataSourceArn = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok {
		relationalTable.Name = aws.String(v)
	}

	if v, ok := tfMap["catelog"].(string); ok {
		relationalTable.Catalog = aws.String(v)
	}

	return relationalTable
}

func expandQuickSightDataSetS3Source(tfMap map[string]interface{}) *quicksight.S3Source {
	if tfMap == nil {
		return nil
	}

	s3Source := &quicksight.S3Source{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
		s3Source.InputColumns = expandQuickSightDataSetInputColumns(v)
	}

	if v, ok := tfMap["upload_settings"].(map[string]interface{}); ok {
		s3Source.UploadSettings = expandQuickSightDataSetUploadSettings(v)
	}

	if v, ok := tfMap["data_source arn"].(string); ok {
		s3Source.DataSourceArn = aws.String(v)
	}

	return s3Source
}

func expandQuickSightDataSetUploadSettings(tfMap map[string]interface{}) *quicksight.UploadSettings {
	if tfMap == nil {
		return nil
	}

	uploadSettings := &quicksight.UploadSettings{}

	if v, ok := tfMap["contains_header"].(bool); ok {
		uploadSettings.ContainsHeader = aws.Bool(v)
	}

	if v, ok := tfMap["delimiter"].(string); ok {
		uploadSettings.Delimiter = aws.String(v)
	}

	if v, ok := tfMap["format"].(string); ok {
		uploadSettings.Format = aws.String(v)
	}

	if v, ok := tfMap["start_from_row"].(int64); ok {
		uploadSettings.StartFromRow = aws.Int64(v)
	}

	if v, ok := tfMap["text_qualifier"].(string); ok {
		uploadSettings.TextQualifier = aws.String(v)
	}

	return uploadSettings
}

func expandQuickSightDataSetResourcePermissions(tfList []interface{}) []*quicksight.ResourcePermission {
	if len(tfList) == 0 {
		return nil
	}

	var resourcePermissions []*quicksight.ResourcePermission

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		resourcePermission := expandQuickSightDataSetResourcePermission(tfMap)

		if resourcePermission == nil {
			continue
		}

		resourcePermissions = append(resourcePermissions, resourcePermission)
	}

	return resourcePermissions
}

func expandQuickSightDataSetResourcePermission(tfMap map[string]interface{}) *quicksight.ResourcePermission {
	if tfMap == nil {
		return nil
	}

	resourcePermission := &quicksight.ResourcePermission{}

	if v, ok := tfMap["actions"].([]string); ok {
		resourcePermission.Actions = aws.StringSlice(v)
	}

	if v, ok := tfMap["permission_policy"].(string); ok {
		resourcePermission.Principal = aws.String(v)
	}

	return resourcePermission
}

func expandQuickSightDataSetRowLevelPermissionDataSet(tfMap map[string]interface{}) *quicksight.RowLevelPermissionDataSet {
	if tfMap == nil {
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

func expandQuickSightDataSetRowLevelPermissionTagConfigurations(tfMap map[string]interface{}) *quicksight.RowLevelPermissionTagConfiguration {
	if tfMap == nil {
		return nil
	}

	rowLevelPermissionTagConfiguration := &quicksight.RowLevelPermissionTagConfiguration{}

	if v, ok := tfMap["tag_rules"].([]interface{}); ok {
		rowLevelPermissionTagConfiguration.TagRules = expandQuickSightDataSetTagRules(v)
	}

	if v, ok := tfMap["status"].(string); ok {
		rowLevelPermissionTagConfiguration.Status = aws.String(v)
	}

	return rowLevelPermissionTagConfiguration
}

func expandQuickSightDataSetTagRules(tfList []interface{}) []*quicksight.RowLevelPermissionTagRule {
	if len(tfList) == 0 {
		return nil
	}

	var tagRules []*quicksight.RowLevelPermissionTagRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		tagRule := expandQuickSightDataSetTagRule(tfMap)

		if tagRule == nil {
			continue
		}

		tagRules = append(tagRules, tagRule)
	}

	return tagRules
}

func expandQuickSightDataSetTagRule(tfMap map[string]interface{}) *quicksight.RowLevelPermissionTagRule {
	if tfMap == nil {
		return nil
	}

	tagRules := &quicksight.RowLevelPermissionTagRule{}

	if v, ok := tfMap["column_name"].(string); ok {
		tagRules.ColumnName = aws.String(v)
	}

	if v, ok := tfMap["tag_key"].(string); ok {
		tagRules.TagKey = aws.String(v)
	}

	if v, ok := tfMap["match_all_value"].(string); ok {
		tagRules.MatchAllValue = aws.String(v)
	}

	if v, ok := tfMap["tag_multi_value_delimiter"].(string); ok {
		tagRules.TagMultiValueDelimiter = aws.String(v)
	}

	return tagRules
}
