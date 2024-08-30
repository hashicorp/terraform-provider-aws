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
		PhysicalTableMap: quicksightschema.ExpandPhysicalTableMap(d.Get("physical_table_map").(*schema.Set).List()),
		Name:             aws.String(d.Get(names.AttrName).(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("column_groups"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ColumnGroups = quicksightschema.ExpandColumnGroups(v.([]interface{}))
	}

	if v, ok := d.GetOk("column_level_permission_rules"); ok && len(v.([]interface{})) > 0 {
		input.ColumnLevelPermissionRules = quicksightschema.ExpandColumnLevelPermissionRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("data_set_usage_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataSetUsageConfiguration = quicksightschema.ExpandDataSetUsageConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("field_folders"); ok && v.(*schema.Set).Len() != 0 {
		input.FieldFolders = quicksightschema.ExpandFieldFolders(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("logical_table_map"); ok && v.(*schema.Set).Len() != 0 {
		input.LogicalTableMap = quicksightschema.ExpandLogicalTableMap(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && v.(*schema.Set).Len() != 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("row_level_permission_data_set"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RowLevelPermissionDataSet = quicksightschema.ExpandRowLevelPermissionDataSet(v.([]interface{}))
	}

	if v, ok := d.GetOk("row_level_permission_tag_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RowLevelPermissionTagConfiguration = quicksightschema.ExpandRowLevelPermissionTagConfiguration(v.([]interface{}))
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
			DataSetRefreshProperties: quicksightschema.ExpandDataSetRefreshProperties(v.([]interface{})),
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
	if err := d.Set("column_groups", quicksightschema.FlattenColumnGroups(dataSet.ColumnGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting column_groups: %s", err)
	}
	if err := d.Set("column_level_permission_rules", quicksightschema.FlattenColumnLevelPermissionRules(dataSet.ColumnLevelPermissionRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting column_level_permission_rules: %s", err)
	}
	d.Set("data_set_id", dataSet.DataSetId)
	if err := d.Set("data_set_usage_configuration", quicksightschema.FlattenDataSetUsageConfiguration(dataSet.DataSetUsageConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_set_usage_configuration: %s", err)
	}
	if err := d.Set("field_folders", quicksightschema.FlattenFieldFolders(dataSet.FieldFolders)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting field_folders: %s", err)
	}
	d.Set("import_mode", dataSet.ImportMode)
	if err := d.Set("logical_table_map", quicksightschema.FlattenLogicalTableMap(dataSet.LogicalTableMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logical_table_map: %s", err)
	}
	d.Set(names.AttrName, dataSet.Name)
	if err := d.Set("output_columns", quicksightschema.FlattenOutputColumns(dataSet.OutputColumns)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting output_columns: %s", err)
	}
	if err := d.Set("physical_table_map", quicksightschema.FlattenPhysicalTableMap(dataSet.PhysicalTableMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_table_map: %s", err)
	}
	if err := d.Set("row_level_permission_data_set", quicksightschema.FlattenRowLevelPermissionDataSet(dataSet.RowLevelPermissionDataSet)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting row_level_permission_data_set: %s", err)
	}
	if err := d.Set("row_level_permission_tag_configuration", quicksightschema.FlattenRowLevelPermissionTagConfiguration(dataSet.RowLevelPermissionTagConfiguration)); err != nil {
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
		if err := d.Set("refresh_properties", quicksightschema.FlattenDataSetRefreshProperties(refreshProperties)); err != nil {
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
			ColumnGroups:                       quicksightschema.ExpandColumnGroups(d.Get("column_groups").([]interface{})),
			ColumnLevelPermissionRules:         quicksightschema.ExpandColumnLevelPermissionRules(d.Get("column_level_permission_rules").([]interface{})),
			DataSetId:                          aws.String(dataSetID),
			DataSetUsageConfiguration:          quicksightschema.ExpandDataSetUsageConfiguration(d.Get("data_set_usage_configuration").([]interface{})),
			FieldFolders:                       quicksightschema.ExpandFieldFolders(d.Get("field_folders").(*schema.Set).List()),
			ImportMode:                         awstypes.DataSetImportMode(d.Get("import_mode").(string)),
			LogicalTableMap:                    quicksightschema.ExpandLogicalTableMap(d.Get("logical_table_map").(*schema.Set).List()),
			Name:                               aws.String(d.Get(names.AttrName).(string)),
			PhysicalTableMap:                   quicksightschema.ExpandPhysicalTableMap(d.Get("physical_table_map").(*schema.Set).List()),
			RowLevelPermissionDataSet:          quicksightschema.ExpandRowLevelPermissionDataSet(d.Get("row_level_permission_data_set").([]interface{})),
			RowLevelPermissionTagConfiguration: quicksightschema.ExpandRowLevelPermissionTagConfiguration(d.Get("row_level_permission_tag_configuration").([]interface{})),
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
				DataSetRefreshProperties: quicksightschema.ExpandDataSetRefreshProperties(new),
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
