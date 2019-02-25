package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsShieldProtection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsShieldProtectionCreate,
		Read:   resourceAwsShieldProtectionRead,
		Delete: resourceAwsShieldProtectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsShieldProtectionImport,
		},

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func resourceAwsShieldProtectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.CreateProtectionInput{
		Name:        aws.String(d.Get("name").(string)),
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	resp, err := conn.CreateProtection(input)
	if err != nil {
		return err
	}
	d.SetId(*resp.ProtectionId)
	return resourceAwsShieldProtectionRead(d, meta)
}

func resourceAwsShieldProtectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	_, err := conn.DescribeProtection(input)
	if err != nil {
		return err
	}
	return nil
}

func resourceAwsShieldProtectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DeleteProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	_, err := conn.DeleteProtection(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case shield.ErrCodeResourceNotFoundException:
				return nil
			default:
				return err
			}
		}
		return err
	}
	return nil
}

func resourceAwsShieldProtectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeProtection(input)
	if err != nil {
		return nil, err
	}

	d.Set("name", resp.Protection.Name)
	d.Set("resource_arn", resp.Protection.ResourceArn)

	return []*schema.ResourceData{d}, nil
}
