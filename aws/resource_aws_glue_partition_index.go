package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfglue "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsGluePartitionIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGluePartitionIndexCreate,
		Read:   resourceAwsGluePartitionIndexRead,
		Delete: resourceAwsGluePartitionIndexDelete,
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
			"partition_index": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"index_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"keys": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceAwsGluePartitionIndexCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(d, meta.(*AWSClient).accountid)
	dbName := d.Get("database_name").(string)
	tableName := d.Get("table_name").(string)

	input := &glue.CreatePartitionIndexInput{
		CatalogId:      aws.String(catalogID),
		DatabaseName:   aws.String(dbName),
		TableName:      aws.String(tableName),
		PartitionIndex: expandAwsGluePartitionIndex(d.Get("partition_index").([]interface{})),
	}

	log.Printf("[DEBUG] Creating Glue Partition Index: %#v", input)
	_, err := conn.CreatePartitionIndex(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Partition Index: %w", err)
	}

	d.SetId(tfglue.CreateAwsGluePartitionIndexID(catalogID, dbName, tableName, aws.StringValue(input.PartitionIndex.IndexName)))

	if _, err := waiter.GluePartitionIndexCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error while waiting for Glue Partition Index (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsGluePartitionIndexRead(d, meta)
}

func resourceAwsGluePartitionIndexRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, tableName, _, tableErr := tfglue.ReadAwsGluePartitionIndexID(d.Id())
	if tableErr != nil {
		return tableErr
	}

	log.Printf("[DEBUG] Reading Glue Partition Index: %s", d.Id())
	partition, err := finder.PartitionIndexByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Partition Index (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Glue Partition Index (%s): %w", d.Id(), err)
	}

	d.Set("table_name", tableName)
	d.Set("catalog_id", catalogID)
	d.Set("database_name", dbName)

	if err := d.Set("partition_index", []map[string]interface{}{flattenGluePartitionIndex(partition)}); err != nil {
		return fmt.Errorf("error setting partition_index: %w", err)
	}

	return nil
}

func resourceAwsGluePartitionIndexDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, tableName, partIndex, tableErr := tfglue.ReadAwsGluePartitionIndexID(d.Id())
	if tableErr != nil {
		return tableErr
	}

	log.Printf("[DEBUG] Deleting Glue Partition Index: %s", d.Id())
	_, err := conn.DeletePartitionIndex(&glue.DeletePartitionIndexInput{
		CatalogId:    aws.String(catalogID),
		TableName:    aws.String(tableName),
		DatabaseName: aws.String(dbName),
		IndexName:    aws.String(partIndex),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting Glue Partition Index: %w", err)
	}

	if _, err := waiter.GluePartitionIndexDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error while waiting for Glue Partition Index (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func expandAwsGluePartitionIndex(l []interface{}) *glue.PartitionIndex {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	parIndex := &glue.PartitionIndex{}

	if v, ok := s["keys"].([]interface{}); ok && len(v) > 0 {
		parIndex.Keys = expandStringList(v)
	}

	if v, ok := s["index_name"].(string); ok && v != "" {
		parIndex.IndexName = aws.String(v)
	}

	return parIndex
}
