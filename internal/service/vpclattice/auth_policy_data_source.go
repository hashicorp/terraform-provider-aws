// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpclattice_auth_policy", name="Auth Policy")
func dataSourceAuthPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAuthPolicyRead,

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAuthPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	resourceID := d.Get("resource_identifier").(string)
	output, err := findAuthPolicyByID(ctx, conn, resourceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Auth Policy (%s): %s", resourceID, err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.Policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(resourceID)
	d.Set(names.AttrPolicy, policyToSet)
	d.Set("resource_identifier", resourceID)

	return diags
}
