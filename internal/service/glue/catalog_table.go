// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_catalog_table")
func ResourceCatalogTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCatalogTableCreate,
		ReadWithoutTimeout:   resourceCatalogTableRead,
		UpdateWithoutTimeout: resourceCatalogTableUpdate,
		DeleteWithoutTimeout: resourceCatalogTableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringDoesNotMatch(regexache.MustCompile(`[A-Z]`), "uppercase characters cannot be used"),
				),
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"partition_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrComment: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 131072),
						},
					},
				},
			},
			"retention": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"storage_descriptor": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_locations": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"bucket_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
						},
						"columns": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrComment: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 255),
									},
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									names.AttrParameters: {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrType: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 131072),
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
						names.AttrLocation: {
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
						names.AttrParameters: {
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
									names.AttrName: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									names.AttrParameters: {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"serialization_library": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"schema_reference": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"schema_id": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"schema_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
													ExactlyOneOf: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn", "storage_descriptor.0.schema_reference.0.schema_id.0.schema_name"},
												},
												"schema_name": {
													Type:         schema.TypeString,
													Optional:     true,
													ExactlyOneOf: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn", "storage_descriptor.0.schema_reference.0.schema_id.0.schema_name"},
												},
												"registry_name": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn"},
												},
											},
										},
									},
									"schema_version_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ExactlyOneOf: []string{"storage_descriptor.0.schema_reference.0.schema_version_id", "storage_descriptor.0.schema_reference.0.schema_id"},
									},
									"schema_version_number": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 100000),
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
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"sort_order": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntInSlice([]int{0, 1}),
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
			"table_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"open_table_format_input": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iceberg_input": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metadata_operation": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"CREATE"}, false),
									},
									names.AttrVersion: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
								},
							},
						},
					},
				},
			},
			"target_table": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrRegion: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"view_original_text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 409600),
			},
			"view_expanded_text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 409600),
			},
			"partition_index": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"keys": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"index_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func ReadTableID(id string) (string, string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format catalog-id:database-name:table-name, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func resourceCatalogTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)
	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	dbName := d.Get(names.AttrDatabaseName).(string)
	name := d.Get(names.AttrName).(string)

	input := &glue.CreateTableInput{
		CatalogId:            aws.String(catalogID),
		DatabaseName:         aws.String(dbName),
		OpenTableFormatInput: expandOpenTableFormat(d),
		TableInput:           expandTableInput(d),
		PartitionIndexes:     expandTablePartitionIndexes(d.Get("partition_index").([]interface{})),
	}

	_, err := conn.CreateTable(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Catalog Table (%s): %s", name, err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", catalogID, dbName, name))

	return append(diags, resourceCatalogTableRead(ctx, d, meta)...)
}

func resourceCatalogTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, name, err := ReadTableID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	table, err := FindTableByName(ctx, conn, catalogID, dbName, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Catalog Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table (%s): %s", d.Id(), err)
	}

	tableArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("table/%s/%s", dbName, aws.ToString(table.Name)),
	}.String()
	d.Set(names.AttrARN, tableArn)
	d.Set(names.AttrCatalogID, catalogID)
	d.Set(names.AttrDatabaseName, dbName)
	d.Set(names.AttrDescription, table.Description)
	d.Set(names.AttrName, table.Name)
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

	if err := d.Set(names.AttrParameters, flattenNonManagedParameters(table)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}

	if table.TargetTable != nil {
		if err := d.Set("target_table", []interface{}{flattenTableTargetTable(table.TargetTable)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_table: %s", err)
		}
	} else {
		d.Set("target_table", nil)
	}

	input := &glue.GetPartitionIndexesInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(name),
	}

	output, err := conn.GetPartitionIndexes(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table (%s) partition indexes: %s", d.Id(), err)
	}

	if output != nil && len(output.PartitionIndexDescriptorList) > 0 {
		if err := d.Set("partition_index", flattenPartitionIndexes(output.PartitionIndexDescriptorList)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting partition_index: %s", err)
		}
	}

	return diags
}

func resourceCatalogTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, name, err := ReadTableID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &glue.UpdateTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableInput:   expandTableInput(d),
	}

	// Add back any managed parameters. See flattenNonManagedParameters.
	table, err := FindTableByName(ctx, conn, catalogID, dbName, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table (%s): %s", d.Id(), err)
	}

	if allParameters := table.Parameters; allParameters["table_type"] == "ICEBERG" {
		for _, k := range []string{"table_type", "metadata_location"} {
			if v := allParameters[k]; v != "" {
				if input.TableInput.Parameters == nil {
					input.TableInput.Parameters = make(map[string]string)
				}
				input.TableInput.Parameters[k] = v
			}
		}
	}

	_, err = conn.UpdateTable(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Glue Catalog Table (%s): %s", d.Id(), err)
	}

	return append(diags, resourceCatalogTableRead(ctx, d, meta)...)
}

func resourceCatalogTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, name, err := ReadTableID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Glue Catalog Table: %s", d.Id())
	_, err = conn.DeleteTable(ctx, &glue.DeleteTableInput{
		CatalogId:    aws.String(catalogID),
		Name:         aws.String(name),
		DatabaseName: aws.String(dbName),
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Catalog Table (%s): %s", d.Id(), err)
	}

	return diags
}

func FindTableByName(ctx context.Context, conn *glue.Client, catalogID, dbName, name string) (*awstypes.Table, error) {
	input := &glue.GetTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		Name:         aws.String(name),
	}

	output, err := conn.GetTable(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Table == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Table, nil
}

