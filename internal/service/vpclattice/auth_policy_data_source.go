// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_vpclattice_auth_policy", name="Auth Policy")
func DataSourceAuthPolicy() *schema.Resource {
	return &schema.Resource{

		ReadWithoutTimeout: dataSourceAuthPolicyRead,

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

const (
	DSNameAuthPolicy = "Auth Policy Data Source"
)

func dataSourceAuthPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	resourceID := d.Get("resource_identifier").(string)
	out, err := findAuthPolicy(ctx, conn, resourceID)

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, DSNameAuthPolicy, resourceID, err)
	}

	d.SetId(resourceID)

	d.Set(names.AttrPolicy, out.Policy)
	d.Set("resource_identifier", resourceID)

	// TIP: Setting a JSON string to avoid errorneous diffs.
	p, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), aws.ToString(out.Policy))
	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionSetting, DSNameAuthPolicy, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, DSNameAuthPolicy, d.Id(), err)
	}

	d.Set(names.AttrPolicy, p)

	return diags
}
