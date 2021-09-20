package glue

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9-_$#]+$`), ""),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRegistryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &glue.CreateRegistryInput{
		RegistryName: aws.String(d.Get("registry_name").(string)),
		Tags:         Tags(tags.IgnoreAws()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Registry: %s", input)
	output, err := conn.CreateRegistry(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Registry: %w", err)
	}
	d.SetId(aws.StringValue(output.RegistryArn))

	return resourceRegistryRead(d, meta)
}

func resourceRegistryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindRegistryByID(conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Registry (%s): %w", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(output.RegistryArn)
	d.Set("arn", arn)
	d.Set("description", output.Description)
	d.Set("registry_name", output.RegistryName)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Registry (%s): %w", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceRegistryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	if d.HasChanges("description") {
		input := &glue.UpdateRegistryInput{
			RegistryId: createAwsRegistryID(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Glue Registry: %#v", input)
		_, err := conn.UpdateRegistry(input)
		if err != nil {
			return fmt.Errorf("error updating Glue Registry (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceRegistryRead(d, meta)
}

func resourceRegistryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	log.Printf("[DEBUG] Deleting Glue Registry: %s", d.Id())
	input := &glue.DeleteRegistryInput{
		RegistryId: createAwsRegistryID(d.Id()),
	}

	_, err := conn.DeleteRegistry(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Glue Registry (%s): %w", d.Id(), err)
	}

	_, err = waitRegistryDeleted(conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error waiting for Glue Registry (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
