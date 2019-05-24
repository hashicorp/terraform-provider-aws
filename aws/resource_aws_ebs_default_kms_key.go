package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsEbsDefaultKmsKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEbsDefaultKmsKeyCreate,
		Read:   resourceAwsEbsDefaultKmsKeyRead,
		Update: resourceAwsEbsDefaultKmsKeyUpdate,
		Delete: resourceAwsEbsDefaultKmsKeyDelete,

		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsEbsDefaultKmsKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.ModifyEbsDefaultKmsKeyId(&ec2.ModifyEbsDefaultKmsKeyIdInput{
		KmsKeyId: aws.String(d.Get("key_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("error creating EBS default KMS key: %s", err)
	}

	d.SetId(resource.UniqueId())

	return resourceAwsEbsDefaultKmsKeyRead(d, meta)
}

func resourceAwsEbsDefaultKmsKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resp, err := conn.GetEbsDefaultKmsKeyId(&ec2.GetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return fmt.Errorf("error reading EBS default KMS key: %s", err)
	}

	d.Set("key_id", aws.StringValue(resp.KmsKeyId))

	return nil
}

func resourceAwsEbsDefaultKmsKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.ModifyEbsDefaultKmsKeyId(&ec2.ModifyEbsDefaultKmsKeyIdInput{
		KmsKeyId: aws.String(d.Get("key_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("error updating EBS default KMS key: %s", err)
	}

	return resourceAwsEbsDefaultKmsKeyRead(d, meta)
}

func resourceAwsEbsDefaultKmsKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.ResetEbsDefaultKmsKeyId(&ec2.ResetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return fmt.Errorf("error deleting EBS default KMS key: %s", err)
	}

	return nil
}
