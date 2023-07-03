// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_s3_account_public_access_block")
func dataSourceAccountPublicAccessBlock() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAccountPublicAccessBlockRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"block_public_acls": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"block_public_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ignore_public_acls": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"restrict_public_buckets": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAccountPublicAccessBlockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	output, err := FindPublicAccessBlockByAccountID(ctx, conn, accountID)

	if err != nil {
		return diag.Errorf("reading S3 Account Public Access Block (%s): %s", accountID, err)
	}

	d.SetId(accountID)
	d.Set("block_public_acls", output.BlockPublicAcls)
	d.Set("block_public_policy", output.BlockPublicPolicy)
	d.Set("ignore_public_acls", output.IgnorePublicAcls)
	d.Set("restrict_public_buckets", output.RestrictPublicBuckets)

	return nil
}
