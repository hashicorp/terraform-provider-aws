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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ebs_encryption_by_default", name="EBS Encryption By Default")
func dataSourceEBSEncryptionByDefault() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEBSEncryptionByDefaultRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}
func dataSourceEBSEncryptionByDefaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	res, err := conn.GetEbsEncryptionByDefault(ctx, &ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading default EBS encryption toggle: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrEnabled, res.EbsEncryptionByDefault)

	return diags
}
