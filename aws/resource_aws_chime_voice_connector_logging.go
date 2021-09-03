package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsChimeVoiceConnectorLogging() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsChimeVoiceConnectorLoggingPut,
		ReadContext:   resourceAwsChimeVoiceConnectorLoggingRead,
		UpdateContext: resourceAwsChimeVoiceConnectorLoggingUpdate,
		DeleteContext: resourceAwsChimeVoiceConnectorLoggingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enable_sip_logs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"voice_connector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsChimeVoiceConnectorLoggingPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	input := &chime.PutVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
		LoggingConfiguration: &chime.LoggingConfiguration{
			EnableSIPLogs: aws.Bool(d.Get("enable_sip_logs").(bool)),
		},
	}

	if _, err := conn.PutVoiceConnectorLoggingConfigurationWithContext(ctx, input); err != nil {
		return diag.Errorf("error creating Chime Voice Connector (%s) logging configuration: %s", vcId, err)
	}

	d.SetId(fmt.Sprintf("logging-%s", vcId))
	return resourceAwsChimeVoiceConnectorLoggingRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorLoggingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	input := &chime.GetVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
	}

	resp, err := conn.GetVoiceConnectorLoggingConfigurationWithContext(ctx, input)
	if !d.IsNewResource() && isAWSErr(err, chime.ErrCodeNotFoundException, "") {
		log.Printf("[WARN]Chime Voice Connector logging configuration %s not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil || resp.LoggingConfiguration == nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) logging configuration: %s", vcId, err)
	}

	d.Set("enable_sip_logs", resp.LoggingConfiguration.EnableSIPLogs)
	return nil
}

func resourceAwsChimeVoiceConnectorLoggingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	if d.HasChange("enable_sip_logs") {
		input := &chime.PutVoiceConnectorLoggingConfigurationInput{
			VoiceConnectorId: aws.String(vcId),
			LoggingConfiguration: &chime.LoggingConfiguration{
				EnableSIPLogs: aws.Bool(d.Get("enable_sip_logs").(bool)),
			},
		}

		if _, err := conn.PutVoiceConnectorLoggingConfigurationWithContext(ctx, input); err != nil {
			if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
				log.Printf("[WARN]Chime Voice Connector logging configuration %s not found", d.Id())
				d.SetId("")
				return nil
			}
			return diag.Errorf("error updating Chime Voice Connector (%s) logging configuration: %s", vcId, err)
		}
	}

	return resourceAwsChimeVoiceConnectorLoggingRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorLoggingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	vcId := d.Get("voice_connector_id").(string)
	input := &chime.PutVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
		LoggingConfiguration: &chime.LoggingConfiguration{
			EnableSIPLogs: aws.Bool(false),
		},
	}

	if _, err := conn.PutVoiceConnectorLoggingConfigurationWithContext(ctx, input); err != nil {
		return diag.Errorf("error deleting Chime Voice Connector (%s) logging configuration: %s", vcId, err)
	}

	return nil
}
