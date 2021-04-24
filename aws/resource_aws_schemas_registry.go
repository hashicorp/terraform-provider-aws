package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	schemas "github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSchemasRegistry() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSchemasRegistryCreate,
		Read:   resourceAwsSchemasRegistryRead,
		Update: resourceAwsSchemasRegistryUpdate,
		Delete: resourceAwsSchemasRegistryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+`), ""),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSchemasRegistryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	input := &schemas.CreateRegistryInput{}
	if name, ok := d.GetOk("name"); ok {
		input.RegistryName = aws.String(name.(string))
	}
	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SchemasTags()
	}

	log.Printf("[DEBUG] Creating Schemas Registry: %v", input)

	_, err := conn.CreateRegistry(input)
	if err != nil {
		return fmt.Errorf("Creating Schemas Registry (%s) failed: %w", *input.RegistryName, err)
	}

	d.SetId(aws.StringValue(input.RegistryName))

	log.Printf("[INFO] Schemas Registry (%s) created", d.Id())

	return resourceAwsSchemasRegistryRead(d, meta)
}

func resourceAwsSchemasRegistryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &schemas.DescribeRegistryInput{
		RegistryName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Schemas Registry (%s)", d.Id())
	output, err := conn.DescribeRegistry(input)
	if isAWSErr(err, schemas.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Schemas Registry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading Schemas Registry: %w", err)
	}

	log.Printf("[DEBUG] Found CloudWatch Event bus: %#v", *output)

	d.Set("arn", output.RegistryArn)
	d.Set("name", output.RegistryName)
	d.Set("description", output.Description)

	tags, err := keyvaluetags.SchemasListTags(conn, *output.RegistryArn)
	if err != nil {
		return fmt.Errorf("error listing tags for Schemas Registry (%s): %w", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSchemasRegistryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	if d.HasChanges("name", "description") {
		input := &schemas.UpdateRegistryInput{
			Description: aws.String(""),
		}
		if name, ok := d.GetOk("name"); ok {
			input.RegistryName = aws.String(name.(string))
		}
		if description, ok := d.GetOk("description"); ok {
			input.Description = aws.String(description.(string))
		}

		log.Printf("[DEBUG] Updating Schemas Registry: %s", input)
		_, err := conn.UpdateRegistry(input)
		if err != nil {
			return fmt.Errorf("error updating Schemas Registry (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.SchemasUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsSchemasRegistryRead(d, meta)
}

func resourceAwsSchemasRegistryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	log.Printf("[INFO] Deleting Schemas Registry (%s)", d.Id())
	_, err := conn.DeleteRegistry(&schemas.DeleteRegistryInput{
		RegistryName: aws.String(d.Id()),
	})
	if isAWSErr(err, schemas.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Schemas Registry (%s) not found", d.Id())
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting Schemas Registry (%s): %w", d.Id(), err)
	}
	log.Printf("[INFO] Schemas Registry (%s) deleted", d.Id())

	return nil
}
