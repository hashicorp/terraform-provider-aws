// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_service_setting", name="Service Setting")
func resourceServiceSetting() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceSettingUpdate,
		ReadWithoutTimeout:   resourceServiceSettingRead,
		UpdateWithoutTimeout: resourceServiceSettingUpdate,
		DeleteWithoutTimeout: resourceServiceSettingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"setting_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"setting_value": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceServiceSettingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	settingID := d.Get("setting_id").(string)
	input := &ssm.UpdateServiceSettingInput{
		SettingId:    aws.String(settingID),
		SettingValue: aws.String(d.Get("setting_value").(string)),
	}

	_, err := conn.UpdateServiceSetting(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SSM Service Setting (%s): %s", settingID, err)
	}

	d.SetId(settingID)

	if _, err := waitServiceSettingUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSM Service Setting (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceServiceSettingRead(ctx, d, meta)...)
}

func resourceServiceSettingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	output, err := findServiceSettingByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Service Setting %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Service Setting (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ARN)
	// AWS SSM service setting API requires the entire ARN as input,
	// but setting_id in the output is only a part of ARN.
	d.Set("setting_id", output.ARN)
	d.Set("setting_value", output.SettingValue)
	d.Set(names.AttrStatus, output.Status)

	return diags
}

func resourceServiceSettingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[DEBUG] Deleting SSM Service Setting: %s", d.Id())
	_, err := conn.ResetServiceSetting(ctx, &ssm.ResetServiceSettingInput{
		SettingId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Service Setting (%s): %s", d.Id(), err)
	}

	if _, err := waitServiceSettingReset(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSM Service Setting (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findServiceSettingByID(ctx context.Context, conn *ssm.Client, id string) (*awstypes.ServiceSetting, error) {
	input := &ssm.GetServiceSettingInput{
		SettingId: aws.String(id),
	}

	output, err := conn.GetServiceSetting(ctx, input)

	if errs.IsA[*awstypes.ServiceSettingNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServiceSetting == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ServiceSetting, nil
}

func statusServiceSetting(ctx context.Context, conn *ssm.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findServiceSettingByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitServiceSettingUpdated(ctx context.Context, conn *ssm.Client, id string, timeout time.Duration) (*awstypes.ServiceSetting, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"PendingUpdate", ""},
		Target:  []string{"Customized", "Default"},
		Refresh: statusServiceSetting(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ServiceSetting); ok {
		return output, err
	}

	return nil, err
}

func waitServiceSettingReset(ctx context.Context, conn *ssm.Client, id string, timeout time.Duration) (*awstypes.ServiceSetting, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"Customized", "PendingUpdate", ""},
		Target:  []string{"Default"},
		Refresh: statusServiceSetting(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ServiceSetting); ok {
		return output, err
	}

	return nil, err
}
