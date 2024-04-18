// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_glue_partition")
func ResourcePartition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePartitionCreate,
		ReadWithoutTimeout:   resourcePartitionRead,
		UpdateWithoutTimeout: resourcePartitionUpdate,
		DeleteWithoutTimeout: resourcePartitionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"database_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"table_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"partition_values": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 1024),
				},
			},
			"storage_descriptor": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"comment": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"compressed": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"input_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"number_of_buckets": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"output_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ser_de_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"parameters": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"serialization_library": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"skewed_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"skewed_column_names": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"skewed_column_values": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"skewed_column_value_location_maps": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"sort_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"column": {
										Type:     schema.TypeString,
										Required: true,
									},
									"sort_order": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"stored_as_sub_directories": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_analyzed_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_accessed_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePartitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)
	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	dbName := d.Get("database_name").(string)
	tableName := d.Get("table_name").(string)
	values := d.Get("partition_values").([]interface{})

	input := &glue.CreatePartitionInput{
		CatalogId:      aws.String(catalogID),
		DatabaseName:   aws.String(dbName),
		TableName:      aws.String(tableName),
		PartitionInput: expandPartitionInput(d),
	}

	log.Printf("[DEBUG] Creating Glue Partition: %#v", input)
	_, err := conn.CreatePartitionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Partition: %s", err)
	}

	d.SetId(createPartitionID(catalogID, dbName, tableName, values))

	return append(diags, resourcePartitionRead(ctx, d, meta)...)
}

func resourcePartitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	log.Printf("[DEBUG] Reading Glue Partition: %s", d.Id())
	partition, err := FindPartitionByValues(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Glue Partition (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Glue Partition: %s", err)
	}

	d.Set("table_name", partition.TableName)
	d.Set("catalog_id", partition.CatalogId)
	d.Set("database_name", partition.DatabaseName)
	d.Set("partition_values", flex.FlattenStringList(partition.Values))

	if partition.LastAccessTime != nil {
		d.Set("last_accessed_time", partition.LastAccessTime.Format(time.RFC3339))
	}

	if partition.LastAnalyzedTime != nil {
		d.Set("last_analyzed_time", partition.LastAnalyzedTime.Format(time.RFC3339))
	}

	if partition.CreationTime != nil {
		d.Set("creation_time", partition.CreationTime.Format(time.RFC3339))
	}

	if err := d.Set("storage_descriptor", flattenStorageDescriptor(partition.StorageDescriptor)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_descriptor: %s", err)
	}

	if err := d.Set("parameters", aws.StringValueMap(partition.Parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}

	return diags
}

func resourcePartitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	catalogID, dbName, tableName, values, err := readPartitionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Glue Partition (%s): %s", d.Id(), err)
	}

	input := &glue.UpdatePartitionInput{
		CatalogId:          aws.String(catalogID),
		DatabaseName:       aws.String(dbName),
		TableName:          aws.String(tableName),
		PartitionInput:     expandPartitionInput(d),
		PartitionValueList: aws.StringSlice(values),
	}

	if _, err := conn.UpdatePartitionWithContext(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Glue Partition (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePartitionRead(ctx, d, meta)...)
}

func resourcePartitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	catalogID, dbName, tableName, values, err := readPartitionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Partition: %s", err)
	}

	log.Printf("[DEBUG] Deleting Glue Partition: %s", d.Id())
	_, err = conn.DeletePartitionWithContext(ctx, &glue.DeletePartitionInput{
		CatalogId:       aws.String(catalogID),
		TableName:       aws.String(tableName),
		DatabaseName:    aws.String(dbName),
		PartitionValues: aws.StringSlice(values),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Partition: %s", err)
	}
	return diags
}

func expandPartitionInput(d *schema.ResourceData) *glue.PartitionInput {
	tableInput := &glue.PartitionInput{}

	if v, ok := d.GetOk("storage_descriptor"); ok {
		tableInput.StorageDescriptor = expandStorageDescriptor(v.([]interface{}))
	}

	if v, ok := d.GetOk("parameters"); ok {
		tableInput.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("partition_values"); ok && len(v.([]interface{})) > 0 {
		tableInput.Values = flex.ExpandStringList(v.([]interface{}))
	}

	return tableInput
}
