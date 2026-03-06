// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package glue

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_catalog_table", name="Catalog Table")
func resourceCatalogTable() *schema.Resource {
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
			names.AttrOwner: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
						"index_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"keys": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
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
												"registry_name": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn"},
												},
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
									"skewed_column_value_location_maps": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"skewed_column_values": {
										Type:     schema.TypeList,
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
			"view_expanded_text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 409600),
			},
			"view_original_text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 409600),
			},
		},
	}
}

func resourceCatalogTableCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GlueClient(ctx)

	catalogID, dbName, name := cmp.Or(d.Get(names.AttrCatalogID).(string), c.AccountID(ctx)), d.Get(names.AttrDatabaseName).(string), d.Get(names.AttrName).(string)
	id := catalogTableCreateResourceID(catalogID, dbName, name)
	input := &glue.CreateTableInput{
		CatalogId:            aws.String(catalogID),
		DatabaseName:         aws.String(dbName),
		OpenTableFormatInput: expandOpenTableFormat(d),
		TableInput:           expandTableInput(d),
	}
	if v, ok := d.GetOk("partition_index"); ok && len(v.([]any)) > 0 {
		input.PartitionIndexes = expandPartitionIndexes(v.([]any))
	}

	_, err := conn.CreateTable(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Catalog Table (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceCatalogTableRead(ctx, d, meta)...)
}

func resourceCatalogTableRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GlueClient(ctx)

	catalogID, dbName, name, err := catalogTableParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	table, err := findTableByThreePartKey(ctx, conn, catalogID, dbName, name)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Glue Catalog Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, tableARN(ctx, c, dbName, name))
	d.Set(names.AttrCatalogID, catalogID)
	d.Set(names.AttrDatabaseName, dbName)
	d.Set(names.AttrDescription, table.Description)
	d.Set(names.AttrName, table.Name)
	d.Set(names.AttrOwner, table.Owner)
	if err := d.Set(names.AttrParameters, flattenNonManagedParameters(table)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}
	if err := d.Set("partition_keys", flattenColumns(table.PartitionKeys)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting partition_keys: %s", err)
	}
	d.Set("retention", table.Retention)
	if err := d.Set("storage_descriptor", flattenStorageDescriptor(table.StorageDescriptor)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_descriptor: %s", err)
	}
	d.Set("table_type", table.TableType)
	if table.TargetTable != nil {
		if err := d.Set("target_table", []any{flattenTableTargetTable(table.TargetTable)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_table: %s", err)
		}
	} else {
		d.Set("target_table", nil)
	}
	d.Set("view_original_text", table.ViewOriginalText)
	d.Set("view_expanded_text", table.ViewExpandedText)

	input := glue.GetPartitionIndexesInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(name),
	}
	partitionIndexes, err := findPartitionIndexes(ctx, conn, &input, tfslices.PredicateTrue[awstypes.PartitionIndexDescriptor]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table (%s) partition indexes: %s", d.Id(), err)
	}

	if err := d.Set("partition_index", flattenPartitionIndexDescriptors(partitionIndexes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting partition_index: %s", err)
	}

	return diags
}

func resourceCatalogTableUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, name, err := catalogTableParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &glue.UpdateTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableInput:   expandTableInput(d),
	}

	// Add back any managed parameters. See flattenNonManagedParameters.
	table, err := findTableByThreePartKey(ctx, conn, catalogID, dbName, name)

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

func resourceCatalogTableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, name, err := catalogTableParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Glue Catalog Table: %s", d.Id())
	input := glue.DeleteTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		Name:         aws.String(name),
	}
	_, err = conn.DeleteTable(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Catalog Table (%s): %s", d.Id(), err)
	}

	return diags
}

const catalogTableResourceIDSeparator = ":"

func catalogTableCreateResourceID(catalogID, dbName, name string) string {
	parts := []string{catalogID, dbName, name}
	id := strings.Join(parts, catalogTableResourceIDSeparator)

	return id
}

func catalogTableParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, catalogTableResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected catalog-id%[2]sdatabase-name%[2]stable-name", id, catalogTableResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findTableByThreePartKey(ctx context.Context, conn *glue.Client, catalogID, dbName, name string) (*awstypes.Table, error) {
	input := glue.GetTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		Name:         aws.String(name),
	}

	return findTable(ctx, conn, &input)
}

