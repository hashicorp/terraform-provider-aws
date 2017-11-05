package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"log"
)

func resourceAwsGlueCatalogDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueCatalogDatabaseCreate,
		Read:   resourceAwsGlueCatalogDatabaseRead,
		Update: resourceAwsGlueCatalogDatabaseUpdate,
		Delete: resourceAwsGlueCatalogDatabaseDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"location_uri": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"parameters": &schema.Schema{
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Optional: true,
			},
			"create_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsGlueCatalogDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	input := &glue.CreateDatabaseInput{
		DatabaseInput: &glue.DatabaseInput{
			Name: aws.String(name),
		},
	}

	_, err := glueconn.CreateDatabase(input)
	if err != nil {
		return fmt.Errorf("Error creating Catalogue Database: %s", err)
	}

	d.SetId(name)

	return resourceAwsGlueCatalogDatabaseUpdate(d, meta)
}

func resourceAwsGlueCatalogDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn
	doUpdate := false
	input := &glue.UpdateDatabaseInput{
		DatabaseInput: &glue.DatabaseInput{
			Name: aws.String(d.Id()),
		},
		Name: aws.String(d.Id()),
	}

	if ok := d.HasChange("description"); ok {
		doUpdate = true
		input.DatabaseInput.Description = aws.String(
			d.Get("description").(string),
		)
	}

	if ok := d.HasChange("location_uri"); ok {
		doUpdate = true
		input.DatabaseInput.LocationUri = aws.String(
			d.Get("location_uri").(string),
		)
	}

	if ok := d.HasChange("parameters"); ok {
		doUpdate = true
		input.DatabaseInput.Parameters = make(map[string]*string)
		for key, value := range d.Get("parameters").(map[string]interface{}) {
			input.DatabaseInput.Parameters[key] = aws.String(value.(string))
		}
	}

	if doUpdate {
		if _, err := glueconn.UpdateDatabase(input); err != nil {
			return err
		}
	}

	return resourceAwsGlueCatalogDatabaseRead(d, meta)
}

func resourceAwsGlueCatalogDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn

	input := &glue.GetDatabaseInput{
		Name: aws.String(d.Id()),
	}

	out, err := glueconn.GetDatabase(input)
	if err != nil {
		return fmt.Errorf("Error reading Glue Cataloge Database: %s", err.Error())
	}

	d.Set("name", d.Id())
	d.Set("create_time", out.Database.CreateTime)
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

	log.Printf("[DEBUG] Glue Catalog Database: %s", d.Id())
	_, err := glueconn.DeleteDatabase(&glue.DeleteDatabaseInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Glue Catalog Database: %s", err.Error())
	}
	return nil
}
