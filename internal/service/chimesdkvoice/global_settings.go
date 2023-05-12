package chimesdkvoice

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
				Type:     schema.TypeSet,
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

func flattenVoiceConnectorSettings(settings *chimesdkvoice.VoiceConnectorSettings) []interface{} {
	var voiceConnectorSettings []interface{}
	r := map[string]interface{}{
		"cdr_bucket": aws.StringValue(settings.CdrBucket),
	}
	voiceConnectorSettings = append(voiceConnectorSettings, r)
	return voiceConnectorSettings
}

func expandVoiceConnectorSettings(data []interface{}) *chimesdkvoice.VoiceConnectorSettings {
	var voiceConnectorSettings *chimesdkvoice.VoiceConnectorSettings
	for _, items := range data {
		item := items.(map[string]interface{})
		voiceConnectorSettings = &chimesdkvoice.VoiceConnectorSettings{
			CdrBucket: aws.String(item["cdr_bucket"].(string)),
		}
	}
	return voiceConnectorSettings
}

func resourceGlobalSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()
	if d.HasChange("voice_connector") {
		createInput := &chimesdkvoice.UpdateGlobalSettingsInput{
			VoiceConnector: expandVoiceConnectorSettings(d.Get("voice_connector").(*schema.Set).List()),
		}

		_, err := conn.UpdateGlobalSettingsWithContext(ctx, createInput)

		if err != nil {
			return diag.Errorf("error updating the global settings : %s", err)
		}
	}
	d.SetId(meta.(*conns.AWSClient).AccountID)
	return resourceGlobalSettingsRead(ctx, d, meta)
}

func resourceGlobalSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()
	createInput := &chimesdkvoice.GetGlobalSettingsInput{}

	resp, err := conn.GetGlobalSettingsWithContext(ctx, createInput)

	if err != nil {
		return diag.Errorf("error getting the global settings : %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	d.Set("voice_connector", flattenVoiceConnectorSettings(resp.VoiceConnector))
	return nil
}

func resourceGlobalSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn()
	createInput := &chimesdkvoice.UpdateGlobalSettingsInput{
		VoiceConnector: &chimesdkvoice.VoiceConnectorSettings{},
	}

	_, err := conn.UpdateGlobalSettingsWithContext(ctx, createInput)

	if err != nil {
		return diag.Errorf("error deleting the global settings : %s", err)
	}
	d.SetId(meta.(*conns.AWSClient).AccountID)
	return resourceGlobalSettingsRead(ctx, d, meta)
}
