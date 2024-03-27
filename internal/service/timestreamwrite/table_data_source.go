// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_timestreamwrite_table", name="Table")
func DataSourceTable() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTableRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"magnetic_store_write_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_magnetic_store_writes": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"magnetic_store_rejected_data_location": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_configuration": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"encryption_option": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"kms_key_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"object_key_prefix": {
													Type:     schema.TypeString,
													Computed: true,
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
			"retention_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"magnetic_store_retention_period_in_days": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory_store_retention_period_in_hours": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"schema": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"composite_partition_key": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enforcement_in_record": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameTable = "Table Data Source"
)

func dataSourceTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	databaseName := d.Get("database_name").(string)
	tableName := d.Get("table_name").(string)
	id := tableCreateResourceID(tableName, databaseName)

	table, err := findTableByTwoPartKey(ctx, conn, databaseName, tableName)
	if err != nil {
		return create.AppendDiagError(diags, names.TimestreamWrite, create.ErrActionReading, DSNameTable, tableName, err)
	}

	d.SetId(id)

	d.Set("arn", table.Arn)
	d.Set("database_name", table.DatabaseName)
	d.Set("table_name", table.TableName)

	if err := d.Set("magnetic_store_write_properties", flattenMagneticStoreWriteProperties(table.MagneticStoreWriteProperties)); err != nil {
		return diag.Errorf("setting magnetic_store_write_properties: %s", err)
	}
	if err := d.Set("retention_properties", flattenRetentionProperties(table.RetentionProperties)); err != nil {
		return diag.Errorf("setting retention_properties: %s", err)
	}
	if table.Schema != nil {
		if err := d.Set("schema", []interface{}{flattenSchema(table.Schema)}); err != nil {
			return diag.Errorf("setting schema: %s", err)
		}
	} else {
		d.Set("schema", nil)
	}

	tags, err := listTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return diag.Errorf("listing tags for timestream table (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return diags
}
