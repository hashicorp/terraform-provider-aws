// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ssm

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
				ValidateFunc: validation.Any(
					verify.ValidARN,
					validation.StringMatch(regexache.MustCompile(`^/ssm/`), "setting_id must begin with '/ssm/'"),
				),
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

func resourceServiceSettingUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.SSMClient(ctx)

	settingID := d.Get("setting_id").(string)
	input := ssm.UpdateServiceSettingInput{
		SettingId:    aws.String(settingID),
		SettingValue: aws.String(d.Get("setting_value").(string)),
	}

	_, err := conn.UpdateServiceSetting(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SSM Service Setting (%s): %s", settingID, err)
	}

	// While settingID can be either a full ARN or an ID with "/ssm/" prefix, id is always ARN.
	if arn.IsARN(settingID) {
		d.SetId(settingID)
	} else {
		d.SetId(c.RegionalARN(ctx, "ssm", "servicesetting"+settingID))
	}

	if _, err := waitServiceSettingUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSM Service Setting (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceServiceSettingRead(ctx, d, meta)...)
}

func resourceServiceSettingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	output, err := findServiceSettingByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SSM Service Setting %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Service Setting (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ARN)
	// setting_id begins with "/ssm/" prefix, according to the AWS documentation
	// https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_GetServiceSetting.html#API_GetServiceSetting_RequestSyntax
	// However, the full ARN format can be accepted by the AWS API as well and the first implementation of this resource assumed the full ARN format for setting_id.
	// For backwards compatibility, support both formats.
	if arn.IsARN(d.Get("setting_id").(string)) {
		d.Set("setting_id", output.ARN)
	} else {
		d.Set("setting_id", output.SettingId)
	}
	d.Set("setting_value", output.SettingValue)
	d.Set(names.AttrStatus, output.Status)

	return diags
}

func resourceServiceSettingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[DEBUG] Deleting SSM Service Setting: %s", d.Id())
	input := ssm.ResetServiceSettingInput{
		SettingId: aws.String(d.Id()),
	}
	_, err := conn.ResetServiceSetting(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Service Setting (%s): %s", d.Id(), err)
	}

	if _, err := waitServiceSettingReset(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSM Service Setting (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findServiceSettingByID(ctx context.Context, conn *ssm.Client, id string) (*awstypes.ServiceSetting, error) {
	input := ssm.GetServiceSettingInput{
		SettingId: aws.String(id),
	}

	output, err := conn.GetServiceSetting(ctx, &input)

	if errs.IsA[*awstypes.ServiceSettingNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServiceSetting == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ServiceSetting, nil
}

func statusServiceSetting(conn *ssm.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findServiceSettingByID(ctx, conn, id)

		if retry.NotFound(err) {
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
		Refresh: statusServiceSetting(conn, id),
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
		Refresh: statusServiceSetting(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ServiceSetting); ok {
		return output, err
	}

	return nil, err
}
