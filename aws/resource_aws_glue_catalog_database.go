package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsGlueCatalogDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueCatalogDatabaseCreate,
		Read:   resourceAwsGlueCatalogDatabaseRead,
		Update: resourceAwsGlueCatalogDatabaseUpdate,
		Delete: resourceAwsGlueCatalogDatabaseDelete,
		Exists: resourceAwsGlueCatalogDatabaseExists,
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
				ForceNew: true,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"location_uri": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAwsGlueCatalogDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(d, meta)
	name := d.Get("name").(string)

	input := &glue.CreateDatabaseInput{
		CatalogId: aws.String(catalogID),
		DatabaseInput: &glue.DatabaseInput{
			Name: aws.String(name),
		},
	}

	_, err := glueconn.CreateDatabase(input)
	if err != nil {
		return fmt.Errorf("Error creating Catalogue Database: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, name))

	return resourceAwsGlueCatalogDatabaseUpdate(d, meta)
}

func resourceAwsGlueCatalogDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn

	catalogID, name := readAwsGlueCatalogID(d.Id())
	input := &glue.UpdateDatabaseInput{
		CatalogId: aws.String(catalogID),
		DatabaseInput: &glue.DatabaseInput{
			Name:        aws.String(name),
			Description: aws.String(d.Get("description").(string)),
			LocationUri: aws.String(d.Get("location_uri").(string)),
			Parameters:  make(map[string]*string),
		},
		Name: aws.String(name),
	}
	for key, value := range d.Get("parameters").(map[string]interface{}) {
		input.DatabaseInput.Parameters[key] = aws.String(value.(string))
	}

	doUpdate := d.HasChange("description")
	doUpdate = doUpdate || d.HasChange("location_uri")
	doUpdate = doUpdate || d.HasChange("parameters")

	if doUpdate {
		if _, err := glueconn.UpdateDatabase(input); err != nil {
			return err
		}
	}

	return resourceAwsGlueCatalogDatabaseRead(d, meta)
}

func resourceAwsGlueCatalogDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn

	catalogID, name := readAwsGlueCatalogID(d.Id())

	input := &glue.GetDatabaseInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	out, err := glueconn.GetDatabase(input)
	if err != nil {
		return fmt.Errorf("Error reading Glue Cataloge Database: %s", err.Error())
	}

	d.Set("name", out.Database.Name)
	d.Set("catalog_id", catalogID)
	d.Set("description", out.Database.Description)
	d.Set("location_uri", out.Database.LocationUri)

	dParams := make(map[string]string)
	if len(out.Database.Parameters) > 0 {
		for key, value := range out.Database.Parameters {
			dParams[key] = *value
		}
	}
	d.Set("parameters", dParams)

	return nil
}

func resourceAwsGlueCatalogDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn
	catalogID, name := readAwsGlueCatalogID(d.Id())

	log.Printf("[DEBUG] Glue Catalog Database: %s:%s", catalogID, name)
	_, err := glueconn.DeleteDatabase(&glue.DeleteDatabaseInput{
		Name: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Glue Catalog Database: %s", err.Error())
	}
	return nil
}

func resourceAwsGlueCatalogDatabaseExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	glueconn := meta.(*AWSClient).glueconn
	catalogID, name := readAwsGlueCatalogID(d.Id())

	input := &glue.GetDatabaseInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	_, err := glueconn.GetDatabase(input)
	return err == nil, err
}

func readAwsGlueCatalogID(id string) (catalogID string, name string) {
	idParts := strings.Split(id, ":")
	return idParts[0], idParts[1]
}

func createAwsGlueCatalogID(d *schema.ResourceData, meta interface{}) (catalogID string) {
	if rawCatalogID, ok := d.GetOkExists("catalog_id"); ok {
		catalogID = rawCatalogID.(string)
	} else {
		catalogID = meta.(*AWSClient).accountid
	}
	return
}
