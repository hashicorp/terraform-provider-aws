// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(chimesdkvoice.NotificationTarget_Values(), false),
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	vcId := d.Get("voice_connector_id").(string)
	input := &chimesdkvoice.PutVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	config := &chimesdkvoice.StreamingConfiguration{
		DataRetentionInHours: aws.Int64(int64(d.Get("data_retention").(int))),
		Disabled:             aws.Bool(d.Get("disabled").(bool)),
	}

	if v, ok := d.GetOk("streaming_notification_targets"); ok && v.(*schema.Set).Len() > 0 {
		config.StreamingNotificationTargets = expandStreamingNotificationTargets(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("media_insights_configuration"); ok && len(v.([]interface{})) > 0 {
		config.MediaInsightsConfiguration = expandMediaInsightsConfiguration(v.([]interface{}))
	}

	input.StreamingConfiguration = config

	if _, err := conn.PutVoiceConnectorStreamingConfigurationWithContext(ctx, input); err != nil {
		return diag.Errorf("creating Chime Voice Connector (%s) streaming configuration: %s", vcId, err)
	}

	d.SetId(vcId)

	return resourceVoiceConnectorStreamingRead(ctx, d, meta)
}

func resourceVoiceConnectorStreamingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	input := &chimesdkvoice.GetVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnectorStreamingConfigurationWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
		log.Printf("[WARN] Chime Voice Connector (%s) streaming not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("getting Chime Voice Connector (%s) streaming: %s", d.Id(), err)
	}

	if resp == nil || resp.StreamingConfiguration == nil {
		return diag.Errorf("getting Chime Voice Connector (%s) streaming: empty response", d.Id())
	}

	d.Set("disabled", resp.StreamingConfiguration.Disabled)
	d.Set("data_retention", resp.StreamingConfiguration.DataRetentionInHours)
	d.Set("voice_connector_id", d.Id())

	if err := d.Set("streaming_notification_targets", flattenStreamingNotificationTargets(resp.StreamingConfiguration.StreamingNotificationTargets)); err != nil {
		return diag.Errorf("setting Chime Voice Connector streaming configuration targets (%s): %s", d.Id(), err)
	}

	if err := d.Set("media_insights_configuration", flattenMediaInsightsConfiguration(resp.StreamingConfiguration.MediaInsightsConfiguration)); err != nil {
		return diag.Errorf("setting Chime Voice Connector streaming configuration media insights configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceVoiceConnectorStreamingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	vcId := d.Get("voice_connector_id").(string)

	if d.HasChanges("data_retention", "disabled", "streaming_notification_targets", "media_insights_configuration") {
		input := &chimesdkvoice.PutVoiceConnectorStreamingConfigurationInput{
			VoiceConnectorId: aws.String(vcId),
		}

		config := &chimesdkvoice.StreamingConfiguration{
			DataRetentionInHours: aws.Int64(int64(d.Get("data_retention").(int))),
			Disabled:             aws.Bool(d.Get("disabled").(bool)),
		}

		if v, ok := d.GetOk("streaming_notification_targets"); ok && v.(*schema.Set).Len() > 0 {
			config.StreamingNotificationTargets = expandStreamingNotificationTargets(v.(*schema.Set).List())
		}

		if v, ok := d.GetOk("media_insights_configuration"); ok && len(v.([]interface{})) > 0 {
			config.MediaInsightsConfiguration = expandMediaInsightsConfiguration(v.([]interface{}))
		}

		input.StreamingConfiguration = config

		if _, err := conn.PutVoiceConnectorStreamingConfigurationWithContext(ctx, input); err != nil {
			return diag.Errorf("updating Chime Voice Connector (%s) streaming configuration: %s", d.Id(), err)
		}
	}

	return resourceVoiceConnectorStreamingRead(ctx, d, meta)
}

func resourceVoiceConnectorStreamingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	input := &chimesdkvoice.DeleteVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVoiceConnectorStreamingConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Chime Voice Connector (%s) streaming configuration: %s", d.Id(), err)
	}

	return nil
}

func expandStreamingNotificationTargets(data []interface{}) []*chimesdkvoice.StreamingNotificationTarget {
	var streamingTargets []*chimesdkvoice.StreamingNotificationTarget

	for _, item := range data {
		streamingTargets = append(streamingTargets, &chimesdkvoice.StreamingNotificationTarget{
			NotificationTarget: aws.String(item.(string)),
		})
	}

	return streamingTargets
}

func expandMediaInsightsConfiguration(tfList []interface{}) *chimesdkvoice.MediaInsightsConfiguration {
	if len(tfList) == 0 {
		return nil
	}
	tfMap := tfList[0].(map[string]interface{})

	mediaInsightsConfiguration := &chimesdkvoice.MediaInsightsConfiguration{}
	if v, ok := tfMap["disabled"]; ok {
		mediaInsightsConfiguration.Disabled = aws.Bool(v.(bool))
	}
	if v, ok := tfMap["configuration_arn"]; ok {
		mediaInsightsConfiguration.ConfigurationArn = aws.String(v.(string))
	}
	return mediaInsightsConfiguration
}

func flattenStreamingNotificationTargets(targets []*chimesdkvoice.StreamingNotificationTarget) []*string {
	var rawTargets []*string

	for _, t := range targets {
		rawTargets = append(rawTargets, t.NotificationTarget)
	}

	return rawTargets
}

func flattenMediaInsightsConfiguration(mediaInsightsConfiguration *chimesdkvoice.MediaInsightsConfiguration) []interface{} {
	if mediaInsightsConfiguration == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"disabled":          mediaInsightsConfiguration.Disabled,
		"configuration_arn": mediaInsightsConfiguration.ConfigurationArn,
	}

	return []interface{}{tfMap}
}
