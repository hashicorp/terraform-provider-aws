package glue

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCatalogTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceCatalogTableCreate,
		Read:   resourceCatalogTableRead,
		Update: resourceCatalogTableUpdate,
		Delete: resourceCatalogTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringDoesNotMatch(regexp.MustCompile(`[A-Z]`), "uppercase characters cannot be used"),
				),
			},
			"owner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"partition_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comment": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"type": {
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
									"comment": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 255),
									},
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"parameters": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"type": {
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"parameters": {
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
			"target_table": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
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

func ReadTableID(id string) (catalogID string, dbName string, name string, error error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format catalog-id:database-name:table-name, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func resourceCatalogTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	dbName := d.Get("database_name").(string)
	name := d.Get("name").(string)

	input := &glue.CreateTableInput{
		CatalogId:        aws.String(catalogID),
		DatabaseName:     aws.String(dbName),
		TableInput:       expandTableInput(d),
		PartitionIndexes: expandTablePartitionIndexes(d.Get("partition_index").([]interface{})),
	}

	log.Printf("[DEBUG] Glue catalog table input: %#v", input)
	_, err := conn.CreateTable(input)
	if err != nil {
		return fmt.Errorf("Error creating Glue Catalog Table: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", catalogID, dbName, name))

	return resourceCatalogTableRead(d, meta)
}

func resourceCatalogTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	catalogID, dbName, name, err := ReadTableID(d.Id())
	if err != nil {
		return err
	}

	out, err := FindTableByName(conn, catalogID, dbName, name)
	if err != nil {

		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Glue Catalog Table (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading Glue Catalog Table: %w", err)
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
		return fmt.Errorf("error setting storage_descriptor: %w", err)
	}

	if err := d.Set("partition_keys", flattenColumns(table.PartitionKeys)); err != nil {
		return fmt.Errorf("error setting partition_keys: %w", err)
	}

	d.Set("view_original_text", table.ViewOriginalText)
	d.Set("view_expanded_text", table.ViewExpandedText)
	d.Set("table_type", table.TableType)

	if err := d.Set("parameters", aws.StringValueMap(table.Parameters)); err != nil {
		return fmt.Errorf("error setting parameters: %w", err)
	}

	if table.TargetTable != nil {
		if err := d.Set("target_table", []interface{}{flattenTableTargetTable(table.TargetTable)}); err != nil {
			return fmt.Errorf("error setting target_table: %w", err)
		}
	} else {
		d.Set("target_table", nil)
	}

	partIndexInput := &glue.GetPartitionIndexesInput{
		CatalogId:    out.Table.CatalogId,
		TableName:    out.Table.Name,
		DatabaseName: out.Table.DatabaseName,
	}
	partOut, err := conn.GetPartitionIndexes(partIndexInput)
	if err != nil {
		return fmt.Errorf("error getting Glue Partition Indexes: %w", err)
	}

	if partOut != nil && len(partOut.PartitionIndexDescriptorList) > 0 {
		if err := d.Set("partition_index", flattenPartitionIndexes(partOut.PartitionIndexDescriptorList)); err != nil {
			return fmt.Errorf("error setting partition_index: %w", err)
		}
	}

	return nil
}

func resourceCatalogTableUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	catalogID, dbName, _, err := ReadTableID(d.Id())
	if err != nil {
		return err
	}

	updateTableInput := &glue.UpdateTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableInput:   expandTableInput(d),
	}

	if _, err := conn.UpdateTable(updateTableInput); err != nil {
		return fmt.Errorf("Error updating Glue Catalog Table: %w", err)
	}

	return resourceCatalogTableRead(d, meta)
}

func resourceCatalogTableDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	catalogID, dbName, name, tableIdErr := ReadTableID(d.Id())
	if tableIdErr != nil {
		return tableIdErr
	}

	log.Printf("[DEBUG] Glue Catalog Table: %s:%s:%s", catalogID, dbName, name)
	_, err := conn.DeleteTable(&glue.DeleteTableInput{
		CatalogId:    aws.String(catalogID),
		Name:         aws.String(name),
		DatabaseName: aws.String(dbName),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting Glue Catalog Table: %w", err)
	}
	return nil
}

