package quicksight

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDataSet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataSetCreate,
		ReadContext:   resourceDataSetRead,
		UpdateContext: resourceDataSetUpdate,
		DeleteContext: resourceDataSetDelete,

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
			},

			"field_folders": {
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
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
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
											},
										},
									},
									"physical_table_id": {
										Type:         schema.TypeString,
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
								},
							},
						},
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
													Computed:     true,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 32),
												},
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
										Computed: true,
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
										Computed: true,
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
																Computed: true,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"text": {
																			Type:         schema.TypeString,
																			Computed:     true,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 500),
																		},
																	},
																},
															},
															"column_geographic_role": {
																Type:         schema.TypeString,
																Computed:     true,
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
										Computed: true,
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
													Type:     schema.TypeList,
													Required: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
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
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 32,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"physical_table_map_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"custom_sql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_source_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
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
										ValidateFunc: verify.ValidARN,
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
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
									"catalog": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 256),
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
							Computed: true,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_source_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
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
										Computed: true,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"contains_header": {
													Type:     schema.TypeBool,
													Computed: true,
													Optional: true,
												},
												"delimiter": {
													Type:         schema.TypeString,
													Computed:     true,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 1),
												},
												"format": {
													Type:         schema.TypeString,
													Computed:     true,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(quicksight.FileFormat_Values(), false),
												},
												"start_from_row": {
													Type:         schema.TypeInt,
													Computed:     true,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"text_qualifier": {
													Type:         schema.TypeString,
													Computed:     true,
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
							ValidateFunc: verify.ValidARN,
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

			"row_level_permission_tag_configuration": {
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
										ValidateFunc: validation.StringLenBetween(1, 10),
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

func resourceDataSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	awsAccountId := meta.(*conns.AWSClient).AccountID
	id := d.Get("data_set_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", awsAccountId, id))

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountId = v.(string)
	}
	params := &quicksight.CreateDataSetInput{
		AwsAccountId:     aws.String(awsAccountId),
		DataSetId:        aws.String(id),
		ImportMode:       aws.String(d.Get("import_mode").(string)),
		PhysicalTableMap: expandDataSetPhysicalTableMap(d.Get("physical_table_map").(*schema.Set)),
		Name:             aws.String(d.Get("name").(string)),
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("column_groups"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.ColumnGroups = expandDataSetColumnGroups(v.([]interface{}))
	}

	if v, ok := d.GetOk("column_level_permission_rules"); ok && len(v.([]interface{})) > 0 {
		params.ColumnLevelPermissionRules = expandDataSetColumnLevelPermissionRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("data_set_usage_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.DataSetUsageConfiguration = expandDataSetUsageConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("field_folders"); ok && len(v.([]interface{})) != 0 {
		params.FieldFolders = expandDataSetFieldFolders(v.([]interface{}))
	}

	if v, ok := d.GetOk("logical_table_map"); ok && len(v.([]interface{})) != 0 {
		params.LogicalTableMap = expandDataSetLogicalTableMap(v.([]interface{}))
	}

	if v, ok := d.GetOk("permissions"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.Permissions = expandDataSetPermissions(v.([]interface{}))
	}

	if v, ok := d.GetOk("row_level_permission_data_set"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.RowLevelPermissionDataSet = expandDataSetRowLevelPermissionDataSet(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("row_level_permission_tag_configurations"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.RowLevelPermissionTagConfiguration = expandDataSetRowLevelPermissionTagConfigurations(v.(map[string]interface{}))
	}

	_, err := conn.CreateDataSetWithContext(ctx, params)
	if err != nil {
		return diag.Errorf("error creating QuickSight Data Set: %s", err)
	}

	if _, err := waitDataSetCreated(ctx, conn, awsAccountId, id); err != nil {
		return diag.Errorf("error waiting from QuickSight Data Set (%s) creation: %s", d.Id(), err)
	}
	return resourceDataSetRead(ctx, d, meta)
}

func resourceDataSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// arn.IsARN(d.Id()

	conn := meta.(*conns.AWSClient).QuickSightConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	awsAccountId, dataSetId, err := ParseDataSetID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	descOpts := &quicksight.DescribeDataSetInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSetId:    aws.String(dataSetId),
	}

	output, err := conn.DescribeDataSetWithContext(ctx, descOpts)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] QuickSight Data Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error describing QuickSight Data Set (%s): %s", d.Id(), err)
	}

	if output == nil || output.DataSet == nil {
		return diag.Errorf("error describing QuickSight Data Set (%s): empty output", d.Id())
	}

	dataSet := output.DataSet

	d.Set("arn", dataSet.Arn)
	d.Set("aws_account_id", awsAccountId)
	d.Set("data_set_id", dataSet.DataSetId)
	d.Set("name", dataSet.Name)
	d.Set("import_mode", dataSet.ImportMode)

	if err := d.Set("column_groups", flattenColumnGroups(dataSet.ColumnGroups)); err != nil {
		return diag.Errorf("error setting column_groups: %s", err)
	}

	if err := d.Set("column_level_permission_rules", flattenColumnLevelPermissionRules(dataSet.ColumnLevelPermissionRules)); err != nil {
		return diag.Errorf("error setting column_level_permission_rules: %s", err)
	}

	if err := d.Set("data_set_usage_configuration", flattenDataSetUsageConfiguration(dataSet.DataSetUsageConfiguration)); err != nil {
		return diag.Errorf("error setting data_set_usage_configuration: %s", err)
	}

	if err := d.Set("field_folders", flattenFieldFolders(dataSet.FieldFolders)); err != nil {
		return diag.Errorf("error setting field_folders: %s", err)
	}

	if err := d.Set("logical_table_map", flattenLogicalTableMap(dataSet.LogicalTableMap)); err != nil {
		return diag.Errorf("error setting logical_table_map: %s", err)
	}

	if err := d.Set("physical_table_map", flattenPhysicalTableMap(dataSet.PhysicalTableMap)); err != nil {
		return diag.Errorf("error setting physical_table_map: %s", err)
	}

	if err := d.Set("row_level_permission_data_set", flattenRowLevelPermissionDataSet(dataSet.RowLevelPermissionDataSet)); err != nil {
		return diag.Errorf("error setting row_level_permission_data_set: %s", err)
	}

	if err := d.Set("row_level_permission_tag_configuration", flattenRowLevelPermissionTagConfiguration(dataSet.RowLevelPermissionTagConfiguration)); err != nil {
		return diag.Errorf("error setting row_level_permission_tag_configuration: %s", err)
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("error listing tags for QuickSight Data Set (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	permsResp, err := conn.DescribeDataSetPermissionsWithContext(ctx, &quicksight.DescribeDataSetPermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSetId:    aws.String(dataSetId),
	})

	if err != nil {
		return diag.Errorf("error describing QuickSight Data Source (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set("permissions", flattenQuickSightPermissions(permsResp.Permissions)); err != nil {
		return diag.Errorf("error setting permissions: %s", err)
	}
	return nil
}

func resourceDataSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightConn

	if d.HasChangesExcept("permissions", "tags", "tags_all") {
		awsAccountId, dataSetId, err := ParseDataSetID(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		params := &quicksight.UpdateDataSetInput{
			AwsAccountId: aws.String(awsAccountId),
			DataSetId:    aws.String(dataSetId),
			Name:         aws.String(d.Get("name").(string)),
		}

		if d.HasChange("column_groups") {
			params.ColumnGroups = expandDataSetColumnGroups(d.Get("column_groups").([]interface{}))
		}

		if d.HasChange("column_level_permission_rules") {
			params.ColumnLevelPermissionRules = expandDataSetColumnLevelPermissionRules(d.Get("column_level_permission_rules").([]interface{}))
		}

		if d.HasChange("data_set_usage_configuration") {
			params.DataSetUsageConfiguration = expandDataSetUsageConfiguration(d.Get("data_set_usage_configuration").([]interface{}))
		}

		if d.HasChange("field_folders") {
			params.FieldFolders = expandDataSetFieldFolders(d.Get("field_folders").([]interface{}))
		}

		if d.HasChange("import_mode") {
			params.ImportMode = aws.String(d.Get("import_mode").(string))
		}

		if d.HasChange("logical_table_map") {
			params.LogicalTableMap = expandDataSetLogicalTableMap(d.Get("logical_table_map").([]interface{}))
		}

		if d.HasChange("physical_table_map") {
			params.PhysicalTableMap = expandDataSetPhysicalTableMap(d.Get("physical_table_map").(*schema.Set))
		}

		if d.HasChange("row_level_permission_data_set") {
			params.RowLevelPermissionDataSet = expandDataSetRowLevelPermissionDataSet(d.Get("row_level_permission_data_set").(map[string]interface{}))
		}

		if d.HasChange("row_level_permission_tag_configuration") {
			params.RowLevelPermissionTagConfiguration = expandDataSetRowLevelPermissionTagConfigurations(d.Get("row_level_permission_tag_configuration").(map[string]interface{}))
		}

		_, err = conn.UpdateDataSetWithContext(ctx, params)
		if err != nil {
			return diag.Errorf("error updating QuickSight Data Set (%s): %s", d.Id(), err)
		}

		// dataSet doesnt have a status, don't know what to do without a status function

		// if _, err := waitUpdated(ctx, conn, awsAccountId, dataSourceId); err != nil {
		// 	return diag.Errorf("error waiting for QuickSight Data Set (%s) to update: %s", d.Id(), err)
		// }
	}

	if d.HasChange("permissions") {
		awsAccountId, dataSetId, err := ParseDataSetID(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		oraw, nraw := d.GetChange("permissions")
		o := oraw.(*schema.Set).List()
		n := nraw.(*schema.Set).List()

		toGrant, toRevoke := DiffPermissions(o, n)

		params := &quicksight.UpdateDataSetPermissionsInput{
			AwsAccountId: aws.String(awsAccountId),
			DataSetId:    aws.String(dataSetId),
		}

		if len(toGrant) > 0 {
			params.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			params.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateDataSetPermissionsWithContext(ctx, params)

		if err != nil {
			return diag.Errorf("error updating QuickSight Data Set (%s) permissions: %s", dataSetId, err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating QuickSight Data Source (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceDataSetRead(ctx, d, meta)
}

func resourceDataSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightConn

	awsAccountId, dataSetId, err := ParseDataSetID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deleteOpts := &quicksight.DeleteDataSetInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSetId:    aws.String(dataSetId),
	}

	_, err = conn.DeleteDataSetWithContext(ctx, deleteOpts)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting QuickSight Data Set (%s): %s", d.Id(), err)
	}

	return nil
}

func expandDataSetColumnGroups(tfList []interface{}) []*quicksight.ColumnGroup {
	if len(tfList) == 0 {
		return nil
	}

	var columnGroups []*quicksight.ColumnGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		columnGroup := expandDataSetColumnGroup(tfMap)

		if columnGroup == nil {
			continue
		}

		columnGroups = append(columnGroups, columnGroup)
	}

	return columnGroups
}

func expandDataSetColumnGroup(tfMap map[string]interface{}) *quicksight.ColumnGroup {
	if len(tfMap) == 0 {
		return nil
	}

	columnGroup := &quicksight.ColumnGroup{}
	if tfMapRaw, ok := tfMap["geo_spatial_column_group"].([]interface{}); ok {
		columnGroup.GeoSpatialColumnGroup = expandDataSetGeoSpatialColumnGroup(tfMapRaw[0].(map[string]interface{}))
	}

	return columnGroup
}

func expandDataSetGeoSpatialColumnGroup(tfMap map[string]interface{}) *quicksight.GeoSpatialColumnGroup {
	if tfMap == nil {
		return nil
	}

	geoSpatialColumnGroup := &quicksight.GeoSpatialColumnGroup{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
		geoSpatialColumnGroup.Columns = flex.ExpandStringList(v)
	}

	if v, ok := tfMap["country_code"].(string); ok && v != "" {
		geoSpatialColumnGroup.CountryCode = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		geoSpatialColumnGroup.Name = aws.String(v)
	}

	return geoSpatialColumnGroup
}

func expandDataSetColumnLevelPermissionRules(tfList []interface{}) []*quicksight.ColumnLevelPermissionRule {
	if len(tfList) == 0 {
		return nil
	}

	var columnLevelPermissionRules []*quicksight.ColumnLevelPermissionRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		var columnLevelPermissionRule *quicksight.ColumnLevelPermissionRule

		if v, ok := tfMap["column_names"].([]interface{}); ok {
			columnLevelPermissionRule.ColumnNames = flex.ExpandStringList(v)
		}

		if v, ok := tfMap["principals"].([]interface{}); ok {
			columnLevelPermissionRule.ColumnNames = flex.ExpandStringList(v)
		}
	}

	return columnLevelPermissionRules
}

func expandDataSetUsageConfiguration(tfList []interface{}) *quicksight.DataSetUsageConfiguration {
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

func expandDataSetFieldFolders(tfList []interface{}) map[string]*quicksight.FieldFolder {
	if len(tfList) == 0 {
		return nil
	}

	fieldFolderMap := make(map[string]*quicksight.FieldFolder)
	for _, v := range tfList {

		tfMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		fieldFolder := &quicksight.FieldFolder{}

		if v, ok := tfMap["columns"].([]interface{}); ok {
			var fin []string
			for _, str := range v {
				fin = append(fin, str.(string))
			}

			fieldFolder.Columns = aws.StringSlice(fin)
		}

		if v, ok := tfMap["description"].(string); ok {
			fieldFolder.Description = aws.String(v)
		}

		fieldFolderID := tfMap["field_folders_id"].(string)
		fieldFolderMap[fieldFolderID] = fieldFolder
	}

	return fieldFolderMap
}

func expandDataSetLogicalTableMap(tfList []interface{}) map[string]*quicksight.LogicalTable {
	if len(tfList) == 0 {
		return nil
	}

	logicalTableMap := make(map[string]*quicksight.LogicalTable)
	for _, v := range tfList {

		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		logicalTable := &quicksight.LogicalTable{}

		if v, ok := vMap["alias"].(string); ok {
			logicalTable.Alias = aws.String(v)
		}

		if v, ok := vMap["source"].([]interface{}); ok {
			logicalTable.Source = expandDataSetLogicalTableSource(v[0].(map[string]interface{}))
		}

		if v, ok := vMap["data_transforms"].([]interface{}); ok {
			logicalTable.DataTransforms = expandDataSetDataTransforms(v)
		}

		logicalTableMap["s3PhysicalTable"] = logicalTable
	}

	return logicalTableMap
}

func expandDataSetLogicalTableSource(tfMap map[string]interface{}) *quicksight.LogicalTableSource {
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
		logicalTableSource.JoinInstruction = expandDataSetJoinInstruction(v)
	}

	return logicalTableSource
}

func expandDataSetJoinInstruction(tfMap map[string]interface{}) *quicksight.JoinInstruction {
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
		joinInstruction.LeftJoinKeyProperties = expandDataSetJoinKeyProperties(v)
	}

	if v, ok := tfMap["right_join_key_properties"].(map[string]interface{}); ok {
		joinInstruction.RightJoinKeyProperties = expandDataSetJoinKeyProperties(v)
	}

	return joinInstruction
}

func expandDataSetJoinKeyProperties(tfMap map[string]interface{}) *quicksight.JoinKeyProperties {
	if tfMap == nil {
		return nil
	}

	joinKeyProperties := &quicksight.JoinKeyProperties{}

	if v, ok := tfMap["unique_key"].(bool); ok {
		joinKeyProperties.UniqueKey = aws.Bool(v)
	}

	return joinKeyProperties
}

func expandDataSetDataTransforms(tfList []interface{}) []*quicksight.TransformOperation {
	if len(tfList) == 0 {
		return nil
	}

	var transformOperations []*quicksight.TransformOperation

	for _, tfMapRaw := range tfList {

		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		transformOperation := expandDataSetDataTransform(tfMap)

		if transformOperation == nil {
			continue
		}

		transformOperations = append(transformOperations, transformOperation)

	}

	return transformOperations
}

func expandDataSetDataTransform(tfMap map[string]interface{}) *quicksight.TransformOperation {
	if tfMap == nil {
		return nil
	}

	transformOperation := &quicksight.TransformOperation{}

	if v, ok := tfMap["cast_column_type_operation"].(map[string]interface{}); ok {
		transformOperation.CastColumnTypeOperation = expandDataSetCastColumnTypeOperation(v)
	}

	if v, ok := tfMap["create_columns_operation"].(map[string]interface{}); ok {
		transformOperation.CreateColumnsOperation = expandDataSetCreateColumnsOperation(v)
	}

	if v, ok := tfMap["filter_operation"].(map[string]interface{}); ok {
		transformOperation.FilterOperation = expandDataSetFilterOperation(v)
	}

	if v, ok := tfMap["project_operation"].(map[string]interface{}); ok {
		transformOperation.ProjectOperation = expandDataSetProjectOperation(v)
	}

	if v, ok := tfMap["rename_column_operation"].(map[string]interface{}); ok {
		transformOperation.RenameColumnOperation = expandDataSetRenameColumnOperation(v)
	}

	if v, ok := tfMap["tag_column_operation"].(map[string]interface{}); ok {
		transformOperation.TagColumnOperation = expandDataSetTagColumnOperation(v)
	}

	if v, ok := tfMap["untag_column_operation"].(map[string]interface{}); ok {
		transformOperation.UntagColumnOperation = expandDataSetUntagColumnOperation(v)
	}

	return transformOperation
}

func expandDataSetCastColumnTypeOperation(tfMap map[string]interface{}) *quicksight.CastColumnTypeOperation {
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

func expandDataSetCreateColumnsOperation(tfMap map[string]interface{}) *quicksight.CreateColumnsOperation {
	if tfMap == nil {
		return nil
	}

	createColumnsOperation := &quicksight.CreateColumnsOperation{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
		createColumnsOperation.Columns = expandDataSetCalculatedColumns(v)
	}

	return createColumnsOperation
}

func expandDataSetCalculatedColumns(tfList []interface{}) []*quicksight.CalculatedColumn {
	if len(tfList) == 0 {
		return nil
	}

	var calculatedColumns []*quicksight.CalculatedColumn

	for _, tfMapRaw := range tfList {

		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		calculatedColumn := expandDataSetCalculatedColumn(tfMap)
		if calculatedColumn == nil {
			continue
		}

		calculatedColumns = append(calculatedColumns, calculatedColumn)
	}

	return calculatedColumns
}

func expandDataSetCalculatedColumn(tfMap map[string]interface{}) *quicksight.CalculatedColumn {
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

func expandDataSetFilterOperation(tfMap map[string]interface{}) *quicksight.FilterOperation {
	if tfMap == nil {
		return nil
	}

	filterOperation := &quicksight.FilterOperation{}

	if v, ok := tfMap["condition_expression"].(string); ok {
		filterOperation.ConditionExpression = aws.String(v)
	}

	return filterOperation
}

func expandDataSetProjectOperation(tfMap map[string]interface{}) *quicksight.ProjectOperation {
	if tfMap == nil {
		return nil
	}

	projectOperation := &quicksight.ProjectOperation{}

	if v, ok := tfMap["projected_columns"].([]string); ok && len(v) > 0 {
		projectOperation.ProjectedColumns = aws.StringSlice(v)
	}

	return projectOperation
}

func expandDataSetRenameColumnOperation(tfMap map[string]interface{}) *quicksight.RenameColumnOperation {
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

func expandDataSetTagColumnOperation(tfMap map[string]interface{}) *quicksight.TagColumnOperation {
	if tfMap == nil {
		return nil
	}

	tagColumnOperation := &quicksight.TagColumnOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		tagColumnOperation.ColumnName = aws.String(v)
	}

	if v, ok := tfMap["tags"].([]interface{}); ok {
		tagColumnOperation.Tags = expandDataSetTags(v)
	}

	return tagColumnOperation
}

func expandDataSetTags(tfList []interface{}) []*quicksight.ColumnTag {
	if len(tfList) == 0 {
		return nil
	}

	var tags []*quicksight.ColumnTag

	for _, tfMapRaw := range tfList {

		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		tag := expandDataSetTag(tfMap)
		if tag == nil {
			continue
		}

		tags = append(tags, tag)
	}

	return tags
}

func expandDataSetTag(tfMap map[string]interface{}) *quicksight.ColumnTag {
	if tfMap == nil {
		return nil
	}

	tag := &quicksight.ColumnTag{}

	if v, ok := tfMap["column_description"].(map[string]interface{}); ok {
		tag.ColumnDescription = expandDataSetColumnDescription(v)
	}

	if v, ok := tfMap["column_geographic_role"].(string); ok {
		tag.ColumnGeographicRole = aws.String(v)
	}

	return tag
}

func expandDataSetColumnDescription(tfMap map[string]interface{}) *quicksight.ColumnDescription {
	if tfMap == nil {
		return nil
	}

	columnDescription := &quicksight.ColumnDescription{}

	if v, ok := tfMap["text"].(string); ok {
		columnDescription.Text = aws.String(v)
	}

	return columnDescription
}

func expandDataSetUntagColumnOperation(tfMap map[string]interface{}) *quicksight.UntagColumnOperation {
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

func expandDataSetPhysicalTableMap(tfSet *schema.Set) map[string]*quicksight.PhysicalTable {
	if tfSet.Len() == 0 {
		return nil
	}

	physicalTableMap := make(map[string]*quicksight.PhysicalTable)
	for _, v := range tfSet.List() {

		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		physicalTable := &quicksight.PhysicalTable{}

		physicalTableMapID := vMap["physical_table_map_id"].(string)

		if customSqlList, ok := vMap["custom_sql"].([]interface{}); ok {
			for _, v := range customSqlList {
				physicalTable.CustomSql = expandDataSetCustomSql(v.(map[string]interface{}))
			}
		}

		if relationalTableList, ok := vMap["relational_table"].([]interface{}); ok {
			for _, v := range relationalTableList {
				physicalTable.RelationalTable = expandDataSetRelationalTable(v.(map[string]interface{}))
			}
		}

		if s3SourceList, ok := vMap["s3_source"].([]interface{}); ok {
			for _, v := range s3SourceList {
				physicalTable.S3Source = expandDataSetS3Source(v.(map[string]interface{}))
			}
		}

		physicalTableMap[physicalTableMapID] = physicalTable
	}

	return physicalTableMap
}

func expandDataSetCustomSql(tfMap map[string]interface{}) *quicksight.CustomSql {
	if tfMap == nil {
		return nil
	}

	customSql := &quicksight.CustomSql{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
		customSql.Columns = expandDataSetInputColumns(v)
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

func expandDataSetInputColumns(tfList []interface{}) []*quicksight.InputColumn {
	if len(tfList) == 0 {
		return nil
	}

	var inputColumns []*quicksight.InputColumn

	for _, tfMapRaw := range tfList {

		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		inputColumn := expandDataSetInputColumn(tfMap)
		if inputColumn == nil {
			continue
		}

		inputColumns = append(inputColumns, inputColumn)
	}
	return inputColumns
}

func expandDataSetInputColumn(tfMap map[string]interface{}) *quicksight.InputColumn {
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

func expandDataSetRelationalTable(tfMap map[string]interface{}) *quicksight.RelationalTable {
	if tfMap == nil {
		return nil
	}

	relationalTable := &quicksight.RelationalTable{}

	if v, ok := tfMap["input_columns"].([]interface{}); ok {
		relationalTable.InputColumns = expandDataSetInputColumns(v)
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

func expandDataSetS3Source(tfMap map[string]interface{}) *quicksight.S3Source {
	if tfMap == nil {
		return nil
	}

	s3Source := &quicksight.S3Source{}

	if v, ok := tfMap["input_columns"].([]interface{}); ok {
		s3Source.InputColumns = expandDataSetInputColumns(v)
	}

	if v, ok := tfMap["upload_settings"].(map[string]interface{}); ok {
		s3Source.UploadSettings = expandDataSetUploadSettings(v)
	}

	if v, ok := tfMap["data_source_arn"].(string); ok {
		s3Source.DataSourceArn = aws.String(v)
	}

	return s3Source
}

func expandDataSetUploadSettings(tfMap map[string]interface{}) *quicksight.UploadSettings {
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

func expandDataSetPermissions(tfList []interface{}) []*quicksight.ResourcePermission {
	permissions := make([]*quicksight.ResourcePermission, len(tfList))

	for i, tfListRaw := range tfList {
		tfMap := tfListRaw.(map[string]interface{})

		permission := &quicksight.ResourcePermission{
			Actions:   flex.ExpandStringList(tfMap["actions"].([]interface{})),
			Principal: aws.String(tfMap["principal"].(string)),
		}

		permissions[i] = permission
	}
	return permissions
}

func expandDataSetRowLevelPermissionDataSet(tfMap map[string]interface{}) *quicksight.RowLevelPermissionDataSet {
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

func expandDataSetRowLevelPermissionTagConfigurations(tfMap map[string]interface{}) *quicksight.RowLevelPermissionTagConfiguration {
	if tfMap == nil {
		return nil
	}

	rowLevelPermissionTagConfiguration := &quicksight.RowLevelPermissionTagConfiguration{}

	if v, ok := tfMap["tag_rules"].([]interface{}); ok {
		rowLevelPermissionTagConfiguration.TagRules = expandDataSetTagRules(v)
	}

	if v, ok := tfMap["status"].(string); ok {
		rowLevelPermissionTagConfiguration.Status = aws.String(v)
	}

	return rowLevelPermissionTagConfiguration
}

func expandDataSetTagRules(tfList []interface{}) []*quicksight.RowLevelPermissionTagRule {
	if len(tfList) == 0 {
		return nil
	}

	var tagRules []*quicksight.RowLevelPermissionTagRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		tagRule := expandDataSetTagRule(tfMap)

		if tagRule == nil {
			continue
		}

		tagRules = append(tagRules, tagRule)
	}

	return tagRules
}

func expandDataSetTagRule(tfMap map[string]interface{}) *quicksight.RowLevelPermissionTagRule {
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

func flattenColumnGroups(groups []*quicksight.ColumnGroup) []interface{} {
	if len(groups) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, group := range groups {
		if group == nil {
			continue
		}

		tfList = append(tfList, flattenColumnGroup(group))
	}

	return tfList
}

func flattenColumnGroup(group *quicksight.ColumnGroup) map[string]interface{} {
	if group == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if group.GeoSpatialColumnGroup != nil {
		tfMap["geo_spatial_column_group"] = flattenGeoSpatialColumnGroup(group.GeoSpatialColumnGroup)
	}

	return tfMap
}

func flattenGeoSpatialColumnGroup(group *quicksight.GeoSpatialColumnGroup) map[string]interface{} {
	if group == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if group.Columns != nil {
		tfMap["columns"] = aws.StringValueSlice(group.Columns)
	}

	if group.CountryCode != nil {
		tfMap["country_code"] = aws.StringValue(group.CountryCode)
	}

	if group.Name != nil {
		tfMap["name"] = aws.StringValue(group.Name)
	}

	return tfMap
}

func flattenColumnLevelPermissionRules(rules []*quicksight.ColumnLevelPermissionRule) []interface{} {
	if len(rules) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		tfList = append(tfList, flattenColumnLevelPermissionRule(rule))
	}

	return tfList
}

func flattenColumnLevelPermissionRule(rule *quicksight.ColumnLevelPermissionRule) map[string]interface{} {
	if rule == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if rule.ColumnNames != nil {
		tfMap["column_names"] = aws.StringValueSlice(rule.ColumnNames)
	}

	if rule.Principals != nil {
		tfMap["principals"] = aws.StringValueSlice(rule.Principals)
	}

	return tfMap
}

func flattenDataSetUsageConfiguration(configuration *quicksight.DataSetUsageConfiguration) []interface{} {
	if configuration == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if configuration.DisableUseAsDirectQuerySource != nil {
		tfMap["disable_use_as_direct_query_source"] = aws.BoolValue(configuration.DisableUseAsDirectQuerySource)
	}

	if configuration.DisableUseAsImportedSource != nil {
		tfMap["disable_use_as_imported_source"] = aws.BoolValue(configuration.DisableUseAsImportedSource)
	}

	tfList := []interface{}{tfMap}
	return tfList
}

func flattenFieldFolders(folders map[string]*quicksight.FieldFolder) []interface{} {
	if len(folders) == 0 {
		return nil
	}

	var tfList []interface{}

	for key, value := range folders {
		if value == nil {
			continue
		}

		tfMap := flattenFieldFolder(key, value)
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFieldFolder(fieldFolderId string, fieldFolder *quicksight.FieldFolder) map[string]interface{} {
	if fieldFolder == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["field_folder_id"] = fieldFolderId

	if fieldFolder.Columns != nil {
		tfMap["columns"] = aws.StringValueSlice(fieldFolder.Columns)
	}

	if fieldFolder.Description != nil {
		tfMap["description"] = aws.StringValue(fieldFolder.Description)
	}

	return tfMap
}

func flattenLogicalTableMap(maps map[string]*quicksight.LogicalTable) []interface{} {
	if len(maps) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, table := range maps {
		if table == nil {
			continue
		}

		tfMap := flattenLogicalTable(table)
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenLogicalTable(table *quicksight.LogicalTable) map[string]interface{} {
	if table == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if table.Alias != nil {
		tfMap["alias"] = aws.StringValue(table.Alias)
	}

	if table.DataTransforms != nil {
		tfMap["data_transforms"] = flattenTransformOperations(table.DataTransforms)
	}

	if table.Source != nil {
		tfMap["source"] = flattenLogicalTableSource(table.Source)
	}

	return tfMap
}

func flattenTransformOperations(operations []*quicksight.TransformOperation) interface{} {
	if len(operations) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, operation := range operations {
		if operation == nil {
			continue
		}

		tfList = append(tfList, flattenTransformOperation(operation))
	}

	return tfList
}

func flattenTransformOperation(operation *quicksight.TransformOperation) map[string]interface{} {
	if operation == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if operation.CastColumnTypeOperation != nil {
		tfMap["cast_column_type_operation"] = flattenCastColumnTypeOperation(operation.CastColumnTypeOperation)
	}

	if operation.CreateColumnsOperation != nil {
		tfMap["create_column_operation"] = flattenCreateColumnOperation(operation.CreateColumnsOperation)
	}

	if operation.FilterOperation != nil {
		tfMap["filter_operation"] = flattenFilterOperation(operation.FilterOperation)
	}

	if operation.ProjectOperation != nil {
		tfMap["project_operation"] = flattenProjectOperation(operation.ProjectOperation)
	}

	if operation.RenameColumnOperation != nil {
		tfMap["rename_column_operation"] = flattenRenameColumnOperation(operation.RenameColumnOperation)
	}

	if operation.TagColumnOperation != nil {
		tfMap["tag_column_operation"] = flattenTagColumnOperation(operation.TagColumnOperation)
	}

	if operation.UntagColumnOperation != nil {
		tfMap["untag_column_operation"] = flattenUntagColumnOperation(operation.UntagColumnOperation)
	}

	return tfMap
}

func flattenCastColumnTypeOperation(operation *quicksight.CastColumnTypeOperation) map[string]interface{} {
	if operation == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if operation.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(operation.ColumnName)
	}

	if operation.Format != nil {
		tfMap["cast_column_type_operation"] = aws.StringValue(operation.ColumnName)
	}

	if operation.NewColumnType != nil {
		tfMap["new_column_type"] = aws.StringValue(operation.NewColumnType)
	}

	return tfMap
}

func flattenCreateColumnOperation(operation *quicksight.CreateColumnsOperation) map[string]interface{} {
	if operation == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if operation.Columns != nil {
		tfMap["columns"] = flattenCalculatedColumns(operation.Columns)
	}

	return tfMap
}

func flattenCalculatedColumns(columns []*quicksight.CalculatedColumn) interface{} {
	if len(columns) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, column := range columns {
		if column == nil {
			continue
		}

		tfList = append(tfList, flattenCalculatedColumn(column))
	}

	return tfList
}

func flattenCalculatedColumn(column *quicksight.CalculatedColumn) map[string]interface{} {
	if column == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if column.ColumnId != nil {
		tfMap["column_id"] = aws.StringValue(column.ColumnId)
	}

	if column.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(column.ColumnName)
	}

	if column.Expression != nil {
		tfMap["column_id"] = aws.StringValue(column.Expression)
	}

	return tfMap
}

func flattenFilterOperation(operation *quicksight.FilterOperation) map[string]interface{} {
	if operation == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if operation.ConditionExpression != nil {
		tfMap["condition_expression"] = aws.StringValue(operation.ConditionExpression)
	}

	return tfMap
}

func flattenProjectOperation(operation *quicksight.ProjectOperation) map[string]interface{} {
	if operation == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if operation.ProjectedColumns != nil {
		tfMap["project_columns"] = aws.StringValueSlice(operation.ProjectedColumns)
	}

	return tfMap
}

func flattenRenameColumnOperation(operation *quicksight.RenameColumnOperation) map[string]interface{} {
	if operation == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if operation.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(operation.ColumnName)
	}

	if operation.NewColumnName != nil {
		tfMap["new_column_name"] = aws.StringValue(operation.NewColumnName)
	}

	return tfMap
}

func flattenTagColumnOperation(operation *quicksight.TagColumnOperation) map[string]interface{} {
	if operation == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if operation.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(operation.ColumnName)
	}

	if operation.Tags != nil {
		tfMap["tags"] = flattenTags(operation.Tags)
	}

	return tfMap
}

func flattenTags(tags []*quicksight.ColumnTag) []interface{} {
	if len(tags) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, tag := range tags {
		if tag == nil {
			continue
		}

		tfList = append(tfList, flattenTag(tag))
	}

	return tfList
}

func flattenTag(tag *quicksight.ColumnTag) map[string]interface{} {
	if tag == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if tag.ColumnDescription != nil {
		tfMap["column_description"] = flattenColumnDescription(tag.ColumnDescription)
	}

	if tag.ColumnGeographicRole != nil {
		tfMap["column_geographic_role"] = aws.StringValue(tag.ColumnGeographicRole)
	}

	return tfMap
}

func flattenColumnDescription(desc *quicksight.ColumnDescription) map[string]interface{} {
	if desc == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if desc.Text != nil {
		tfMap["text"] = aws.StringValue(desc.Text)
	}

	return tfMap
}

func flattenUntagColumnOperation(operation *quicksight.UntagColumnOperation) map[string]interface{} {
	if operation == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if operation.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(operation.ColumnName)
	}

	if operation.TagNames != nil {
		tfMap["tag_names"] = aws.StringValueSlice(operation.TagNames)
	}

	return tfMap
}

func flattenLogicalTableSource(source *quicksight.LogicalTableSource) []interface{} {
	if source == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if source.DataSetArn != nil {
		tfMap["data_set_arn"] = aws.StringValue(source.DataSetArn)
	}

	if source.JoinInstruction != nil {
		tfMap["join_instruction"] = flattenJoinInstruction(source.JoinInstruction)
	}

	if source.PhysicalTableId != nil {
		tfMap["physical_table_id"] = aws.StringValue(source.PhysicalTableId)
	}

	return []interface{}{tfMap}
}

func flattenJoinInstruction(instruction *quicksight.JoinInstruction) map[string]interface{} {
	if instruction == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if instruction.LeftJoinKeyProperties != nil {
		tfMap["left_join_key_properties"] = flattenJoinKeyProperties(instruction.LeftJoinKeyProperties)
	}

	if instruction.LeftOperand != nil {
		tfMap["left_operand"] = aws.StringValue(instruction.LeftOperand)
	}

	if instruction.OnClause != nil {
		tfMap["on_clause"] = aws.StringValue(instruction.OnClause)
	}

	if instruction.RightJoinKeyProperties != nil {
		tfMap["right_join_key_properties"] = flattenJoinKeyProperties(instruction.RightJoinKeyProperties)
	}

	if instruction.RightOperand != nil {
		tfMap["right_operand"] = aws.StringValue(instruction.RightOperand)
	}

	if instruction.Type != nil {
		tfMap["type"] = aws.StringValue(instruction.Type)
	}

	return tfMap
}

func flattenJoinKeyProperties(prop *quicksight.JoinKeyProperties) map[string]interface{} {
	if prop == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if prop.UniqueKey != nil {
		tfMap["unique_key"] = aws.BoolValue(prop.UniqueKey)
	}

	return tfMap
}

func flattenPhysicalTableMap(maps map[string]*quicksight.PhysicalTable) *schema.Set {
	if len(maps) == 0 {
		return nil
	}

	tfSet := schema.NewSet(physicalTableMapHash, []interface{}{})

	for k, v := range maps {
		if v == nil {
			continue
		}

		tfMap := flattenPhysicalTable(k, v)
		tfSet.Add(tfMap)
	}
	return tfSet
}

func physicalTableMapHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["custom_sql"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", customSqlHash(v)))
	}
	if v, ok := m["relational_table"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", customRelationalTableHash(v)))
	}
	if v, ok := m["s3_source"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", customS3SourceHash(v)))
	}

	return create.StringHashcode(buf.String())
}

func customSqlHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["data_source_arn"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["sql_query"].(string)))
	if v, ok := m["columns"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", customColumnsHash(v)))
	}
	return create.StringHashcode(buf.String())
}

func customColumnsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["type"].(string)))
	return create.StringHashcode(buf.String())
}

func customRelationalTableHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["data_source_arn"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	if v, ok := m["catalog"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["input_columns"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", customColumnsHash(v)))
	}
	return create.StringHashcode(buf.String())
}

func customS3SourceHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["data_source_arn"].(string)))
	if v, ok := m["upload_settings"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", customUploadSettingsHash(v)))
	}
	buf.WriteString(fmt.Sprintf("%d-", customColumnsHash(m["input_columns"])))
	return create.StringHashcode(buf.String())
}

func customUploadSettingsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.([]interface{})[0].(map[string]interface{})

	if v, ok := m["contains_header"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}
	if v, ok := m["delimiter"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["format"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["start_from_row"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["text_qualifier"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}

func flattenPhysicalTable(key string, table *quicksight.PhysicalTable) map[string]interface{} {
	if table == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["physical_table_map_id"] = key

	if table.CustomSql != nil {
		tfMap["custom_sql"] = flattenCustomSql(table.CustomSql)
	}

	if table.RelationalTable != nil {
		tfMap["relational_table"] = flattenRelationalTable(table.RelationalTable)
	}

	if table.S3Source != nil {
		tfMap["s3_source"] = flattenS3Source(table.S3Source)
	}

	return tfMap
}

func flattenCustomSql(sql *quicksight.CustomSql) map[string]interface{} {
	if sql == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if sql.Columns != nil {
		tfMap["columns"] = flattenInputColumns(sql.Columns)
	}

	if sql.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.StringValue(sql.DataSourceArn)
	}

	if sql.Name != nil {
		tfMap["name"] = aws.StringValue(sql.Name)
	}

	if sql.SqlQuery != nil {
		tfMap["sql_query"] = aws.StringValue(sql.SqlQuery)
	}

	return tfMap
}

func flattenInputColumns(columns []*quicksight.InputColumn) []interface{} {
	if len(columns) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, column := range columns {
		if column == nil {
			continue
		}

		tfList = append(tfList, flattenInputColumn(column))
	}

	return tfList
}

func flattenInputColumn(column *quicksight.InputColumn) map[string]interface{} {
	if column == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if column.Name != nil {
		tfMap["name"] = aws.StringValue(column.Name)
	}

	if column.Type != nil {
		tfMap["type"] = aws.StringValue(column.Type)
	}

	return tfMap
}

func flattenRelationalTable(table *quicksight.RelationalTable) map[string]interface{} {
	if table == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if table.Catalog != nil {
		tfMap["catalog"] = aws.StringValue(table.Catalog)
	}

	if table.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.StringValue(table.DataSourceArn)
	}

	if table.InputColumns != nil {
		tfMap["input_columns"] = flattenInputColumns(table.InputColumns)
	}

	if table.Name != nil {
		tfMap["name"] = aws.StringValue(table.Name)
	}

	if table.Schema != nil {
		tfMap["schema"] = aws.StringValue(table.Schema)
	}

	return tfMap
}

func flattenS3Source(source *quicksight.S3Source) []interface{} {
	if source == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if source.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.StringValue(source.DataSourceArn)
	}

	if source.InputColumns != nil {
		tfMap["input_columns"] = flattenInputColumns(source.InputColumns)
	}

	if source.UploadSettings != nil {
		tfMap["upload_settings"] = flattenUploadSettings(source.UploadSettings)
	}

	return []interface{}{tfMap}
}

func flattenUploadSettings(settings *quicksight.UploadSettings) []interface{} {
	if settings == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if settings.ContainsHeader != nil {
		tfMap["contains_header"] = aws.BoolValue(settings.ContainsHeader)
	}

	if settings.Delimiter != nil {
		tfMap["contains_header"] = aws.StringValue(settings.Delimiter)
	}

	if settings.Format != nil {
		tfMap["format"] = aws.StringValue(settings.Format)
	}

	if settings.StartFromRow != nil {
		tfMap["start_from_row"] = aws.Int64Value(settings.StartFromRow)
	}

	if settings.TextQualifier != nil {
		tfMap["text_qualifier"] = aws.StringValue(settings.TextQualifier)
	}

	return []interface{}{tfMap}
}

func flattenRowLevelPermissionDataSet(set *quicksight.RowLevelPermissionDataSet) []interface{} {
	if set == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if set.Arn != nil {
		tfMap["arn"] = aws.StringValue(set.Arn)
	}

	if set.FormatVersion != nil {
		tfMap["format_version"] = aws.StringValue(set.FormatVersion)
	}

	if set.Namespace != nil {
		tfMap["namespace"] = aws.StringValue(set.Namespace)
	}

	if set.PermissionPolicy != nil {
		tfMap["permission_policy"] = aws.StringValue(set.PermissionPolicy)
	}

	if set.Status != nil {
		tfMap["status"] = aws.StringValue(set.Status)
	}

	return []interface{}{tfMap}
}

func flattenRowLevelPermissionTagConfiguration(configuration *quicksight.RowLevelPermissionTagConfiguration) []interface{} {
	if configuration == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if configuration.Status != nil {
		tfMap["status"] = aws.StringValue(configuration.Status)
	}

	if configuration.TagRules != nil {
		tfMap["tag_rules"] = flattenTagRules(configuration.TagRules)
	}

	return []interface{}{tfMap}
}

func flattenTagRules(rules []*quicksight.RowLevelPermissionTagRule) []interface{} {
	if len(rules) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		tfList = append(tfList, flattenTagRule(rule))
	}

	return tfList
}

func flattenTagRule(rule *quicksight.RowLevelPermissionTagRule) map[string]interface{} {
	if rule == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if rule.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(rule.ColumnName)
	}

	if rule.MatchAllValue != nil {
		tfMap["match_all_value"] = aws.StringValue(rule.MatchAllValue)
	}

	if rule.TagKey != nil {
		tfMap["tag_key"] = aws.StringValue(rule.TagKey)
	}

	if rule.TagMultiValueDelimiter != nil {
		tfMap["tag_multi_value_delimiter"] = aws.StringValue(rule.TagMultiValueDelimiter)
	}

	return tfMap
}

func ParseDataSetID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/DATA_SOURCE_ID", id)
	}
	return parts[0], parts[1], nil
}
