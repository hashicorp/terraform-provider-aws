// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_eks_cluster_auth")
func dataSourceClusterAuth() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterAuthRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceClusterAuthRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).STSClient(ctx)

	name := d.Get(names.AttrName).(string)
	generator, err := NewGenerator(false, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	token, err := generator.GetWithSTS(ctx, name, conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Cluster (%s) Authentication Token: %s", name, err)
	}

	d.SetId(name)
	d.Set("token", token.Token)

	return diags
}
