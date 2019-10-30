package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsGreengrassGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassGroupCreate,
		Read:   resourceAwsGreengrassGroupRead,
		Update: resourceAwsGreengrassGroupUpdate,
		Delete: resourceAwsGreengrassGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsGreengrassGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.CreateGroupInput{
		Name: aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Creating Greengrass Group: %s", params)
	out, err := conn.CreateGroup(params)
	if err != nil {
		return err
	}

	d.SetId(*out.Id)

	return resourceAwsGreengrassGroupRead(d, meta)
}

func resourceAwsGreengrassGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.GetGroupInput{
		GroupId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrass Group: %s", params)
	out, err := conn.GetGroup(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Group: %s", out)

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)
	d.Set("group_id", out.Id)

	return nil
}

func resourceAwsGreengrassGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.UpdateGroupInput{
		Name:    aws.String(d.Get("name").(string)),
		GroupId: aws.String(d.Get("group_id").(string)),
	}

	_, err := conn.UpdateGroup(params)
	if err != nil {
		return err
	}

	return resourceAwsGreengrassGroupRead(d, meta)
}

func resourceAwsGreengrassGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.DeleteGroupInput{
		GroupId: aws.String(d.Get("group_id").(string)),
	}
	log.Printf("[DEBUG] Deleting Greengrass Group: %s", params)

	_, err := conn.DeleteGroup(params)
	if err != nil {
		return nil
	}

	return nil
}
