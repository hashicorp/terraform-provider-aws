package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"strings"
)

func resourceAwsGluePartition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGluePartitionCreate,
		Read:   resourceAwsGluePartitionRead,
		Update: resourceAwsGluePartitionUpdate,
		Delete: resourceAwsGluePartitionDelete,
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
			"database_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"table_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
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
			"values": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func readAwsGluePartitionID(id string) (catalogID string, dbName string, tableName string, error error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format catalog-id:database-name:table-name, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func resourceAwsGluePartitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(d, meta.(*AWSClient).accountid)
	dbName := d.Get("database_name").(string)
	tableName := d.Get("table_name").(string)

	input := &glue.CreatePartitionInput{
		CatalogId:      aws.String(catalogID),
		DatabaseName:   aws.String(dbName),
		TableName:      aws.String(tableName),
		PartitionInput: expandGluePartitionInput(d),
	}

	_, err := conn.CreatePartition(input)
	if err != nil {
		return fmt.Errorf("Error creating Glue Partition: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s:", catalogID, dbName, tableName))

	return resourceAwsGluePartitionRead(d, meta)
}

func resourceAwsGluePartitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, tableName, err := readAwsGluePartitionID(d.Id())
	if err != nil {
		return err
	}

	input := &glue.GetPartitionInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
	}

	out, err := conn.GetPartition(input)
	if err != nil {

		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Partition (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading Glue Partition: %s", err)
	}

	partition := out.Partition

	d.Set("table_name", partition.TableName)
	d.Set("catalog_id", catalogID)
	d.Set("database_name", partition.DatabaseName)
	d.Set("values", flattenStringSet(partition.Values))

	if err := d.Set("storage_descriptor", flattenGlueStorageDescriptor(partition.StorageDescriptor)); err != nil {
		return fmt.Errorf("error setting storage_descriptor: %s", err)
	}

	if err := d.Set("parameters", aws.StringValueMap(partition.Parameters)); err != nil {
		return fmt.Errorf("error setting parameters: %s", err)
	}

	return nil
}

func resourceAwsGluePartitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, tableName, err := readAwsGluePartitionID(d.Id())
	if err != nil {
		return err
	}

	input := &glue.UpdatePartitionInput{
		CatalogId:      aws.String(catalogID),
		DatabaseName:   aws.String(dbName),
		TableName:      aws.String(tableName),
		PartitionInput: expandGluePartitionInput(d),
	}

	if _, err := conn.UpdatePartition(input); err != nil {
		return fmt.Errorf("Error updating Glue Partition: %s", err)
	}

	return resourceAwsGluePartitionRead(d, meta)
}

func resourceAwsGluePartitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, tableName, tableErr := readAwsGluePartitionID(d.Id())
	if tableErr != nil {
		return tableErr
	}

	log.Printf("[DEBUG] Glue Partition: %s:%s:%s", catalogID, dbName, tableName)
	_, err := conn.DeletePartition(&glue.DeletePartitionInput{
		CatalogId:    aws.String(catalogID),
		TableName:    aws.String(tableName),
		DatabaseName: aws.String(dbName),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Glue Partition: %s", err.Error())
	}
	return nil
}

func expandGluePartitionInput(d *schema.ResourceData) *glue.PartitionInput {
	tableInput := &glue.PartitionInput{}

	if v, ok := d.GetOk("storage_descriptor"); ok {
		tableInput.StorageDescriptor = expandGlueStorageDescriptor(v.([]interface{}))
	}

	if v, ok := d.GetOk("parameters"); ok {
		paramsMap := map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			paramsMap[key] = value.(string)
		}
		tableInput.Parameters = aws.StringMap(paramsMap)
	}

	if v, ok := d.GetOk("values"); ok && v.(*schema.Set).Len() > 0 {
		tableInput.Values = expandStringSet(v.(*schema.Set))
	}

	return tableInput
}
