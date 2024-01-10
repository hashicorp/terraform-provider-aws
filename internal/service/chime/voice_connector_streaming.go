// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_chime_voice_connector_streaming")
func ResourceVoiceConnectorStreaming() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVoiceConnectorStreamingCreate,
		ReadWithoutTimeout:   resourceVoiceConnectorStreamingRead,
		UpdateWithoutTimeout: resourceVoiceConnectorStreamingUpdate,
		DeleteWithoutTimeout: resourceVoiceConnectorStreamingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"data_retention": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"media_insights_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configuration_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"disabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"streaming_notification_targets": {
				Type:     schema.TypeSet,
				MinItems: 1,
				MaxItems: 3,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.NotificationTarget](),
				},
			},
			"voice_connector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVoiceConnectorStreamingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	vcId := d.Get("voice_connector_id").(string)
	input := &chimesdkvoice.PutVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	config := &awstypes.StreamingConfiguration{
		DataRetentionInHours: aws.Int32(int32(d.Get("data_retention").(int))),
		Disabled:             aws.Bool(d.Get("disabled").(bool)),
	}

	if v, ok := d.GetOk("streaming_notification_targets"); ok && v.(*schema.Set).Len() > 0 {
		config.StreamingNotificationTargets = expandStreamingNotificationTargets(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("media_insights_configuration"); ok && len(v.([]interface{})) > 0 {
		config.MediaInsightsConfiguration = expandMediaInsightsConfiguration(v.([]interface{}))
	}

	input.StreamingConfiguration = config

	if _, err := conn.PutVoiceConnectorStreamingConfiguration(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Voice Connector (%s) streaming configuration: %s", vcId, err)
	}

	d.SetId(vcId)

	return append(diags, resourceVoiceConnectorStreamingRead(ctx, d, meta)...)
}

func resourceVoiceConnectorStreamingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.GetVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnectorStreamingConfiguration(ctx, input)
	if !d.IsNewResource() && errs.IsA[*awstypes.NotFoundException](err) {
		log.Printf("[WARN] Chime Voice Connector (%s) streaming not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Chime Voice Connector (%s) streaming: %s", d.Id(), err)
	}

	if resp == nil || resp.StreamingConfiguration == nil {
		return sdkdiag.AppendErrorf(diags, "getting Chime Voice Connector (%s) streaming: empty response", d.Id())
	}

	d.Set("disabled", resp.StreamingConfiguration.Disabled)
	d.Set("data_retention", resp.StreamingConfiguration.DataRetentionInHours)
	d.Set("voice_connector_id", d.Id())

	if err := d.Set("streaming_notification_targets", flattenStreamingNotificationTargets(resp.StreamingConfiguration.StreamingNotificationTargets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Chime Voice Connector streaming configuration targets (%s): %s", d.Id(), err)
	}

	if err := d.Set("media_insights_configuration", flattenMediaInsightsConfiguration(resp.StreamingConfiguration.MediaInsightsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Chime Voice Connector streaming configuration media insights configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceVoiceConnectorStreamingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	vcId := d.Get("voice_connector_id").(string)

	if d.HasChanges("data_retention", "disabled", "streaming_notification_targets", "media_insights_configuration") {
		input := &chimesdkvoice.PutVoiceConnectorStreamingConfigurationInput{
			VoiceConnectorId: aws.String(vcId),
		}

		config := &awstypes.StreamingConfiguration{
			DataRetentionInHours: aws.Int32(int32(d.Get("data_retention").(int))),
			Disabled:             aws.Bool(d.Get("disabled").(bool)),
		}

		if v, ok := d.GetOk("streaming_notification_targets"); ok && v.(*schema.Set).Len() > 0 {
			config.StreamingNotificationTargets = expandStreamingNotificationTargets(v.(*schema.Set).List())
		}

		if v, ok := d.GetOk("media_insights_configuration"); ok && len(v.([]interface{})) > 0 {
			config.MediaInsightsConfiguration = expandMediaInsightsConfiguration(v.([]interface{}))
		}

		input.StreamingConfiguration = config

		if _, err := conn.PutVoiceConnectorStreamingConfiguration(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Chime Voice Connector (%s) streaming configuration: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVoiceConnectorStreamingRead(ctx, d, meta)...)
}

func resourceVoiceConnectorStreamingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.DeleteVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVoiceConnectorStreamingConfiguration(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Chime Voice Connector (%s) streaming configuration: %s", d.Id(), err)
	}

	return diags
}

func expandStreamingNotificationTargets(data []interface{}) []awstypes.StreamingNotificationTarget {
	var streamingTargets []awstypes.StreamingNotificationTarget

	for _, item := range data {
		streamingTargets = append(streamingTargets, awstypes.StreamingNotificationTarget{
			NotificationTarget: awstypes.NotificationTarget(item.(string)),
		})
	}

	return streamingTargets
}

func expandMediaInsightsConfiguration(tfList []interface{}) *awstypes.MediaInsightsConfiguration {
	if len(tfList) == 0 {
		return nil
	}
	tfMap := tfList[0].(map[string]interface{})

	mediaInsightsConfiguration := &awstypes.MediaInsightsConfiguration{}
	if v, ok := tfMap["disabled"]; ok {
		mediaInsightsConfiguration.Disabled = aws.Bool(v.(bool))
	}
	if v, ok := tfMap["configuration_arn"]; ok {
		mediaInsightsConfiguration.ConfigurationArn = aws.String(v.(string))
	}
	return mediaInsightsConfiguration
}

func flattenStreamingNotificationTargets(targets []awstypes.StreamingNotificationTarget) []*string {
	var rawTargets []*string

	for _, t := range targets {
		rawTargets = append(rawTargets, aws.String(string(t.NotificationTarget)))
	}

	return rawTargets
}

func flattenMediaInsightsConfiguration(mediaInsightsConfiguration *awstypes.MediaInsightsConfiguration) []interface{} {
	if mediaInsightsConfiguration == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"disabled":          mediaInsightsConfiguration.Disabled,
		"configuration_arn": mediaInsightsConfiguration.ConfigurationArn,
	}

	return []interface{}{tfMap}
}
