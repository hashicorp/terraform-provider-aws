// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_data_set", name="Data Set")
// @Tags(identifierAttribute="arn")
func ResourceDataSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataSetCreate,
		ReadWithoutTimeout:   resourceDataSetRead,
		UpdateWithoutTimeout: resourceDataSetUpdate,
		DeleteWithoutTimeout: resourceDataSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
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
										names.AttrName: {
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
							names.AttrDescription: {
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
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					MaxItems: 64,
					Elem:     logicalTableMapSchema(),
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"output_columns": {
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
				},
				names.AttrPermissions: {
					Type:     schema.TypeSet,
					Optional: true,
					MinItems: 1,
					MaxItems: 64,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrActions: {
								Type:     schema.TypeSet,
								Required: true,
								MinItems: 1,
								MaxItems: 20,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							names.AttrPrincipal: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
						},
					},
				},
				"physical_table_map": {
					Type:     schema.TypeSet,
					Optional: true,
					MaxItems: 32,
					Elem:     physicalTableMapSchema(),
				},
				"row_level_permission_data_set": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"format_version": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringInSlice(quicksight.RowLevelPermissionFormatVersion_Values(), false),
							},
							names.AttrNamespace: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 64),
							},
							"permission_policy": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice(quicksight.RowLevelPermissionPolicy_Values(), false),
							},
							names.AttrStatus: {
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
							names.AttrStatus: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
							},
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
										"match_all_value": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(1, 256),
										},
										"tag_key": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										"tag_multi_value_delimiter": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(1, 10),
										},
									},
								},
							},
						},
					},
				},
				"refresh_properties": {
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
																"size_unit": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringInSlice(quicksight.LookbackWindowSizeUnit_Values(), false),
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
					},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: customdiff.All(
			func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
				mode := diff.Get("import_mode").(string)
				if v, ok := diff.Get("refresh_properties").([]interface{}); ok && v != nil && len(v) > 0 && mode == "DIRECT_QUERY" {
					return fmt.Errorf("refresh_properties cannot be set when import_mode is 'DIRECT_QUERY'")
				}
				return nil
			},
			verify.SetTagsDiff,
		),
	}
}

func logicalTableMapSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrAlias: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
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
									names.AttrFormat: {
										Type:         schema.TypeString,
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 32),
									},
									"new_column_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(quicksight.ColumnDataType_Values(), false),
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
												names.AttrExpression: {
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
									"right_operand": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
									names.AttrType: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(quicksight.JoinType_Values(), false),
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
		},
	}
}

func physicalTableMapSchema() *schema.Resource {
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
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									names.AttrType: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(quicksight.InputColumnDataType_Values(), false),
									},
								},
							},
						},
						"data_source_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
						"sql_query": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 65536),
						},
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
						"catalog": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
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
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									names.AttrType: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(quicksight.InputColumnDataType_Values(), false),
									},
								},
							},
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
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
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									names.AttrType: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(quicksight.InputColumnDataType_Values(), false),
									},
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
									"delimiter": {
										Type:         schema.TypeString,
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1),
									},
									names.AttrFormat: {
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
	}
}

func resourceDataSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountId = v.(string)
	}
	dataSetID := d.Get("data_set_id").(string)

	d.SetId(createDataSetID(awsAccountId, dataSetID))

	input := &quicksight.CreateDataSetInput{
		AwsAccountId:     aws.String(awsAccountId),
		DataSetId:        aws.String(dataSetID),
		ImportMode:       aws.String(d.Get("import_mode").(string)),
		PhysicalTableMap: expandDataSetPhysicalTableMap(d.Get("physical_table_map").(*schema.Set)),
		Name:             aws.String(d.Get(names.AttrName).(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("column_groups"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ColumnGroups = expandDataSetColumnGroups(v.([]interface{}))
	}

	if v, ok := d.GetOk("column_level_permission_rules"); ok && len(v.([]interface{})) > 0 {
		input.ColumnLevelPermissionRules = expandDataSetColumnLevelPermissionRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("data_set_usage_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataSetUsageConfiguration = expandDataSetUsageConfiguration(v.([]interface{}))
	}

	if v, ok := d.Get("field_folders").(*schema.Set); ok && v.Len() > 0 {
		input.FieldFolders = expandDataSetFieldFolders(v.List())
	}

	if v, ok := d.GetOk("logical_table_map"); ok && v.(*schema.Set).Len() != 0 {
		input.LogicalTableMap = expandDataSetLogicalTableMap(v.(*schema.Set))
	}

	if v, ok := d.Get(names.AttrPermissions).(*schema.Set); ok && v.Len() > 0 {
		input.Permissions = expandResourcePermissions(v.List())
	}

	if v, ok := d.GetOk("row_level_permission_data_set"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RowLevelPermissionDataSet = expandDataSetRowLevelPermissionDataSet(v.([]interface{}))
	}

	if v, ok := d.GetOk("row_level_permission_tag_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RowLevelPermissionTagConfiguration = expandDataSetRowLevelPermissionTagConfigurations(v.([]interface{}))
	}

	_, err := conn.CreateDataSetWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Data Set: %s", err)
	}

	if v, ok := d.GetOk("refresh_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &quicksight.PutDataSetRefreshPropertiesInput{
			AwsAccountId:             aws.String(awsAccountId),
			DataSetId:                aws.String(dataSetID),
			DataSetRefreshProperties: expandDataSetRefreshProperties(v.([]interface{})),
		}

		_, err := conn.PutDataSetRefreshPropertiesWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting QuickSight Data Set Refresh Properties: %s", err)
		}
	}

	return append(diags, resourceDataSetRead(ctx, d, meta)...)
}

func resourceDataSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, dataSetId, err := ParseDataSetID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	descOpts := &quicksight.DescribeDataSetInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSetId:    aws.String(dataSetId),
	}

	output, err := conn.DescribeDataSetWithContext(ctx, descOpts)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] QuickSight Data Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Data Set (%s): %s", d.Id(), err)
	}

	if output == nil || output.DataSet == nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Data Set (%s): empty output", d.Id())
	}

	dataSet := output.DataSet

	d.Set(names.AttrARN, dataSet.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountId)
	d.Set("data_set_id", dataSet.DataSetId)
	d.Set(names.AttrName, dataSet.Name)
	d.Set("import_mode", dataSet.ImportMode)

	if err := d.Set("column_groups", flattenColumnGroups(dataSet.ColumnGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting column_groups: %s", err)
	}

	if err := d.Set("column_level_permission_rules", flattenColumnLevelPermissionRules(dataSet.ColumnLevelPermissionRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting column_level_permission_rules: %s", err)
	}

	if err := d.Set("data_set_usage_configuration", flattenDataSetUsageConfiguration(dataSet.DataSetUsageConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_set_usage_configuration: %s", err)
	}

	if err := d.Set("field_folders", flattenFieldFolders(dataSet.FieldFolders)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting field_folders: %s", err)
	}

	if err := d.Set("logical_table_map", flattenLogicalTableMap(dataSet.LogicalTableMap, logicalTableMapSchema())); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logical_table_map: %s", err)
	}

	if err := d.Set("output_columns", flattenOutputColumns(dataSet.OutputColumns)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting output_columns: %s", err)
	}

	if err := d.Set("physical_table_map", flattenPhysicalTableMap(dataSet.PhysicalTableMap, physicalTableMapSchema())); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_table_map: %s", err)
	}

	if err := d.Set("row_level_permission_data_set", flattenRowLevelPermissionDataSet(dataSet.RowLevelPermissionDataSet)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting row_level_permission_data_set: %s", err)
	}

	if err := d.Set("row_level_permission_tag_configuration", flattenRowLevelPermissionTagConfiguration(dataSet.RowLevelPermissionTagConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting row_level_permission_tag_configuration: %s", err)
	}

	permsResp, err := conn.DescribeDataSetPermissionsWithContext(ctx, &quicksight.DescribeDataSetPermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSetId:    aws.String(dataSetId),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Data Source (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, flattenPermissions(permsResp.Permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	propsResp, err := conn.DescribeDataSetRefreshPropertiesWithContext(ctx, &quicksight.DescribeDataSetRefreshPropertiesInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSetId:    aws.String(dataSetId),
	})

	if err != nil && !(tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, quicksight.ErrCodeInvalidParameterValueException, "not a SPICE dataset")) {
		return sdkdiag.AppendErrorf(diags, "describing refresh properties (%s): %s", d.Id(), err)
	}

	if err == nil {
		if err := d.Set("refresh_properties", flattenRefreshProperties(propsResp.DataSetRefreshProperties)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting refresh properties: %s", err)
		}
	}

	return diags
}

func resourceDataSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll, "refresh_properties") {
		awsAccountId, dataSetId, err := ParseDataSetID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		params := &quicksight.UpdateDataSetInput{
			AwsAccountId:     aws.String(awsAccountId),
			DataSetId:        aws.String(dataSetId),
			ImportMode:       aws.String(d.Get("import_mode").(string)),
			PhysicalTableMap: expandDataSetPhysicalTableMap(d.Get("physical_table_map").(*schema.Set)),
			Name:             aws.String(d.Get(names.AttrName).(string)),
		}

		params.ColumnGroups = expandDataSetColumnGroups(d.Get("column_groups").([]interface{}))

		params.ColumnLevelPermissionRules = expandDataSetColumnLevelPermissionRules(d.Get("column_level_permission_rules").([]interface{}))

		params.DataSetUsageConfiguration = expandDataSetUsageConfiguration(d.Get("data_set_usage_configuration").([]interface{}))

		params.FieldFolders = expandDataSetFieldFolders(d.Get("field_folders").(*schema.Set).List())

		params.LogicalTableMap = expandDataSetLogicalTableMap(d.Get("logical_table_map").(*schema.Set))

		params.RowLevelPermissionDataSet = expandDataSetRowLevelPermissionDataSet(d.Get("row_level_permission_data_set").([]interface{}))

		params.RowLevelPermissionTagConfiguration = expandDataSetRowLevelPermissionTagConfigurations(d.Get("row_level_permission_tag_configuration").([]interface{}))

		_, err = conn.UpdateDataSetWithContext(ctx, params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Set (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		awsAccountId, dataSetId, err := ParseDataSetID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		oraw, nraw := d.GetChange(names.AttrPermissions)
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)

		toGrant, toRevoke := DiffPermissions(o.List(), n.List())

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
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Set (%s) permissions: %s", dataSetId, err)
		}
	}

	if d.HasChange("refresh_properties") {
		awsAccountId, dataSetId, err := ParseDataSetID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		oldraw, newraw := d.GetChange("refresh_properties")
		old := oldraw.([]interface{})
		new := newraw.([]interface{})
		if len(old) == 1 && len(new) == 0 {
			_, err := conn.DeleteDataSetRefreshPropertiesWithContext(ctx, &quicksight.DeleteDataSetRefreshPropertiesInput{
				AwsAccountId: aws.String(awsAccountId),
				DataSetId:    aws.String(dataSetId),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting QuickSight Data Set Refresh Properties (%s): %s", d.Id(), err)
			}
		} else {
			_, err = conn.PutDataSetRefreshPropertiesWithContext(ctx, &quicksight.PutDataSetRefreshPropertiesInput{
				AwsAccountId:             aws.String(awsAccountId),
				DataSetId:                aws.String(dataSetId),
				DataSetRefreshProperties: expandDataSetRefreshProperties(d.Get("refresh_properties").([]interface{})),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Set Refresh Properties (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDataSetRead(ctx, d, meta)...)
}

func resourceDataSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	log.Printf("[INFO] Deleting QuickSight Data Set %s", d.Id())
	awsAccountId, dataSetId, err := ParseDataSetID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	deleteOpts := &quicksight.DeleteDataSetInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSetId:    aws.String(dataSetId),
	}

	_, err = conn.DeleteDataSetWithContext(ctx, deleteOpts)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Data Set (%s): %s", d.Id(), err)
	}

	return diags
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
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		geoSpatialColumnGroup.Name = aws.String(v)
	}

	return geoSpatialColumnGroup
}

func expandDataSetColumnLevelPermissionRules(tfList []interface{}) []*quicksight.ColumnLevelPermissionRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObject []*quicksight.ColumnLevelPermissionRule
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := &quicksight.ColumnLevelPermissionRule{}
		if v, ok := tfMap["column_names"].([]interface{}); ok {
			rule.ColumnNames = flex.ExpandStringList(v)
		}
		if v, ok := tfMap["principals"].([]interface{}); ok {
			rule.Principals = flex.ExpandStringList(v)
		}
		apiObject = append(apiObject, rule)
	}

	return apiObject
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
		if v, ok := tfMap[names.AttrDescription].(string); ok {
			fieldFolder.Description = aws.String(v)
		}

		fieldFolderID := tfMap["field_folders_id"].(string)
		fieldFolderMap[fieldFolderID] = fieldFolder
	}

	return fieldFolderMap
}

