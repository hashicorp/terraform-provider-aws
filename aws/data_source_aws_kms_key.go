package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/aws/aws-sdk-go/service/kms"
	"fmt"
)

func dataSourceAwsKmsKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKmsKeyRead,
		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateKmsKey,
			},
			"grant_tokens": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"deletion_date": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"expiration_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_manager": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_usage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_to": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsKmsKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn
	keyId := d.Get("key_id")
	input := &kms.DescribeKeyInput{
		KeyId: keyId,
		GrantTokens: d.Get("grant_tokens"),
	}
	output, err := conn.DescribeKey(input)
	if err != nil {
		return fmt.Errorf("Error while describing key [%s]: %s", keyId, err)
	}
	d.SetId(keyId)
	d.Set("arn", output.KeyMetadata.Arn)
	return nil
}
