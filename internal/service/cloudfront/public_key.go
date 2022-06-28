package cloudfront

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourcePublicKey() *schema.Resource {
	return &schema.Resource{
		Create: resourcePublicKeyCreate,
		Read:   resourcePublicKeyRead,
		Update: resourcePublicKeyUpdate,
		Delete: resourcePublicKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encoded_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validPublicKeyName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validPublicKeyNamePrefix,
			},
		},
	}
}

func resourcePublicKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

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
	return resourcePublicKeyRead(d, meta)
}

func resourcePublicKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn
	request := &cloudfront.GetPublicKeyInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetPublicKey(request)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchPublicKey) {
		names.LogNotFoundRemoveState(names.CloudFront, names.ErrActionReading, ResPublicKey, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CloudFront, names.ErrActionReading, ResPublicKey, d.Id(), err)
	}

	if !d.IsNewResource() && (output == nil || output.PublicKey == nil || output.PublicKey.PublicKeyConfig == nil) {
		names.LogNotFoundRemoveState(names.CloudFront, names.ErrActionReading, ResPublicKey, d.Id())
		d.SetId("")
		return nil
	}

	if d.IsNewResource() && (output == nil || output.PublicKey == nil || output.PublicKey.PublicKeyConfig == nil) {
		return names.Error(names.CloudFront, names.ErrActionReading, ResPublicKey, d.Id(), errors.New("empty response after creation"))
	}

	publicKeyConfig := output.PublicKey.PublicKeyConfig
	d.Set("encoded_key", publicKeyConfig.EncodedKey)
	d.Set("name", publicKeyConfig.Name)
	d.Set("comment", publicKeyConfig.Comment)
	d.Set("caller_reference", publicKeyConfig.CallerReference)
	d.Set("etag", output.ETag)

	return nil
}

func resourcePublicKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	request := &cloudfront.UpdatePublicKeyInput{
		Id:              aws.String(d.Id()),
		PublicKeyConfig: expandPublicKeyConfig(d),
		IfMatch:         aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdatePublicKey(request)
	if err != nil {
		return fmt.Errorf("error updating CloudFront PublicKey (%s): %s", d.Id(), err)
	}

	return resourcePublicKeyRead(d, meta)
}

func resourcePublicKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	request := &cloudfront.DeletePublicKeyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeletePublicKey(request)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchPublicKey) {
			return nil
		}
		return err
	}

	return nil
}

func expandPublicKeyConfig(d *schema.ResourceData) *cloudfront.PublicKeyConfig {
	publicKeyConfig := &cloudfront.PublicKeyConfig{
		EncodedKey: aws.String(d.Get("encoded_key").(string)),
		Name:       aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		publicKeyConfig.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("caller_reference"); ok {
		publicKeyConfig.CallerReference = aws.String(v.(string))
	} else {
		publicKeyConfig.CallerReference = aws.String(resource.UniqueId())
	}

	return publicKeyConfig
}