func expandDataSetLogicalTableMap(tfSet *schema.Set) map[string]*quicksight.LogicalTable {
	if tfSet.Len() == 0 {
		return nil
	}

	logicalTableMap := make(map[string]*quicksight.LogicalTable)
	for _, v := range tfSet.List() {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		logicalTable := &quicksight.LogicalTable{}
		logicalTableMapID := vMap["logical_table_map_id"].(string)

		if v, ok := vMap[names.AttrAlias].(string); ok {
			logicalTable.Alias = aws.String(v)
		}
		if v, ok := vMap[names.AttrSource].([]interface{}); ok {
			logicalTable.Source = expandDataSetLogicalTableSource(v[0].(map[string]interface{}))
		}
		if v, ok := vMap["data_transforms"].([]interface{}); ok {
			logicalTable.DataTransforms = expandDataSetDataTransforms(v)
		}

		logicalTableMap[logicalTableMapID] = logicalTable
	}

	return logicalTableMap
}

func expandDataSetLogicalTableSource(tfMap map[string]interface{}) *quicksight.LogicalTableSource {
	if tfMap == nil {
		return nil
	}

	logicalTableSource := &quicksight.LogicalTableSource{}
	if v, ok := tfMap["data_set_arn"].(string); ok && v != "" {
		logicalTableSource.DataSetArn = aws.String(v)
	}
	if v, ok := tfMap["physical_table_id"].(string); ok && v != "" {
		logicalTableSource.PhysicalTableId = aws.String(v)
	}
	if v, ok := tfMap["join_instruction"].([]interface{}); ok && len(v) > 0 {
		logicalTableSource.JoinInstruction = expandDataSetJoinInstruction(v[0].(map[string]interface{}))
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
	if v, ok := tfMap[names.AttrType].(string); ok {
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
	if v, ok := tfMap["cast_column_type_operation"].([]interface{}); ok && len(v) > 0 {
		transformOperation.CastColumnTypeOperation = expandDataSetCastColumnTypeOperation(v)
	}
	if v, ok := tfMap["create_columns_operation"].([]interface{}); ok && len(v) > 0 {
		transformOperation.CreateColumnsOperation = expandDataSetCreateColumnsOperation(v)
	}
	if v, ok := tfMap["filter_operation"].([]interface{}); ok && len(v) > 0 {
		transformOperation.FilterOperation = expandDataSetFilterOperation(v)
	}
	if v, ok := tfMap["project_operation"].([]interface{}); ok && len(v) > 0 {
		transformOperation.ProjectOperation = expandDataSetProjectOperation(v)
	}
	if v, ok := tfMap["rename_column_operation"].([]interface{}); ok && len(v) > 0 {
		transformOperation.RenameColumnOperation = expandDataSetRenameColumnOperation(v)
	}
	if v, ok := tfMap["tag_column_operation"].([]interface{}); ok && len(v) > 0 {
		transformOperation.TagColumnOperation = expandDataSetTagColumnOperation(v)
	}
	if v, ok := tfMap["untag_column_operation"].([]interface{}); ok && len(v) > 0 {
		transformOperation.UntagColumnOperation = expandDataSetUntagColumnOperation(v)
	}

	return transformOperation
}

func expandDataSetCastColumnTypeOperation(tfList []interface{}) *quicksight.CastColumnTypeOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	castColumnTypeOperation := &quicksight.CastColumnTypeOperation{}
	if v, ok := tfMap["column_name"].(string); ok {
		castColumnTypeOperation.ColumnName = aws.String(v)
	}
	if v, ok := tfMap["new_column_type"].(string); ok {
		castColumnTypeOperation.NewColumnType = aws.String(v)
	}
	if v, ok := tfMap[names.AttrFormat].(string); ok {
		castColumnTypeOperation.Format = aws.String(v)
	}

	return castColumnTypeOperation
}

func expandDataSetCreateColumnsOperation(tfList []interface{}) *quicksight.CreateColumnsOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
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
	if v, ok := tfMap[names.AttrExpression].(string); ok {
		calculatedColumn.Expression = aws.String(v)
	}

	return calculatedColumn
}

