package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCloudFrontPublicKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFrontPublicKeyCreate,
		Read:   resourceAwsCloudFrontPublicKeyRead,
		Update: resourceAwsCloudFrontPublicKeyUpdate,
		Delete: resourceAwsCloudFrontPublicKeyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateCloudFrontPublicKeyName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateCloudFrontPublicKeyNamePrefix,
			},
			"encoded_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloudFrontPublicKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	if v, ok := d.GetOk("name"); ok {
		d.Set("name", v.(string))
	} else if v, ok := d.GetOk("name_prefix"); ok {
		d.Set("name", resource.PrefixedUniqueId(v.(string)))
	} else {
		d.Set("name", resource.PrefixedUniqueId("tf-"))
	}

	request := &cloudfront.CreatePublicKeyInput{
		PublicKeyConfig: expandPublicKeyConfig(d),
	}

	log.Println("[DEBUG] Create CloudFront PublicKey:", request)

	output, err := conn.CreatePublicKey(request)
	if err != nil {
		return fmt.Errorf("error creating CloudFront PublicKey: %s", err)
	}

	d.SetId(aws.StringValue(output.PublicKey.Id))
	return resourceAwsCloudFrontPublicKeyRead(d, meta)
}

func expandPublicKeyConfig(d *schema.ResourceData) *cloudfront.PublicKeyConfig {
	publicKeyConfig := &cloudfront.PublicKeyConfig{
		CallerReference: aws.String(resource.UniqueId()),
		EncodedKey:      aws.String(d.Get("encoded_key").(string)),
		Name:            aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		publicKeyConfig.Comment = aws.String(v.(string))
	}

	return publicKeyConfig
}
