// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kms_ciphertext", name="Ciphertext")
func resourceCiphertext() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCiphertextCreate,
		ReadWithoutTimeout:   schema.NoopContext,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"ciphertext_blob": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"context": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrKeyID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"plaintext": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Sensitive:    true,
				ExactlyOneOf: []string{"plaintext", "plaintext_wo"},
			},
			"plaintext_wo": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				WriteOnly:    true,
				ExactlyOneOf: []string{"plaintext", "plaintext_wo"},
				RequiredWith: []string{"plaintext_wo_version"},
			},
			"plaintext_wo_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"plaintext_wo"},
			},
		},
	}
}

func resourceCiphertextCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	plaintextWO, di := flex.GetWriteOnlyStringValue(d, cty.GetAttrPath("plaintext_wo"))
	diags = append(diags, di...)
	if diags.HasError() {
		return diags
	}

	plaintext := d.Get("plaintext").(string)
	if plaintextWO != "" {
		plaintext = plaintextWO
	}

	keyID := d.Get(names.AttrKeyID).(string)
	input := &kms.EncryptInput{
		KeyId:     aws.String(d.Get(names.AttrKeyID).(string)),
		Plaintext: []byte(plaintext),
	}

	if v, ok := d.GetOk("context"); ok && len(v.(map[string]any)) > 0 {
		input.EncryptionContext = flex.ExpandStringValueMap(v.(map[string]any))
	}

	output, err := conn.Encrypt(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "encrypting with KMS Key (%s): %s", keyID, err)
	}

	//lintignore:R017 // Allow legacy unstable ID usage in managed resource
	d.SetId(time.Now().UTC().String())
	d.Set("ciphertext_blob", inttypes.Base64Encode(output.CiphertextBlob))

	return diags
}
