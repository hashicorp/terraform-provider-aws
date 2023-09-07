// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_networkfirewall_resource_policy")
func DataSourceFirewallResourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallResourcePolicyRead,

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func dataSourceFirewallResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn(ctx)

	resourceARN := d.Get("resource_arn").(string)
	policy, err := FindResourcePolicy(ctx, conn, resourceARN)

	if err != nil {
		return diag.Errorf("reading NetworkFirewall Resource Policy (%s): %s", resourceARN, err)
	}

	if policy == nil {
		return diag.Errorf("reading NetworkFirewall Resource Policy (%s): empty output", resourceARN)
	}

	d.SetId(resourceARN)
	d.Set("policy", policy)
	d.Set("resource_arn", resourceARN)

	return nil
}
