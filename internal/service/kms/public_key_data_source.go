// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"encoding/pem"

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

// @SDKDataSource("aws_kms_public_key", name="Public Key")
func dataSourcePublicKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePublicKeyRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			"grant_tokens": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrKeyID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateKeyOrAlias,
			},
			"key_usage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPublicKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key_pem": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_algorithms": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourcePublicKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	keyID := d.Get(names.AttrKeyID).(string)
	input := &kms.GetPublicKeyInput{
		KeyId: aws.String(keyID),
	}

	if v, ok := d.GetOk("grant_tokens"); ok && len(v.([]interface{})) > 0 {
		input.GrantTokens = flex.ExpandStringValueList(v.([]interface{}))
	}

	output, err := conn.GetPublicKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Public Key (%s): %s", keyID, err)
	}

	d.SetId(aws.ToString(output.KeyId))
	d.Set(names.AttrARN, output.KeyId)
	d.Set("customer_master_key_spec", output.CustomerMasterKeySpec)
	d.Set("encryption_algorithms", output.EncryptionAlgorithms)
	d.Set("key_usage", output.KeyUsage)
	d.Set(names.AttrPublicKey, itypes.Base64Encode(output.PublicKey))
	d.Set("public_key_pem", string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: output.PublicKey,
	})))
	d.Set("signing_algorithms", output.SigningAlgorithms)

	return diags
}