func expandTableInput(d *schema.ResourceData) *glue.TableInput {
	tableInput := &glue.TableInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		tableInput.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("owner"); ok {
		tableInput.Owner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retention"); ok {
		tableInput.Retention = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("storage_descriptor"); ok {
		tableInput.StorageDescriptor = expandStorageDescriptor(v.([]interface{}))
	}

	if v, ok := d.GetOk("partition_keys"); ok {
		tableInput.PartitionKeys = expandColumns(v.([]interface{}))
	} else {
		tableInput.PartitionKeys = []*glue.Column{}
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

	if v, ok := d.GetOk("parameters"); ok {
		tableInput.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("target_table"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tableInput.TargetTable = expandTableTargetTable(v.([]interface{})[0].(map[string]interface{}))
	}

	return tableInput
}

func expandTablePartitionIndexes(a []interface{}) []*glue.PartitionIndex {
	partitionIndexes := make([]*glue.PartitionIndex, 0, len(a))

	for _, m := range a {
		partitionIndexes = append(partitionIndexes, expandTablePartitionIndex(m.(map[string]interface{})))
	}

	return partitionIndexes
}

func expandTablePartitionIndex(m map[string]interface{}) *glue.PartitionIndex {
	partitionIndex := &glue.PartitionIndex{
		IndexName: aws.String(m["index_name"].(string)),
		Keys:      flex.ExpandStringList(m["keys"].([]interface{})),
	}

	return partitionIndex
}

func expandStorageDescriptor(l []interface{}) *glue.StorageDescriptor {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	storageDescriptor := &glue.StorageDescriptor{}

	if v, ok := s["columns"]; ok {
		storageDescriptor.Columns = expandColumns(v.([]interface{}))
	}

	if v, ok := s["location"]; ok {
		storageDescriptor.Location = aws.String(v.(string))
	}

	if v, ok := s["input_format"]; ok {
		storageDescriptor.InputFormat = aws.String(v.(string))
	}

	if v, ok := s["output_format"]; ok {
		storageDescriptor.OutputFormat = aws.String(v.(string))
	}

	if v, ok := s["compressed"]; ok {
		storageDescriptor.Compressed = aws.Bool(v.(bool))
	}

	if v, ok := s["number_of_buckets"]; ok {
		storageDescriptor.NumberOfBuckets = aws.Int64(int64(v.(int)))
	}

	if v, ok := s["ser_de_info"]; ok {
		storageDescriptor.SerdeInfo = expandSerDeInfo(v.([]interface{}))
	}

	if v, ok := s["bucket_columns"]; ok {
		storageDescriptor.BucketColumns = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := s["sort_columns"]; ok {
		storageDescriptor.SortColumns = expandSortColumns(v.([]interface{}))
	}

	if v, ok := s["skewed_info"]; ok {
		storageDescriptor.SkewedInfo = expandSkewedInfo(v.([]interface{}))
	}

	if v, ok := s["parameters"]; ok {
		storageDescriptor.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := s["stored_as_sub_directories"]; ok {
		storageDescriptor.StoredAsSubDirectories = aws.Bool(v.(bool))
	}

	if v, ok := s["schema_reference"]; ok && len(v.([]interface{})) > 0 {
		storageDescriptor.Columns = nil
		storageDescriptor.SchemaReference = expandTableSchemaReference(v.([]interface{}))
	}

	return storageDescriptor
}

func expandColumns(columns []interface{}) []*glue.Column {
	columnSlice := []*glue.Column{}
	for _, element := range columns {
		elementMap := element.(map[string]interface{})

		column := &glue.Column{
			Name: aws.String(elementMap["name"].(string)),
		}

		if v, ok := elementMap["comment"]; ok {
			column.Comment = aws.String(v.(string))
		}

		if v, ok := elementMap["type"]; ok {
			column.Type = aws.String(v.(string))
		}

		if v, ok := elementMap["parameters"]; ok {
			column.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
		}

		columnSlice = append(columnSlice, column)
	}

	return columnSlice
}

func expandSerDeInfo(l []interface{}) *glue.SerDeInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	serDeInfo := &glue.SerDeInfo{}

	if v := s["name"]; len(v.(string)) > 0 {
		serDeInfo.Name = aws.String(v.(string))
	}

	if v := s["parameters"]; len(v.(map[string]interface{})) > 0 {
		serDeInfo.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v := s["serialization_library"]; len(v.(string)) > 0 {
		serDeInfo.SerializationLibrary = aws.String(v.(string))
	}

	return serDeInfo
}

func expandSortColumns(columns []interface{}) []*glue.Order {
	orderSlice := make([]*glue.Order, len(columns))

	for i, element := range columns {
		elementMap := element.(map[string]interface{})

		order := &glue.Order{
			Column: aws.String(elementMap["column"].(string)),
		}

		if v, ok := elementMap["sort_order"]; ok {
			order.SortOrder = aws.Int64(int64(v.(int)))
		}

		orderSlice[i] = order
	}

	return orderSlice
}

func expandSkewedInfo(l []interface{}) *glue.SkewedInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	skewedInfo := &glue.SkewedInfo{}

	if v, ok := s["skewed_column_names"]; ok {
		skewedInfo.SkewedColumnNames = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := s["skewed_column_value_location_maps"]; ok {
		skewedInfo.SkewedColumnValueLocationMaps = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := s["skewed_column_values"]; ok {
		skewedInfo.SkewedColumnValues = flex.ExpandStringList(v.([]interface{}))
	}

	return skewedInfo
}

func expandTableSchemaReference(l []interface{}) *glue.SchemaReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	schemaRef := &glue.SchemaReference{}

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

func expandTableSchemaReferenceSchemaID(l []interface{}) *glue.SchemaId {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	schemaID := &glue.SchemaId{}

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

func flattenStorageDescriptor(s *glue.StorageDescriptor) []map[string]interface{} {
	if s == nil {
		storageDescriptors := make([]map[string]interface{}, 0)
		return storageDescriptors
	}

	storageDescriptors := make([]map[string]interface{}, 1)

	storageDescriptor := make(map[string]interface{})

	storageDescriptor["columns"] = flattenColumns(s.Columns)
	storageDescriptor["location"] = aws.StringValue(s.Location)
	storageDescriptor["input_format"] = aws.StringValue(s.InputFormat)
	storageDescriptor["output_format"] = aws.StringValue(s.OutputFormat)
	storageDescriptor["compressed"] = aws.BoolValue(s.Compressed)
	storageDescriptor["number_of_buckets"] = aws.Int64Value(s.NumberOfBuckets)
	storageDescriptor["ser_de_info"] = flattenSerDeInfo(s.SerdeInfo)
	storageDescriptor["bucket_columns"] = flex.FlattenStringList(s.BucketColumns)
	storageDescriptor["sort_columns"] = flattenOrders(s.SortColumns)
	storageDescriptor["parameters"] = aws.StringValueMap(s.Parameters)
	storageDescriptor["skewed_info"] = flattenSkewedInfo(s.SkewedInfo)
	storageDescriptor["stored_as_sub_directories"] = aws.BoolValue(s.StoredAsSubDirectories)

	if s.SchemaReference != nil {
		storageDescriptor["schema_reference"] = flattenTableSchemaReference(s.SchemaReference)
	}

	storageDescriptors[0] = storageDescriptor

	return storageDescriptors
}

func flattenColumns(cs []*glue.Column) []map[string]interface{} {
	columnsSlice := make([]map[string]interface{}, len(cs))
	if len(cs) > 0 {
		for i, v := range cs {
			columnsSlice[i] = flattenColumn(v)
		}
	}

	return columnsSlice
}

func flattenColumn(c *glue.Column) map[string]interface{} {
	column := make(map[string]interface{})

	if c == nil {
		return column
	}

	if v := aws.StringValue(c.Name); v != "" {
		column["name"] = v
	}

	if v := aws.StringValue(c.Type); v != "" {
		column["type"] = v
	}

	if v := aws.StringValue(c.Comment); v != "" {
		column["comment"] = v
	}

	if v := c.Parameters; v != nil {
		column["parameters"] = aws.StringValueMap(v)
	}

	return column
}

func flattenPartitionIndexes(cs []*glue.PartitionIndexDescriptor) []map[string]interface{} {
	partitionIndexSlice := make([]map[string]interface{}, len(cs))
	if len(cs) > 0 {
		for i, v := range cs {
			partitionIndexSlice[i] = flattenPartitionIndex(v)
		}
	}

	return partitionIndexSlice
}

func flattenPartitionIndex(c *glue.PartitionIndexDescriptor) map[string]interface{} {
	partitionIndex := make(map[string]interface{})

	if c == nil {
		return partitionIndex
	}

	if v := aws.StringValue(c.IndexName); v != "" {
		partitionIndex["index_name"] = v
	}

	if v := aws.StringValue(c.IndexStatus); v != "" {
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

func flattenSerDeInfo(s *glue.SerDeInfo) []map[string]interface{} {
	if s == nil {
		serDeInfos := make([]map[string]interface{}, 0)
		return serDeInfos
	}

	serDeInfos := make([]map[string]interface{}, 1)
	serDeInfo := make(map[string]interface{})

	if v := aws.StringValue(s.Name); v != "" {
		serDeInfo["name"] = v
	}

	serDeInfo["parameters"] = aws.StringValueMap(s.Parameters)

	if v := aws.StringValue(s.SerializationLibrary); v != "" {
		serDeInfo["serialization_library"] = v
	}

	serDeInfos[0] = serDeInfo
	return serDeInfos
}

func flattenOrders(os []*glue.Order) []map[string]interface{} {
	orders := make([]map[string]interface{}, len(os))
	for i, v := range os {
		order := make(map[string]interface{})
		order["column"] = aws.StringValue(v.Column)
		order["sort_order"] = int(aws.Int64Value(v.SortOrder))
		orders[i] = order
	}

	return orders
}

func flattenSkewedInfo(s *glue.SkewedInfo) []map[string]interface{} {
	if s == nil {
		skewedInfoSlice := make([]map[string]interface{}, 0)
		return skewedInfoSlice
	}

	skewedInfoSlice := make([]map[string]interface{}, 1)

	skewedInfo := make(map[string]interface{})
	skewedInfo["skewed_column_names"] = flex.FlattenStringList(s.SkewedColumnNames)
	skewedInfo["skewed_column_value_location_maps"] = aws.StringValueMap(s.SkewedColumnValueLocationMaps)
	skewedInfo["skewed_column_values"] = flex.FlattenStringList(s.SkewedColumnValues)
	skewedInfoSlice[0] = skewedInfo

	return skewedInfoSlice
}

func flattenTableSchemaReference(s *glue.SchemaReference) []map[string]interface{} {
	if s == nil {
		schemaReferenceInfoSlice := make([]map[string]interface{}, 0)
		return schemaReferenceInfoSlice
	}

	schemaReferenceInfoSlice := make([]map[string]interface{}, 1)

	schemaReferenceInfo := make(map[string]interface{})

	if s.SchemaVersionId != nil {
		schemaReferenceInfo["schema_version_id"] = aws.StringValue(s.SchemaVersionId)
	}

	if s.SchemaVersionNumber != nil {
		schemaReferenceInfo["schema_version_number"] = aws.Int64Value(s.SchemaVersionNumber)
	}

	if s.SchemaId != nil {
		schemaReferenceInfo["schema_id"] = flattenTableSchemaReferenceSchemaID(s.SchemaId)
	}

	schemaReferenceInfoSlice[0] = schemaReferenceInfo

	return schemaReferenceInfoSlice
}

func flattenTableSchemaReferenceSchemaID(s *glue.SchemaId) []map[string]interface{} {
	if s == nil {
		schemaIDInfoSlice := make([]map[string]interface{}, 0)
		return schemaIDInfoSlice
	}

	schemaIDInfoSlice := make([]map[string]interface{}, 1)

	schemaIDInfo := make(map[string]interface{})

	if s.RegistryName != nil {
		schemaIDInfo["registry_name"] = aws.StringValue(s.RegistryName)
	}

	if s.SchemaArn != nil {
		schemaIDInfo["schema_arn"] = aws.StringValue(s.SchemaArn)
	}

	if s.SchemaName != nil {
		schemaIDInfo["schema_name"] = aws.StringValue(s.SchemaName)
	}

	schemaIDInfoSlice[0] = schemaIDInfo

	return schemaIDInfoSlice
}

func expandTableTargetTable(tfMap map[string]interface{}) *glue.TableIdentifier {
	if tfMap == nil {
		return nil
	}

	apiObject := &glue.TableIdentifier{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func flattenTableTargetTable(apiObject *glue.TableIdentifier) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap["database_name"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}
