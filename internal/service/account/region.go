// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/account"
	"github.com/aws/aws-sdk-go-v2/service/account/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_account_region")
func resourceRegion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegionUpdate,
		ReadWithoutTimeout:   resourceRegionRead,
		UpdateWithoutTimeout: resourceRegionUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"opt_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

const (
	regionResourceIDPartCount = 2
)

func resourceRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccountClient(ctx)

	var id string
	region := d.Get("region_name").(string)
	accountID := ""
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
		id = errs.Must(flex.FlattenResourceId([]string{accountID, region}, regionResourceIDPartCount, false))
	} else {
		id = region
	}

	output, err := findRegionOptStatus(ctx, conn, accountID, region)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Account Region (%s): %s", d.Id(), err)
	}

	if v := d.Get("enabled").(bool); v {
		if output.RegionOptStatus == types.RegionOptStatusDisabled {
			input := &account.EnableRegionInput{
				RegionName: aws.String(region),
			}
			if accountID != "" {
				input.AccountId = aws.String(accountID)
			}
			_, err := conn.EnableRegion(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "Enabling account region (%s): %s", id, err)
			}
		}
	} else {
		if output.RegionOptStatus == types.RegionOptStatusEnabledByDefault {
			return sdkdiag.AppendErrorf(diags, "cannot disable region (%s) that is enabled by default", id)
		}
		input := &account.DisableRegionInput{
			RegionName: aws.String(region),
		}
		if accountID != "" {
			input.AccountId = aws.String(accountID)
		}
		_, err = conn.DisableRegion(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Disabling account region (%s): %s", id, err)
		}
	}

	// TODO Wait for the region to be enabled/disabled.

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourceRegionRead(ctx, d, meta)...)
}

func resourceRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccountClient(ctx)

	var accountID, region string
	if strings.Contains(d.Id(), flex.ResourceIdSeparator) {
		parts, err := flex.ExpandResourceId(d.Id(), regionResourceIDPartCount, false)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		accountID = parts[0]
		region = parts[1]
	} else {
		region = d.Id()
	}

	output, err := FindRegionOptStatus(ctx, conn, accountID, region)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Account Region (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("enabled", output.RegionOptStatus == types.RegionOptStatusEnabled)
	d.Set("opt_status", string(output.RegionOptStatus))
	d.Set("region_name", region)

	return diags
}

func findRegionOptStatus(ctx context.Context, conn *account.Client, accountID, region string) (*account.GetRegionOptStatusOutput, error) {
	input := &account.GetRegionOptStatusInput{
		RegionName: aws.String(region),
	}
	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	output, err := conn.GetRegionOptStatus(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
