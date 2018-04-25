package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsGlueCatalogTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueCatalogTableCreate,
		Read:   resourceAwsGlueCatalogTableRead,
		Update: resourceAwsGlueCatalogTableUpdate,
		Delete: resourceAwsGlueCatalogTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"retention": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"storage_descriptor": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"comment": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"input_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"output_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"compressed": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"number_of_buckets": {
							Type:     schema.TypeInt,
							Optional: true,
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
						"bucket_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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
						"stored_as_sub_directories": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"partition_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"comment": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"view_original_text": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"view_expanded_text": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"table_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func readAwsGlueTableID(id string) (catalogID string, dbName string, name string) {
	idParts := strings.Split(id, ":")
	return idParts[0], idParts[1], idParts[2]
}

func resourceAwsGlueCatalogTableCreate(t *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(t, meta.(*AWSClient).accountid)
	dbName := t.Get("database_name").(string)
	name := t.Get("name").(string)

	input := &glue.CreateTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableInput:   expandGlueTableInput(t),
	}

	_, err := glueconn.CreateTable(input)
	if err != nil {
		return fmt.Errorf("Error creating Catalog Table: %s", err)
	}

	t.SetId(fmt.Sprintf("%s:%s:%s", catalogID, dbName, name))

	return resourceAwsGlueCatalogTableRead(t, meta)
}

func resourceAwsGlueCatalogTableRead(t *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn

	catalogID, dbName, name := readAwsGlueTableID(t.Id())

	input := &glue.GetTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		Name:         aws.String(name),
	}

	out, err := glueconn.GetTable(input)
	if err != nil {

		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Catalog Table (%s) not found, removing from state", t.Id())
			t.SetId("")
		}

		return fmt.Errorf("Error reading Glue Catalog Table: %s", err)
	}

	t.Set("name", out.Table.Name)
	t.Set("catalog_id", catalogID)
	t.Set("database_name", dbName)
	t.Set("description", out.Table.Description)
	t.Set("owner", out.Table.Owner)
	t.Set("retention", out.Table.Retention)
	t.Set("storage_descriptor", flattenStorageDescriptor(out.Table.StorageDescriptor))
	t.Set("partition_keys", flattenGlueColumns(out.Table.PartitionKeys))
	t.Set("view_original_text", out.Table.ViewOriginalText)
	t.Set("view_expanded_text", out.Table.ViewExpandedText)
	t.Set("table_type", out.Table.TableType)
	t.Set("parameters", flattenStringParameters(out.Table.Parameters))

	return nil
}

func resourceAwsGlueCatalogTableUpdate(t *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn

	catalogID, dbName, _ := readAwsGlueTableID(t.Id())

	updateTableInput := &glue.UpdateTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableInput:   expandGlueTableInput(t),
	}

	if _, err := glueconn.UpdateTable(updateTableInput); err != nil {
		return fmt.Errorf("Error updating Glue Catalog Table: %s", err)
	}

	return resourceAwsGlueCatalogTableRead(t, meta)
}

