package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsKmsPublicKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKmsPublicKeyRead,
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
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_master_key_spec": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_algorithms": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"key_usage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_algorithms": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsKmsPublicKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn
	keyId := d.Get("key_id")
	var grantTokens []*string
	if v, ok := d.GetOk("grant_tokens"); ok {
		grantTokens = aws.StringSlice(v.([]string))
	}
	input := &kms.GetPublicKeyInput{
		KeyId:       aws.String(keyId.(string)),
		GrantTokens: grantTokens,
	}
	output, err := conn.GetPublicKey(input)
	if err != nil {
		return fmt.Errorf("error while describing key [%s]: %w", keyId, err)
	}
	d.SetId(aws.StringValue(output.KeyId))
	d.Set("customer_master_key_spec", output.CustomerMasterKeySpec)
	d.Set("encryption_algorithms", output.EncryptionAlgorithms)
	d.Set("id", output.KeyId)
	d.Set("key_usage", output.KeyUsage)
	d.Set("public_key", string(output.PublicKey))
	d.Set("signing_algorithms", output.SigningAlgorithms)

	return nil
}
