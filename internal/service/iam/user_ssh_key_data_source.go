// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_user_ssh_key", name="User SSH Key")
func dataSourceUserSSHKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserSSHKeyRead,
		Schema: map[string]*schema.Schema{
			"encoding": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.EncodingType](),
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPublicKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_public_key_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrUsername: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	encoding := d.Get("encoding").(string)
	sshPublicKeyId := d.Get("ssh_public_key_id").(string)
	username := d.Get(names.AttrUsername).(string)

	request := &iam.GetSSHPublicKeyInput{
		Encoding:       awstypes.EncodingType(encoding),
		SSHPublicKeyId: aws.String(sshPublicKeyId),
		UserName:       aws.String(username),
	}

	response, err := conn.GetSSHPublicKey(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User SSH Key: %s", err)
	}

	publicKey := response.SSHPublicKey
	publicKeyBody := publicKey.SSHPublicKeyBody
	if encoding == string(awstypes.EncodingTypeSsh) {
		publicKeyBody = aws.String(cleanSSHKey(aws.ToString(publicKeyBody)))
	}

	d.SetId(aws.ToString(publicKey.SSHPublicKeyId))
	d.Set("fingerprint", publicKey.Fingerprint)
	d.Set(names.AttrPublicKey, publicKeyBody)
	d.Set(names.AttrStatus, publicKey.Status)

	return diags
}