func resourceAwsGlueCatalogTableDelete(t *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn
	catalogID, dbName, name := readAwsGlueTableID(t.Id())

	log.Printf("[DEBUG] Glue Catalog Table: %s:%s:%s", catalogID, dbName, name)
	_, err := glueconn.DeleteTable(&glue.DeleteTableInput{
		CatalogId:    aws.String(catalogID),
		Name:         aws.String(name),
		DatabaseName: aws.String(dbName),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Glue Catalog Table: %s", err.Error())
	}
	return nil
}

func expandGlueTableInput(t *schema.ResourceData) *glue.TableInput {
	tableInput := &glue.TableInput{
		Name: aws.String(t.Get("name").(string)),
	}

	if v, ok := t.GetOk("description"); ok {
		tableInput.Description = aws.String(v.(string))
	}

	if v, ok := t.GetOk("owner"); ok {
		tableInput.Owner = aws.String(v.(string))
	}

	if v, ok := t.GetOk("retention"); ok {
		tableInput.Retention = aws.Int64(int64(v.(int)))
	}

	if v, ok := t.GetOk("storage_descriptor"); ok {
		for _, elem := range v.([]interface{}) {
			tableInput.StorageDescriptor = expandGlueStorageDescriptor(elem.(map[string]interface{}))
		}
	}

	if v, ok := t.GetOk("partition_keys"); ok {
		columns := expandGlueColumns(v.([]interface{}))
		tableInput.PartitionKeys = columns
	}

	if v, ok := t.GetOk("view_original_text"); ok {
		tableInput.ViewOriginalText = aws.String(v.(string))
	}

	if v, ok := t.GetOk("view_expanded_text"); ok {
		tableInput.ViewExpandedText = aws.String(v.(string))
	}

	if v, ok := t.GetOk("table_type"); ok {
		tableInput.TableType = aws.String(v.(string))
	}

	if v, ok := t.GetOk("parameters"); ok {
		paramsMap := map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			paramsMap[key] = value.(string)
		}
		tableInput.Parameters = aws.StringMap(paramsMap)
	}

	return tableInput
}

func expandGlueStorageDescriptor(s map[string]interface{}) *glue.StorageDescriptor {
	storageDescriptor := &glue.StorageDescriptor{}

	if v, ok := s["columns"]; ok {
		columns := expandGlueColumns(v.([]interface{}))
		storageDescriptor.Columns = columns
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
		for _, elem := range v.([]interface{}) {
			storageDescriptor.SerdeInfo = expandSerDeInfo(elem.(map[string]interface{}))
		}
	}

	if v, ok := s["bucket_columns"]; ok {
		bucketColumns := make([]string, len(v.([]interface{})))
		for i, item := range v.([]interface{}) {
			bucketColumns[i] = fmt.Sprint(item)
		}
		storageDescriptor.BucketColumns = aws.StringSlice(bucketColumns)
	}

	if v, ok := s["sort_columns"]; ok {
		storageDescriptor.SortColumns = expandSortColumns(v.([]interface{}))
	}

	if v, ok := s["skewed_info"]; ok {
		for _, elem := range v.([]interface{}) {
			storageDescriptor.SkewedInfo = expandSkewedInfo(elem.(map[string]interface{}))
		}
	}

	if v, ok := s["parameters"]; ok {
		paramsMap := map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			paramsMap[key] = value.(string)
		}
		storageDescriptor.Parameters = aws.StringMap(paramsMap)
	}

	if v, ok := s["stored_as_sub_directories"]; ok {
		storageDescriptor.StoredAsSubDirectories = aws.Bool(v.(bool))
	}

	return storageDescriptor
}

func expandGlueColumns(columns []interface{}) []*glue.Column {
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

		columnSlice = append(columnSlice, column)
	}

	return columnSlice
}

func expandSerDeInfo(s map[string]interface{}) *glue.SerDeInfo {
	serDeInfo := &glue.SerDeInfo{}

	if v, ok := s["name"]; ok {
		serDeInfo.Name = aws.String(v.(string))
	}

	if v, ok := s["parameters"]; ok {
		paramsMap := map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			paramsMap[key] = value.(string)
		}
		serDeInfo.Parameters = aws.StringMap(paramsMap)
	}

	if v, ok := s["serialization_library"]; ok {
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

func expandSkewedInfo(s map[string]interface{}) *glue.SkewedInfo {
	skewedInfo := &glue.SkewedInfo{}

	if v, ok := s["skewed_column_names"]; ok {
		columnsSlice := make([]string, len(v.([]interface{})))
		for i, item := range v.([]interface{}) {
			columnsSlice[i] = fmt.Sprint(item)
		}
		skewedInfo.SkewedColumnNames = aws.StringSlice(columnsSlice)
	}

	if v, ok := s["skewed_column_value_location_maps"]; ok {
		typeMap := map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			typeMap[key] = value.(string)
		}
		skewedInfo.SkewedColumnValueLocationMaps = aws.StringMap(typeMap)
	}

	if v, ok := s["skewed_column_values"]; ok {
		columnsSlice := make([]string, len(v.([]interface{})))
		for i, item := range v.([]interface{}) {
			columnsSlice[i] = fmt.Sprint(item)
		}
		skewedInfo.SkewedColumnValues = aws.StringSlice(columnsSlice)
	}

	return skewedInfo
}

