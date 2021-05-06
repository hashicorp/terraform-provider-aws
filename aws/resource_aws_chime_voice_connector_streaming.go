package aws

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsChimeVoiceConnectorStreaming() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsChimeVoiceConnectorStreamingPut,
		ReadContext:   resourceAwsChimeVoiceConnectorStreamingRead,
		UpdateContext: resourceAwsChimeVoiceConnectorStreamingUpdate,
		DeleteContext: resourceAwsChimeVoiceConnectorStreamingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"voice_connector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"data_retention": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"streaming_notification_targets": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 3,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(chime.NotificationTarget_Values(), false),
				},
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceAwsChimeVoiceConnectorStreamingPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	input := &chime.PutVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	config := &chime.StreamingConfiguration{
		DataRetentionInHours: aws.Int64(int64(d.Get("data_retention").(int))),
		Disabled:             aws.Bool(d.Get("disabled").(bool)),
	}

	if v, ok := d.GetOk("streaming_notification_targets"); ok && len(v.([]interface{})) > 0 {
		config.StreamingNotificationTargets = expandStreamingNotificationTargets(v.([]interface{}))
	}

	input.StreamingConfiguration = config

	if _, err := conn.PutVoiceConnectorStreamingConfigurationWithContext(ctx, input); err != nil {
		return diag.Errorf("error creating voice connector streaming configuration (%s): %s, %v+", vcId, err, input)
	}

	d.SetId(resource.UniqueId())

	return resourceAwsChimeVoiceConnectorStreamingRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorStreamingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	input := &chime.GetVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	resp, err := conn.GetVoiceConnectorStreamingConfigurationWithContext(ctx, input)
	if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] error getting voice connector streaming configuration")
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error getting streaming configuration (%s): %s", vcId, err)
	}

	d.Set("disabled", resp.StreamingConfiguration.Disabled)
	d.Set("data_retention", resp.StreamingConfiguration.DataRetentionInHours)

	if err := d.Set("streaming_notification_targets", flattenStreamingNotificationTargets(resp.StreamingConfiguration.StreamingNotificationTargets)); err != nil {
		return diag.Errorf("error setting streaming configuration targets (%s): %s", vcId, err)
	}

	return nil
}

func resourceAwsChimeVoiceConnectorStreamingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	if d.HasChanges("data_retention", "disabled", "streaming_notification_targets") {
		input := &chime.PutVoiceConnectorStreamingConfigurationInput{
			VoiceConnectorId: aws.String(vcId),
		}

		config := &chime.StreamingConfiguration{
			DataRetentionInHours: aws.Int64(int64(d.Get("data_retention").(int))),
			Disabled:             aws.Bool(d.Get("disabled").(bool)),
		}

		if v, ok := d.GetOk("streaming_notification_targets"); ok {
			config.StreamingNotificationTargets = expandStreamingNotificationTargets(v.([]interface{}))
		}

		input.StreamingConfiguration = config

		if _, err := conn.PutVoiceConnectorStreamingConfigurationWithContext(ctx, input); err != nil {
			if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
				log.Printf("[WARN] error getting voice connector streaming configuration")
				d.SetId("")
				return nil
			}
			return diag.Errorf("error updating voice connector streaming configuration: (%s), %s", vcId, err)
		}
	}

	return resourceAwsChimeVoiceConnectorStreamingRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorStreamingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	input := &chime.DeleteVoiceConnectorStreamingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	if _, err := conn.DeleteVoiceConnectorStreamingConfigurationWithContext(ctx, input); err != nil {
		return diag.Errorf("error deleting voice connector (%s) streaming configuration (%s): %s", vcId, d.Id(), err)
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

func flattenStreamingNotificationTargets(targets []*chime.StreamingNotificationTarget) []string {
	var rawTargets []string

	for _, t := range targets {
		rawTargets = append(rawTargets, t.String())
	}

	return rawTargets
}
