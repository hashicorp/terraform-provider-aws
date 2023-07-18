// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_iam_user_ssh_key")
func DataSourceUserSSHKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserSSHKeyRead,
		Schema: map[string]*schema.Schema{
			"encoding": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					iam.EncodingTypeSsh,
					iam.EncodingTypePem,
				}, false),
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_public_key_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	encoding := d.Get("encoding").(string)
	sshPublicKeyId := d.Get("ssh_public_key_id").(string)
	username := d.Get("username").(string)

	request := &iam.GetSSHPublicKeyInput{
		Encoding:       aws.String(encoding),
		SSHPublicKeyId: aws.String(sshPublicKeyId),
		UserName:       aws.String(username),
	}

	response, err := conn.GetSSHPublicKeyWithContext(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User SSH Key: %s", err)
	}

	publicKey := response.SSHPublicKey
	publicKeyBody := publicKey.SSHPublicKeyBody
	if encoding == iam.EncodingTypeSsh {
		publicKeyBody = aws.String(cleanSSHKey(aws.StringValue(publicKeyBody)))
	}

	d.SetId(aws.StringValue(publicKey.SSHPublicKeyId))
	d.Set("fingerprint", publicKey.Fingerprint)
	d.Set("public_key", publicKeyBody)
	d.Set("status", publicKey.Status)

	return diags
}
