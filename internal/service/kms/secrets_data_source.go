// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKDataSource("aws_kms_secrets")
func DataSourceSecrets() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecretsRead,

		Schema: map[string]*schema.Schema{
			"secret": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"context": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"encryption_algorithm": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(kms.EncryptionAlgorithmSpec_Values(), false),
						},
						"grant_tokens": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"key_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"payload": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"plaintext": {
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSecretsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	secrets := d.Get("secret").(*schema.Set).List()
	plaintext := make(map[string]string, len(secrets))

	for _, v := range secrets {
		secret := v.(map[string]interface{})
		name := secret["name"].(string)

		// base64 decode the payload
		payload, err := base64.StdEncoding.DecodeString(secret["payload"].(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "invalid base64 value for secret (%s): %s", name, err)
		}

		// build the kms decrypt input
		input := &kms.DecryptInput{
			CiphertextBlob: payload,
		}

		if v, ok := secret["context"].(map[string]interface{}); ok && len(v) > 0 {
			input.EncryptionContext = flex.ExpandStringMap(v)
		}

		if v, ok := secret["encryption_algorithm"].(string); ok && v != "" {
			input.EncryptionAlgorithm = aws.String(v)
		}

		if v, ok := secret["grant_tokens"].([]interface{}); ok && len(v) > 0 {
			input.GrantTokens = flex.ExpandStringList(v)
		}

		if v, ok := secret["key_id"].(string); ok && v != "" {
			input.KeyId = aws.String(v)
		}

		// decrypt
		output, err := conn.DecryptWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "decrypting secret (%s): %s", name, err)
		}

		// Set the secret via the name
		plaintext[name] = string(output.Plaintext)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("plaintext", plaintext)

	return diags
}
