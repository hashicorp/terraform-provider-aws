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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_account_region", name="Region")
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
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Required: true,
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
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
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
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
		id = errs.Must(flex.FlattenResourceId([]string{accountID, region}, regionResourceIDPartCount, false))
	} else {
		id = region
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	if !d.IsNewResource() {
		timeout = d.Timeout(schema.TimeoutUpdate)
	}

	if v := d.Get(names.AttrEnabled).(bool); v {
		input := &account.EnableRegionInput{
			RegionName: aws.String(region),
		}
		if accountID != "" {
			input.AccountId = aws.String(accountID)
		}

		_, err := conn.EnableRegion(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Account Region (%s): %s", id, err)
		}

		if _, err := waitRegionEnabled(ctx, conn, accountID, region, timeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Account Region (%s) enable: %s", d.Id(), err)
		}
	} else {
		input := &account.DisableRegionInput{
			RegionName: aws.String(region),
		}
		if accountID != "" {
			input.AccountId = aws.String(accountID)
		}

		_, err := conn.DisableRegion(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Account Region (%s): %s", id, err)
		}

		if _, err := waitRegionDisabled(ctx, conn, accountID, region, timeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Account Region (%s) disable: %s", d.Id(), err)
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

	output, err := findRegionOptStatus(ctx, conn, accountID, region)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Account Region (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrEnabled, output.RegionOptStatus == types.RegionOptStatusEnabled || output.RegionOptStatus == types.RegionOptStatusEnabledByDefault)
	d.Set("opt_status", string(output.RegionOptStatus))
	d.Set("region_name", output.RegionName)

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

func statusRegionOptStatus(ctx context.Context, conn *account.Client, accountID, region string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRegionOptStatus(ctx, conn, accountID, region)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.RegionOptStatus), nil
	}
}

func waitRegionEnabled(ctx context.Context, conn *account.Client, accountID, region string, timeout time.Duration) (*account.GetRegionOptStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(types.RegionOptStatusEnabling),
		Target:       enum.Slice(types.RegionOptStatusEnabled),
		Refresh:      statusRegionOptStatus(ctx, conn, accountID, region),
		Timeout:      timeout,
		Delay:        1 * time.Minute,
		PollInterval: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*account.GetRegionOptStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRegionDisabled(ctx context.Context, conn *account.Client, accountID, region string, timeout time.Duration) (*account.GetRegionOptStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(types.RegionOptStatusDisabling),
		Target:       enum.Slice(types.RegionOptStatusDisabled),
		Refresh:      statusRegionOptStatus(ctx, conn, accountID, region),
		Timeout:      timeout,
		Delay:        1 * time.Minute,
		PollInterval: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*account.GetRegionOptStatusOutput); ok {
		return output, err
	}

	return nil, err
}
