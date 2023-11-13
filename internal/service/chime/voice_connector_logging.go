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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_chime_voice_connector_logging")
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
			"enable_media_metric_logs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	vcId := d.Get("voice_connector_id").(string)
	input := &chimesdkvoice.PutVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
		LoggingConfiguration: &chimesdkvoice.LoggingConfiguration{
			EnableMediaMetricLogs: aws.Bool(d.Get("enable_media_metric_logs").(bool)),
			EnableSIPLogs:         aws.Bool(d.Get("enable_sip_logs").(bool)),
		},
	}

	if _, err := conn.PutVoiceConnectorLoggingConfigurationWithContext(ctx, input); err != nil {
		return diag.Errorf("creating Chime Voice Connector (%s) logging configuration: %s", vcId, err)
	}

	d.SetId(vcId)
	return resourceVoiceConnectorLoggingRead(ctx, d, meta)
}

func resourceVoiceConnectorLoggingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	resp, err := FindVoiceConnectorResourceWithRetry(ctx, d.IsNewResource(), func() (*chimesdkvoice.LoggingConfiguration, error) {
		return findVoiceConnectorLoggingByID(ctx, conn, d.Id())
	})

	if tfresource.TimedOut(err) {
		resp, err = findVoiceConnectorLoggingByID(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Voice Connector logging configuration %s not found", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("enable_media_metric_logs", resp.EnableMediaMetricLogs)
	d.Set("enable_sip_logs", resp.EnableSIPLogs)
	d.Set("voice_connector_id", d.Id())

	return nil
}

func resourceVoiceConnectorLoggingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	if d.HasChanges("enable_sip_logs", "enable_media_metric_logs") {
		input := &chimesdkvoice.PutVoiceConnectorLoggingConfigurationInput{
			VoiceConnectorId: aws.String(d.Id()),
			LoggingConfiguration: &chimesdkvoice.LoggingConfiguration{
				EnableMediaMetricLogs: aws.Bool(d.Get("enable_media_metric_logs").(bool)),
				EnableSIPLogs:         aws.Bool(d.Get("enable_sip_logs").(bool)),
			},
		}

		if _, err := conn.PutVoiceConnectorLoggingConfigurationWithContext(ctx, input); err != nil {
			return diag.Errorf("updating Chime Voice Connector (%s) logging configuration: %s", d.Id(), err)
		}
	}

	return resourceVoiceConnectorLoggingRead(ctx, d, meta)
}

func resourceVoiceConnectorLoggingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	input := &chimesdkvoice.PutVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
		LoggingConfiguration: &chimesdkvoice.LoggingConfiguration{
			EnableSIPLogs:         aws.Bool(false),
			EnableMediaMetricLogs: aws.Bool(false),
		},
	}

	_, err := conn.PutVoiceConnectorLoggingConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Chime Voice Connector (%s) logging configuration: %s", d.Id(), err)
	}

	return nil
}

func findVoiceConnectorLoggingByID(ctx context.Context, conn *chimesdkvoice.ChimeSDKVoice, id string) (*chimesdkvoice.LoggingConfiguration, error) {
	in := &chimesdkvoice.GetVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(id),
	}

	resp, err := conn.GetVoiceConnectorLoggingConfigurationWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if resp == nil || resp.LoggingConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if err != nil {
		return nil, err
	}

	return resp.LoggingConfiguration, nil
}
