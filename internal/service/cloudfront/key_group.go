package cloudfront

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceKeyGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceKeyGroupCreate,
		Read:   resourceKeyGroupRead,
		Update: resourceKeyGroupUpdate,
		Delete: resourceKeyGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"items": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceKeyGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	input := &cloudfront.CreateKeyGroupInput{
		KeyGroupConfig: expandKeyGroupConfig(d),
	}

	log.Println("[DEBUG] Create CloudFront Key Group:", input)

	output, err := conn.CreateKeyGroup(input)
	if err != nil {
		return fmt.Errorf("error creating CloudFront Key Group: %w", err)
	}

	if output == nil || output.KeyGroup == nil {
		return fmt.Errorf("error creating CloudFront Key Group: empty response")
	}

	d.SetId(aws.StringValue(output.KeyGroup.Id))
	return resourceKeyGroupRead(d, meta)
}

func resourceKeyGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn
	input := &cloudfront.GetKeyGroupInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetKeyGroup(input)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchResource) {
			log.Printf("[WARN] No key group found: %s, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading CloudFront Key Group (%s): %w", d.Id(), err)
	}

	if output == nil || output.KeyGroup == nil || output.KeyGroup.KeyGroupConfig == nil {
		return fmt.Errorf("error reading CloudFront Key Group: empty response")
	}

	keyGroupConfig := output.KeyGroup.KeyGroupConfig

	d.Set("name", keyGroupConfig.Name)
	d.Set("comment", keyGroupConfig.Comment)
	d.Set("items", flex.FlattenStringSet(keyGroupConfig.Items))
	d.Set("etag", output.ETag)

	return nil
}

func resourceKeyGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	input := &cloudfront.UpdateKeyGroupInput{
		Id:             aws.String(d.Id()),
		KeyGroupConfig: expandKeyGroupConfig(d),
		IfMatch:        aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateKeyGroup(input)
	if err != nil {
		return fmt.Errorf("error updating CloudFront Key Group (%s): %w", d.Id(), err)
	}

	return resourceKeyGroupRead(d, meta)
}

func resourceKeyGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	input := &cloudfront.DeleteKeyGroupInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteKeyGroup(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchResource) {
			return nil
		}
		return fmt.Errorf("error deleting CloudFront Key Group (%s): %w", d.Id(), err)
	}

	return nil
}

func expandKeyGroupConfig(d *schema.ResourceData) *cloudfront.KeyGroupConfig {
	keyGroupConfig := &cloudfront.KeyGroupConfig{
		Items: flex.ExpandStringSet(d.Get("items").(*schema.Set)),
		Name:  aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		keyGroupConfig.Comment = aws.String(v.(string))
	}

	return keyGroupConfig
}
