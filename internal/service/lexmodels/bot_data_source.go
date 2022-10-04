package lexmodels

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceBot() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceBotRead,

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locale": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validBotName,
			},
			"nlu_intent_confidence_threshold": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
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

func dataSourceBotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelsConn

	name := d.Get("name").(string)
	version := d.Get("version").(string)
	output, err := FindBotVersionByName(conn, name, version)

	if err != nil {
		return fmt.Errorf("error reading Lex Bot (%s/%s): %w", name, version, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "lex",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("bot:%s", name),
	}
	d.Set("arn", arn.String())
	d.Set("checksum", output.Checksum)
	d.Set("child_directed", output.ChildDirected)
	d.Set("created_date", output.CreatedDate.Format(time.RFC3339))
	d.Set("description", output.Description)
	d.Set("detect_sentiment", output.DetectSentiment)
	d.Set("enable_model_improvements", output.EnableModelImprovements)
	d.Set("failure_reason", output.FailureReason)
	d.Set("idle_session_ttl_in_seconds", output.IdleSessionTTLInSeconds)
	d.Set("last_updated_date", output.LastUpdatedDate.Format(time.RFC3339))
	d.Set("locale", output.Locale)
	d.Set("name", output.Name)
	d.Set("nlu_intent_confidence_threshold", output.NluIntentConfidenceThreshold)
	d.Set("status", output.Status)
	d.Set("version", output.Version)
	d.Set("voice_id", output.VoiceId)

	d.SetId(name)

	return nil
}
