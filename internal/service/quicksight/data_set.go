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
		params.ColumnGroups = expandQuickSightDataSetColumnGroups(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("column_level_permission_rules"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.ColumnLevelPermissionRules = expandQuickSightDataSetColumnLevelPermissionRules(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("data_set_usage_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.DataSetUsageConfiguration = expandQuickSightDataSetUsageConfiguration(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("field_folders"); ok && len(v.(map[string]interface{})) != 0 {
		params.FieldFolders = expandQuickSightDataSourceFieldFolders(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("logical_table_map"); ok && len(v.(map[string]interface{})) != 0 {
		params.LogicalTableMap = expandQuickSightDataSourceLogicalTableMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("permissions"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.Permissions = expandQuickSightDataSetResourcePermissions(v.(map[string]interface{}))
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

func expandQuickSightDataSetPhysicalTableMap(tfMap map[string]interface{}) map[string]*quicksight.PhysicalTable {
	if len(tfMap) == 0 {
		return nil
	}

	physicalTableMap := make(map[string]*quicksight.PhysicalTable)
	for key, value := range tfMap {

		ptm, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		physicalTable := &quicksight.PhysicalTable{}

		if csm, ok := ptm["custom_sql"].(map[string]interface{}); ok {

			customSql := &quicksight.CustomSql{}

			if columnsMap, ok := csm["columns"].(map[string]interface{}); ok && len(columnsMap) > 0 {
				expandInputColumns(columnsMap)
			}

			if v, ok := csm["data_source_arn"].(string); ok {
				customSql.DataSourceArn = aws.String(v)
			}

			if v, ok := csm["name"].(string); ok {
				customSql.Name = aws.String(v)
			}

			if v, ok := csm["sql_query"].(string); ok {
				customSql.SqlQuery = aws.String(v)
			}

			physicalTable.CustomSql = customSql
		}

		if rtm, ok := ptm["relational_table"].(map[string]interface{}); ok {

			relationalTable := &quicksight.RelationalTable{}

			if columnsMap, ok := rtm["columns"].(map[string]interface{}); ok && len(columnsMap) > 0 {
				relationalTable.InputColumns = expandInputColumns(columnsMap)
			}

			if v, ok := rtm["catelog"].(string); ok {
				relationalTable.Catalog = aws.String(v)
			}

			if v, ok := rtm["data_source_arn"].(string); ok {
				relationalTable.DataSourceArn = aws.String(v)
			}

			if v, ok := rtm["name"].(string); ok {
				relationalTable.Name = aws.String(v)
			}

			if v, ok := rtm["catelog"].(string); ok {
				relationalTable.Catalog = aws.String(v)
			}

			physicalTable.RelationalTable = relationalTable
		}

		if ssm, ok := ptm["s3_source"].(map[string]interface{}); ok {

			s3Source := &quicksight.S3Source{}

			if columnsMap, ok := ssm["columns"].(map[string]interface{}); ok && len(columnsMap) > 0 {
				s3Source.InputColumns = expandInputColumns(columnsMap)
			}

			if uploadSettingsMap, ok := ssm["upload_settings"].(map[string]interface{}); ok {

				uploadSettings := &quicksight.UploadSettings{}

				if v, ok := uploadSettingsMap["contains_header"].(bool); ok {
					uploadSettings.ContainsHeader = aws.Bool(v)
				}

				if v, ok := uploadSettingsMap["delimiter"].(string); ok {
					uploadSettings.Delimiter = aws.String(v)
				}

				if v, ok := uploadSettingsMap["format"].(string); ok {
					uploadSettings.Format = aws.String(v)
				}

				if v, ok := uploadSettingsMap["start_from_row"].(int64); ok {
					uploadSettings.StartFromRow = aws.Int64(v)
				}

				if v, ok := uploadSettingsMap["text_qualifier"].(string); ok {
					uploadSettings.TextQualifier = aws.String(v)
				}

				s3Source.UploadSettings = uploadSettings
			}

			if v, ok := ssm["data_source arn"].(string); ok {
				s3Source.DataSourceArn = aws.String(v)
			}

			physicalTable.S3Source = s3Source
		}

		physicalTableMap[key] = physicalTable
	}

	return physicalTableMap
}

func expandInputColumns(tfMap map[string]interface{}) []*quicksight.InputColumn {

	var inputColumns []*quicksight.InputColumn
	for _, columnMap := range tfMap {
		cm, ok := columnMap.(map[string]interface{})
		if !ok {
			continue
		}

		inputColumn := &quicksight.InputColumn{}

		if v, ok := cm["name"].(string); ok {
			inputColumn.Name = aws.String(v)
		}
		if v, ok := cm["type"].(string); ok {
			inputColumn.Type = aws.String(v)
		}

		inputColumns = append(inputColumns, inputColumn)
	}
	return inputColumns
}

func expandQuickSightDataSetColumnGroups(tfMap map[string]interface{}) []*quicksight.ColumnGroup {
	if len(tfMap) == 0 {
		return nil
	}

	var columnGroups []*quicksight.ColumnGroup

	for _, tfMapRaw := range tfMap {
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

	if v, ok := tfMap["geo_spacial_column_group"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			geoSpactialColumnGroup := &quicksight.GeoSpatialColumnGroup{}

			if v, ok := m["columns"].([]string); ok {
				geoSpactialColumnGroup.Columns = aws.StringSlice(v)
			}

			if v, ok := m["country_code"].(string); ok {
				geoSpactialColumnGroup.CountryCode = aws.String(v)
			}

			if v, ok := m["name"].(string); ok {
				geoSpactialColumnGroup.Name = aws.String(v)
			}
		}
	}

	return columnGroup
}

func expandQuickSightDataSetColumnLevelPermissionRules(tfMap map[string]interface{}) []*quicksight.ColumnLevelPermissionRule {
	if len(tfMap) == 0 {
		return nil
	}

	var columnLevelPermissionRules []*quicksight.ColumnLevelPermissionRule

	for _, tfMapRaw := range tfMap {
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

func expandQuickSightDataSourceFieldFolders(tfMap map[string]interface{}) map[string]*quicksight.FieldFolder {
	if len(tfMap) == 0 {
		return nil
	}

	fieldFolderMap := make(map[string]*quicksight.FieldFolder)
	for k, v := range tfMap {

		x, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		fieldFolder := &quicksight.FieldFolder{}

		if y, ok := x["columns"].([]string); ok {
			fieldFolder.Columns = aws.StringSlice(y)
		}

		if y, ok := x["description"].(string); ok {
			fieldFolder.Description = aws.String(y)
		}

		fieldFolderMap[k] = fieldFolder
	}

	return fieldFolderMap
}

func expandQuickSightDataSourceLogicalTableMap(tfMap map[string]interface{}) map[string]*quicksight.LogicalTable {
	if len(tfMap) == 0 {
		return nil
	}

	logicalTableMap := make(map[string]*quicksight.LogicalTable)
	for key, value := range tfMap {

		ltm, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		logicalTable := &quicksight.LogicalTable{}

		if v, ok := ltm["alias"].(string); ok {
			logicalTable.Alias = aws.String(v)
		}

		if ltsm, ok := ltm["logical_table_source"].(map[string]interface{}); ok {

			logicalTableSource := &quicksight.LogicalTableSource{}

			if v, ok := ltsm["data_set_arn"].(string); ok {
				logicalTableSource.DataSetArn = aws.String(v)
			}

			if v, ok := ltsm["physical_table_id"].(string); ok {
				logicalTableSource.PhysicalTableId = aws.String(v)
			}

			if jim := ltsm["join_instruction"].(map[string]interface{}); ok {

				joinInstruction := &quicksight.JoinInstruction{}

				if v, ok := jim["left_operand"].(string); ok {
					joinInstruction.LeftOperand = aws.String(v)
				}

				if v, ok := jim["on_clause"].(string); ok {
					joinInstruction.OnClause = aws.String(v)
				}

				if v, ok := jim["right_operand"].(string); ok {
					joinInstruction.RightOperand = aws.String(v)
				}

				if v, ok := jim["type"].(string); ok {
					joinInstruction.Type = aws.String(v)
				}

				if ljkpm, ok := jim["left_join_key_properties"].(map[string]interface{}); ok {

					leftJoinKeyProperties := &quicksight.JoinKeyProperties{}

					if v, ok := ljkpm["unique_key"].(bool); ok {
						leftJoinKeyProperties.UniqueKey = aws.Bool(v)
					}

					joinInstruction.LeftJoinKeyProperties = leftJoinKeyProperties
				}

				if rjkpm, ok := jim["right_join_key_properties"].(map[string]interface{}); ok {

					rightJoinKeyProperties := &quicksight.JoinKeyProperties{}

					if v, ok := rjkpm["unique_key"].(bool); ok {
						rightJoinKeyProperties.UniqueKey = aws.Bool(v)
					}

					joinInstruction.RightJoinKeyProperties = rightJoinKeyProperties
				}

				logicalTableSource.JoinInstruction = joinInstruction
			}

			logicalTable.Source = logicalTableSource
		}

		if dtm, ok := ltm["data_transforms"].([]map[string]interface{}); ok {

			var transformOperations []*quicksight.TransformOperation

			for _, transformOperationMap := range dtm {

				transformOperation := &quicksight.TransformOperation{}

				if cctom, ok := transformOperationMap["cast_column_type_operation"].(map[string]interface{}); ok {

					castColumnTypeOperation := &quicksight.CastColumnTypeOperation{}

					if v, ok := cctom["column_name"].(string); ok {
						castColumnTypeOperation.ColumnName = aws.String(v)
					}

					if v, ok := cctom["new_column_type"].(string); ok {
						castColumnTypeOperation.NewColumnType = aws.String(v)
					}

					if v, ok := cctom["format"].(string); ok {
						castColumnTypeOperation.Format = aws.String(v)
					}

					transformOperation.CastColumnTypeOperation = castColumnTypeOperation
				}

				if ccom, ok := transformOperationMap["create_columns_operation"].(map[string]interface{}); ok {

					createColumnsOperation := &quicksight.CreateColumnsOperation{}

					if columnsMap, ok := ccom["columns"].([]map[string]interface{}); ok {

						var columns []*quicksight.CalculatedColumn

						for _, columnMap := range columnsMap {

							column := &quicksight.CalculatedColumn{}

							if v, ok := columnMap["column_id"].(string); ok {
								column.ColumnId = aws.String(v)
							}

							if v, ok := columnMap["column_name"].(string); ok {
								column.ColumnName = aws.String(v)
							}

							if v, ok := columnMap["expression"].(string); ok {
								column.Expression = aws.String(v)
							}

							columns = append(columns, column)
						}

						createColumnsOperation.Columns = columns
					}

					transformOperation.CreateColumnsOperation = createColumnsOperation
				}

				if fom, ok := transformOperationMap["filter_operation"].(map[string]interface{}); ok {

					filterOperation := &quicksight.FilterOperation{}

					if v, ok := fom["condition_expression"].(string); ok {
						filterOperation.ConditionExpression = aws.String(v)
					}

					transformOperation.FilterOperation = filterOperation
				}

				if pom, ok := transformOperationMap["project_operation"].(map[string]interface{}); ok {

					projectOperation := &quicksight.ProjectOperation{}

					if v, ok := pom["projected_columns"].([]string); ok && len(v) > 0 {
						projectOperation.ProjectedColumns = aws.StringSlice(v)
					}

					transformOperation.ProjectOperation = projectOperation
				}

				if rcom, ok := transformOperationMap["rename_column_operation"].(map[string]interface{}); ok {

					renameColumnOperation := &quicksight.RenameColumnOperation{}

					if v, ok := rcom["column_name"].(string); ok {
						renameColumnOperation.ColumnName = aws.String(v)
					}

					if v, ok := rcom["new_column_name"].(string); ok {
						renameColumnOperation.NewColumnName = aws.String(v)
					}

					transformOperation.RenameColumnOperation = renameColumnOperation
				}

				if tcom, ok := transformOperationMap["tag_column_operation"].(map[string]interface{}); ok {

					tagColumnOperation := &quicksight.TagColumnOperation{}

					if v, ok := tcom["column_name"].(string); ok {
						tagColumnOperation.ColumnName = aws.String(v)
					}

					if tagsMap, ok := tcom["tags"].([]map[string]interface{}); ok {

						var tags []*quicksight.ColumnTag

						for _, tagMap := range tagsMap {

							tag := &quicksight.ColumnTag{}

							if cdm, ok := tagMap["column_description"].(map[string]interface{}); ok {

								columnDescription := &quicksight.ColumnDescription{}

								if v, ok := cdm["text"].(string); ok {
									columnDescription.Text = aws.String(v)
								}

								tag.ColumnDescription = columnDescription
							}

							if v, ok := tagMap["column_geographic_role"].(string); ok {
								tag.ColumnGeographicRole = aws.String(v)
							}

							tags = append(tags, tag)
						}

						tagColumnOperation.Tags = tags
					}

					transformOperation.TagColumnOperation = tagColumnOperation
				}

				if ucom, ok := transformOperationMap["untag_column_operation"].(map[string]interface{}); ok {

					untagColumnOperation := &quicksight.UntagColumnOperation{}

					if v, ok := ucom["column_name"].(string); ok {
						untagColumnOperation.ColumnName = aws.String(v)
					}

					if v, ok := ucom["tag_names"].([]string); ok {
						untagColumnOperation.TagNames = aws.StringSlice(v)
					}

					transformOperation.UntagColumnOperation = untagColumnOperation
				}

				transformOperations = append(transformOperations, transformOperation)
			}

			logicalTable.DataTransforms = transformOperations
		}

		logicalTableMap[key] = logicalTable
	}

	return logicalTableMap
}

func expandQuickSightDataSetResourcePermissions(tfMap map[string]interface{}) []*quicksight.ResourcePermission {
	if len(tfMap) == 0 {
		return nil
	}

	var resourcePermissions []*quicksight.ResourcePermission

	for _, tfMapRaw := range tfMap {
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
	if len(tfMap) == 0 {
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
	if len(tfMap) == 0 {
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
	if len(tfMap) == 0 {
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
