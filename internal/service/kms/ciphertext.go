// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
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
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCiphertextCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	keyID := d.Get(names.AttrKeyID).(string)
	input := &kms.EncryptInput{
		KeyId:     aws.String(d.Get(names.AttrKeyID).(string)),
		Plaintext: []byte(d.Get("plaintext").(string)),
	}

	if v, ok := d.GetOk("context"); ok && len(v.(map[string]interface{})) > 0 {
		input.EncryptionContext = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	output, err := conn.Encrypt(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "encrypting with KMS Key (%s): %s", keyID, err)
	}

	//lintignore:R017 // Allow legacy unstable ID usage in managed resource
	d.SetId(time.Now().UTC().String())
	d.Set("ciphertext_blob", itypes.Base64Encode(output.CiphertextBlob))

	return diags
}
