package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func dataSourceAwsLexBot() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexBotRead,

		Schema: map[string]*schema.Schema{
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexNameMinLength, lexNameMaxLength),
					validation.StringMatch(regexp.MustCompile(lexNameRegex), ""),
				),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength),
					validation.StringMatch(regexp.MustCompile(lexVersionRegex), ""),
				),
			},
			"voice_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLexBotRead(d *schema.ResourceData, meta interface{}) error {
	botName := d.Get("name").(string)
	botVersion := "$LATEST"
	if v, ok := d.GetOk("version"); ok {
		botVersion = v.(string)
	}

	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
		Name:           aws.String(botName),
		VersionOrAlias: aws.String(botVersion),
	})
	if err != nil {
		return fmt.Errorf("error getting bot %s: %s", botName, err)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("child_directed", resp.ChildDirected)
	d.Set("created_date", resp.CreatedDate.UTC().String())
	d.Set("description", resp.Description)
	d.Set("failure_reason", resp.FailureReason)
	d.Set("idle_session_ttl_in_seconds", resp.IdleSessionTTLInSeconds)
	d.Set("last_updated_date", resp.LastUpdatedDate.UTC().String())
	d.Set("locale", resp.Locale)
	d.Set("name", resp.Name)
	d.Set("status", resp.Status)
	d.Set("version", resp.Version)
	d.Set("voice_id", resp.VoiceId)

	d.SetId(botName)

	return nil
}
