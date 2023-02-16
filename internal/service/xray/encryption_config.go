package xray

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEncryptionConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEncryptionPutConfig,
		ReadWithoutTimeout:   resourceEncryptionConfigRead,
		UpdateWithoutTimeout: resourceEncryptionPutConfig,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					xray.EncryptionTypeKms,
					xray.EncryptionTypeNone,
				}, false),
			},
		},
	}
}

func resourceEncryptionPutConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	input := &xray.PutEncryptionConfigInput{
		Type: aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("key_id"); ok {
		input.KeyId = aws.String(v.(string))
	}

	_, err := conn.PutEncryptionConfigWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating XRay Encryption Config: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if _, err := waitEncryptionConfigAvailable(ctx, conn); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Xray Encryption Config (%s) to Available: %s", d.Id(), err)
	}

	return append(diags, resourceEncryptionConfigRead(ctx, d, meta)...)
}

func resourceEncryptionConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	config, err := conn.GetEncryptionConfigWithContext(ctx, &xray.GetEncryptionConfigInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading XRay Encryption Config: %s", err)
	}

	d.Set("key_id", config.EncryptionConfig.KeyId)
	d.Set("type", config.EncryptionConfig.Type)

	return diags
}
