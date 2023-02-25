package cloudfront

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourcePublicKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePublicKeyCreate,
		ReadWithoutTimeout:   resourcePublicKeyRead,
		UpdateWithoutTimeout: resourcePublicKeyUpdate,
		DeleteWithoutTimeout: resourcePublicKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourcePublicKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

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

	output, err := conn.CreatePublicKeyWithContext(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront PublicKey: %s", err)
	}

	d.SetId(aws.StringValue(output.PublicKey.Id))
	return append(diags, resourcePublicKeyRead(ctx, d, meta)...)
}

func resourcePublicKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()
	request := &cloudfront.GetPublicKeyInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetPublicKeyWithContext(ctx, request)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchPublicKey) {
		create.LogNotFoundRemoveState(names.CloudFront, create.ErrActionReading, ResNamePublicKey, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CloudFront, create.ErrActionReading, ResNamePublicKey, d.Id(), err)
	}

	if !d.IsNewResource() && (output == nil || output.PublicKey == nil || output.PublicKey.PublicKeyConfig == nil) {
		create.LogNotFoundRemoveState(names.CloudFront, create.ErrActionReading, ResNamePublicKey, d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && (output == nil || output.PublicKey == nil || output.PublicKey.PublicKeyConfig == nil) {
		return create.DiagError(names.CloudFront, create.ErrActionReading, ResNamePublicKey, d.Id(), errors.New("empty response after creation"))
	}

	publicKeyConfig := output.PublicKey.PublicKeyConfig
	d.Set("encoded_key", publicKeyConfig.EncodedKey)
	d.Set("name", publicKeyConfig.Name)
	d.Set("comment", publicKeyConfig.Comment)
	d.Set("caller_reference", publicKeyConfig.CallerReference)
	d.Set("etag", output.ETag)

	return diags
}

func resourcePublicKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	request := &cloudfront.UpdatePublicKeyInput{
		Id:              aws.String(d.Id()),
		PublicKeyConfig: expandPublicKeyConfig(d),
		IfMatch:         aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdatePublicKeyWithContext(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront PublicKey (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePublicKeyRead(ctx, d, meta)...)
}

func resourcePublicKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	request := &cloudfront.DeletePublicKeyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeletePublicKeyWithContext(ctx, request)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchPublicKey) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront PublicKey (%s): %s", d.Id(), err)
	}

	return diags
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