func expandDataSetFilterOperation(tfList []interface{}) *quicksight.FilterOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filterOperation := &quicksight.FilterOperation{}
	if v, ok := tfMap["condition_expression"].(string); ok {
		filterOperation.ConditionExpression = aws.String(v)
	}

	return filterOperation
}

func expandDataSetProjectOperation(tfList []interface{}) *quicksight.ProjectOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	projectOperation := &quicksight.ProjectOperation{}
	if v, ok := tfMap["projected_columns"].([]interface{}); ok && len(v) > 0 {
		projectOperation.ProjectedColumns = flex.ExpandStringList(v)
	}

	return projectOperation
}

func expandDataSetRenameColumnOperation(tfList []interface{}) *quicksight.RenameColumnOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
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

func expandDataSetTagColumnOperation(tfList []interface{}) *quicksight.TagColumnOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	tagColumnOperation := &quicksight.TagColumnOperation{}
	if v, ok := tfMap["column_name"].(string); ok {
		tagColumnOperation.ColumnName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrTags].([]interface{}); ok {
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

func expandDataSetUntagColumnOperation(tfList []interface{}) *quicksight.UntagColumnOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
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
				physicalTable.CustomSql = expandDataSetCustomSQL(v.(map[string]interface{}))
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

func expandDataSetCustomSQL(tfMap map[string]interface{}) *quicksight.CustomSql {
	if tfMap == nil {
		return nil
	}

	customSQL := &quicksight.CustomSql{}
	if v, ok := tfMap["columns"].([]interface{}); ok {
		customSQL.Columns = expandDataSetInputColumns(v)
	}
	if v, ok := tfMap["data_source_arn"].(string); ok {
		customSQL.DataSourceArn = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok {
		customSQL.Name = aws.String(v)
	}
	if v, ok := tfMap["sql_query"].(string); ok {
		customSQL.SqlQuery = aws.String(v)
	}

	return customSQL
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
	if v, ok := tfMap[names.AttrName].(string); ok {
		inputColumn.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok {
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
	if v, ok := tfMap["catalog"].(string); ok {
		relationalTable.Catalog = aws.String(v)
	}
	if v, ok := tfMap["data_source_arn"].(string); ok {
		relationalTable.DataSourceArn = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok {
		relationalTable.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrSchema].(string); ok {
		relationalTable.Schema = aws.String(v)
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
	if v, ok := tfMap[names.AttrFormat].(string); ok {
		uploadSettings.Format = aws.String(v)
	}
	if v, ok := tfMap["start_from_row"].(int); ok {
		uploadSettings.StartFromRow = aws.Int64(int64(v))
	}
	if v, ok := tfMap["text_qualifier"].(string); ok {
		uploadSettings.TextQualifier = aws.String(v)
	}

	return uploadSettings
}

func expandDataSetRowLevelPermissionDataSet(tfList []interface{}) *quicksight.RowLevelPermissionDataSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	rowLevelPermission := &quicksight.RowLevelPermissionDataSet{}
	if v, ok := tfMap[names.AttrARN].(string); ok {
		rowLevelPermission.Arn = aws.String(v)
	}
	if v, ok := tfMap["permission_policy"].(string); ok {
		rowLevelPermission.PermissionPolicy = aws.String(v)
	}
	if v, ok := tfMap["format_version"].(string); ok {
		rowLevelPermission.FormatVersion = aws.String(v)
	}
	if v, ok := tfMap[names.AttrNamespace].(string); ok {
		rowLevelPermission.Namespace = aws.String(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok {
		rowLevelPermission.Status = aws.String(v)
	}

	return rowLevelPermission
}

func expandDataSetRowLevelPermissionTagConfigurations(tfList []interface{}) *quicksight.RowLevelPermissionTagConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	rowLevelPermissionTagConfiguration := &quicksight.RowLevelPermissionTagConfiguration{}
	if v, ok := tfMap["tag_rules"].([]interface{}); ok {
		rowLevelPermissionTagConfiguration.TagRules = expandDataSetTagRules(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok {
		rowLevelPermissionTagConfiguration.Status = aws.String(v)
	}

	return rowLevelPermissionTagConfiguration
}

func expandDataSetRefreshProperties(tfList []interface{}) *quicksight.DataSetRefreshProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}
	props := &quicksight.DataSetRefreshProperties{}
	if v, ok := tfMap["refresh_configuration"].([]interface{}); ok {
		props.RefreshConfiguration = expandDataSetRefreshConfiguration(v)
	}
	return props
}

func expandDataSetRefreshConfiguration(tfList []interface{}) *quicksight.RefreshConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}
	config := &quicksight.RefreshConfiguration{}
	if v, ok := tfMap["incremental_refresh"].([]interface{}); ok {
		config.IncrementalRefresh = expandIncrementalRefresh(v)
	}
	return config
}

func expandIncrementalRefresh(tfList []interface{}) *quicksight.IncrementalRefresh {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}
	refresh := &quicksight.IncrementalRefresh{}
	if v, ok := tfMap["lookback_window"].([]interface{}); ok {
		refresh.LookbackWindow = expandLookbackWindow(v)
	}
	return refresh
}

func expandLookbackWindow(tfList []interface{}) *quicksight.LookbackWindow {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}
	window := &quicksight.LookbackWindow{}
	if v, ok := tfMap["column_name"].(string); ok {
		window.ColumnName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrSize].(int); ok {
		window.Size = aws.Int64(int64(v))
	}
	if v, ok := tfMap["size_unit"].(string); ok {
		window.SizeUnit = aws.String(v)
	}
	return window
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

func flattenColumnGroups(apiObject []*quicksight.ColumnGroup) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, group := range apiObject {
		if group == nil {
			continue
		}

		item := map[string]interface{}{}
		if group.GeoSpatialColumnGroup != nil {
			item["geo_spatial_column_group"] = flattenGeoSpatialColumnGroup(group.GeoSpatialColumnGroup)
		}
		tfList = append(tfList, item)
	}

	return tfList
}

func flattenOutputColumns(apiObject []*quicksight.OutputColumn) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, column := range apiObject {
		if column == nil {
			continue
		}

		item := map[string]interface{}{}

		if column.Description != nil {
			item[names.AttrDescription] = aws.StringValue(column.Description)
		}
		if column.Name != nil {
			item[names.AttrName] = aws.StringValue(column.Name)
		}
		if column.Type != nil {
			item[names.AttrType] = aws.StringValue(column.Type)
		}

		tfList = append(tfList, item)
	}

	return tfList
}

func flattenGeoSpatialColumnGroup(apiObject *quicksight.GeoSpatialColumnGroup) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Columns != nil {
		tfMap["columns"] = aws.StringValueSlice(apiObject.Columns)
	}
	if apiObject.CountryCode != nil {
		tfMap["country_code"] = aws.StringValue(apiObject.CountryCode)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}

	return []interface{}{tfMap}
}