func flattenStorageDescriptor(s *glue.StorageDescriptor) []map[string]interface{} {
	if s == nil {
		storageDescriptors := make([]map[string]interface{}, 0)
		return storageDescriptors
	}

	storageDescriptors := make([]map[string]interface{}, 1)

	storageDescriptor := make(map[string]interface{})

	storageDescriptor["columns"] = flattenGlueColumns(s.Columns)
	storageDescriptor["location"] = *s.Location
	storageDescriptor["input_format"] = *s.InputFormat
	storageDescriptor["output_format"] = *s.OutputFormat
	storageDescriptor["compressed"] = *s.Compressed
	storageDescriptor["number_of_buckets"] = *s.NumberOfBuckets
	storageDescriptor["ser_de_info"] = flattenSerDeInfo(s.SerdeInfo)
	storageDescriptor["bucket_columns"] = flattenStringList(s.BucketColumns)
	storageDescriptor["sort_columns"] = flattenOrders(s.SortColumns)
	storageDescriptor["parameters"] = flattenStringParameters(s.Parameters)
	storageDescriptor["skewed_info"] = flattenSkewedInfo(s.SkewedInfo)
	storageDescriptor["stored_as_sub_directories"] = *s.StoredAsSubDirectories

	storageDescriptors[0] = storageDescriptor

	return storageDescriptors
}

func flattenGlueColumns(cs []*glue.Column) []map[string]string {
	columnsSlice := make([]map[string]string, len(cs))
	if len(cs) > 0 {
		for i, v := range cs {
			columnsSlice[i] = flattenGlueColumn(v)
		}
	}

	return columnsSlice
}

func flattenGlueColumn(c *glue.Column) map[string]string {
	column := make(map[string]string)

	if v := *c.Name; v != "" {
		column["name"] = v
	}

	if v := *c.Type; v != "" {
		column["type"] = v
	}

	if v := *c.Comment; v != "" {
		column["comment"] = v
	}

	return column
}

func flattenStringParameters(p map[string]*string) map[string]string {
	tParams := make(map[string]string)
	if len(p) > 0 {
		for key, value := range p {
			tParams[key] = *value
		}
	}

	return tParams
}

func flattenSerDeInfo(s *glue.SerDeInfo) []map[string]interface{} {
	if s == nil {
		serDeInfos := make([]map[string]interface{}, 0)
		return serDeInfos
	}

	serDeInfos := make([]map[string]interface{}, 1)
	serDeInfo := make(map[string]interface{})

	serDeInfo["name"] = *s.Name
	serDeInfo["parameters"] = flattenStringParameters(s.Parameters)
	serDeInfo["serialization_library"] = *s.SerializationLibrary

	serDeInfos[0] = serDeInfo
	return serDeInfos
}

func flattenOrders(os []*glue.Order) []map[string]interface{} {
	orders := make([]map[string]interface{}, len(os))
	for i, v := range os {
		order := make(map[string]interface{})
		order["column"] = *v.Column
		order["sort_order"] = *v.SortOrder
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
	skewedInfo["skewed_column_names"] = flattenStringList(s.SkewedColumnNames)
	skewedInfo["skewed_column_value_location_maps"] = flattenStringParameters(s.SkewedColumnValueLocationMaps)
	skewedInfo["skewed_column_values"] = flattenStringList(s.SkewedColumnValues)
	skewedInfoSlice[0] = skewedInfo

	return skewedInfoSlice
}
