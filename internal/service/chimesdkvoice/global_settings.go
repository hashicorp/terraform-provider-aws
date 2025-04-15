// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
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

// @SDKResource("aws_chimesdkvoice_global_settings", name="Global Settings")
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

func resourceGlobalSettingsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	// Include retry handling to allow for propagation of the Global Settings
	// logging bucket configuration
	var out *chimesdkvoice.GetGlobalSettingsOutput
	err := tfresource.Retry(ctx, globalSettingsPropagationTimeout, func() *retry.RetryError {
		var getErr error
		input := chimesdkvoice.GetGlobalSettingsInput{}
		out, getErr = conn.GetGlobalSettings(ctx, &input)

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
		return create.AppendDiagError(diags, names.ChimeSDKVoice, create.ErrActionReading, ResNameGlobalSettings, d.Id(), err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID(ctx))
	d.Set("voice_connector", flattenVoiceConnectorSettings(out.VoiceConnector))

	return diags
}

func resourceGlobalSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	if d.HasChange("voice_connector") {
		input := &chimesdkvoice.UpdateGlobalSettingsInput{
			VoiceConnector: expandVoiceConnectorSettings(d.Get("voice_connector").([]any)),
		}

		_, err := conn.UpdateGlobalSettings(ctx, input)
		if err != nil {
			return create.AppendDiagError(diags, names.ChimeSDKVoice, create.ErrActionUpdating, ResNameGlobalSettings, d.Id(), err)
		}
	}
	d.SetId(meta.(*conns.AWSClient).AccountID(ctx))

	return append(diags, resourceGlobalSettingsRead(ctx, d, meta)...)
}

func resourceGlobalSettingsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := chimesdkvoice.UpdateGlobalSettingsInput{
		VoiceConnector: &awstypes.VoiceConnectorSettings{},
	}
	_, err := conn.UpdateGlobalSettings(ctx, &input)
	if err != nil {
		return create.AppendDiagError(diags, names.ChimeSDKVoice, create.ErrActionDeleting, ResNameGlobalSettings, d.Id(), err)
	}

	return diags
}

func expandVoiceConnectorSettings(tfList []any) *awstypes.VoiceConnectorSettings {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	return &awstypes.VoiceConnectorSettings{
		CdrBucket: aws.String(tfMap["cdr_bucket"].(string)),
	}
}

func flattenVoiceConnectorSettings(apiObject *awstypes.VoiceConnectorSettings) []any {
	m := map[string]any{
		"cdr_bucket": aws.ToString(apiObject.CdrBucket),
	}
	return []any{m}
}