func flattenColumnLevelPermissionRules(apiObject []*quicksight.ColumnLevelPermissionRule) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, rule := range apiObject {
		if rule == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if rule.ColumnNames != nil {
			tfMap["column_names"] = aws.StringValueSlice(rule.ColumnNames)
		}
		if rule.Principals != nil {
			tfMap["principals"] = aws.StringValueSlice(rule.Principals)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataSetUsageConfiguration(apiObject *quicksight.DataSetUsageConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DisableUseAsDirectQuerySource != nil {
		tfMap["disable_use_as_direct_query_source"] = aws.BoolValue(apiObject.DisableUseAsDirectQuerySource)
	}
	if apiObject.DisableUseAsImportedSource != nil {
		tfMap["disable_use_as_imported_source"] = aws.BoolValue(apiObject.DisableUseAsImportedSource)
	}

	return []interface{}{tfMap}
}

func flattenFieldFolders(apiObject map[string]*quicksight.FieldFolder) *schema.Set {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for key, value := range apiObject {
		if value == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"field_folders_id": key,
		}
		if len(value.Columns) > 0 {
			tfMap["columns"] = flex.FlattenStringList(value.Columns)
		}
		if value.Description != nil {
			tfMap[names.AttrDescription] = aws.StringValue(value.Description)
		}
		tfList = append(tfList, tfMap)
	}

	return schema.NewSet(fieldFoldersHash, tfList)
}

func fieldFoldersHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["field_folders_id"].(string)))
	if v, ok := m["columns"]; ok {
		if sl, ok := v.([]string); ok {
			buf.WriteString(fmt.Sprintf("%s-", sl))
		}
	}
	if v, ok := m[names.AttrDescription]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}

