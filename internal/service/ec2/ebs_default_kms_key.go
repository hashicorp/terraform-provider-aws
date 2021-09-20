package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEBSDefaultKMSKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceEBSDefaultKMSKeyCreate,
		Read:   resourceEBSDefaultKMSKeyRead,
		Delete: resourceEBSDefaultKMSKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"key_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceEBSDefaultKMSKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	resp, err := conn.ModifyEbsDefaultKmsKeyId(&ec2.ModifyEbsDefaultKmsKeyIdInput{
		KmsKeyId: aws.String(d.Get("key_arn").(string)),
	})
	if err != nil {
		return fmt.Errorf("error creating EBS default KMS key: %s", err)
	}

	d.SetId(aws.StringValue(resp.KmsKeyId))

	return resourceEBSDefaultKMSKeyRead(d, meta)
}

func resourceEBSDefaultKMSKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	resp, err := conn.GetEbsDefaultKmsKeyId(&ec2.GetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return fmt.Errorf("error reading EBS default KMS key: %s", err)
	}

	d.Set("key_arn", resp.KmsKeyId)

	return nil
}

func resourceEBSDefaultKMSKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	_, err := conn.ResetEbsDefaultKmsKeyId(&ec2.ResetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return fmt.Errorf("error deleting EBS default KMS key: %s", err)
	}

	return nil
}
