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
				Required: true,
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
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
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

func resourceAwsGlueCatalogTableUpdate(t *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn

	catalogID, dbName, _ := readAwsGlueTableID(t.Id())

	updateTableInput := &glue.UpdateTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableInput:   expandGlueTableInput(t),
	}

	if t.HasChange("table_input") {
		if _, err := glueconn.UpdateTable(updateTableInput); err != nil {
			return fmt.Errorf("Error updating Glue Catalog Table: %s", err)
		}
	}

	return nil
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
	t.Set("description", out.Table.Description)

	return nil
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
		tableInput.Retention = aws.Int64(v.(int64))
	}

	if v, ok := t.GetOk("storage_descriptor"); ok {
		tableInput.StorageDescriptor = expandGlueStorageDescriptor(v.(map[string]interface{}))
	}

	if v, ok := t.GetOk("partition_keys"); ok {
		columns := expandGlueColumns(v.(map[string]schema.ResourceData))
		tableInput.PartitionKeys = columns
	}

	if v, ok := t.GetOk("view_original_text"); ok {
		tableInput.Owner = aws.String(v.(string))
	}

	if v, ok := t.GetOk("view_expanded_text"); ok {
		tableInput.Owner = aws.String(v.(string))
	}

	if v, ok := t.GetOk("table_type"); ok {
		tableInput.Owner = aws.String(v.(string))
	}

	if v, ok := t.GetOk("parameters"); ok {
		tableInput.Parameters = aws.StringMap(v.(map[string]string))
	}

	return tableInput
}

func expandGlueStorageDescriptor(s map[string]interface{}) *glue.StorageDescriptor {
	storageDescriptor := &glue.StorageDescriptor{}

	return storageDescriptor
}

func expandGlueColumns(columns map[string]schema.ResourceData) []*glue.Column {
	columnSlice := []*glue.Column{}
	for _, element := range columns {
		column := &glue.Column{
			Name: aws.String(element.Get("name").(string)),
		}

		if v, ok := element.GetOk("comment"); ok {
			column.Comment = aws.String(v.(string))
		}

		if v, ok := element.GetOk("type"); ok {
			column.Type = aws.String(v.(string))
		}

		columnSlice = append(columnSlice, column)
	}

	return columnSlice
}