func flattenLogicalTableMap(apiObject map[string]*quicksight.LogicalTable, resourceSchema *schema.Resource) *schema.Set {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for key, table := range apiObject {
		if table == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"logical_table_map_id": key,
		}
		if table.Alias != nil {
			tfMap[names.AttrAlias] = aws.StringValue(table.Alias)
		}
		if table.DataTransforms != nil {
			tfMap["data_transforms"] = flattenDataTransforms(table.DataTransforms)
		}
		if table.Source != nil {
			tfMap[names.AttrSource] = flattenLogicalTableSource(table.Source)
		}
		tfList = append(tfList, tfMap)
	}

	return schema.NewSet(schema.HashResource(resourceSchema), tfList)
}

func flattenDataTransforms(apiObject []*quicksight.TransformOperation) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, operation := range apiObject {
		if operation == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if operation.CastColumnTypeOperation != nil {
			tfMap["cast_column_type_operation"] = flattenCastColumnTypeOperation(operation.CastColumnTypeOperation)
		}
		if operation.CreateColumnsOperation != nil {
			tfMap["create_columns_operation"] = flattenCreateColumnsOperation(operation.CreateColumnsOperation)
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
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCastColumnTypeOperation(apiObject *quicksight.CastColumnTypeOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(apiObject.ColumnName)
	}
	if apiObject.Format != nil {
		tfMap[names.AttrFormat] = aws.StringValue(apiObject.Format)
	}
	if apiObject.NewColumnType != nil {
		tfMap["new_column_type"] = aws.StringValue(apiObject.NewColumnType)
	}

	return []interface{}{tfMap}
}

func flattenCreateColumnsOperation(apiObject *quicksight.CreateColumnsOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Columns != nil {
		tfMap["columns"] = flattenCalculatedColumns(apiObject.Columns)
	}

	return []interface{}{tfMap}
}

func flattenCalculatedColumns(apiObject []*quicksight.CalculatedColumn) interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, column := range apiObject {
		if column == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if column.ColumnId != nil {
			tfMap["column_id"] = aws.StringValue(column.ColumnId)
		}
		if column.ColumnName != nil {
			tfMap["column_name"] = aws.StringValue(column.ColumnName)
		}
		if column.Expression != nil {
			tfMap[names.AttrExpression] = aws.StringValue(column.Expression)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFilterOperation(apiObject *quicksight.FilterOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ConditionExpression != nil {
		tfMap["condition_expression"] = aws.StringValue(apiObject.ConditionExpression)
	}

	return []interface{}{tfMap}
}

func flattenProjectOperation(apiObject *quicksight.ProjectOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ProjectedColumns != nil {
		tfMap["projected_columns"] = flex.FlattenStringList(apiObject.ProjectedColumns)
	}

	return []interface{}{tfMap}
}

func flattenRenameColumnOperation(apiObject *quicksight.RenameColumnOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(apiObject.ColumnName)
	}
	if apiObject.NewColumnName != nil {
		tfMap["new_column_name"] = aws.StringValue(apiObject.NewColumnName)
	}

	return []interface{}{tfMap}
}

func flattenTagColumnOperation(apiObject *quicksight.TagColumnOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(apiObject.ColumnName)
	}
	if apiObject.Tags != nil {
		tfMap[names.AttrTags] = flattenTags(apiObject.Tags)
	}

	return []interface{}{tfMap}
}

func flattenTags(apiObject []*quicksight.ColumnTag) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, tag := range apiObject {
		if tag == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if tag.ColumnDescription != nil {
			tfMap["column_description"] = flattenColumnDescription(tag.ColumnDescription)
		}
		if tag.ColumnGeographicRole != nil {
			tfMap["column_geographic_role"] = aws.StringValue(tag.ColumnGeographicRole)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnDescription(apiObject *quicksight.ColumnDescription) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Text != nil {
		tfMap["text"] = aws.StringValue(apiObject.Text)
	}

	return []interface{}{tfMap}
}

func flattenUntagColumnOperation(apiObject *quicksight.UntagColumnOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(apiObject.ColumnName)
	}
	if apiObject.TagNames != nil {
		tfMap["tag_names"] = aws.StringValueSlice(apiObject.TagNames)
	}

	return []interface{}{tfMap}
}

func flattenLogicalTableSource(apiObject *quicksight.LogicalTableSource) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DataSetArn != nil {
		tfMap["data_set_arn"] = aws.StringValue(apiObject.DataSetArn)
	}
	if apiObject.JoinInstruction != nil {
		tfMap["join_instruction"] = flattenJoinInstruction(apiObject.JoinInstruction)
	}
	if apiObject.PhysicalTableId != nil {
		tfMap["physical_table_id"] = aws.StringValue(apiObject.PhysicalTableId)
	}

	return []interface{}{tfMap}
}

