package aws

import (
	"fmt"
	"log"
	"regexp"

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
				MinItems: 1,
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
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexPromptResource,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      lexDescriptionDefault,
				ValidateFunc: validation.StringLenBetween(lexDescriptionMinLength, lexDescriptionMaxLength),
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idle_session_ttl_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      lexBotIdleSessionTtlDefault,
				ValidateFunc: validation.IntBetween(lexBotIdleSessionTtlMin, lexBotIdleSessionTtlMax),
			},
			"intent": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: lexBotMinIntents,
				MaxItems: lexBotMaxIntents,
				Elem:     lexIntentResource,
			},
			"locale": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  lexmodelbuildingservice.LocaleEnUs,
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
			"process_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  lexmodelbuildingservice.ProcessBehaviorSave,
				ValidateFunc: validation.StringInSlice([]string{
					lexmodelbuildingservice.ProcessBehaviorBuild,
					lexmodelbuildingservice.ProcessBehaviorSave,
				}, false),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  lexVersionDefault,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength),
					validation.StringMatch(regexp.MustCompile(lexVersionRegex), ""),
				),
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
		AbortStatement:          expandLexStatement(expandLexObject(d.Get("abort_statement"))),
		ChildDirected:           aws.Bool(d.Get("child_directed").(bool)),
		ClarificationPrompt:     expandLexPrompt(expandLexObject(d.Get("clarification_prompt"))),
		IdleSessionTTLInSeconds: aws.Int64(int64(d.Get("idle_session_ttl_in_seconds").(int))),
		Intents:                 expandLexIntents(expandLexSet(d.Get("intent").(*schema.Set))),
		Locale:                  aws.String(d.Get("locale").(string)),
		Name:                    aws.String(name),
		ProcessBehavior:         aws.String(d.Get("process_behavior").(string)),
	}

	// optional attributes

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("voice_id"); ok {
		input.VoiceId = aws.String(v.(string))
	}

	if _, err := conn.PutBot(input); err != nil {
		return fmt.Errorf("error creating bot %s: %s", name, err)
	}

	d.SetId(name)

	return resourceAwsLexBotRead(d, meta)
}

func resourceAwsLexBotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	version := "$LATEST"
	if v, ok := d.GetOk("version"); ok {
		version = v.(string)
	}

	resp, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
		Name:           aws.String(d.Id()),
		VersionOrAlias: aws.String(version),
	})
	if err != nil {
		if isAWSErr(err, "NotFoundException", "") {
			log.Printf("[WARN] Bot (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error getting bot: %s", err)
	}

	// Process behavior is not returned from the API but is used for create and update.
	// Manually write to state file to avoid un-expected diffs.
	processBehavior := lexmodelbuildingservice.ProcessBehaviorSave
	if v, ok := d.GetOk("process_behavior"); ok {
		processBehavior = v.(string)
	}

	d.Set("abort_statement", flattenLexObject(flattenLexStatement(resp.AbortStatement)))
	d.Set("checksum", resp.Checksum)
	d.Set("child_directed", resp.ChildDirected)
	d.Set("clarification_prompt", flattenLexObject(flattenLexPrompt(resp.ClarificationPrompt)))
	d.Set("failure_reason", resp.FailureReason)
	d.Set("idle_session_ttl_in_seconds", resp.IdleSessionTTLInSeconds)
	d.Set("intent", flattenLexIntents(resp.Intents))
	d.Set("locale", resp.Locale)
	d.Set("name", resp.Name)
	d.Set("process_behavior", processBehavior)
	d.Set("status", resp.Status)
	d.Set("version", resp.Version)

	// optional attributes

	if resp.Description != nil {
		d.Set("description", resp.Description)
	}

	if resp.VoiceId != nil {
		d.Set("voice_id", resp.VoiceId)
	}

	return nil
}

func resourceAwsLexBotUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.PutBotInput{
		AbortStatement:          expandLexStatement(expandLexObject(d.Get("abort_statement"))),
		Checksum:                aws.String(d.Get("checksum").(string)),
		ChildDirected:           aws.Bool(d.Get("child_directed").(bool)),
		ClarificationPrompt:     expandLexPrompt(expandLexObject(d.Get("clarification_prompt"))),
		IdleSessionTTLInSeconds: aws.Int64(int64(d.Get("idle_session_ttl_in_seconds").(int))),
		Intents:                 expandLexIntents(expandLexSet(d.Get("intent").(*schema.Set))),
		Locale:                  aws.String(d.Get("locale").(string)),
		Name:                    aws.String(d.Id()),
		ProcessBehavior:         aws.String(d.Get("process_behavior").(string)),
	}

	// optional attributes

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("voice_id"); ok {
		input.VoiceId = aws.String(v.(string))
	}

	_, err := RetryOnAwsCodes([]string{"ConflictException"}, func() (interface{}, error) {
		return conn.PutBot(input)
	})
	if err != nil {
		return fmt.Errorf("error updating bot %s: %s", d.Id(), err)
	}

	return resourceAwsLexBotRead(d, meta)
}

func resourceAwsLexBotDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	_, err := RetryOnAwsCodes([]string{"ConflictException"}, func() (interface{}, error) {
		return conn.DeleteBot(&lexmodelbuildingservice.DeleteBotInput{
			Name: aws.String(d.Id()),
		})
	})
	if err != nil {
		return fmt.Errorf("error deleteing bot %s: %s", d.Id(), err)
	}

	return nil
}
