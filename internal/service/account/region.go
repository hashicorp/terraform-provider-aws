// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account

import (
	"context"
	"log"
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
			"region": {
				Type:     schema.TypeString,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"opt_in_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AccountClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	region := d.Get("region").(string)

	id := AccountRegionResourceID(accountID, region)

	var err error

	output, err := FindRegionOptInStatus(ctx, conn, accountID, region)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Account Region (%s): %s", d.Id(), err)
	}

	if v := d.Get("enabled").(bool); v {
		if output.RegionOptStatus != accounttypes.RegionOptStatusEnabledByDefault {
			_, err := conn.EnableRegion(ctx, &account.EnableRegionInput{
				RegionName: aws.String(region),
				AccountId:  aws.String(accountID),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "Enabling account region (%s): %s", id, err)
			}
		}
	} else {

		if output.RegionOptStatus == accounttypes.RegionOptStatusEnabledByDefault {
			return sdkdiag.AppendErrorf(diags, "cannot disable region (%s) that is enabled by default", id)
		}

		_, err = conn.DisableRegion(ctx, &account.DisableRegionInput{
			RegionName: aws.String(region),
			AccountId:  aws.String(accountID),
		})

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

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	region := d.Get("region").(string)

	output, err := FindRegionOptInStatus(ctx, conn, accountID, region)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Account Region (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Account Region (%s): %s", d.Id(), err)
	}

	if status := output.RegionOptStatus; status == accounttypes.RegionOptStatusEnabled || status == accounttypes.RegionOptStatusEnabling || status == accounttypes.RegionOptStatusEnabledByDefault {
		d.Set("enabled", true)
	} else {
		d.Set("enabled", false)
	}

	d.Set("opt_in_status", string(output.RegionOptStatus))
	d.Set("account_id", d.Get("account_id"))

	return diags
}

const AccountRegionResourceIDSeparator = ","

func AccountRegionResourceID(accountID, region string) string {
	parts := []string{accountID, region}
	id := strings.Join(parts, AccountRegionResourceIDSeparator)

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