func flattenJoinInstruction(apiObject *quicksight.JoinInstruction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LeftJoinKeyProperties != nil {
		tfMap["left_join_key_properties"] = flattenJoinKeyProperties(apiObject.LeftJoinKeyProperties)
	}
	if apiObject.LeftOperand != nil {
		tfMap["left_operand"] = aws.StringValue(apiObject.LeftOperand)
	}
	if apiObject.OnClause != nil {
		tfMap["on_clause"] = aws.StringValue(apiObject.OnClause)
	}
	if apiObject.RightJoinKeyProperties != nil {
		tfMap["right_join_key_properties"] = flattenJoinKeyProperties(apiObject.RightJoinKeyProperties)
	}
	if apiObject.RightOperand != nil {
		tfMap["right_operand"] = aws.StringValue(apiObject.RightOperand)
	}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}

	return []interface{}{tfMap}
}

func flattenJoinKeyProperties(apiObject *quicksight.JoinKeyProperties) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.UniqueKey != nil {
		tfMap["unique_key"] = aws.BoolValue(apiObject.UniqueKey)
	}

	return tfMap
}

func flattenPhysicalTableMap(apiObject map[string]*quicksight.PhysicalTable, resourceSchema *schema.Resource) *schema.Set {
	var tfList []interface{}
	for k, v := range apiObject {
		if v == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"physical_table_map_id": k,
		}
		if v.CustomSql != nil {
			tfMap["custom_sql"] = flattenCustomSQL(v.CustomSql)
		}
		if v.RelationalTable != nil {
			tfMap["relational_table"] = flattenRelationalTable(v.RelationalTable)
		}
		if v.S3Source != nil {
			tfMap["s3_source"] = flattenS3Source(v.S3Source)
		}
		tfList = append(tfList, tfMap)
	}

	return schema.NewSet(schema.HashResource(resourceSchema), tfList)
}

func flattenCustomSQL(apiObject *quicksight.CustomSql) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Columns != nil {
		tfMap["columns"] = flattenInputColumns(apiObject.Columns)
	}
	if apiObject.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.StringValue(apiObject.DataSourceArn)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.SqlQuery != nil {
		tfMap["sql_query"] = aws.StringValue(apiObject.SqlQuery)
	}

	return []interface{}{tfMap}
}