func findTable(ctx context.Context, conn *glue.Client, input *glue.GetTableInput) (*awstypes.Table, error) {
	output, err := conn.GetTable(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Table == nil {
		return nil, tfresource.NewEmptyResultError()
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
		tableInput.StorageDescriptor = expandStorageDescriptor(v.([]any))
	}

	if v, ok := d.GetOk("partition_keys"); ok {
		tableInput.PartitionKeys = expandColumns(v.([]any))
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
		tableInput.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := d.GetOk("target_table"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		tableInput.TargetTable = expandTableTargetTable(v.([]any)[0].(map[string]any))
	}

	return tableInput
}

func expandOpenTableFormat(s *schema.ResourceData) *awstypes.OpenTableFormatInput {
	if v, ok := s.GetOk("open_table_format_input"); ok {
		openTableFormatInput := &awstypes.OpenTableFormatInput{
			IcebergInput: expandIcebergInput(v.([]any)[0].(map[string]any)),
		}
		return openTableFormatInput
	}
	return nil
}

func expandIcebergInput(s map[string]any) *awstypes.IcebergInput {
	var iceberg = s["iceberg_input"].([]any)[0].(map[string]any)
	icebergInput := &awstypes.IcebergInput{
		MetadataOperation: awstypes.MetadataOperation(iceberg["metadata_operation"].(string)),
	}
	if v, ok := iceberg[names.AttrVersion].(string); ok && v != "" {
		icebergInput.Version = aws.String(v)
	}
	return icebergInput
}

func expandPartitionIndexes(tfList []any) []awstypes.PartitionIndex {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PartitionIndex

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandPartitionIndex(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandStorageDescriptor(l []any) *awstypes.StorageDescriptor {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]any)
	storageDescriptor := &awstypes.StorageDescriptor{}

	if v, ok := s["additional_locations"]; ok {
		storageDescriptor.AdditionalLocations = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := s["columns"]; ok {
		storageDescriptor.Columns = expandColumns(v.([]any))
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
		storageDescriptor.SerdeInfo = expandSerDeInfo(v.([]any))
	}

	if v, ok := s["bucket_columns"]; ok {
		storageDescriptor.BucketColumns = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := s["sort_columns"]; ok {
		storageDescriptor.SortColumns = expandSortColumns(v.([]any))
	}

	if v, ok := s["skewed_info"]; ok {
		storageDescriptor.SkewedInfo = expandSkewedInfo(v.([]any))
	}

	if v, ok := s[names.AttrParameters]; ok {
		storageDescriptor.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := s["stored_as_sub_directories"]; ok {
		storageDescriptor.StoredAsSubDirectories = v.(bool)
	}

	if v, ok := s["schema_reference"]; ok && len(v.([]any)) > 0 {
		storageDescriptor.Columns = nil
		storageDescriptor.SchemaReference = expandTableSchemaReference(v.([]any))
	}

	return storageDescriptor
}

func expandColumns(columns []any) []awstypes.Column {
	columnSlice := []awstypes.Column{}
	for _, element := range columns {
		elementMap := element.(map[string]any)

		column := awstypes.Column{
			Name: aws.String(elementMap[names.AttrName].(string)),
		}

		if v, ok := elementMap[names.AttrComment]; ok {
			column.Comment = aws.String(v.(string))
		}

		if v, ok := elementMap[names.AttrParameters]; ok {
			column.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
		}

		if v, ok := elementMap[names.AttrType]; ok {
			column.Type = aws.String(v.(string))
		}

		columnSlice = append(columnSlice, column)
	}

	return columnSlice
}

func expandSerDeInfo(l []any) *awstypes.SerDeInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]any)
	serDeInfo := &awstypes.SerDeInfo{}

	if v := s[names.AttrName]; len(v.(string)) > 0 {
		serDeInfo.Name = aws.String(v.(string))
	}

	if v := s[names.AttrParameters]; len(v.(map[string]any)) > 0 {
		serDeInfo.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v := s["serialization_library"]; len(v.(string)) > 0 {
		serDeInfo.SerializationLibrary = aws.String(v.(string))
	}

	return serDeInfo
}

func expandSortColumns(columns []any) []awstypes.Order {
	orderSlice := make([]awstypes.Order, len(columns))

	for i, element := range columns {
		elementMap := element.(map[string]any)

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

func expandSkewedInfo(l []any) *awstypes.SkewedInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]any)
	skewedInfo := &awstypes.SkewedInfo{}

	if v, ok := s["skewed_column_names"]; ok {
		skewedInfo.SkewedColumnNames = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := s["skewed_column_value_location_maps"]; ok {
		skewedInfo.SkewedColumnValueLocationMaps = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := s["skewed_column_values"]; ok {
		skewedInfo.SkewedColumnValues = flex.ExpandStringValueList(v.([]any))
	}

	return skewedInfo
}

func expandTableSchemaReference(l []any) *awstypes.SchemaReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]any)
	schemaRef := &awstypes.SchemaReference{}

	if v, ok := s["schema_version_id"].(string); ok && v != "" {
		schemaRef.SchemaVersionId = aws.String(v)
	}

	if v, ok := s["schema_id"]; ok {
		schemaRef.SchemaId = expandTableSchemaReferenceSchemaID(v.([]any))
	}

	if v, ok := s["schema_version_number"].(int); ok {
		schemaRef.SchemaVersionNumber = aws.Int64(int64(v))
	}

	return schemaRef
}

func expandTableSchemaReferenceSchemaID(l []any) *awstypes.SchemaId {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]any)
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

