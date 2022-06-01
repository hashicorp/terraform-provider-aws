package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

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
			"streaming_notification_targets": {
				Type:     schema.TypeSet,
				MinItems: 1,
				MaxItems: 3,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(chime.NotificationTarget_Values(), false),
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
	conn := meta.(*conns.AWSClient).ChimeConn

	vcId := d.Get("voice_connector_id").(string)
	input := &chime.PutVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	config := &chime.StreamingConfiguration{
		DataRetentionInHours: aws.Int64(int64(d.Get("data_retention").(int))),
		Disabled:             aws.Bool(d.Get("disabled").(bool)),
	}

	if v, ok := d.GetOk("streaming_notification_targets"); ok && v.(*schema.Set).Len() > 0 {
		config.StreamingNotificationTargets = expandStreamingNotificationTargets(v.(*schema.Set).List())
	}

	input.StreamingConfiguration = config

	if _, err := conn.PutVoiceConnectorStreamingConfigurationWithContext(ctx, input); err != nil {
		return diag.Errorf("error creating Chime Voice Connector (%s) streaming configuration: %s", vcId, err)
	}

	d.SetId(vcId)

	return resourceVoiceConnectorStreamingRead(ctx, d, meta)
}

func resourceVoiceConnectorStreamingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	input := &chime.GetVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnectorStreamingConfigurationWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
		log.Printf("[WARN] Chime Voice Connector (%s) streaming not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) streaming: %s", d.Id(), err)
	}

	if resp == nil || resp.StreamingConfiguration == nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) streaming: empty response", d.Id())
	}

	d.Set("disabled", resp.StreamingConfiguration.Disabled)
	d.Set("data_retention", resp.StreamingConfiguration.DataRetentionInHours)
	d.Set("voice_connector_id", d.Id())

	if err := d.Set("streaming_notification_targets", flattenStreamingNotificationTargets(resp.StreamingConfiguration.StreamingNotificationTargets)); err != nil {
		return diag.Errorf("error setting Chime Voice Connector streaming configuration targets (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceVoiceConnectorStreamingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	vcId := d.Get("voice_connector_id").(string)

	if d.HasChanges("data_retention", "disabled", "streaming_notification_targets") {
		input := &chime.PutVoiceConnectorStreamingConfigurationInput{
			VoiceConnectorId: aws.String(vcId),
		}

		config := &chime.StreamingConfiguration{
			DataRetentionInHours: aws.Int64(int64(d.Get("data_retention").(int))),
			Disabled:             aws.Bool(d.Get("disabled").(bool)),
		}

		if v, ok := d.GetOk("streaming_notification_targets"); ok && v.(*schema.Set).Len() > 0 {
			config.StreamingNotificationTargets = expandStreamingNotificationTargets(v.(*schema.Set).List())
		}

		input.StreamingConfiguration = config

		if _, err := conn.PutVoiceConnectorStreamingConfigurationWithContext(ctx, input); err != nil {
			return diag.Errorf("error updating Chime Voice Connector (%s) streaming configuration: %s", d.Id(), err)
		}
	}

	return resourceVoiceConnectorStreamingRead(ctx, d, meta)
}

func resourceVoiceConnectorStreamingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	input := &chime.DeleteVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVoiceConnectorStreamingConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Chime Voice Connector (%s) streaming configuration: %s", d.Id(), err)
	}

	return nil
}

func expandStreamingNotificationTargets(data []interface{}) []*chime.StreamingNotificationTarget {
	var streamingTargets []*chime.StreamingNotificationTarget

	for _, item := range data {
		streamingTargets = append(streamingTargets, &chime.StreamingNotificationTarget{
			NotificationTarget: aws.String(item.(string)),
		})
	}

	return streamingTargets
}

func flattenStreamingNotificationTargets(targets []*chime.StreamingNotificationTarget) []*string {
	var rawTargets []*string

	for _, t := range targets {
		rawTargets = append(rawTargets, t.NotificationTarget)
	}

	return rawTargets
}
