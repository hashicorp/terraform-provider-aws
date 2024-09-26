// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_ebs_default_kms_key", name="EBS Default KMS Key")
func dataSourceEBSDefaultKMSKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEBSDefaultKMSKeyRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func dataSourceEBSDefaultKMSKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	res, err := conn.GetEbsDefaultKmsKeyId(ctx, &ec2.GetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS default KMS key: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("key_arn", res.KmsKeyId)

	return diags
}