func expandTableInput(d *schema.ResourceData) *awstypes.TableInput {
	tableInput := &awstypes.TableInput{
		Name: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		tableInput.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrOwner); ok {
		tableInput.Owner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retention"); ok {
		tableInput.Retention = int32(v.(int))
	}

	if v, ok := d.GetOk("storage_descriptor"); ok {
		tableInput.StorageDescriptor = expandStorageDescriptor(v.([]interface{}))
	}

	if v, ok := d.GetOk("partition_keys"); ok {
		tableInput.PartitionKeys = expandColumns(v.([]interface{}))
	} else if _, ok = d.GetOk("open_table_format_input"); !ok {
		tableInput.PartitionKeys = []awstypes.Column{}
	}

	if v, ok := d.GetOk("view_original_text"); ok {
		tableInput.ViewOriginalText = aws.String(v.(string))
	}

	if v, ok := d.GetOk("view_expanded_text"); ok {
		tableInput.ViewExpandedText = aws.String(v.(string))
	}

	if v, ok := d.GetOk("table_type"); ok {
		tableInput.TableType = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrParameters); ok {
		tableInput.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("target_table"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tableInput.TargetTable = expandTableTargetTable(v.([]interface{})[0].(map[string]interface{}))
	}

	return tableInput
}

func expandOpenTableFormat(s *schema.ResourceData) *awstypes.OpenTableFormatInput {
	if v, ok := s.GetOk("open_table_format_input"); ok {
		openTableFormatInput := &awstypes.OpenTableFormatInput{
			IcebergInput: expandIcebergInput(v.([]interface{})[0].(map[string]interface{})),
		}
		return openTableFormatInput
	}
	return nil
}

func expandIcebergInput(s map[string]interface{}) *awstypes.IcebergInput {
	var iceberg = s["iceberg_input"].([]interface{})[0].(map[string]interface{})
	icebergInput := &awstypes.IcebergInput{
		MetadataOperation: awstypes.MetadataOperation(iceberg["metadata_operation"].(string)),
	}
	if v, ok := iceberg[names.AttrVersion].(string); ok && v != "" {
		icebergInput.Version = aws.String(v)
	}
	return icebergInput
}

func expandTablePartitionIndexes(a []interface{}) []awstypes.PartitionIndex {
	partitionIndexes := make([]awstypes.PartitionIndex, 0, len(a))

	for _, m := range a {
		partitionIndexes = append(partitionIndexes, expandTablePartitionIndex(m.(map[string]interface{})))
	}

	return partitionIndexes
}

func expandTablePartitionIndex(m map[string]interface{}) awstypes.PartitionIndex {
	partitionIndex := awstypes.PartitionIndex{
		IndexName: aws.String(m["index_name"].(string)),
		Keys:      flex.ExpandStringValueList(m["keys"].([]interface{})),
	}

	return partitionIndex
}

func expandStorageDescriptor(l []interface{}) *awstypes.StorageDescriptor {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	storageDescriptor := &awstypes.StorageDescriptor{}

	if v, ok := s["additional_locations"]; ok {
		storageDescriptor.AdditionalLocations = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := s["columns"]; ok {
		storageDescriptor.Columns = expandColumns(v.([]interface{}))
	}

	if v, ok := s[names.AttrLocation]; ok {
		storageDescriptor.Location = aws.String(v.(string))
	}

	if v, ok := s["input_format"]; ok {
		storageDescriptor.InputFormat = aws.String(v.(string))
	}

	if v, ok := s["output_format"]; ok {
		storageDescriptor.OutputFormat = aws.String(v.(string))
	}

	if v, ok := s["compressed"]; ok {
		storageDescriptor.Compressed = v.(bool)
	}

	if v, ok := s["number_of_buckets"]; ok {
		storageDescriptor.NumberOfBuckets = int32(v.(int))
	}

	if v, ok := s["ser_de_info"]; ok {
		storageDescriptor.SerdeInfo = expandSerDeInfo(v.([]interface{}))
	}

	if v, ok := s["bucket_columns"]; ok {
		storageDescriptor.BucketColumns = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := s["sort_columns"]; ok {
		storageDescriptor.SortColumns = expandSortColumns(v.([]interface{}))
	}

	if v, ok := s["skewed_info"]; ok {
		storageDescriptor.SkewedInfo = expandSkewedInfo(v.([]interface{}))
	}

	if v, ok := s[names.AttrParameters]; ok {
		storageDescriptor.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := s["stored_as_sub_directories"]; ok {
		storageDescriptor.StoredAsSubDirectories = v.(bool)
	}

	if v, ok := s["schema_reference"]; ok && len(v.([]interface{})) > 0 {
		storageDescriptor.Columns = nil
		storageDescriptor.SchemaReference = expandTableSchemaReference(v.([]interface{}))
	}

	return storageDescriptor
}

func expandColumns(columns []interface{}) []awstypes.Column {
	columnSlice := []awstypes.Column{}
	for _, element := range columns {
		elementMap := element.(map[string]interface{})

		column := awstypes.Column{
			Name: aws.String(elementMap[names.AttrName].(string)),
		}

		if v, ok := elementMap[names.AttrComment]; ok {
			column.Comment = aws.String(v.(string))
		}

		if v, ok := elementMap[names.AttrType]; ok {
			column.Type = aws.String(v.(string))
		}

		if v, ok := elementMap[names.AttrParameters]; ok {
			column.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
		}

		columnSlice = append(columnSlice, column)
	}

	return columnSlice
}

func expandSerDeInfo(l []interface{}) *awstypes.SerDeInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	serDeInfo := &awstypes.SerDeInfo{}

	if v := s[names.AttrName]; len(v.(string)) > 0 {
		serDeInfo.Name = aws.String(v.(string))
	}

	if v := s[names.AttrParameters]; len(v.(map[string]interface{})) > 0 {
		serDeInfo.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v := s["serialization_library"]; len(v.(string)) > 0 {
		serDeInfo.SerializationLibrary = aws.String(v.(string))
	}

	return serDeInfo
}

func expandSortColumns(columns []interface{}) []awstypes.Order {
	orderSlice := make([]awstypes.Order, len(columns))

	for i, element := range columns {
		elementMap := element.(map[string]interface{})

		order := awstypes.Order{
			Column: aws.String(elementMap["column"].(string)),
		}

		if v, ok := elementMap["sort_order"]; ok {
			order.SortOrder = int32(v.(int))
		}

		orderSlice[i] = order
	}

	return orderSlice
}

func expandSkewedInfo(l []interface{}) *awstypes.SkewedInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	skewedInfo := &awstypes.SkewedInfo{}

	if v, ok := s["skewed_column_names"]; ok {
		skewedInfo.SkewedColumnNames = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := s["skewed_column_value_location_maps"]; ok {
		skewedInfo.SkewedColumnValueLocationMaps = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := s["skewed_column_values"]; ok {
		skewedInfo.SkewedColumnValues = flex.ExpandStringValueList(v.([]interface{}))
	}

	return skewedInfo
}

func expandTableSchemaReference(l []interface{}) *awstypes.SchemaReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	schemaRef := &awstypes.SchemaReference{}

	if v, ok := s["schema_version_id"].(string); ok && v != "" {
		schemaRef.SchemaVersionId = aws.String(v)
	}

	if v, ok := s["schema_id"]; ok {
		schemaRef.SchemaId = expandTableSchemaReferenceSchemaID(v.([]interface{}))
	}

	if v, ok := s["schema_version_number"].(int); ok {
		schemaRef.SchemaVersionNumber = aws.Int64(int64(v))
	}

	return schemaRef
}

func expandTableSchemaReferenceSchemaID(l []interface{}) *awstypes.SchemaId {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	schemaID := &awstypes.SchemaId{}

	if v, ok := s["registry_name"].(string); ok && v != "" {
		schemaID.RegistryName = aws.String(v)
	}

	if v, ok := s["schema_name"].(string); ok && v != "" {
		schemaID.SchemaName = aws.String(v)
	}

	if v, ok := s["schema_arn"].(string); ok && v != "" {
		schemaID.SchemaArn = aws.String(v)
	}

	return schemaID
}

func flattenStorageDescriptor(s *awstypes.StorageDescriptor) []map[string]interface{} {
	if s == nil {
		storageDescriptors := make([]map[string]interface{}, 0)
		return storageDescriptors
	}

	storageDescriptors := make([]map[string]interface{}, 1)

	storageDescriptor := make(map[string]interface{})

	storageDescriptor["additional_locations"] = flex.FlattenStringValueList(s.AdditionalLocations)
	storageDescriptor["columns"] = flattenColumns(s.Columns)
	storageDescriptor[names.AttrLocation] = aws.ToString(s.Location)
	storageDescriptor["input_format"] = aws.ToString(s.InputFormat)
	storageDescriptor["output_format"] = aws.ToString(s.OutputFormat)
	storageDescriptor["compressed"] = s.Compressed
	storageDescriptor["number_of_buckets"] = s.NumberOfBuckets
	storageDescriptor["ser_de_info"] = flattenSerDeInfo(s.SerdeInfo)
	storageDescriptor["bucket_columns"] = flex.FlattenStringValueList(s.BucketColumns)
	storageDescriptor["sort_columns"] = flattenOrders(s.SortColumns)
	storageDescriptor[names.AttrParameters] = s.Parameters
	storageDescriptor["skewed_info"] = flattenSkewedInfo(s.SkewedInfo)
	storageDescriptor["stored_as_sub_directories"] = s.StoredAsSubDirectories

	if s.SchemaReference != nil {
		storageDescriptor["schema_reference"] = flattenTableSchemaReference(s.SchemaReference)
	}

	storageDescriptors[0] = storageDescriptor

	return storageDescriptors
}

func flattenColumns(cs []awstypes.Column) []map[string]interface{} {
	columnsSlice := make([]map[string]interface{}, len(cs))
	if len(cs) > 0 {
		for i, v := range cs {
			columnsSlice[i] = flattenColumn(v)
		}
	}

	return columnsSlice
}

func flattenColumn(c awstypes.Column) map[string]interface{} {
	column := make(map[string]interface{})

	if v := aws.ToString(c.Name); v != "" {
		column[names.AttrName] = v
	}

	if v := aws.ToString(c.Type); v != "" {
		column[names.AttrType] = v
	}

	if v := aws.ToString(c.Comment); v != "" {
		column[names.AttrComment] = v
	}

	if v := c.Parameters; v != nil {
		column[names.AttrParameters] = v
	}

	return column
}

func flattenPartitionIndexes(cs []awstypes.PartitionIndexDescriptor) []map[string]interface{} {
	partitionIndexSlice := make([]map[string]interface{}, len(cs))
	if len(cs) > 0 {
		for i, v := range cs {
			partitionIndexSlice[i] = flattenPartitionIndex(v)
		}
	}

	return partitionIndexSlice
}

func flattenPartitionIndex(c awstypes.PartitionIndexDescriptor) map[string]interface{} {
	partitionIndex := make(map[string]interface{})

	if v := aws.ToString(c.IndexName); v != "" {
		partitionIndex["index_name"] = v
	}

	if v := string(c.IndexStatus); v != "" {
		partitionIndex["index_status"] = v
	}

	if c.Keys != nil {
		names := make([]*string, 0, len(c.Keys))
		for _, key := range c.Keys {
			names = append(names, key.Name)
		}
		partitionIndex["keys"] = flex.FlattenStringList(names)
	}

	return partitionIndex
}

func flattenSerDeInfo(s *awstypes.SerDeInfo) []map[string]interface{} {
	if s == nil {
		serDeInfos := make([]map[string]interface{}, 0)
		return serDeInfos
	}

	serDeInfos := make([]map[string]interface{}, 1)
	serDeInfo := make(map[string]interface{})

	if v := aws.ToString(s.Name); v != "" {
		serDeInfo[names.AttrName] = v
	}
	serDeInfo[names.AttrParameters] = s.Parameters
	if v := aws.ToString(s.SerializationLibrary); v != "" {
		serDeInfo["serialization_library"] = v
	}

	serDeInfos[0] = serDeInfo
	return serDeInfos
}

func flattenOrders(os []awstypes.Order) []map[string]interface{} {
	orders := make([]map[string]interface{}, len(os))
	for i, v := range os {
		order := make(map[string]interface{})
		order["column"] = aws.ToString(v.Column)
		order["sort_order"] = int(v.SortOrder)
		orders[i] = order
	}

	return orders
}

func flattenSkewedInfo(s *awstypes.SkewedInfo) []map[string]interface{} {
	if s == nil {
		skewedInfoSlice := make([]map[string]interface{}, 0)
		return skewedInfoSlice
	}

	skewedInfoSlice := make([]map[string]interface{}, 1)

	skewedInfo := make(map[string]interface{})
	skewedInfo["skewed_column_names"] = flex.FlattenStringValueList(s.SkewedColumnNames)
	skewedInfo["skewed_column_value_location_maps"] = s.SkewedColumnValueLocationMaps
	skewedInfo["skewed_column_values"] = flex.FlattenStringValueList(s.SkewedColumnValues)
	skewedInfoSlice[0] = skewedInfo

	return skewedInfoSlice
}

func flattenTableSchemaReference(s *awstypes.SchemaReference) []map[string]interface{} {
	if s == nil {
		schemaReferenceInfoSlice := make([]map[string]interface{}, 0)
		return schemaReferenceInfoSlice
	}

	schemaReferenceInfoSlice := make([]map[string]interface{}, 1)

	schemaReferenceInfo := make(map[string]interface{})

	if s.SchemaVersionId != nil {
		schemaReferenceInfo["schema_version_id"] = aws.ToString(s.SchemaVersionId)
	}

	if s.SchemaVersionNumber != nil {
		schemaReferenceInfo["schema_version_number"] = aws.ToInt64(s.SchemaVersionNumber)
	}

	if s.SchemaId != nil {
		schemaReferenceInfo["schema_id"] = flattenTableSchemaReferenceSchemaID(s.SchemaId)
	}

	schemaReferenceInfoSlice[0] = schemaReferenceInfo

	return schemaReferenceInfoSlice
}

func flattenTableSchemaReferenceSchemaID(s *awstypes.SchemaId) []map[string]interface{} {
	if s == nil {
		schemaIDInfoSlice := make([]map[string]interface{}, 0)
		return schemaIDInfoSlice
	}

	schemaIDInfoSlice := make([]map[string]interface{}, 1)

	schemaIDInfo := make(map[string]interface{})

	if s.RegistryName != nil {
		schemaIDInfo["registry_name"] = aws.ToString(s.RegistryName)
	}

	if s.SchemaArn != nil {
		schemaIDInfo["schema_arn"] = aws.ToString(s.SchemaArn)
	}

	if s.SchemaName != nil {
		schemaIDInfo["schema_name"] = aws.ToString(s.SchemaName)
	}

	schemaIDInfoSlice[0] = schemaIDInfo

	return schemaIDInfoSlice
}

func expandTableTargetTable(tfMap map[string]interface{}) *awstypes.TableIdentifier {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TableIdentifier{}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDatabaseName].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func flattenTableTargetTable(apiObject *awstypes.TableIdentifier) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap[names.AttrDatabaseName] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	return tfMap
}

func flattenNonManagedParameters(table *awstypes.Table) map[string]string {
	allParameters := table.Parameters
	if allParameters["table_type"] == "ICEBERG" {
		delete(allParameters, "table_type")
		delete(allParameters, "metadata_location")
	}
	return allParameters
}
