// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_glue_catalog_tables", name="Catalog Tables")
func dataSourceCatalogTables() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCatalogTablesRead,

		Schema: map[string]*schema.Schema{
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"expression": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regular expression to filter the list of table names",
			},
			"table_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The type of tables to return. Valid values are EXTERNAL_TABLE, MANAGED_TABLE, VIRTUAL_VIEW",
			},
			"tables": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrOwner: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrParameters: {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"partition_keys": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrComment: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrParameters: {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrType: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"retention": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"storage_descriptor": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"additional_locations": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"bucket_columns": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"columns": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrComment: {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrName: {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrParameters: {
													Type:     schema.TypeMap,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												names.AttrType: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"compressed": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"input_format": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrLocation: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"number_of_buckets": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"output_format": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrParameters: {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"ser_de_info": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrName: {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrParameters: {
													Type:     schema.TypeMap,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"serialization_library": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"schema_reference": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"schema_id": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"registry_name": {
																Type:     schema.TypeString,
																Computed: true,
															},
															"schema_arn": {
																Type:     schema.TypeString,
																Computed: true,
															},
															"schema_name": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
												},
												"schema_version_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"schema_version_number": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"skewed_info": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"skewed_column_names": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"skewed_column_value_location_maps": {
													Type:     schema.TypeMap,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"skewed_column_values": {
													Type:     schema.TypeList,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"sort_columns": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"column": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"sort_order": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"stored_as_sub_directories": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"table_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_table": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrCatalogID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrDatabaseName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"view_original_text": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"view_expanded_text": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCatalogTablesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID(ctx))
	dbName := d.Get(names.AttrDatabaseName).(string)

	d.SetId(fmt.Sprintf("%s:%s", catalogID, dbName))

	input := &glue.GetTablesInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
	}

	if v, ok := d.GetOk("expression"); ok {
		input.Expression = aws.String(v.(string))
	}

	var tables []awstypes.Table
	paginator := glue.NewGetTablesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Tables for database %s: %s", dbName, err)
		}

		tables = append(tables, page.TableList...)
	}

	// Filter by table type if specified
	if v, ok := d.GetOk("table_type"); ok {
		tableType := v.(string)
		var filteredTables []awstypes.Table
		for _, table := range tables {
			if aws.ToString(table.TableType) == tableType {
				filteredTables = append(filteredTables, table)
			}
		}
		tables = filteredTables
	}

	err := d.Set(names.AttrCatalogID, catalogID)
	if err != nil {
		return nil
	}
	err = d.Set(names.AttrDatabaseName, dbName)
	if err != nil {
		return nil
	}

	if err := d.Set("tables", flattenCatalogTables(ctx, tables, catalogID, dbName, meta.(*conns.AWSClient))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tables: %s", err)
	}

	return diags
}

func flattenCatalogTables(ctx context.Context, tables []awstypes.Table, catalogID, dbName string, awsClient *conns.AWSClient) []map[string]any {
	if len(tables) == 0 {
		return nil
	}

	var tfList []map[string]any

	for _, table := range tables {
		tfMap := map[string]any{}

		tableArn := arn.ARN{
			Partition: awsClient.Partition(ctx),
			Service:   "glue",
			Region:    awsClient.Region(ctx),
			AccountID: awsClient.AccountID(ctx),
			Resource:  fmt.Sprintf("table/%s/%s", dbName, aws.ToString(table.Name)),
		}.String()
		tfMap[names.AttrARN] = tableArn

		tfMap[names.AttrCatalogID] = catalogID
		tfMap[names.AttrDatabaseName] = dbName
		tfMap[names.AttrName] = aws.ToString(table.Name)
		tfMap[names.AttrDescription] = aws.ToString(table.Description)
		tfMap[names.AttrOwner] = aws.ToString(table.Owner)
		tfMap["retention"] = table.Retention
		tfMap[names.AttrParameters] = table.Parameters
		tfMap["table_type"] = aws.ToString(table.TableType)
		tfMap["view_original_text"] = aws.ToString(table.ViewOriginalText)
		tfMap["view_expanded_text"] = aws.ToString(table.ViewExpandedText)

		if table.StorageDescriptor != nil {
			tfMap["storage_descriptor"] = flattenStorageDescriptor(table.StorageDescriptor)
		}

		if len(table.PartitionKeys) > 0 {
			tfMap["partition_keys"] = flattenColumns(table.PartitionKeys)
		}

		if table.TargetTable != nil {
			tfMap["target_table"] = []map[string]any{flattenTableTargetTable(table.TargetTable)}
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
