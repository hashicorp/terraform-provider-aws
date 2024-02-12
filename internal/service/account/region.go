// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/account"
	accounttypes "github.com/aws/aws-sdk-go-v2/service/account/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
			"region_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
		},
	}
}

func resourceRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AccountClient(ctx)

	accountID := ""
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	region := d.Get("region_name").(string)

	id := RegionResourceID(accountID, region)

	output, err := FindRegionOptInStatus(ctx, conn, accountID, region)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Account Region (%s): %s", d.Id(), err)
	}

	if v := d.Get("enabled").(bool); v {
		if output.RegionOptStatus != accounttypes.RegionOptStatusEnabledByDefault {
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

		if output.RegionOptStatus == accounttypes.RegionOptStatusEnabledByDefault {
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

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourceRegionRead(ctx, d, meta)...)
}

func resourceRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AccountClient(ctx)

	accountID := ""
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	region := d.Get("region_name").(string)

	output, err := FindRegionOptInStatus(ctx, conn, accountID, region)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Account Region (%s): %s", d.Id(), err)
	}

	if status := output.RegionOptStatus; status == accounttypes.RegionOptStatusEnabled || status == accounttypes.RegionOptStatusEnabling || status == accounttypes.RegionOptStatusEnabledByDefault {
		d.Set("enabled", true)
	} else {
		d.Set("enabled", false)
	}

	d.Set("opt_status", string(output.RegionOptStatus))
	d.Set("account_id", d.Get("account_id"))

	return diags
}

const RegionResourceIDSeparator = ","

func RegionResourceID(accountID string, region string) string {
	parts := []string{accountID, region}
	id := strings.Join(parts, RegionResourceIDSeparator)

	return id
}

func FindRegionOptInStatus(ctx context.Context, conn *account.Client, accountID, region string) (*account.GetRegionOptStatusOutput, error) {
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
