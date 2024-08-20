// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lex_bot")
func ResourceBot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBotCreate,
		ReadWithoutTimeout:   resourceBotRead,
		UpdateWithoutTimeout: resourceBotUpdate,
		DeleteWithoutTimeout: resourceBotDelete,

		// TODO add to other lex resources
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				if _, ok := d.GetOk("create_version"); !ok {
					d.Set("create_version", false)
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"abort_statement": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     statementResource,
			},
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
				Required: true,
			},
			"clarification_prompt": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     promptResource,
			},
			"create_version": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
				MaxItems: 250,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"intent_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexache.MustCompile(`^([A-Za-z]_?)+$`), ""),
							),
						},
						"intent_version": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringMatch(regexache.MustCompile(`\$LATEST|[0-9]+`), ""),
							),
						},
					},
				},
			},
			names.AttrLastUpdatedDate: {
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
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validBotName,
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
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVersion: {
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
		d.SetNewComputed(names.AttrVersion)
	}
	return nil
}

func hasBotConfigChanges(d sdkv2.ResourceDiffer) bool {
	for _, key := range []string{
		names.AttrDescription,
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

var validBotName = validation.All(
	validation.StringLenBetween(2, 50),
	validation.StringMatch(regexache.MustCompile(`^([A-Za-z]_?)+$`), ""),
)

var validBotVersion = validation.All(
	validation.StringLenBetween(1, 64),
	validation.StringMatch(regexache.MustCompile(`\$LATEST|[0-9]+`), ""),
)

func resourceBotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &lexmodelbuildingservice.PutBotInput{
		AbortStatement:          expandStatement(d.Get("abort_statement")),
		ChildDirected:           aws.Bool(d.Get("child_directed").(bool)),
		CreateVersion:           aws.Bool(d.Get("create_version").(bool)),
		Description:             aws.String(d.Get(names.AttrDescription).(string)),
		EnableModelImprovements: aws.Bool(d.Get("enable_model_improvements").(bool)),
		IdleSessionTTLInSeconds: aws.Int64(int64(d.Get("idle_session_ttl_in_seconds").(int))),
		Intents:                 expandIntents(d.Get("intent").(*schema.Set).List()),
		Name:                    aws.String(name),
	}

	if v, ok := d.GetOk("clarification_prompt"); ok {
		input.ClarificationPrompt = expandPrompt(v)
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

	var output *lexmodelbuildingservice.PutBotOutput
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		var err error

		if output != nil {
			input.Checksum = output.Checksum
		}
		output, err = conn.PutBotWithContext(ctx, input)

		return output, err
	}, lexmodelbuildingservice.ErrCodeConflictException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lex Bot (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Name))

	if _, err := waitBotVersionCreated(ctx, conn, name, BotVersionLatest, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lex Bot (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBotRead(ctx, d, meta)...)
}

func resourceBotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	output, err := FindBotVersionByName(ctx, conn, d.Id(), BotVersionLatest)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lex Bot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lex Bot (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "lex",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("bot:%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	// Process behavior is not returned from the API but is used for create and update.
	// Manually write to state file to avoid un-expected diffs.
	processBehavior := lexmodelbuildingservice.ProcessBehaviorSave
	if v, ok := d.GetOk("process_behavior"); ok {
		processBehavior = v.(string)
	}

	d.Set("checksum", output.Checksum)
	d.Set("child_directed", output.ChildDirected)
	d.Set(names.AttrCreatedDate, output.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set("detect_sentiment", output.DetectSentiment)
	d.Set("enable_model_improvements", output.EnableModelImprovements)
	d.Set("failure_reason", output.FailureReason)
	d.Set("idle_session_ttl_in_seconds", output.IdleSessionTTLInSeconds)
	d.Set("intent", flattenIntents(output.Intents))
	d.Set(names.AttrLastUpdatedDate, output.LastUpdatedDate.Format(time.RFC3339))
	d.Set("locale", output.Locale)
	d.Set(names.AttrName, output.Name)
	d.Set("nlu_intent_confidence_threshold", output.NluIntentConfidenceThreshold)
	d.Set("process_behavior", processBehavior)
	d.Set(names.AttrStatus, output.Status)

	if output.AbortStatement != nil {
		d.Set("abort_statement", flattenStatement(output.AbortStatement))
	}

	if output.ClarificationPrompt != nil {
		d.Set("clarification_prompt", flattenPrompt(output.ClarificationPrompt))
	}

	version, err := FindLatestBotVersionByName(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lex Bot (%s) latest version: %s", d.Id(), err)
	}

	d.Set(names.AttrVersion, version)
	d.Set("voice_id", output.VoiceId)

	return diags
}

func resourceBotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	input := &lexmodelbuildingservice.PutBotInput{
		Checksum:                     aws.String(d.Get("checksum").(string)),
		ChildDirected:                aws.Bool(d.Get("child_directed").(bool)),
		CreateVersion:                aws.Bool(d.Get("create_version").(bool)),
		Description:                  aws.String(d.Get(names.AttrDescription).(string)),
		DetectSentiment:              aws.Bool(d.Get("detect_sentiment").(bool)),
		EnableModelImprovements:      aws.Bool(d.Get("enable_model_improvements").(bool)),
		IdleSessionTTLInSeconds:      aws.Int64(int64(d.Get("idle_session_ttl_in_seconds").(int))),
		Intents:                      expandIntents(d.Get("intent").(*schema.Set).List()),
		Locale:                       aws.String(d.Get("locale").(string)),
		Name:                         aws.String(d.Id()),
		NluIntentConfidenceThreshold: aws.Float64(d.Get("nlu_intent_confidence_threshold").(float64)),
		ProcessBehavior:              aws.String(d.Get("process_behavior").(string)),
	}

	if v, ok := d.GetOk("abort_statement"); ok {
		input.AbortStatement = expandStatement(v)
	}

	if v, ok := d.GetOk("clarification_prompt"); ok {
		input.ClarificationPrompt = expandPrompt(v)
	}

	if v, ok := d.GetOk("voice_id"); ok {
		input.VoiceId = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
		return conn.PutBotWithContext(ctx, input)
	}, lexmodelbuildingservice.ErrCodeConflictException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lex Bot (%s): %s", d.Id(), err)
	}

	if _, err = waitBotVersionCreated(ctx, conn, d.Id(), BotVersionLatest, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lex Bot (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceBotRead(ctx, d, meta)...)
}

func resourceBotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	input := &lexmodelbuildingservice.DeleteBotInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Lex Bot: (%s)", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteBotWithContext(ctx, input)
	}, lexmodelbuildingservice.ErrCodeConflictException)

	if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lex Bot (%s): %s", d.Id(), err)
	}

	if _, err = waitBotDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lex Bot (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func flattenIntents(intents []*lexmodelbuildingservice.Intent) (flattenedIntents []map[string]interface{}) {
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
func expandIntents(rawValues []interface{}) []*lexmodelbuildingservice.Intent {
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
