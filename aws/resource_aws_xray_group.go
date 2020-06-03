package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsXrayGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsXrayGroupCreate,
		Read:   resourceAwsXrayGroupRead,
		Update: resourceAwsXrayGroupUpdate,
		Delete: resourceAwsXrayGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter_expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsXrayGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn
	input := &xray.CreateGroupInput{
		GroupName:        aws.String(d.Get("group_name").(string)),
		FilterExpression: aws.String(d.Get("filter_expression").(string)),
	}

	out, err := conn.CreateGroup(input)
	if err != nil {
		return fmt.Errorf("error creating XRay Group: %s", err)
	}

	d.SetId(aws.StringValue(out.Group.GroupARN))

	return resourceAwsXrayGroupRead(d, meta)
}

func resourceAwsXrayGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn

	input := &xray.GetGroupInput{
		GroupARN: aws.String(d.Id()),
	}

	group, err := conn.GetGroup(input)

	if err != nil {
		if isAWSErr(err, xray.ErrCodeInvalidRequestException, "Group not found") {
			log.Printf("[WARN] XRay Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading XRay Group (%s): %s", d.Id(), err)
	}

	d.Set("arn", group.Group.GroupARN)
	d.Set("group_name", group.Group.GroupName)
	d.Set("filter_expression", group.Group.FilterExpression)

	return nil
}

func resourceAwsXrayGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn

	input := &xray.UpdateGroupInput{
		GroupARN:         aws.String(d.Id()),
		FilterExpression: aws.String(d.Get("filter_expression").(string)),
	}

	_, err := conn.UpdateGroup(input)
	if err != nil {
		return fmt.Errorf("error updating XRay Group (%s): %s", d.Id(), err)
	}

	return resourceAwsXrayGroupRead(d, meta)
}

func resourceAwsXrayGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn

	log.Printf("[INFO] Deleting XRay Group: %s", d.Id())

	params := &xray.DeleteGroupInput{
		GroupARN: aws.String(d.Id()),
	}
	_, err := conn.DeleteGroup(params)
	if err != nil {
		return fmt.Errorf("error deleting XRay Group: %s", d.Id())
	}

	return nil
}
