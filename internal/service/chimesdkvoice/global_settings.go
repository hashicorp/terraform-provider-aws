// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameGlobalSettings            = "Global Settings"
	globalSettingsPropagationTimeout = 20 * time.Second
)

// @SDKResource("aws_chimesdkvoice_global_settings")
func ResourceGlobalSettings() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGlobalSettingsUpdate,
		ReadWithoutTimeout:   resourceGlobalSettingsRead,
		UpdateWithoutTimeout: resourceGlobalSettingsUpdate,
		DeleteWithoutTimeout: resourceGlobalSettingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"voice_connector": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cdr_bucket": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceGlobalSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	// Include retry handling to allow for propagation of the Global Settings
	// logging bucket configuration
	var out *chimesdkvoice.GetGlobalSettingsOutput
	err := tfresource.Retry(ctx, globalSettingsPropagationTimeout, func() *retry.RetryError {
		var getErr error
		out, getErr = conn.GetGlobalSettingsWithContext(ctx, &chimesdkvoice.GetGlobalSettingsInput{})

		if getErr != nil {
			return retry.NonRetryableError(getErr)
		}

		if out.VoiceConnector == nil || out.VoiceConnector.CdrBucket == nil {
			return retry.RetryableError(tfresource.NewEmptyResultError(&chimesdkvoice.GetGlobalSettingsInput{}))
		}

		return nil
	})

	var ere *tfresource.EmptyResultError
	if !d.IsNewResource() && errors.As(err, &ere) {
		log.Printf("[WARN] ChimeSDKVoice GlobalSettings (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.ChimeSDKVoice, create.ErrActionReading, ResNameGlobalSettings, d.Id(), err)...)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)
	d.Set("voice_connector", flattenVoiceConnectorSettings(out.VoiceConnector))

	return diags
}

func resourceGlobalSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	if d.HasChange("voice_connector") {
		input := &chimesdkvoice.UpdateGlobalSettingsInput{
			VoiceConnector: expandVoiceConnectorSettings(d.Get("voice_connector").([]interface{})),
		}

		_, err := conn.UpdateGlobalSettingsWithContext(ctx, input)
		if err != nil {
			return append(diags, create.DiagError(names.ChimeSDKVoice, create.ErrActionUpdating, ResNameGlobalSettings, d.Id(), err)...)
		}
	}
	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceGlobalSettingsRead(ctx, d, meta)...)
}

func resourceGlobalSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	_, err := conn.UpdateGlobalSettingsWithContext(ctx, &chimesdkvoice.UpdateGlobalSettingsInput{
		VoiceConnector: &chimesdkvoice.VoiceConnectorSettings{},
	})
	if err != nil {
		return append(diags, create.DiagError(names.ChimeSDKVoice, create.ErrActionDeleting, ResNameGlobalSettings, d.Id(), err)...)
	}

	return diags
}

func expandVoiceConnectorSettings(tfList []interface{}) *chimesdkvoice.VoiceConnectorSettings {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &chimesdkvoice.VoiceConnectorSettings{
		CdrBucket: aws.String(tfMap["cdr_bucket"].(string)),
	}
}

func flattenVoiceConnectorSettings(apiObject *chimesdkvoice.VoiceConnectorSettings) []interface{} {
	m := map[string]interface{}{
		"cdr_bucket": aws.StringValue(apiObject.CdrBucket),
	}
	return []interface{}{m}
}
