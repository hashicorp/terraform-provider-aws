// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_kms_secrets", name="Secrets)
func dataSourceSecrets() *schema.Resource {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.EncryptionAlgorithmSpec](),
						},
						"grant_tokens": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrKeyID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrName: {
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
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	tfList := d.Get("secret").(*schema.Set).List()
	plaintext := make(map[string]string, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		name := tfMap[names.AttrName].(string)

		// base64 decode the payload
		payload, err := itypes.Base64Decode(tfMap["payload"].(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "invalid base64 value for secret (%s): %s", name, err)
		}

		input := &kms.DecryptInput{
			CiphertextBlob: payload,
		}

		if v, ok := tfMap["context"].(map[string]interface{}); ok && len(v) > 0 {
			input.EncryptionContext = flex.ExpandStringValueMap(v)
		}

		if v, ok := tfMap["encryption_algorithm"].(string); ok && v != "" {
			input.EncryptionAlgorithm = awstypes.EncryptionAlgorithmSpec(v)
		}

		if v, ok := tfMap["grant_tokens"].([]interface{}); ok && len(v) > 0 {
			input.GrantTokens = flex.ExpandStringValueList(v)
		}

		if v, ok := tfMap[names.AttrKeyID].(string); ok && v != "" {
			input.KeyId = aws.String(v)
		}

		output, err := conn.Decrypt(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "decrypting KMS Secret (%s): %s", name, err)
		}

		// Set the secret via the name
		plaintext[name] = string(output.Plaintext)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("plaintext", plaintext)

	return diags
}
