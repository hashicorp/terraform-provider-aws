package glue

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceCatalogTable() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCatalogTableRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringDoesNotMatch(regexp.MustCompile(`[A-Z]`), "uppercase characters cannot be used"),
				),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameters": {
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
						"comment": {
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
									"comment": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"parameters": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"type": {
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
						"location": {
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
						"parameters": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ser_de_info": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"parameters": {
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
						"catalog_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"database_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
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
	conn := meta.(*conns.AWSClient).GlueConn()

	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	dbName := d.Get("database_name").(string)
	name := d.Get("name").(string)

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

	out, err := conn.GetTableWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return diag.Errorf("No Glue table %s found for catalog_id: %s, database_name: %s", name, catalogID,
				dbName)
		}

		return diag.Errorf("Error reading Glue Catalog Table: %s", err)
	}

	table := out.Table
	tableArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("table/%s/%s", dbName, aws.StringValue(table.Name)),
	}.String()
	d.Set("arn", tableArn)

	d.Set("name", table.Name)
	d.Set("catalog_id", catalogID)
	d.Set("database_name", dbName)
	d.Set("description", table.Description)
	d.Set("owner", table.Owner)
	d.Set("retention", table.Retention)

	if err := d.Set("storage_descriptor", flattenStorageDescriptor(table.StorageDescriptor)); err != nil {
		return diag.Errorf("error setting storage_descriptor: %s", err)
	}

	if err := d.Set("partition_keys", flattenColumns(table.PartitionKeys)); err != nil {
		return diag.Errorf("error setting partition_keys: %s", err)
	}

	d.Set("view_original_text", table.ViewOriginalText)
	d.Set("view_expanded_text", table.ViewExpandedText)
	d.Set("table_type", table.TableType)

	if err := d.Set("parameters", aws.StringValueMap(table.Parameters)); err != nil {
		return diag.Errorf("error setting parameters: %s", err)
	}

	if table.TargetTable != nil {
		if err := d.Set("target_table", []interface{}{flattenTableTargetTable(table.TargetTable)}); err != nil {
			return diag.Errorf("error setting target_table: %s", err)
		}
	} else {
		d.Set("target_table", nil)
	}

	partIndexInput := &glue.GetPartitionIndexesInput{
		CatalogId:    out.Table.CatalogId,
		TableName:    out.Table.Name,
		DatabaseName: out.Table.DatabaseName,
	}
	partOut, err := conn.GetPartitionIndexesWithContext(ctx, partIndexInput)
	if err != nil {
		return diag.Errorf("error getting Glue Partition Indexes: %s", err)
	}

	if partOut != nil && len(partOut.PartitionIndexDescriptorList) > 0 {
		if err := d.Set("partition_index", flattenPartitionIndexes(partOut.PartitionIndexDescriptorList)); err != nil {
			return diag.Errorf("error setting partition_index: %s", err)
		}
	}

	return nil
}
