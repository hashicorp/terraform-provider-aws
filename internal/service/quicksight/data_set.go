// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_data_set", name="Data Set")
// @Tags(identifierAttribute="arn")
func resourceDataSet() *schema.Resource {
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
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.GeoSpatialCountryCode](),
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
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.DataSetImportMode](),
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
				names.AttrPermissions: quicksightschema.PermissionsSchema(),
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
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.RowLevelPermissionFormatVersion](),
							},
							names.AttrNamespace: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 64),
							},
							"permission_policy": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.RowLevelPermissionPolicy](),
							},
							names.AttrStatus: {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[awstypes.LookbackWindowSizeUnit](),
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ColumnDataType](),
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
													Type:             schema.TypeString,
													Computed:         true,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.GeoSpatialDataRole](),
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
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[awstypes.ColumnTagName](),
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.JoinType](),
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.InputColumnDataType](),
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.InputColumnDataType](),
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.InputColumnDataType](),
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
										Type:             schema.TypeString,
										Computed:         true,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.FileFormat](),
									},
									"start_from_row": {
										Type:         schema.TypeInt,
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"text_qualifier": {
										Type:             schema.TypeString,
										Computed:         true,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.TextQualifier](),
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
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	dataSetID := d.Get("data_set_id").(string)
	id := dataSetCreateResourceID(awsAccountID, dataSetID)
	input := &quicksight.CreateDataSetInput{
		AwsAccountId:     aws.String(awsAccountID),
		DataSetId:        aws.String(dataSetID),
		ImportMode:       awstypes.DataSetImportMode(d.Get("import_mode").(string)),
		PhysicalTableMap: expandPhysicalTableMap(d.Get("physical_table_map").(*schema.Set).List()),
		Name:             aws.String(d.Get(names.AttrName).(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("column_groups"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ColumnGroups = expandColumnGroups(v.([]interface{}))
	}

	if v, ok := d.GetOk("column_level_permission_rules"); ok && len(v.([]interface{})) > 0 {
		input.ColumnLevelPermissionRules = expandColumnLevelPermissionRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("data_set_usage_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataSetUsageConfiguration = expandDataSetUsageConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("field_folders"); ok && v.(*schema.Set).Len() != 0 {
		input.FieldFolders = expandFieldFolders(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("logical_table_map"); ok && v.(*schema.Set).Len() != 0 {
		input.LogicalTableMap = expandLogicalTableMap(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && v.(*schema.Set).Len() != 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("row_level_permission_data_set"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RowLevelPermissionDataSet = expandRowLevelPermissionDataSet(v.([]interface{}))
	}

	if v, ok := d.GetOk("row_level_permission_tag_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RowLevelPermissionTagConfiguration = expandRowLevelPermissionTagConfiguration(v.([]interface{}))
	}

	_, err := conn.CreateDataSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Data Set (%s): %s", id, err)
	}

	d.SetId(id)

	if v, ok := d.GetOk("refresh_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &quicksight.PutDataSetRefreshPropertiesInput{
			AwsAccountId:             aws.String(awsAccountID),
			DataSetId:                aws.String(dataSetID),
			DataSetRefreshProperties: expandDataSetRefreshProperties(v.([]interface{})),
		}

		_, err := conn.PutDataSetRefreshProperties(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting QuickSight Data Set (%s) refresh properties: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataSetRead(ctx, d, meta)...)
}

func resourceDataSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dataSetID, err := dataSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dataSet, err := findDataSetByTwoPartKey(ctx, conn, awsAccountID, dataSetID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Data Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, dataSet.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	if err := d.Set("column_groups", flattenColumnGroups(dataSet.ColumnGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting column_groups: %s", err)
	}
	if err := d.Set("column_level_permission_rules", flattenColumnLevelPermissionRules(dataSet.ColumnLevelPermissionRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting column_level_permission_rules: %s", err)
	}
	d.Set("data_set_id", dataSet.DataSetId)
	if err := d.Set("data_set_usage_configuration", flattenDataSetUsageConfiguration(dataSet.DataSetUsageConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_set_usage_configuration: %s", err)
	}
	if err := d.Set("field_folders", flattenFieldFolders(dataSet.FieldFolders)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting field_folders: %s", err)
	}
	d.Set("import_mode", dataSet.ImportMode)
	if err := d.Set("logical_table_map", flattenLogicalLogicalTableMap(dataSet.LogicalTableMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logical_table_map: %s", err)
	}
	d.Set(names.AttrName, dataSet.Name)
	if err := d.Set("output_columns", flattenOutputColumns(dataSet.OutputColumns)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting output_columns: %s", err)
	}
	if err := d.Set("physical_table_map", flattenPhysicalTableMap(dataSet.PhysicalTableMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_table_map: %s", err)
	}
	if err := d.Set("row_level_permission_data_set", flattenRowLevelPermissionDataSet(dataSet.RowLevelPermissionDataSet)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting row_level_permission_data_set: %s", err)
	}
	if err := d.Set("row_level_permission_tag_configuration", flattenRowLevelPermissionTagConfiguration(dataSet.RowLevelPermissionTagConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting row_level_permission_tag_configuration: %s", err)
	}

	permissions, err := findDataSetPermissionsByTwoPartKey(ctx, conn, awsAccountID, dataSetID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Set (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	refreshProperties, err := findDataSetRefreshPropertiesByTwoPartKey(ctx, conn, awsAccountID, dataSetID)

	switch {
	case tfresource.NotFound(err):
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Set (%s) refresh properties: %s", d.Id(), err)
	default:
		if err := d.Set("refresh_properties", flattenDataSetRefreshProperties(refreshProperties)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting refresh properties: %s", err)
		}
	}

	return diags
}

func resourceDataSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dataSetID, err := dataSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll, "refresh_properties") {
		input := &quicksight.UpdateDataSetInput{
			AwsAccountId:                       aws.String(awsAccountID),
			ColumnGroups:                       expandColumnGroups(d.Get("column_groups").([]interface{})),
			ColumnLevelPermissionRules:         expandColumnLevelPermissionRules(d.Get("column_level_permission_rules").([]interface{})),
			DataSetId:                          aws.String(dataSetID),
			DataSetUsageConfiguration:          expandDataSetUsageConfiguration(d.Get("data_set_usage_configuration").([]interface{})),
			FieldFolders:                       expandFieldFolders(d.Get("field_folders").(*schema.Set).List()),
			ImportMode:                         awstypes.DataSetImportMode(d.Get("import_mode").(string)),
			LogicalTableMap:                    expandLogicalTableMap(d.Get("logical_table_map").(*schema.Set).List()),
			Name:                               aws.String(d.Get(names.AttrName).(string)),
			PhysicalTableMap:                   expandPhysicalTableMap(d.Get("physical_table_map").(*schema.Set).List()),
			RowLevelPermissionDataSet:          expandRowLevelPermissionDataSet(d.Get("row_level_permission_data_set").([]interface{})),
			RowLevelPermissionTagConfiguration: expandRowLevelPermissionTagConfiguration(d.Get("row_level_permission_tag_configuration").([]interface{})),
		}

		_, err = conn.UpdateDataSet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Set (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		o, n := d.GetChange(names.AttrPermissions)
		os, ns := o.(*schema.Set), n.(*schema.Set)
		toGrant, toRevoke := quicksightschema.DiffPermissions(os.List(), ns.List())

		input := &quicksight.UpdateDataSetPermissionsInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSetId:    aws.String(dataSetID),
		}

		if len(toGrant) > 0 {
			input.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			input.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateDataSetPermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Set (%s) permissions: %s", d.Id(), err)
		}
	}

	if d.HasChange("refresh_properties") {
		o, n := d.GetChange("refresh_properties")

		if old, new := o.([]interface{}), n.([]interface{}); len(old) == 1 && len(new) == 0 {
			input := &quicksight.DeleteDataSetRefreshPropertiesInput{
				AwsAccountId: aws.String(awsAccountID),
				DataSetId:    aws.String(dataSetID),
			}

			_, err := conn.DeleteDataSetRefreshProperties(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting QuickSight Data Set (%s) refresh properties: %s", d.Id(), err)
			}
		} else {
			input := &quicksight.PutDataSetRefreshPropertiesInput{
				AwsAccountId:             aws.String(awsAccountID),
				DataSetId:                aws.String(dataSetID),
				DataSetRefreshProperties: expandDataSetRefreshProperties(new),
			}

			_, err = conn.PutDataSetRefreshProperties(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "putting QuickSight Data Set (%s) refresh properties (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDataSetRead(ctx, d, meta)...)
}

func resourceDataSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dataSetID, err := dataSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Data Set: %s", d.Id())
	_, err = conn.DeleteDataSet(ctx, &quicksight.DeleteDataSetInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Data Set (%s): %s", d.Id(), err)
	}

	return diags
}

const dataSetResourceIDSeparator = ","

func dataSetCreateResourceID(awsAccountID, dataSetID string) string {
	parts := []string{awsAccountID, dataSetID}
	id := strings.Join(parts, dataSetResourceIDSeparator)

	return id
}

func dataSetParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, dataSetResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sDATA_SET_ID", id, dataSetResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findDataSetByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSetID string) (*awstypes.DataSet, error) {
	input := &quicksight.DescribeDataSetInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
	}

	return findDataSet(ctx, conn, input)
}

func findDataSet(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDataSetInput) (*awstypes.DataSet, error) {
	output, err := conn.DescribeDataSet(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DataSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DataSet, nil
}

func findDataSetRefreshPropertiesByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSetID string) (*awstypes.DataSetRefreshProperties, error) {
	input := &quicksight.DescribeDataSetRefreshPropertiesInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
	}

	return findDataSetRefreshProperties(ctx, conn, input)
}

func findDataSetRefreshProperties(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDataSetRefreshPropertiesInput) (*awstypes.DataSetRefreshProperties, error) {
	output, err := conn.DescribeDataSetRefreshProperties(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "not a SPICE dataset") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DataSetRefreshProperties == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DataSetRefreshProperties, nil
}

func findDataSetPermissionsByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSetID string) ([]awstypes.ResourcePermission, error) {
	input := &quicksight.DescribeDataSetPermissionsInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
	}

	return findDataSetPermissions(ctx, conn, input)
}

func findDataSetPermissions(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDataSetPermissionsInput) ([]awstypes.ResourcePermission, error) {
	output, err := conn.DescribeDataSetPermissions(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Permissions, nil
}

func expandColumnGroups(tfList []interface{}) []awstypes.ColumnGroup {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
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

func expandColumnGroup(tfMap map[string]interface{}) *awstypes.ColumnGroup {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &awstypes.ColumnGroup{}

	if tfMapRaw, ok := tfMap["geo_spatial_column_group"].([]interface{}); ok {
		apiObject.GeoSpatialColumnGroup = expandGeoSpatialColumnGroup(tfMapRaw[0].(map[string]interface{}))
	}

	return apiObject
}

func expandGeoSpatialColumnGroup(tfMap map[string]interface{}) *awstypes.GeoSpatialColumnGroup {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.GeoSpatialColumnGroup{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
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

func expandColumnLevelPermissionRules(tfList []interface{}) []awstypes.ColumnLevelPermissionRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnLevelPermissionRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.ColumnLevelPermissionRule{}

		if v, ok := tfMap["column_names"].([]interface{}); ok {
			apiObject.ColumnNames = flex.ExpandStringValueList(v)
		}
		if v, ok := tfMap["principals"].([]interface{}); ok {
			apiObject.Principals = flex.ExpandStringValueList(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDataSetUsageConfiguration(tfList []interface{}) *awstypes.DataSetUsageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
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

func expandFieldFolders(tfList []interface{}) map[string]awstypes.FieldFolder {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]awstypes.FieldFolder)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.FieldFolder{}

		if v, ok := tfMap["columns"].([]interface{}); ok && len(v) > 0 {
			apiObject.Columns = flex.ExpandStringValueList(v)
		}
		if v, ok := tfMap[names.AttrDescription].(string); ok {
			apiObject.Description = aws.String(v)
		}

		apiObjects[tfMap["field_folders_id"].(string)] = apiObject
	}

	return apiObjects
}

func expandLogicalTableMap(tfList []interface{}) map[string]awstypes.LogicalTable {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]awstypes.LogicalTable)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.LogicalTable{}

		if v, ok := tfMap[names.AttrAlias].(string); ok {
			apiObject.Alias = aws.String(v)
		}
		if v, ok := tfMap[names.AttrSource].([]interface{}); ok {
			apiObject.Source = expandLogicalTableSource(v[0].(map[string]interface{}))
		}
		if v, ok := tfMap["data_transforms"].([]interface{}); ok {
			apiObject.DataTransforms = expandTransformOperations(v)
		}

		apiObjects[tfMap["logical_table_map_id"].(string)] = apiObject
	}

	return apiObjects
}

func expandLogicalTableSource(tfMap map[string]interface{}) *awstypes.LogicalTableSource {
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
	if v, ok := tfMap["join_instruction"].([]interface{}); ok && len(v) > 0 {
		apiObject.JoinInstruction = expandJoinInstruction(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandJoinInstruction(tfMap map[string]interface{}) *awstypes.JoinInstruction {
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
	if v, ok := tfMap["left_join_key_properties"].(map[string]interface{}); ok {
		apiObject.LeftJoinKeyProperties = expandJoinKeyProperties(v)
	}
	if v, ok := tfMap["right_join_key_properties"].(map[string]interface{}); ok {
		apiObject.RightJoinKeyProperties = expandJoinKeyProperties(v)
	}

	return apiObject
}

func expandJoinKeyProperties(tfMap map[string]interface{}) *awstypes.JoinKeyProperties {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.JoinKeyProperties{}

	if v, ok := tfMap["unique_key"].(bool); ok {
		apiObject.UniqueKey = aws.Bool(v)
	}

	return apiObject
}

func expandTransformOperations(tfList []interface{}) []awstypes.TransformOperation {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.TransformOperation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
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

func expandTransformOperation(tfMap map[string]interface{}) awstypes.TransformOperation {
	if tfMap == nil {
		return nil
	}

	var apiObject awstypes.TransformOperation

	if v, ok := tfMap["cast_column_type_operation"].([]interface{}); ok && len(v) > 0 {
		if v := expandCastColumnTypeOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberCastColumnTypeOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["create_columns_operation"].([]interface{}); ok && len(v) > 0 {
		if v := expandCreateColumnsOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberCreateColumnsOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["filter_operation"].([]interface{}); ok && len(v) > 0 {
		if v := expandFilterOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberFilterOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["project_operation"].([]interface{}); ok && len(v) > 0 {
		if v := expandProjectOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberProjectOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["rename_column_operation"].([]interface{}); ok && len(v) > 0 {
		if v := expandRenameColumnOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberRenameColumnOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["tag_column_operation"].([]interface{}); ok && len(v) > 0 {
		if v := expandTagColumnOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberTagColumnOperation{
				Value: *v,
			}
		}
	}
	if v, ok := tfMap["untag_column_operation"].([]interface{}); ok && len(v) > 0 {
		if v := expandUntagColumnOperation(v); v != nil {
			apiObject = &awstypes.TransformOperationMemberUntagColumnOperation{
				Value: *v,
			}
		}
	}

	return apiObject
}

func expandCastColumnTypeOperation(tfList []interface{}) *awstypes.CastColumnTypeOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
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

func expandCreateColumnsOperation(tfList []interface{}) *awstypes.CreateColumnsOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CreateColumnsOperation{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
		apiObject.Columns = expandCalculatedColumns(v)
	}

	return apiObject
}

func expandCalculatedColumns(tfList []interface{}) []awstypes.CalculatedColumn {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CalculatedColumn

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
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

func expandCalculatedColumn(tfMap map[string]interface{}) *awstypes.CalculatedColumn {
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

func expandFilterOperation(tfList []interface{}) *awstypes.FilterOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterOperation{}

	if v, ok := tfMap["condition_expression"].(string); ok {
		apiObject.ConditionExpression = aws.String(v)
	}

	return apiObject
}

func expandProjectOperation(tfList []interface{}) *awstypes.ProjectOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ProjectOperation{}

	if v, ok := tfMap["projected_columns"].([]interface{}); ok && len(v) > 0 {
		apiObject.ProjectedColumns = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandRenameColumnOperation(tfList []interface{}) *awstypes.RenameColumnOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
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

func expandTagColumnOperation(tfList []interface{}) *awstypes.TagColumnOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TagColumnOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		apiObject.ColumnName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrTags].([]interface{}); ok {
		apiObject.Tags = expandColumnTags(v)
	}

	return apiObject
}

func expandColumnTags(tfList []interface{}) []awstypes.ColumnTag {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnTag

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
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

func expandColumnTag(tfMap map[string]interface{}) *awstypes.ColumnTag {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ColumnTag{}

	if v, ok := tfMap["column_description"].(map[string]interface{}); ok {
		apiObject.ColumnDescription = expandColumnDescription(v)
	}
	if v, ok := tfMap["column_geographic_role"].(string); ok {
		apiObject.ColumnGeographicRole = awstypes.GeoSpatialDataRole(v)
	}

	return apiObject
}

func expandColumnDescription(tfMap map[string]interface{}) *awstypes.ColumnDescription {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ColumnDescription{}

	if v, ok := tfMap["text"].(string); ok {
		apiObject.Text = aws.String(v)
	}

	return apiObject
}

func expandUntagColumnOperation(tfList []interface{}) *awstypes.UntagColumnOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.UntagColumnOperation{}

	if v, ok := tfMap["column_name"].(string); ok {
		apiObject.ColumnName = aws.String(v)
	}
	if v, ok := tfMap["tag_names"].([]interface{}); ok {
		apiObject.TagNames = flex.ExpandStringyValueList[awstypes.ColumnTagName](v)
	}

	return apiObject
}

func expandPhysicalTableMap(tfList []interface{}) map[string]awstypes.PhysicalTable {
	apiObjects := make(map[string]awstypes.PhysicalTable)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		var apiObject awstypes.PhysicalTable

		if v, ok := tfMap["custom_sql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			if v := expandCustomSQL(v[0].(map[string]interface{})); v != nil {
				apiObject = &awstypes.PhysicalTableMemberCustomSql{
					Value: *v,
				}
			}
		}
		if v, ok := tfMap["relational_table"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			if v := expandRelationalTable(v[0].(map[string]interface{})); v != nil {
				apiObject = &awstypes.PhysicalTableMemberRelationalTable{
					Value: *v,
				}
			}
		}
		if v, ok := tfMap["s3_source"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			if v := expandS3Source(v[0].(map[string]interface{})); v != nil {
				apiObject = &awstypes.PhysicalTableMemberS3Source{
					Value: *v,
				}
			}
		}

		apiObjects[tfMap["physical_table_map_id"].(string)] = apiObject
	}

	return apiObjects
}

func expandCustomSQL(tfMap map[string]interface{}) *awstypes.CustomSql {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CustomSql{}

	if v, ok := tfMap["columns"].([]interface{}); ok {
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

func expandInputColumns(tfList []interface{}) []awstypes.InputColumn {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.InputColumn

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
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

func expandInputColumn(tfMap map[string]interface{}) *awstypes.InputColumn {
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

func expandRelationalTable(tfMap map[string]interface{}) *awstypes.RelationalTable {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RelationalTable{}

	if v, ok := tfMap["input_columns"].([]interface{}); ok {
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

func expandS3Source(tfMap map[string]interface{}) *awstypes.S3Source {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3Source{}

	if v, ok := tfMap["input_columns"].([]interface{}); ok {
		apiObject.InputColumns = expandInputColumns(v)
	}
	if v, ok := tfMap["upload_settings"].(map[string]interface{}); ok {
		apiObject.UploadSettings = expandUploadSettings(v)
	}
	if v, ok := tfMap["data_source_arn"].(string); ok {
		apiObject.DataSourceArn = aws.String(v)
	}

	return apiObject
}

func expandUploadSettings(tfMap map[string]interface{}) *awstypes.UploadSettings {
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

func expandRowLevelPermissionDataSet(tfList []interface{}) *awstypes.RowLevelPermissionDataSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.RowLevelPermissionDataSet{}

	if v, ok := tfMap[names.AttrARN].(string); ok {
		apiObject.Arn = aws.String(v)
	}
	if v, ok := tfMap["permission_policy"].(string); ok {
		apiObject.PermissionPolicy = awstypes.RowLevelPermissionPolicy(v)
	}
	if v, ok := tfMap["format_version"].(string); ok {
		apiObject.FormatVersion = awstypes.RowLevelPermissionFormatVersion(v)
	}
	if v, ok := tfMap[names.AttrNamespace].(string); ok {
		apiObject.Namespace = aws.String(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok {
		apiObject.Status = awstypes.Status(v)
	}

	return apiObject
}

func expandRowLevelPermissionTagConfiguration(tfList []interface{}) *awstypes.RowLevelPermissionTagConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.RowLevelPermissionTagConfiguration{}

	if v, ok := tfMap["tag_rules"].([]interface{}); ok {
		apiObject.TagRules = expandRowLevelPermissionTagRules(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok {
		apiObject.Status = awstypes.Status(v)
	}

	return apiObject
}

func expandDataSetRefreshProperties(tfList []interface{}) *awstypes.DataSetRefreshProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataSetRefreshProperties{}

	if v, ok := tfMap["refresh_configuration"].([]interface{}); ok {
		apiObject.RefreshConfiguration = expandRefreshConfiguration(v)
	}

	return apiObject
}

func expandRefreshConfiguration(tfList []interface{}) *awstypes.RefreshConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.RefreshConfiguration{}

	if v, ok := tfMap["incremental_refresh"].([]interface{}); ok {
		apiObject.IncrementalRefresh = expandIncrementalRefresh(v)
	}

	return apiObject
}

func expandIncrementalRefresh(tfList []interface{}) *awstypes.IncrementalRefresh {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.IncrementalRefresh{}

	if v, ok := tfMap["lookback_window"].([]interface{}); ok {
		apiObject.LookbackWindow = expandLookbackWindow(v)
	}

	return apiObject
}

func expandLookbackWindow(tfList []interface{}) *awstypes.LookbackWindow {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
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

func expandRowLevelPermissionTagRules(tfList []interface{}) []awstypes.RowLevelPermissionTagRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.RowLevelPermissionTagRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
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

func expandRowLevelPermissionTagRule(tfMap map[string]interface{}) *awstypes.RowLevelPermissionTagRule {
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

func flattenColumnGroups(apiObjects []awstypes.ColumnGroup) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.GeoSpatialColumnGroup != nil {
			tfMap["geo_spatial_column_group"] = flattenGeoSpatialColumnGroup(apiObject.GeoSpatialColumnGroup)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenOutputColumns(apiObjects []awstypes.OutputColumn) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

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

func flattenGeoSpatialColumnGroup(apiObject *awstypes.GeoSpatialColumnGroup) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Columns != nil {
		tfMap["columns"] = apiObject.Columns
	}
	tfMap["country_code"] = apiObject.CountryCode
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}

	return []interface{}{tfMap}
}

func flattenColumnLevelPermissionRules(apiObjects []awstypes.ColumnLevelPermissionRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

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

func flattenDataSetUsageConfiguration(apiObject *awstypes.DataSetUsageConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["disable_use_as_direct_query_source"] = apiObject.DisableUseAsDirectQuerySource
	tfMap["disable_use_as_imported_source"] = apiObject.DisableUseAsImportedSource

	return []interface{}{tfMap}
}

func flattenFieldFolders(apiObjects map[string]awstypes.FieldFolder) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for k, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
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

func flattenLogicalLogicalTableMap(apiObjects map[string]awstypes.LogicalTable) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for k, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
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

func flattenTransformOperations(apiObjects []awstypes.TransformOperation) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

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

func flattenCastColumnTypeOperation(apiObject *awstypes.CastColumnTypeOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.Format != nil {
		tfMap[names.AttrFormat] = aws.ToString(apiObject.Format)
	}
	tfMap["new_column_type"] = apiObject.NewColumnType

	return []interface{}{tfMap}
}

func flattenCreateColumnsOperation(apiObject *awstypes.CreateColumnsOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Columns != nil {
		tfMap["columns"] = flattenCalculatedColumns(apiObject.Columns)
	}

	return []interface{}{tfMap}
}

func flattenCalculatedColumns(apiObjects []awstypes.CalculatedColumn) interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

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

func flattenFilterOperation(apiObject *awstypes.FilterOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ConditionExpression != nil {
		tfMap["condition_expression"] = aws.ToString(apiObject.ConditionExpression)
	}

	return []interface{}{tfMap}
}

func flattenProjectOperation(apiObject *awstypes.ProjectOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ProjectedColumns != nil {
		tfMap["projected_columns"] = apiObject.ProjectedColumns
	}

	return []interface{}{tfMap}
}

func flattenRenameColumnOperation(apiObject *awstypes.RenameColumnOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.NewColumnName != nil {
		tfMap["new_column_name"] = aws.ToString(apiObject.NewColumnName)
	}

	return []interface{}{tfMap}
}

func flattenTagColumnOperation(apiObject *awstypes.TagColumnOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.Tags != nil {
		tfMap[names.AttrTags] = flattenColumnTags(apiObject.Tags)
	}

	return []interface{}{tfMap}
}

func flattenColumnTags(apiObjects []awstypes.ColumnTag) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.ColumnDescription != nil {
			tfMap["column_description"] = flattenColumnDescription(apiObject.ColumnDescription)
		}
		tfMap["column_geographic_role"] = apiObject.ColumnGeographicRole

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnDescription(apiObject *awstypes.ColumnDescription) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Text != nil {
		tfMap["text"] = aws.ToString(apiObject.Text)
	}

	return []interface{}{tfMap}
}

func flattenUntagColumnOperation(apiObject *awstypes.UntagColumnOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.TagNames != nil {
		tfMap["tag_names"] = apiObject.TagNames
	}

	return []interface{}{tfMap}
}

func flattenLogicalTableSource(apiObject *awstypes.LogicalTableSource) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DataSetArn != nil {
		tfMap["data_set_arn"] = aws.ToString(apiObject.DataSetArn)
	}
	if apiObject.JoinInstruction != nil {
		tfMap["join_instruction"] = flattenJoinInstruction(apiObject.JoinInstruction)
	}
	if apiObject.PhysicalTableId != nil {
		tfMap["physical_table_id"] = aws.ToString(apiObject.PhysicalTableId)
	}

	return []interface{}{tfMap}
}

func flattenJoinInstruction(apiObject *awstypes.JoinInstruction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

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

	return []interface{}{tfMap}
}

func flattenJoinKeyProperties(apiObject *awstypes.JoinKeyProperties) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.UniqueKey != nil {
		tfMap["unique_key"] = aws.ToBool(apiObject.UniqueKey)
	}

	return tfMap
}

func flattenPhysicalTableMap(apiObjects map[string]awstypes.PhysicalTable) []interface{} {
	var tfList []interface{}

	for k, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
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

func flattenCustomSQL(apiObject *awstypes.CustomSql) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

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

	return []interface{}{tfMap}
}

func flattenInputColumns(apiObjects []awstypes.InputColumn) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.Name != nil {
			tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		}
		tfMap[names.AttrType] = apiObject.Type

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRelationalTable(apiObject *awstypes.RelationalTable) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

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

	return []interface{}{tfMap}
}

func flattenS3Source(apiObject *awstypes.S3Source) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DataSourceArn != nil {
		tfMap["data_source_arn"] = aws.ToString(apiObject.DataSourceArn)
	}
	if apiObject.InputColumns != nil {
		tfMap["input_columns"] = flattenInputColumns(apiObject.InputColumns)
	}
	if apiObject.UploadSettings != nil {
		tfMap["upload_settings"] = flattenUploadSettings(apiObject.UploadSettings)
	}

	return []interface{}{tfMap}
}

func flattenUploadSettings(apiObject *awstypes.UploadSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

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

	return []interface{}{tfMap}
}

func flattenRowLevelPermissionDataSet(apiObject *awstypes.RowLevelPermissionDataSet) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Arn != nil {
		tfMap[names.AttrARN] = aws.ToString(apiObject.Arn)
	}
	tfMap["format_version"] = apiObject.FormatVersion
	if apiObject.Namespace != nil {
		tfMap[names.AttrNamespace] = aws.ToString(apiObject.Namespace)
	}
	tfMap["permission_policy"] = apiObject.PermissionPolicy
	tfMap[names.AttrStatus] = apiObject.Status

	return []interface{}{tfMap}
}

func flattenRowLevelPermissionTagConfiguration(apiObject *awstypes.RowLevelPermissionTagConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrStatus] = apiObject.Status
	if apiObject.TagRules != nil {
		tfMap["tag_rules"] = flattenRowLevelPermissionTagRules(apiObject.TagRules)
	}

	return []interface{}{tfMap}
}

func flattenDataSetRefreshProperties(apiObject *awstypes.DataSetRefreshProperties) interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.RefreshConfiguration != nil {
		tfMap["refresh_configuration"] = flattenRefreshConfiguration(apiObject.RefreshConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenRefreshConfiguration(apiObject *awstypes.RefreshConfiguration) interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.IncrementalRefresh != nil {
		tfMap["incremental_refresh"] = flattenIncrementalRefresh(apiObject.IncrementalRefresh)
	}

	return []interface{}{tfMap}
}

func flattenIncrementalRefresh(apiObject *awstypes.IncrementalRefresh) interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.LookbackWindow != nil {
		tfMap["lookback_window"] = flattenLookbackWindow(apiObject.LookbackWindow)
	}

	return []interface{}{tfMap}
}

func flattenLookbackWindow(apiObject *awstypes.LookbackWindow) interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.Size != nil {
		tfMap[names.AttrSize] = aws.ToInt64(apiObject.Size)
	}
	tfMap["size_unit"] = apiObject.SizeUnit

	return []interface{}{tfMap}
}

func flattenRowLevelPermissionTagRules(apiObjects []awstypes.RowLevelPermissionTagRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

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
