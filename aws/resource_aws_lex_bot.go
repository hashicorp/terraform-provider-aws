package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsLexBot() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLexBotCreate,
		Read:   resourceAwsLexBotRead,
		Update: resourceAwsLexBotUpdate,
		Delete: resourceAwsLexBotDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"abort_statement": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     lexStatementResource,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"child_directed": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"clarification_prompt": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     lexPromptResource,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(0, lexDescriptionMaxLength),
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idle_session_ttl_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntBetween(lexBotIdleSessionTtlMin, lexBotIdleSessionTtlMax),
			},
			"intent": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: lexBotMaxIntents,
				Elem:     lexIntentResource,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locale": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "en-US",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateLexName,
			},
			"process_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "SAVE",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"voice_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAwsLexBotCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	name := d.Get("name").(string)

	input := &lexmodelbuildingservice.PutBotInput{
		AbortStatement:          expandLexStatement(d.Get("abort_statement")),
		ChildDirected:           aws.Bool(d.Get("child_directed").(bool)),
		ClarificationPrompt:     expandLexPrompt(d.Get("clarification_prompt")),
		Description:             aws.String(d.Get("description").(string)),
		IdleSessionTTLInSeconds: aws.Int64(int64(d.Get("idle_session_ttl_in_seconds").(int))),
		Intents:                 expandLexIntents(d.Get("intent")),
		Locale:                  aws.String(d.Get("locale").(string)),
		Name:                    aws.String(name),
		ProcessBehavior:         aws.String(d.Get("process_behavior").(string)),
		VoiceId:                 aws.String(d.Get("voice_id").(string)),
	}

	_, err := conn.PutBot(input)
	if err != nil {
		return fmt.Errorf("error creating Lex bot %s: %s", name, err)
	}

	d.SetId(name)

	return resourceAwsLexBotRead(d, meta)
}

func resourceAwsLexBotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
		Name:           aws.String(d.Id()),
		VersionOrAlias: aws.String("$LATEST"),
	})
	if err != nil {
		return fmt.Errorf("error getting Lex bot: %s", err)
	}

	if resp.AbortStatement != nil {
		d.Set("abort_statement", flattenLexStatement(resp.AbortStatement))
	}

	if resp.ClarificationPrompt != nil {
		d.Set("clarification_prompt", flattenLexPrompt(resp.ClarificationPrompt))
	}

	if resp.Intents != nil {
		d.Set("intent", flattenLexIntents(resp.Intents))
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

	// Process is not returned from the API but is used for create and update.
	// Manually write to state file.
	processBehavior := d.Get("process_behavior")
	if processBehavior == "" {
		processBehavior = "SAVE"
	}
	d.Set("process_behavior", processBehavior)

	return nil
}

func resourceAwsLexBotUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	hasChanges := false

	input := &lexmodelbuildingservice.PutBotInput{
		Name:          aws.String(d.Id()),
		Checksum:      aws.String(d.Get("checksum").(string)),
		ChildDirected: aws.Bool(d.Get("child_directed").(bool)),
		Locale:        aws.String(d.Get("locale").(string)),
	}

	if d.HasChange("child_directed") {
		hasChanges = true
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
		hasChanges = true
	}

	if d.HasChange("idle_session_ttl_in_seconds") {
		input.IdleSessionTTLInSeconds = aws.Int64(int64(d.Get("idle_session_ttl_in_seconds").(int)))
		hasChanges = true
	}

	if d.HasChange("locale") {
		hasChanges = true
	}

	if d.HasChange("abort_statement") {
		input.AbortStatement = expandLexStatement(d.Get("abort_statement"))
		hasChanges = true
	}

	if d.HasChange("clarification_prompt") {
		input.ClarificationPrompt = expandLexPrompt(d.Get("clarification_prompt"))
		hasChanges = true
	}

	if d.HasChange("intent") {
		input.Intents = expandLexIntents(d.Get("intent"))
		hasChanges = true
	}

	if d.HasChange("process_behavior") {
		input.ProcessBehavior = aws.String(d.Get("process_behavior").(string))
		hasChanges = true
	}

	if d.HasChange("voice_id") {
		input.VoiceId = aws.String(d.Get("voice_id").(string))
		hasChanges = true
	}

	if hasChanges {
		_, err := conn.PutBot(input)
		if err != nil {
			return fmt.Errorf("error updating Lex bot %s: %s", d.Id(), err)
		}
	}

	return resourceAwsLexBotRead(d, meta)
}

func resourceAwsLexBotDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	_, err := conn.DeleteBot(&lexmodelbuildingservice.DeleteBotInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleteing Lex bot %s: %s", d.Id(), err)
	}

	return nil
}