func flattenStorageDescriptor(s *awstypes.StorageDescriptor) []map[string]any {
	if s == nil {
		storageDescriptors := make([]map[string]any, 0)
		return storageDescriptors
	}

	storageDescriptors := make([]map[string]any, 1)

	storageDescriptor := make(map[string]any)

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

func flattenColumns(cs []awstypes.Column) []map[string]any {
	columnsSlice := make([]map[string]any, len(cs))
	if len(cs) > 0 {
		for i, v := range cs {
			columnsSlice[i] = flattenColumn(v)
		}
	}

	return columnsSlice
}

func flattenColumn(c awstypes.Column) map[string]any {
	column := make(map[string]any)

	if v := aws.ToString(c.Comment); v != "" {
		column[names.AttrComment] = v
	}

	if v := aws.ToString(c.Name); v != "" {
		column[names.AttrName] = v
	}

	if v := c.Parameters; v != nil {
		column[names.AttrParameters] = v
	}

	if v := aws.ToString(c.Type); v != "" {
		column[names.AttrType] = v
	}

	return column
}

func flattenPartitionIndexDescriptors(apiObjects []awstypes.PartitionIndexDescriptor) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPartitionIndexDescriptor(&apiObject))
	}

	return tfList
}

func flattenSerDeInfo(s *awstypes.SerDeInfo) []map[string]any {
	if s == nil {
		serDeInfos := make([]map[string]any, 0)
		return serDeInfos
	}

	serDeInfos := make([]map[string]any, 1)
	serDeInfo := make(map[string]any)

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

func flattenOrders(os []awstypes.Order) []map[string]any {
	orders := make([]map[string]any, len(os))
	for i, v := range os {
		order := make(map[string]any)
		order["column"] = aws.ToString(v.Column)
		order["sort_order"] = int(v.SortOrder)
		orders[i] = order
	}

	return orders
}

func flattenSkewedInfo(s *awstypes.SkewedInfo) []map[string]any {
	if s == nil {
		skewedInfoSlice := make([]map[string]any, 0)
		return skewedInfoSlice
	}

	skewedInfoSlice := make([]map[string]any, 1)

	skewedInfo := make(map[string]any)
	skewedInfo["skewed_column_names"] = flex.FlattenStringValueList(s.SkewedColumnNames)
	skewedInfo["skewed_column_value_location_maps"] = s.SkewedColumnValueLocationMaps
	skewedInfo["skewed_column_values"] = flex.FlattenStringValueList(s.SkewedColumnValues)
	skewedInfoSlice[0] = skewedInfo

	return skewedInfoSlice
}

func flattenTableSchemaReference(s *awstypes.SchemaReference) []map[string]any {
	if s == nil {
		schemaReferenceInfoSlice := make([]map[string]any, 0)
		return schemaReferenceInfoSlice
	}

	schemaReferenceInfoSlice := make([]map[string]any, 1)

	schemaReferenceInfo := make(map[string]any)

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

func flattenTableSchemaReferenceSchemaID(s *awstypes.SchemaId) []map[string]any {
	if s == nil {
		schemaIDInfoSlice := make([]map[string]any, 0)
		return schemaIDInfoSlice
	}

	schemaIDInfoSlice := make([]map[string]any, 1)

	schemaIDInfo := make(map[string]any)

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

func expandTableTargetTable(tfMap map[string]any) *awstypes.TableIdentifier {
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

func flattenTableTargetTable(apiObject *awstypes.TableIdentifier) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

func tableARN(ctx context.Context, c *conns.AWSClient, dbName, name string) string {
	return c.RegionalARN(ctx, "glue", "table/"+dbName+"/"+name)
}
