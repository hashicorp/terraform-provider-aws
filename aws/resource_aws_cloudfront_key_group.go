package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsCloudFrontKeyGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFrontKeyGroupCreate,
		Read:   resourceAwsCloudFrontKeyGroupRead,
		Update: resourceAwsCloudFrontKeyGroupUpdate,
		Delete: resourceAwsCloudFrontKeyGroupDelete,
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

func resourceAwsCloudFrontKeyGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	input := &cloudfront.CreateKeyGroupInput{
		KeyGroupConfig: expandCloudFrontKeyGroupConfig(d),
	}

	log.Println("[DEBUG] Create CloudFront key group:", input)

	output, err := conn.CreateKeyGroup(input)
	if err != nil {
		return fmt.Errorf("error creating CloudFront key group: %s", err)
	}

	d.SetId(aws.StringValue(output.KeyGroup.Id))
	return resourceAwsCloudFrontKeyGroupRead(d, meta)
}

func resourceAwsCloudFrontKeyGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn
	input := &cloudfront.GetKeyGroupInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetKeyGroup(input)
	if err != nil {
		if isAWSErr(err, cloudfront.ErrCodeNoSuchResource, "") {
			log.Printf("[WARN] No key group found: %s, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if output == nil || output.KeyGroup == nil || output.KeyGroup.KeyGroupConfig == nil {
		log.Printf("[WARN] No key group found: %s, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	keyGroupConfig := output.KeyGroup.KeyGroupConfig

	d.Set("name", keyGroupConfig.Name)
	d.Set("comment", keyGroupConfig.Comment)
	d.Set("items", flattenStringSet(keyGroupConfig.Items))
	d.Set("etag", output.ETag)

	return nil
}

func resourceAwsCloudFrontKeyGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	input := &cloudfront.UpdateKeyGroupInput{
		Id:             aws.String(d.Id()),
		KeyGroupConfig: expandCloudFrontKeyGroupConfig(d),
		IfMatch:        aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateKeyGroup(input)
	if err != nil {
		return fmt.Errorf("error updating CloudFront key group (%s): %s", d.Id(), err)
	}

	return resourceAwsCloudFrontKeyGroupRead(d, meta)
}

func resourceAwsCloudFrontKeyGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	input := &cloudfront.DeleteKeyGroupInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteKeyGroup(input)
	if err != nil {
		if isAWSErr(err, cloudfront.ErrCodeNoSuchResource, "") {
			return nil
		}
		return err
	}

	return nil
}

func expandCloudFrontKeyGroupConfig(d *schema.ResourceData) *cloudfront.KeyGroupConfig {
	keyGroupConfig := &cloudfront.KeyGroupConfig{
		Items: expandStringSet(d.Get("items").(*schema.Set)),
		Name:  aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		keyGroupConfig.Comment = aws.String(v.(string))
	}

	return keyGroupConfig
}
