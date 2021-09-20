package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/schemas/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceRegistry() *schema.Resource {
	return &schema.Resource{
		Create: resourceRegistryCreate,
		Read:   resourceRegistryRead,
		Update: resourceRegistryUpdate,
		Delete: resourceRegistryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+`), ""),
				),
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRegistryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SchemasConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &schemas.CreateRegistryInput{
		RegistryName: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().SchemasTags()
	}

	log.Printf("[DEBUG] Creating EventBridge Schemas Registry: %s", input)
	_, err := conn.CreateRegistry(input)

	if err != nil {
		return fmt.Errorf("error creating EventBridge Schemas Registry (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(input.RegistryName))

	return resourceRegistryRead(d, meta)
}

func resourceRegistryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SchemasConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := finder.RegistryByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Registry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EventBridge Schemas Registry (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.RegistryArn)
	d.Set("description", output.Description)
	d.Set("name", output.RegistryName)

	tags, err := tftags.SchemasListTags(conn, d.Get("arn").(string))

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err != nil {
		return fmt.Errorf("error listing tags for EventBridge Schemas Registry (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceRegistryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SchemasConn

	if d.HasChanges("description") {
		input := &schemas.UpdateRegistryInput{
			Description:  aws.String(d.Get("description").(string)),
			RegistryName: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating EventBridge Schemas Registry: %s", input)
		_, err := conn.UpdateRegistry(input)

		if err != nil {
			return fmt.Errorf("error updating EventBridge Schemas Registry (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.SchemasUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceRegistryRead(d, meta)
}

func resourceRegistryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SchemasConn

	log.Printf("[INFO] Deleting EventBridge Schemas Registry (%s)", d.Id())
	_, err := conn.DeleteRegistry(&schemas.DeleteRegistryInput{
		RegistryName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EventBridge Schemas Registry (%s): %w", d.Id(), err)
	}

	return nil
}