func flattenInputColumns(apiObject []*quicksight.InputColumn) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, column := range apiObject {
		if column == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if column.Name != nil {
			tfMap[names.AttrName] = aws.StringValue(column.Name)
		}
		if column.Type != nil {
			tfMap[names.AttrType] = aws.StringValue(column.Type)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRelationalTable(apiObject *quicksight.RelationalTable) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Catalog != nil {
		tfMap["catalog"] = aws.StringValue(apiObject.Catalog)
	}
	if apiObject.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.StringValue(apiObject.DataSourceArn)
	}
	if apiObject.InputColumns != nil {
		tfMap["input_columns"] = flattenInputColumns(apiObject.InputColumns)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.Schema != nil {
		tfMap[names.AttrSchema] = aws.StringValue(apiObject.Schema)
	}

	return []interface{}{tfMap}
}

func flattenS3Source(apiObject *quicksight.S3Source) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.StringValue(apiObject.DataSourceArn)
	}
	if apiObject.InputColumns != nil {
		tfMap["input_columns"] = flattenInputColumns(apiObject.InputColumns)
	}

	if apiObject.UploadSettings != nil {
		tfMap["upload_settings"] = flattenUploadSettings(apiObject.UploadSettings)
	}

	return []interface{}{tfMap}
}

func flattenUploadSettings(apiObject *quicksight.UploadSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ContainsHeader != nil {
		tfMap["contains_header"] = aws.BoolValue(apiObject.ContainsHeader)
	}
	if apiObject.Delimiter != nil {
		tfMap["delimiter"] = aws.StringValue(apiObject.Delimiter)
	}
	if apiObject.Format != nil {
		tfMap[names.AttrFormat] = aws.StringValue(apiObject.Format)
	}
	if apiObject.StartFromRow != nil {
		tfMap["start_from_row"] = int(aws.Int64Value(apiObject.StartFromRow))
	}
	if apiObject.TextQualifier != nil {
		tfMap["text_qualifier"] = aws.StringValue(apiObject.TextQualifier)
	}

	return []interface{}{tfMap}
}

func flattenRowLevelPermissionDataSet(apiObject *quicksight.RowLevelPermissionDataSet) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Arn != nil {
		tfMap[names.AttrARN] = aws.StringValue(apiObject.Arn)
	}
	if apiObject.FormatVersion != nil {
		tfMap["format_version"] = aws.StringValue(apiObject.FormatVersion)
	}
	if apiObject.Namespace != nil {
		tfMap[names.AttrNamespace] = aws.StringValue(apiObject.Namespace)
	}
	if apiObject.PermissionPolicy != nil {
		tfMap["permission_policy"] = aws.StringValue(apiObject.PermissionPolicy)
	}
	if apiObject.Status != nil {
		tfMap[names.AttrStatus] = aws.StringValue(apiObject.Status)
	}

	return []interface{}{tfMap}
}

func flattenRowLevelPermissionTagConfiguration(apiObject *quicksight.RowLevelPermissionTagConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Status != nil {
		tfMap[names.AttrStatus] = aws.StringValue(apiObject.Status)
	}
	if apiObject.TagRules != nil {
		tfMap["tag_rules"] = flattenTagRules(apiObject.TagRules)
	}

	return []interface{}{tfMap}
}

func flattenRefreshProperties(apiObject *quicksight.DataSetRefreshProperties) interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.RefreshConfiguration != nil {
		tfMap["refresh_configuration"] = flattenRefreshConfiguration(apiObject.RefreshConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenRefreshConfiguration(apiObject *quicksight.RefreshConfiguration) interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.IncrementalRefresh != nil {
		tfMap["incremental_refresh"] = flattenIncrementalRefresh(apiObject.IncrementalRefresh)
	}

	return []interface{}{tfMap}
}

func flattenIncrementalRefresh(apiObject *quicksight.IncrementalRefresh) interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LookbackWindow != nil {
		tfMap["lookback_window"] = flattenLookbackWindow(apiObject.LookbackWindow)
	}

	return []interface{}{tfMap}
}

func flattenLookbackWindow(apiObject *quicksight.LookbackWindow) interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(apiObject.ColumnName)
	}
	if apiObject.Size != nil {
		tfMap[names.AttrSize] = aws.Int64Value(apiObject.Size)
	}
	if apiObject.SizeUnit != nil {
		tfMap["size_unit"] = aws.StringValue(apiObject.SizeUnit)
	}

	return []interface{}{tfMap}
}

func flattenTagRules(apiObject []*quicksight.RowLevelPermissionTagRule) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, rule := range apiObject {
		if rule == nil {
			continue
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
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func ParseDataSetID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,DATA_SET_ID", id)
	}
	return parts[0], parts[1], nil
}

func createDataSetID(awsAccountID, dataSetID string) string {
	return fmt.Sprintf("%s,%s", awsAccountID, dataSetID)
}
