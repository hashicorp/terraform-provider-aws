package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	schemas "github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSchemasDiscoverer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSchemasDiscovererCreate,
		Read:   resourceAwsSchemasDiscovererRead,
		Update: resourceAwsSchemasDiscovererUpdate,
		Delete: resourceAwsSchemasDiscovererDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"source_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
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
			"discoverer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSchemasDiscovererCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	input := &schemas.CreateDiscovererInput{}
	if source, ok := d.GetOk("source_arn"); ok {
		input.SourceArn = aws.String(source.(string))
	}
	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SchemasTags()
	}

	log.Printf("[DEBUG] Creating Schemas Discoverer: %v", input)

	output, err := conn.CreateDiscoverer(input)
	if err != nil {
		return fmt.Errorf("Creating Schemas Discoverer (%s) failed: %w", *input.SourceArn, err)
	}

	d.SetId(aws.StringValue(output.DiscovererId))

	log.Printf("[INFO] Schemas Discoverer (%s) created", d.Id())

	return resourceAwsSchemasDiscovererRead(d, meta)
}

func resourceAwsSchemasDiscovererRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &schemas.DescribeDiscovererInput{
		DiscovererId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Schemas Discoverer (%s)", d.Id())
	output, err := conn.DescribeDiscoverer(input)
	if isAWSErr(err, schemas.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Schemas Discoverer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading Schemas Discoverer: %w", err)
	}

	log.Printf("[DEBUG] Found CloudWatch Event bus: %#v", *output)

	d.Set("arn", output.DiscovererArn)
	d.Set("source_arn", output.SourceArn)
	d.Set("description", output.Description)
	d.Set("discoverer_id", output.DiscovererId)

	tags, err := keyvaluetags.SchemasListTags(conn, *output.DiscovererArn)
	if err != nil {
		return fmt.Errorf("error listing tags for Schemas Discoverer (%s): %w", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSchemasDiscovererUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	if d.HasChange("description") {
		input := &schemas.UpdateDiscovererInput{
			DiscovererId: aws.String(d.Id()),
			Description:  aws.String(""),
		}
		if description, ok := d.GetOk("description"); ok {
			input.Description = aws.String(description.(string))
		}

		log.Printf("[DEBUG] Updating Schemas Discoverer: %s", input)
		_, err := conn.UpdateDiscoverer(input)
		if err != nil {
			return fmt.Errorf("error updating Schemas Discoverer (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.SchemasUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsSchemasDiscovererRead(d, meta)
}

func resourceAwsSchemasDiscovererDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	log.Printf("[INFO] Deleting Schemas Discoverer (%s)", d.Id())
	_, err := conn.DeleteDiscoverer(&schemas.DeleteDiscovererInput{
		DiscovererId: aws.String(d.Id()),
	})
	if isAWSErr(err, schemas.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Schemas Discoverer (%s) not found", d.Id())
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting Schemas Discoverer (%s): %w", d.Id(), err)
	}
	log.Printf("[INFO] Schemas Discoverer (%s) deleted", d.Id())

	return nil
}
