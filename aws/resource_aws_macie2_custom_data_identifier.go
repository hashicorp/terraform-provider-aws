package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsMacie2CustomDataIdentifier() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMacie2CustomDataIdentifierCreate,
		Read:   resourceAwsMacie2CustomDataIdentifierRead,
		Delete: resourceAwsMacie2CustomDataIdentifierDelete,
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
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"ignore_words": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(4, 90),
				},
			},
			"keywords": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(4, 90),
				},
			},
			"maximum_match_distance": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      50,
				ValidateFunc: validation.IntBetween(1, 300),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"regex": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 500),
			},
			"tags": tagsSchemaForceNew(),
		},
	}
}

func resourceAwsMacie2CustomDataIdentifierCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.CreateCustomDataIdentifierInput{
		ClientToken: aws.String(resource.UniqueId()),
		Name:        aws.String(d.Get("name").(string)),
		Regex:       aws.String(d.Get("regex").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("ignore_words"); ok {
		input.IgnoreWords = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("keywords"); ok {
		input.Keywords = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("maximum_match_distance"); ok {
		input.MaximumMatchDistance = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().Macie2Tags()
	}

	log.Printf("[DEBUG] Creating Macie2 Custom Data Identifier: %s", input)
	out, err := conn.CreateCustomDataIdentifier(input)
	if err != nil {
		return fmt.Errorf("Failed to create custom data identifier: %s", err)
	}

	d.SetId(*out.CustomDataIdentifierId)

	return resourceAwsMacie2CustomDataIdentifierRead(d, meta)
}

func resourceAwsMacie2CustomDataIdentifierRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).macie2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &macie2.GetCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetCustomDataIdentifier(input)

	if err != nil {
		return fmt.Errorf("error getting Custom Data Identifier (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("description", output.Description)
	d.Set("ignore_words", output.IgnoreWords)
	d.Set("keywords", output.Keywords)
	d.Set("maximum_match_distance", output.MaximumMatchDistance)
	d.Set("name", output.Name)
	d.Set("regex", output.Regex)

	if err := d.Set("tags", keyvaluetags.Macie2KeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsMacie2CustomDataIdentifierDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	}
	_, err := conn.DeleteCustomDataIdentifier(input)
	if err != nil {
		return fmt.Errorf("Error deleting Custom Data Identifier: %s", err)
	}

	return nil
}
