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
		Delete: resourceAwsGlueCatalogDatabaseDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"create_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"location_uri": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsGlueCatalogDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn

	input := &glue.CreateDatabaseInput{
		DatabaseInput: &glue.DatabaseInput{
			Name: aws.String(d.Get("name").(string)),
		},
	}

	if description, ok := d.GetOk("description"); ok {
		input.DatabaseInput.Description = aws.String(description.(string))
	}

	_, err := glueconn.CreateDatabase(input)
	if err != nil {
		return fmt.Errorf("Error creating Catalogue Database: %s", err)
	}

	d.SetId(d.Get("name").(string))

	return resourceAwsGlueCatalogDatabaseUpdate(d, meta)
}

func resourceAwsGlueCatalogDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	//glueconn := meta.(*AWSClient).glueconn

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

	d.Set("create_time", out.Database.CreateTime)
	d.Set("description", out.Database.Description)
	d.Set("location_uri", out.Database.LocationUri)

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
