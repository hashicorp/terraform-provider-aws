package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	schemas "github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSchemasSchema() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSchemasSchemaCreate,
		Read:   resourceAwsSchemasSchemaRead,
		Update: resourceAwsSchemasSchemaUpdate,
		Delete: resourceAwsSchemasSchemaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 385),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z@]+`), ""),
				),
			},
			"content": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"registry": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(schemas.Type_Values(), true),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSchemasSchemaCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	input := &schemas.CreateSchemaInput{}
	if name, ok := d.GetOk("name"); ok {
		input.SchemaName = aws.String(name.(string))
	}
	if registry, ok := d.GetOk("registry"); ok {
		input.RegistryName = aws.String(registry.(string))
	}
	if schemaType, ok := d.GetOk("type"); ok {
		input.Type = aws.String(schemaType.(string))
	}
	if content, ok := d.GetOk("content"); ok {
		input.Content = aws.String(content.(string))
	}
	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SchemasTags()
	}

	log.Printf("[DEBUG] Creating Schemas Schema: %v", input)

	_, err := conn.CreateSchema(input)
	if err != nil {
		return fmt.Errorf("Creating Schemas Schema (%s) failed: %w", *input.SchemaName, err)
	}

	id := fmt.Sprintf("%s/%s", aws.StringValue(input.SchemaName), aws.StringValue(input.RegistryName))
	d.SetId(id)

	log.Printf("[INFO] Schemas Schema (%s) created", d.Id())

	return resourceAwsSchemasSchemaRead(d, meta)
}

func resourceAwsSchemasSchemaRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	schemaName, registryName, err := parseSchemaID(d.Id())
	if err != nil {
		return err
	}

	input := &schemas.DescribeSchemaInput{
		SchemaName:   aws.String(schemaName),
		RegistryName: aws.String(registryName),
	}

	log.Printf("[DEBUG] Reading Schemas Schema (%s)", d.Id())
	output, err := conn.DescribeSchema(input)
	if isAWSErr(err, schemas.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Schemas Schema (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading Schemas Schema: %w", err)
	}

	log.Printf("[DEBUG] Found CloudWatch Event bus: %#v", *output)

	d.Set("arn", output.SchemaArn)
	d.Set("name", output.SchemaName)
	d.Set("registry", aws.StringValue(input.RegistryName))
	d.Set("content", output.Content)
	d.Set("description", output.Description)
	d.Set("type", output.Type)
	d.Set("version", output.SchemaVersion)
	d.Set("version_created_date", output.VersionCreatedDate.String())
	d.Set("last_modified", output.LastModified.String())

	tags, err := keyvaluetags.SchemasListTags(conn, *output.SchemaArn)
	if err != nil {
		return fmt.Errorf("error listing tags for Schemas Schema (%s): %w", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSchemasSchemaUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	if d.HasChanges("name", "registry", "type", "content", "description") {
		input := &schemas.UpdateSchemaInput{
			Description: aws.String(""),
		}
		if name, ok := d.GetOk("name"); ok {
			input.SchemaName = aws.String(name.(string))
		}
		if registry, ok := d.GetOk("registry"); ok {
			input.RegistryName = aws.String(registry.(string))
		}
		if schemaType, ok := d.GetOk("type"); ok {
			input.Type = aws.String(schemaType.(string))
		}
		if content, ok := d.GetOk("content"); ok {
			input.Content = aws.String(content.(string))
		}
		if description, ok := d.GetOk("description"); ok {
			input.Description = aws.String(description.(string))
		}
		log.Printf("[DEBUG] Updating Schemas Schema: %s", input)
		_, err := conn.UpdateSchema(input)
		if err != nil {
			return fmt.Errorf("error updating Schemas Schema (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.SchemasUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsSchemasSchemaRead(d, meta)
}

func resourceAwsSchemasSchemaDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	schemaName, registryName, err := parseSchemaID(d.Id())
	if err != nil {
		return err
	}

	input := &schemas.DeleteSchemaInput{
		SchemaName:   aws.String(schemaName),
		RegistryName: aws.String(registryName),
	}

	log.Printf("[INFO] Deleting Schemas Schema (%s)", d.Id())
	_, err = conn.DeleteSchema(input)
	if isAWSErr(err, schemas.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Schemas Schema (%s) not found", d.Id())
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting Schemas Schema (%s): %w", d.Id(), err)
	}
	log.Printf("[INFO] Schemas Schema (%s) deleted", d.Id())

	return nil
}

func parseSchemaID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%q), expected SCHEMA_NAME:REGISTRY_NAME", id)
	}
	return parts[0], parts[1], nil
}
