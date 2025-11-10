// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lex_bot", name="Bot")
func dataSourceBot() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBotRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"child_directed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"detect_sentiment": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_model_improvements": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idle_session_ttl_in_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrLastUpdatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locale": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validBotName,
			},
			"nlu_intent_confidence_threshold": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVersion: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      BotVersionLatest,
				ValidateFunc: validBotVersion,
			},
			"voice_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBotRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsClient(ctx)

	name := d.Get(names.AttrName).(string)
	version := d.Get(names.AttrVersion).(string)
	output, err := findBotVersionByName(ctx, conn, name, version)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lex Bot (%s/%s): %s", name, version, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Service:   "lex",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("bot:%s", name),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set("checksum", output.Checksum)
	d.Set("child_directed", output.ChildDirected)
	d.Set(names.AttrCreatedDate, output.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set("detect_sentiment", output.DetectSentiment)
	d.Set("enable_model_improvements", output.EnableModelImprovements)
	d.Set("failure_reason", output.FailureReason)
	d.Set("idle_session_ttl_in_seconds", output.IdleSessionTTLInSeconds)
	d.Set(names.AttrLastUpdatedDate, output.LastUpdatedDate.Format(time.RFC3339))
	d.Set("locale", output.Locale)
	d.Set(names.AttrName, output.Name)
	d.Set("nlu_intent_confidence_threshold", output.NluIntentConfidenceThreshold)
	d.Set(names.AttrStatus, output.Status)
	d.Set(names.AttrVersion, output.Version)
	d.Set("voice_id", output.VoiceId)

	d.SetId(name)

	return diags
}
