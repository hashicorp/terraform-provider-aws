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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	vcId := d.Get("voice_connector_id").(string)
	input := &chimesdkvoice.PutVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(vcId),
		LoggingConfiguration: &awstypes.LoggingConfiguration{
			EnableMediaMetricLogs: aws.Bool(d.Get("enable_media_metric_logs").(bool)),
			EnableSIPLogs:         aws.Bool(d.Get("enable_sip_logs").(bool)),
		},
	}

	if _, err := conn.PutVoiceConnectorLoggingConfiguration(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Voice Connector (%s) logging configuration: %s", vcId, err)
	}

	d.SetId(vcId)
	return append(diags, resourceVoiceConnectorLoggingRead(ctx, d, meta)...)
}

func resourceVoiceConnectorLoggingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	resp, err := FindVoiceConnectorResourceWithRetry(ctx, d.IsNewResource(), func() (*awstypes.LoggingConfiguration, error) {
		return findVoiceConnectorLoggingByID(ctx, conn, d.Id())
	})

	if tfresource.TimedOut(err) {
		resp, err = findVoiceConnectorLoggingByID(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Voice Connector logging configuration %s not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Chime Voice Connector logging configuration (%s): %s", d.Id(), err)
	}

	d.Set("enable_media_metric_logs", resp.EnableMediaMetricLogs)
	d.Set("enable_sip_logs", resp.EnableSIPLogs)
	d.Set("voice_connector_id", d.Id())

	return diags
}

func resourceVoiceConnectorLoggingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	if d.HasChanges("enable_sip_logs", "enable_media_metric_logs") {
		input := &chimesdkvoice.PutVoiceConnectorLoggingConfigurationInput{
			VoiceConnectorId: aws.String(d.Id()),
			LoggingConfiguration: &awstypes.LoggingConfiguration{
				EnableMediaMetricLogs: aws.Bool(d.Get("enable_media_metric_logs").(bool)),
				EnableSIPLogs:         aws.Bool(d.Get("enable_sip_logs").(bool)),
			},
		}

		if _, err := conn.PutVoiceConnectorLoggingConfiguration(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Chime Voice Connector (%s) logging configuration: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVoiceConnectorLoggingRead(ctx, d, meta)...)
}

func resourceVoiceConnectorLoggingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.PutVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(d.Id()),
		LoggingConfiguration: &awstypes.LoggingConfiguration{
			EnableSIPLogs:         aws.Bool(false),
			EnableMediaMetricLogs: aws.Bool(false),
		},
	}

	_, err := conn.PutVoiceConnectorLoggingConfiguration(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Chime Voice Connector (%s) logging configuration: %s", d.Id(), err)
	}

	return diags
}

func findVoiceConnectorLoggingByID(ctx context.Context, conn *chimesdkvoice.Client, id string) (*awstypes.LoggingConfiguration, error) {
	in := &chimesdkvoice.GetVoiceConnectorLoggingConfigurationInput{
		VoiceConnectorId: aws.String(id),
	}

	resp, err := conn.GetVoiceConnectorLoggingConfiguration(ctx, in)

	if errs.IsA[*awstypes.NotFoundException](err) {
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
