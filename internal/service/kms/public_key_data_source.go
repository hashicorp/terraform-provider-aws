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
)

// @SDKDataSource("aws_kms_public_key")
func DataSourcePublicKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePublicKeyRead,
		Schema: map[string]*schema.Schema{
			"arn": {
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
			"key_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: ValidateKeyOrAlias,
			},
			"key_usage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key": {
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
	keyId := d.Get("key_id").(string)

	input := &kms.GetPublicKeyInput{
		KeyId: aws.String(keyId),
	}

	if v, ok := d.GetOk("grant_tokens"); ok {
		input.GrantTokens = v.([]string)
	}

	output, err := conn.GetPublicKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while describing KMS public key (%s): %s", keyId, err)
	}

	d.SetId(aws.ToString(output.KeyId))

	d.Set("arn", output.KeyId)
	d.Set("customer_master_key_spec", output.CustomerMasterKeySpec)
	d.Set("key_usage", output.KeyUsage)
	d.Set("public_key", itypes.Base64Encode(output.PublicKey))
	d.Set("public_key_pem", string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: output.PublicKey,
	})))

	if err := d.Set("encryption_algorithms", flex.FlattenStringyValueList(output.EncryptionAlgorithms)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_algorithms: %s", err)
	}

	if err := d.Set("signing_algorithms", flex.FlattenStringyValueList(output.SigningAlgorithms)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signing_algorithms: %s", err)
	}

	return diags
}
