// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_codeartifact_authorization_token", name="Authoiration Token")
func dataSourceAuthorizationToken() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAuthorizationTokenRead,

		Schema: map[string]*schema.Schema{
			"authorization_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"duration_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: validation.Any(
					validation.IntBetween(900, 43200),
					validation.IntInSlice([]int{0}),
				),
			},
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAuthorizationTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	domainName := d.Get(names.AttrDomain).(string)
	var domainOwner string
	if v, ok := d.GetOk("domain_owner"); ok {
		domainOwner = v.(string)
	} else {
		domainOwner = meta.(*conns.AWSClient).AccountID
	}
	input := &codeartifact.GetAuthorizationTokenInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
	}

	if v, ok := d.GetOkExists("duration_seconds"); ok {
		input.DurationSeconds = aws.Int64(int64(v.(int)))
	}

	output, err := conn.GetAuthorizationToken(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeArtifact Authorization Token: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", domainOwner, domainName))
	d.Set("authorization_token", output.AuthorizationToken)
	d.Set("domain_owner", domainOwner)
	d.Set("expiration", aws.ToTime(output.Expiration).Format(time.RFC3339))

	return diags
}
