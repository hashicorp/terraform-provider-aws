package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceVoiceConnectorLogging() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVoiceConnectorLoggingCreate,
		ReadWithoutTimeout:   resourceVoiceConnectorLoggingRead,
		UpdateWithoutTimeout: resourceVoiceConnectorLoggingUpdate,
		DeleteWithoutTimeout: resourceVoiceConnectorLoggingDelete,

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

func resourceVoiceConnectorLoggingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

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

	d.SetId(vcId)
	return resourceVoiceConnectorLoggingRead(ctx, d, meta)
}

func resourceVoiceConnectorLoggingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	input := &chime.GetVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnectorLoggingConfigurationWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
		log.Printf("[WARN] Chime Voice Connector logging configuration %s not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil || resp.LoggingConfiguration == nil {
		return diag.Errorf("error getting Chime Voice Connector (%s) logging configuration: %s", d.Id(), err)
	}

	d.Set("enable_sip_logs", resp.LoggingConfiguration.EnableSIPLogs)
	d.Set("voice_connector_id", d.Id())

	return nil
}

func resourceVoiceConnectorLoggingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	if d.HasChange("enable_sip_logs") {
		input := &chime.PutVoiceConnectorLoggingConfigurationInput{
			VoiceConnectorId: aws.String(d.Id()),
			LoggingConfiguration: &chime.LoggingConfiguration{
				EnableSIPLogs: aws.Bool(d.Get("enable_sip_logs").(bool)),
			},
		}

		if _, err := conn.PutVoiceConnectorLoggingConfigurationWithContext(ctx, input); err != nil {
			return diag.Errorf("error updating Chime Voice Connector (%s) logging configuration: %s", d.Id(), err)
		}
	}

	return resourceVoiceConnectorLoggingRead(ctx, d, meta)
}

func resourceVoiceConnectorLoggingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeConn

	input := &chime.PutVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
		LoggingConfiguration: &chime.LoggingConfiguration{
			EnableSIPLogs: aws.Bool(false),
		},
	}

	_, err := conn.PutVoiceConnectorLoggingConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Chime Voice Connector (%s) logging configuration: %s", d.Id(), err)
	}

	return nil
}
