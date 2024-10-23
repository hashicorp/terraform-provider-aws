// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_glue_catalog_table")
func DataSourceCatalogTable() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCatalogTableRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringDoesNotMatch(regexache.MustCompile(`[A-Z]`), "uppercase characters cannot be used"),
				),
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
			"partition_index": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"index_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"keys": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
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
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"query_as_of_time": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"transaction_id"},
				ValidateFunc:  validation.IsRFC3339Time,
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
			"transaction_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"query_as_of_time"},
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
	}
}

func dataSourceCatalogTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	dbName := d.Get(names.AttrDatabaseName).(string)
	name := d.Get(names.AttrName).(string)

	d.SetId(fmt.Sprintf("%s:%s:%s", catalogID, dbName, name))

	input := &glue.GetTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		Name:         aws.String(name),
	}

	if v, ok := d.GetOk("query_as_of_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.QueryAsOfTime = aws.Time(t)
	}
	if v, ok := d.GetOk("transaction_id"); ok {
		input.TransactionId = aws.String(v.(string))
	}

	out, err := conn.GetTable(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return sdkdiag.AppendErrorf(diags, "No Glue table %s found for catalog_id: %s, database_name: %s", name, catalogID,
				dbName)
		}

		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table: %s", err)
	}

	table := out.Table
	tableArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("table/%s/%s", dbName, aws.ToString(table.Name)),
	}.String()
	d.Set(names.AttrARN, tableArn)

	d.Set(names.AttrName, table.Name)
	d.Set(names.AttrCatalogID, catalogID)
	d.Set(names.AttrDatabaseName, dbName)
	d.Set(names.AttrDescription, table.Description)
	d.Set(names.AttrOwner, table.Owner)
	d.Set("retention", table.Retention)

	if err := d.Set("storage_descriptor", flattenStorageDescriptor(table.StorageDescriptor)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_descriptor: %s", err)
	}

	if err := d.Set("partition_keys", flattenColumns(table.PartitionKeys)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting partition_keys: %s", err)
	}

	d.Set("view_original_text", table.ViewOriginalText)
	d.Set("view_expanded_text", table.ViewExpandedText)
	d.Set("table_type", table.TableType)

	if err := d.Set(names.AttrParameters, table.Parameters); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}

	if table.TargetTable != nil {
		if err := d.Set("target_table", []interface{}{flattenTableTargetTable(table.TargetTable)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_table: %s", err)
		}
	} else {
		d.Set("target_table", nil)
	}

	partIndexInput := &glue.GetPartitionIndexesInput{
		CatalogId:    out.Table.CatalogId,
		TableName:    out.Table.Name,
		DatabaseName: out.Table.DatabaseName,
	}
	partOut, err := conn.GetPartitionIndexes(ctx, partIndexInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Glue Partition Indexes: %s", err)
	}

	if partOut != nil && len(partOut.PartitionIndexDescriptorList) > 0 {
		if err := d.Set("partition_index", flattenPartitionIndexes(partOut.PartitionIndexDescriptorList)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting partition_index: %s", err)
		}
	}

	return diags
}
