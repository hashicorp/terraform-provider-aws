package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceProtection() *schema.Resource {
	return &schema.Resource{
		Create: resourceProtectionCreate,
		Update: resourceProtectionUpdate,
		Read:   resourceProtectionRead,
		Delete: resourceProtectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
		CustomizeDiff: SetTagsDiff,
	}
}

func resourceProtectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.ShieldUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceProtectionRead(d, meta)
}

func resourceProtectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &shield.CreateProtectionInput{
		Name:        aws.String(d.Get("name").(string)),
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
		Tags:        tags.IgnoreAws().ShieldTags(),
	}

	resp, err := conn.CreateProtection(input)
	if err != nil {
		return fmt.Errorf("error creating Shield Protection: %s", err)
	}
	d.SetId(aws.StringValue(resp.ProtectionId))
	return resourceProtectionRead(d, meta)
}

func resourceProtectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeProtection(input)

	if tfawserr.ErrMessageContains(err, shield.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Shield Protection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Shield Protection (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.Protection.ProtectionArn)
	d.Set("arn", arn)
	d.Set("name", resp.Protection.Name)
	d.Set("resource_arn", resp.Protection.ResourceArn)

	tags, err := keyvaluetags.ShieldListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Shield Protection (%s): %s", arn, err)
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

func resourceProtectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	input := &shield.DeleteProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	_, err := conn.DeleteProtection(input)

	if tfawserr.ErrMessageContains(err, shield.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Shield Protection (%s): %s", d.Id(), err)
	}
	return nil
}
