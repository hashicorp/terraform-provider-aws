// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpclattice_resource_policy", name="Resource Policy")
func DataSourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourcePolicyRead,

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

const (
	DSNameResourcePolicy = "Resource Policy Data Source"
)

func dataSourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	resourceArn := d.Get("resource_arn").(string)

	out, err := findResourcePolicyByID(ctx, conn, resourceArn)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameResourcePolicy, d.Id(), err)
	}

	if out == nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameResourcePolicy, d.Id(), err)
	}

	d.SetId(resourceArn)
	d.Set("policy", out.Policy)

	return nil
}
