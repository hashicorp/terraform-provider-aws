package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/lex/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	LexBotCreateTimeout = 1 * time.Minute
	LexBotUpdateTimeout = 1 * time.Minute
	LexBotDeleteTimeout = 5 * time.Minute
	LexBotVersionLatest = "$LATEST"
)

func ResourceBot() *schema.Resource {
	return &schema.Resource{
		Create: resourceBotCreate,
		Read:   resourceBotRead,
		Update: resourceBotUpdate,
		Delete: resourceBotDelete,

		// TODO add to other lex resources
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				if _, ok := d.GetOk("create_version"); !ok {
					d.Set("create_version", false)
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(LexBotCreateTimeout),
			Update: schema.DefaultTimeout(LexBotUpdateTimeout),
			Delete: schema.DefaultTimeout(LexBotDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"abort_statement": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexStatementResource,
			},
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
				Required: true,
			},
			"clarification_prompt": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexPromptResource,
			},
			"create_version": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"detect_sentiment": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_model_improvements": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idle_session_ttl_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntBetween(60, 86400),
			},
			"intent": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"intent_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
							),
						},
						"intent_version": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringMatch(regexp.MustCompile(`\$LATEST|[0-9]+`), ""),
							),
						},
					},
				},
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locale": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      lexmodelbuildingservice.LocaleEnUs,
				ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.Locale_Values(), false),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateLexBotName,
			},
			"nlu_intent_confidence_threshold": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.FloatBetween(0, 1),
			},
			"process_behavior": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      lexmodelbuildingservice.ProcessBehaviorSave,
				ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.ProcessBehavior_Values(), false),
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
		CustomizeDiff: updateComputedAttributesOnBotCreateVersion,
	}
}

func updateComputedAttributesOnBotCreateVersion(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	createVersion := d.Get("create_version").(bool)
	if createVersion && hasBotConfigChanges(d) {
		d.SetNewComputed("version")
	}
	return nil
}

func hasBotConfigChanges(d verify.ResourceDiffer) bool {
	for _, key := range []string{
		"description",
		"child_directed",
		"detect_sentiment",
		"enable_model_improvements",
		"idle_session_ttl_in_seconds",
		"intent",
		"locale",
		"nlu_intent_confidence_threshold",
		"abort_statement.0.response_card",
		"abort_statement.0.message",
		"clarification_prompt",
		"process_behavior",
		"voice_id",
	} {
		if d.HasChange(key) {
			return true
		}
	}
	return false
}

var validateLexBotName = validation.All(
	validation.StringLenBetween(2, 50),
	validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
)

var validateLexBotVersion = validation.All(
	validation.StringLenBetween(1, 64),
	validation.StringMatch(regexp.MustCompile(`\$LATEST|[0-9]+`), ""),
)

func resourceBotCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelBuildingConn
	name := d.Get("name").(string)

	input := &lexmodelbuildingservice.PutBotInput{
		AbortStatement:          expandLexStatement(d.Get("abort_statement")),
		ChildDirected:           aws.Bool(d.Get("child_directed").(bool)),
		CreateVersion:           aws.Bool(d.Get("create_version").(bool)),
		Description:             aws.String(d.Get("description").(string)),
		EnableModelImprovements: aws.Bool(d.Get("enable_model_improvements").(bool)),
		IdleSessionTTLInSeconds: aws.Int64(int64(d.Get("idle_session_ttl_in_seconds").(int))),
		Intents:                 expandLexIntents(d.Get("intent").(*schema.Set).List()),
		Name:                    aws.String(name),
	}

	if v, ok := d.GetOk("clarification_prompt"); ok {
		input.ClarificationPrompt = expandLexPrompt(v)
	}

	if v, ok := d.GetOk("locale"); ok {
		input.Locale = aws.String(v.(string))
	}

	if v, ok := d.GetOk("process_behavior"); ok {
		input.ProcessBehavior = aws.String(v.(string))
	}

	if v, ok := d.GetOk("voice_id"); ok {
		input.VoiceId = aws.String(v.(string))
	}

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err := conn.PutBot(input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			input.Checksum = output.Checksum
			return resource.RetryableError(fmt.Errorf("%q bot still creating, another operation is pending: %s", d.Id(), err))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) { // nosemgrep: helper-schema-TimeoutError-check-doesnt-return-output
		_, err = conn.PutBot(input)
	}

	if err != nil {
		return fmt.Errorf("error creating bot %s: %w", name, err)
	}

	d.SetId(name)

	return resourceBotRead(d, meta)
}

func resourceBotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelBuildingConn

	resp, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
		Name:           aws.String(d.Id()),
		VersionOrAlias: aws.String(LexBotVersionLatest),
	})
	if tfawserr.ErrMessageContains(err, lexmodelbuildingservice.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Bot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting intent %s: %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "lex",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("bot:%s", d.Id()),
	}
	d.Set("arn", arn.String())

	// Process behavior is not returned from the API but is used for create and update.
	// Manually write to state file to avoid un-expected diffs.
	processBehavior := lexmodelbuildingservice.ProcessBehaviorSave
	if v, ok := d.GetOk("process_behavior"); ok {
		processBehavior = v.(string)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("child_directed", resp.ChildDirected)
	d.Set("created_date", resp.CreatedDate.Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("detect_sentiment", resp.DetectSentiment)
	d.Set("enable_model_improvements", resp.EnableModelImprovements)
	d.Set("failure_reason", resp.FailureReason)
	d.Set("idle_session_ttl_in_seconds", resp.IdleSessionTTLInSeconds)
	d.Set("intent", flattenLexIntents(resp.Intents))
	d.Set("last_updated_date", resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set("locale", resp.Locale)
	d.Set("name", resp.Name)
	d.Set("nlu_intent_confidence_threshold", resp.NluIntentConfidenceThreshold)
	d.Set("process_behavior", processBehavior)
	d.Set("status", resp.Status)

	if resp.AbortStatement != nil {
		d.Set("abort_statement", flattenLexStatement(resp.AbortStatement))
	}

	if resp.ClarificationPrompt != nil {
		d.Set("clarification_prompt", flattenLexPrompt(resp.ClarificationPrompt))
	}

	version, err := getLatestLexBotVersion(conn, &lexmodelbuildingservice.GetBotVersionsInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error reading Lex Bot (%s) version: %w", d.Id(), err)
	}
	d.Set("version", version)

	if resp.VoiceId != nil {
		d.Set("voice_id", resp.VoiceId)
	}

	return nil
}

func resourceBotUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelBuildingConn

	input := &lexmodelbuildingservice.PutBotInput{
		Checksum:                     aws.String(d.Get("checksum").(string)),
		ChildDirected:                aws.Bool(d.Get("child_directed").(bool)),
		CreateVersion:                aws.Bool(d.Get("create_version").(bool)),
		Description:                  aws.String(d.Get("description").(string)),
		DetectSentiment:              aws.Bool(d.Get("detect_sentiment").(bool)),
		EnableModelImprovements:      aws.Bool(d.Get("enable_model_improvements").(bool)),
		IdleSessionTTLInSeconds:      aws.Int64(int64(d.Get("idle_session_ttl_in_seconds").(int))),
		Intents:                      expandLexIntents(d.Get("intent").(*schema.Set).List()),
		Locale:                       aws.String(d.Get("locale").(string)),
		Name:                         aws.String(d.Id()),
		NluIntentConfidenceThreshold: aws.Float64(d.Get("nlu_intent_confidence_threshold").(float64)),
		ProcessBehavior:              aws.String(d.Get("process_behavior").(string)),
	}

	if v, ok := d.GetOk("abort_statement"); ok {
		input.AbortStatement = expandLexStatement(v)
	}

	if v, ok := d.GetOk("clarification_prompt"); ok {
		input.ClarificationPrompt = expandLexPrompt(v)
	}

	if v, ok := d.GetOk("voice_id"); ok {
		input.VoiceId = aws.String(v.(string))
	}

	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := conn.PutBot(input)

		if tfawserr.ErrMessageContains(err, lexmodelbuildingservice.ErrCodeConflictException, "") {
			return resource.RetryableError(fmt.Errorf("%q: bot still updating", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBot(input)
	}

	if err != nil {
		return fmt.Errorf("error updating bot %s: %w", d.Id(), err)
	}

	return resourceBotRead(d, meta)
}

func resourceBotDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelBuildingConn

	input := &lexmodelbuildingservice.DeleteBotInput{
		Name: aws.String(d.Id()),
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteBot(input)

		if tfawserr.ErrMessageContains(err, lexmodelbuildingservice.ErrCodeConflictException, "") {
			return resource.RetryableError(fmt.Errorf("%q: there is a pending operation, bot still deleting", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteBot(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting bot %s: %w", d.Id(), err)
	}

	_, err = waiter.LexBotDeleted(conn, d.Id())

	return err
}

func getLatestLexBotVersion(conn *lexmodelbuildingservice.LexModelBuildingService, input *lexmodelbuildingservice.GetBotVersionsInput) (string, error) {
	version := LexBotVersionLatest

	for {
		page, err := conn.GetBotVersions(input)
		if err != nil {
			return "", err
		}

		// At least 1 version will always be returned.
		if len(page.Bots) == 1 {
			break
		}

		for _, bot := range page.Bots {
			if *bot.Version == LexBotVersionLatest {
				continue
			}
			if *bot.Version > version {
				version = *bot.Version
			}
		}

		if page.NextToken == nil {
			break
		}
		input.NextToken = page.NextToken
	}

	return version, nil
}

func flattenLexIntents(intents []*lexmodelbuildingservice.Intent) (flattenedIntents []map[string]interface{}) {
	for _, intent := range intents {
		flattenedIntents = append(flattenedIntents, map[string]interface{}{
			"intent_name":    aws.StringValue(intent.IntentName),
			"intent_version": aws.StringValue(intent.IntentVersion),
		})
	}

	return
}

// Expects a slice of maps representing the Lex objects.
// The value passed into this function should have been run through the expandLexSet function.
// Example: []map[intent_name: OrderFlowers intent_version: $LATEST]
func expandLexIntents(rawValues []interface{}) []*lexmodelbuildingservice.Intent {
	intents := make([]*lexmodelbuildingservice.Intent, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}

		intents = append(intents, &lexmodelbuildingservice.Intent{
			IntentName:    aws.String(value["intent_name"].(string)),
			IntentVersion: aws.String(value["intent_version"].(string)),
		})
	}

	return intents
}
