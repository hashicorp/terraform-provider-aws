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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	intentCreateTimeout = 1 * time.Minute
	intentUpdateTimeout = 1 * time.Minute
	intentDeleteTimeout = 5 * time.Minute
)

// @SDKResource("aws_lex_intent")
func ResourceIntent() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIntentCreate,
		ReadWithoutTimeout:   resourceIntentRead,
		UpdateWithoutTimeout: resourceIntentUpdate,
		DeleteWithoutTimeout: resourceIntentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(intentCreateTimeout),
			Update: schema.DefaultTimeout(intentUpdateTimeout),
			Delete: schema.DefaultTimeout(intentDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"conclusion_statement": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				MaxItems:      1,
				ConflictsWith: []string{"follow_up_prompt"},
				Elem:          statementResource,
			},
			"confirmation_prompt": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				MaxItems:     1,
				RequiredWith: []string{"rejection_statement"},
				Elem:         promptResource,
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
			"dialog_code_hook": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     codeHookResource,
			},
			"follow_up_prompt": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				MaxItems:      1,
				ConflictsWith: []string{"conclusion_statement"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prompt": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     promptResource,
						},
						"rejection_statement": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     statementResource,
						},
					},
				},
			},
			"fulfillment_activity": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code_hook": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     codeHookResource,
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.FulfillmentActivityType_Values(), false),
						},
					},
				},
			},
			names.AttrLastUpdatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`^([A-Za-z]_?)+$`), ""),
				),
			},
			"parent_intent_signature": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rejection_statement": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				MaxItems:     1,
				RequiredWith: []string{"confirmation_prompt"},
				Elem:         statementResource,
			},
			"sample_utterances": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				MaxItems: 1500,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 200),
				},
			},
			"slot": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDescription: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "",
							ValidateFunc: validation.StringLenBetween(0, 200),
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexache.MustCompile(`^([A-Za-z]_?)+$`), ""),
							),
						},
						names.AttrPriority: {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 100),
						},
						"response_card": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 50000),
						},
						"sample_utterances": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 10,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 200),
							},
						},
						"slot_constraint": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.SlotConstraint_Values(), false),
						},
						"slot_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexache.MustCompile(`^((AMAZON\.)_?|[A-Za-z]_?)+`), ""),
							),
						},
						"slot_type_version": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringMatch(regexache.MustCompile(`\$LATEST|[0-9]+`), ""),
							),
						},
						"value_elicitation_prompt": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     promptResource,
						},
					},
				},
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: updateComputedAttributesOnIntentCreateVersion,
	}
}

func updateComputedAttributesOnIntentCreateVersion(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	createVersion := d.Get("create_version").(bool)
	if createVersion && hasIntentConfigChanges(d) {
		d.SetNewComputed(names.AttrVersion)
	}
	return nil
}

func hasIntentConfigChanges(d sdkv2.ResourceDiffer) bool {
	for _, key := range []string{
		names.AttrDescription,
		"conclusion_statement",
		"confirmation_prompt",
		"dialog_code_hook",
		"follow_up_prompt",
		"fulfillment_activity",
		"parent_intent_signature",
		"rejection_statement",
		"sample_utterances",
		"slot",
	} {
		if d.HasChange(key) {
			return true
		}
	}
	return false
}

func resourceIntentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)
	name := d.Get(names.AttrName).(string)

	input := &lexmodelbuildingservice.PutIntentInput{
		CreateVersion: aws.Bool(d.Get("create_version").(bool)),
		Description:   aws.String(d.Get(names.AttrDescription).(string)),
		Name:          aws.String(name),
	}

	if v, ok := d.GetOk("conclusion_statement"); ok {
		input.ConclusionStatement = expandStatement(v)
	}

	if v, ok := d.GetOk("confirmation_prompt"); ok {
		input.ConfirmationPrompt = expandPrompt(v)
	}

	if v, ok := d.GetOk("dialog_code_hook"); ok {
		input.DialogCodeHook = expandCodeHook(v)
	}

	if v, ok := d.GetOk("follow_up_prompt"); ok {
		input.FollowUpPrompt = expandFollowUpPrompt(v)
	}

	if v, ok := d.GetOk("fulfillment_activity"); ok {
		input.FulfillmentActivity = expandFulfilmentActivity(v)
	}

	if v, ok := d.GetOk("parent_intent_signature"); ok {
		input.ParentIntentSignature = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rejection_statement"); ok {
		input.RejectionStatement = expandStatement(v)
	}

	if v, ok := d.GetOk("sample_utterances"); ok {
		input.SampleUtterances = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("slot"); ok {
		input.Slots = expandSlots(v.(*schema.Set).List())
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		output, err := conn.PutIntentWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			input.Checksum = output.Checksum
			return retry.RetryableError(fmt.Errorf("%q intent still creating, another operation is pending: %s", d.Id(), err))
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) { // nosemgrep:ci.helper-schema-TimeoutError-check-doesnt-return-output
		_, err = conn.PutIntentWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating intent %s: %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceIntentRead(ctx, d, meta)...)
}

func resourceIntentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	resp, err := conn.GetIntentWithContext(ctx, &lexmodelbuildingservice.GetIntentInput{
		Name:    aws.String(d.Id()),
		Version: aws.String(IntentVersionLatest),
	})
	if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
		log.Printf("[WARN] Intent (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting intent %s: %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "lex",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("intent:%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	d.Set("checksum", resp.Checksum)
	d.Set(names.AttrCreatedDate, resp.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, resp.Description)
	d.Set(names.AttrLastUpdatedDate, resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set(names.AttrName, resp.Name)

	version, err := FindLatestIntentVersionByName(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lex Intent (%s) latest version: %s", d.Id(), err)
	}

	d.Set(names.AttrVersion, version)

	if resp.ConclusionStatement != nil {
		d.Set("conclusion_statement", flattenStatement(resp.ConclusionStatement))
	}

	if resp.ConfirmationPrompt != nil {
		d.Set("confirmation_prompt", flattenPrompt(resp.ConfirmationPrompt))
	}

	if resp.DialogCodeHook != nil {
		d.Set("dialog_code_hook", flattenCodeHook(resp.DialogCodeHook))
	}

	if resp.FollowUpPrompt != nil {
		d.Set("follow_up_prompt", flattenFollowUpPrompt(resp.FollowUpPrompt))
	}

	if resp.FulfillmentActivity != nil {
		d.Set("fulfillment_activity", flattenFulfilmentActivity(resp.FulfillmentActivity))
	}

	d.Set("parent_intent_signature", resp.ParentIntentSignature)

	if resp.RejectionStatement != nil {
		d.Set("rejection_statement", flattenStatement(resp.RejectionStatement))
	}

	d.Set("sample_utterances", resp.SampleUtterances)

	if resp.Slots != nil {
		d.Set("slot", flattenSlots(resp.Slots))
	}

	return diags
}

func resourceIntentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	input := &lexmodelbuildingservice.PutIntentInput{
		Checksum:      aws.String(d.Get("checksum").(string)),
		CreateVersion: aws.Bool(d.Get("create_version").(bool)),
		Description:   aws.String(d.Get(names.AttrDescription).(string)),
		Name:          aws.String(d.Id()),
	}

	if v, ok := d.GetOk("conclusion_statement"); ok {
		input.ConclusionStatement = expandStatement(v)
	}

	if v, ok := d.GetOk("confirmation_prompt"); ok {
		input.ConfirmationPrompt = expandPrompt(v)
	}

	if v, ok := d.GetOk("dialog_code_hook"); ok {
		input.DialogCodeHook = expandCodeHook(v)
	}

	if v, ok := d.GetOk("follow_up_prompt"); ok {
		input.FollowUpPrompt = expandFollowUpPrompt(v)
	}

	if v, ok := d.GetOk("fulfillment_activity"); ok {
		input.FulfillmentActivity = expandFulfilmentActivity(v)
	}

	if v, ok := d.GetOk("parent_intent_signature"); ok {
		input.ParentIntentSignature = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rejection_statement"); ok {
		input.RejectionStatement = expandStatement(v)
	}

	if v, ok := d.GetOk("sample_utterances"); ok {
		input.SampleUtterances = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("slot"); ok {
		input.Slots = expandSlots(v.(*schema.Set).List())
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
		_, err := conn.PutIntentWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return retry.RetryableError(fmt.Errorf("%q: intent still updating", d.Id()))
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutIntentWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating intent %s: %s", d.Id(), err)
	}

	return append(diags, resourceIntentRead(ctx, d, meta)...)
}

func resourceIntentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	input := &lexmodelbuildingservice.DeleteIntentInput{
		Name: aws.String(d.Id()),
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := conn.DeleteIntentWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return retry.RetryableError(fmt.Errorf("%q: there is a pending operation, intent still deleting", d.Id()))
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteIntentWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lex Model Intent (%s): %s", d.Id(), err)
	}

	if _, err := waitIntentDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lex Model Intent (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

var codeHookResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"message_version": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 5),
		},
		names.AttrURI: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: verify.ValidARN,
		},
	},
}

var messageResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		names.AttrContent: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 1000),
		},
		names.AttrContentType: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.ContentType_Values(), false),
		},
		"group_number": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(1, 5),
		},
	},
}

var promptResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"max_attempts": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(1, 5),
		},
		names.AttrMessage: {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			MaxItems: 15,
			Elem:     messageResource,
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 50000),
		},
	},
}

var statementResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		names.AttrMessage: {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			MaxItems: 15,
			Elem:     messageResource,
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 50000),
		},
	},
}

func flattenCodeHook(hook *lexmodelbuildingservice.CodeHook) (flattened []map[string]interface{}) {
	return []map[string]interface{}{
		{
			"message_version": aws.StringValue(hook.MessageVersion),
			names.AttrURI:     aws.StringValue(hook.Uri),
		},
	}
}

func expandCodeHook(rawObject interface{}) (hook *lexmodelbuildingservice.CodeHook) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	return &lexmodelbuildingservice.CodeHook{
		MessageVersion: aws.String(m["message_version"].(string)),
		Uri:            aws.String(m[names.AttrURI].(string)),
	}
}

func flattenFollowUpPrompt(followUp *lexmodelbuildingservice.FollowUpPrompt) (flattened []map[string]interface{}) {
	return []map[string]interface{}{
		{
			"prompt":              flattenPrompt(followUp.Prompt),
			"rejection_statement": flattenStatement(followUp.RejectionStatement),
		},
	}
}

func expandFollowUpPrompt(rawObject interface{}) (followUp *lexmodelbuildingservice.FollowUpPrompt) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	return &lexmodelbuildingservice.FollowUpPrompt{
		Prompt:             expandPrompt(m["prompt"]),
		RejectionStatement: expandStatement(m["rejection_statement"]),
	}
}

func flattenFulfilmentActivity(activity *lexmodelbuildingservice.FulfillmentActivity) (flattened []map[string]interface{}) {
	flattened = []map[string]interface{}{
		{
			names.AttrType: aws.StringValue(activity.Type),
		},
	}

	if activity.CodeHook != nil {
		flattened[0]["code_hook"] = flattenCodeHook(activity.CodeHook)
	}

	return
}

func expandFulfilmentActivity(rawObject interface{}) (activity *lexmodelbuildingservice.FulfillmentActivity) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	activity = &lexmodelbuildingservice.FulfillmentActivity{}
	activity.Type = aws.String(m[names.AttrType].(string))

	if v, ok := m["code_hook"]; ok && len(v.([]interface{})) != 0 {
		activity.CodeHook = expandCodeHook(v)
	}

	return
}

func flattenMessages(messages []*lexmodelbuildingservice.Message) (flattenedMessages []map[string]interface{}) {
	for _, message := range messages {
		flattenedMessages = append(flattenedMessages, map[string]interface{}{
			names.AttrContent:     aws.StringValue(message.Content),
			names.AttrContentType: aws.StringValue(message.ContentType),
			"group_number":        aws.Int64Value(message.GroupNumber),
		})
	}

	return
}

// Expects a slice of maps representing the Lex objects.
// The value passed into this function should have been run through the expandLexSet function.
// Example: []map[content: test content_type: PlainText group_number: 1]
func expandMessages(rawValues []interface{}) []*lexmodelbuildingservice.Message {
	messages := make([]*lexmodelbuildingservice.Message, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}

		message := &lexmodelbuildingservice.Message{
			Content:     aws.String(value[names.AttrContent].(string)),
			ContentType: aws.String(value[names.AttrContentType].(string)),
		}

		if v, ok := value["group_number"]; ok && v != 0 {
			message.GroupNumber = aws.Int64(int64(v.(int)))
		}

		messages = append(messages, message)
	}

	return messages
}

func flattenPrompt(prompt *lexmodelbuildingservice.Prompt) (flattened []map[string]interface{}) {
	flattened = []map[string]interface{}{
		{
			"max_attempts":    aws.Int64Value(prompt.MaxAttempts),
			names.AttrMessage: flattenMessages(prompt.Messages),
		},
	}

	if prompt.ResponseCard != nil {
		flattened[0]["response_card"] = aws.StringValue(prompt.ResponseCard)
	}

	return
}

func expandPrompt(rawObject interface{}) (prompt *lexmodelbuildingservice.Prompt) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	prompt = &lexmodelbuildingservice.Prompt{}
	prompt.MaxAttempts = aws.Int64(int64(m["max_attempts"].(int)))
	prompt.Messages = expandMessages(m[names.AttrMessage].(*schema.Set).List())

	if v, ok := m["response_card"]; ok && v != "" {
		prompt.ResponseCard = aws.String(v.(string))
	}

	return
}

func flattenSlots(slots []*lexmodelbuildingservice.Slot) (flattenedSlots []map[string]interface{}) {
	for _, slot := range slots {
		flattenedSlot := map[string]interface{}{
			names.AttrName:     aws.StringValue(slot.Name),
			names.AttrPriority: aws.Int64Value(slot.Priority),
			"slot_constraint":  aws.StringValue(slot.SlotConstraint),
			"slot_type":        aws.StringValue(slot.SlotType),
		}

		if slot.Description != nil {
			flattenedSlot[names.AttrDescription] = aws.StringValue(slot.Description)
		}

		if slot.ResponseCard != nil {
			flattenedSlot["response_card"] = aws.StringValue(slot.ResponseCard)
		}

		if slot.SampleUtterances != nil {
			flattenedSlot["sample_utterances"] = flex.FlattenStringList(slot.SampleUtterances)
		}

		if slot.SlotTypeVersion != nil {
			flattenedSlot["slot_type_version"] = aws.StringValue(slot.SlotTypeVersion)
		}

		if slot.ValueElicitationPrompt != nil {
			flattenedSlot["value_elicitation_prompt"] = flattenPrompt(slot.ValueElicitationPrompt)
		}

		flattenedSlots = append(flattenedSlots, flattenedSlot)
	}

	return flattenedSlots
}

// Expects a slice of maps representing the Lex objects.
// The value passed into this function should have been run through the expandLexSet function.
// Example: []map[name: test priority: 0 ...]
func expandSlots(rawValues []interface{}) []*lexmodelbuildingservice.Slot {
	slots := make([]*lexmodelbuildingservice.Slot, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}

		slot := &lexmodelbuildingservice.Slot{
			Name:           aws.String(value[names.AttrName].(string)),
			Priority:       aws.Int64(int64(value[names.AttrPriority].(int))),
			SlotConstraint: aws.String(value["slot_constraint"].(string)),
			SlotType:       aws.String(value["slot_type"].(string)),
		}

		if v, ok := value[names.AttrDescription]; ok && v != "" {
			slot.Description = aws.String(v.(string))
		}

		if v, ok := value["response_card"]; ok && v != "" {
			slot.ResponseCard = aws.String(v.(string))
		}

		if v, ok := value["response_card"]; ok && v != "" {
			slot.ResponseCard = aws.String(v.(string))
		}

		if v, ok := value["sample_utterances"]; ok && len(v.([]interface{})) != 0 {
			slot.SampleUtterances = flex.ExpandStringList(v.([]interface{}))
		}

		if v, ok := value["slot_type_version"]; ok && v != "" {
			slot.SlotTypeVersion = aws.String(v.(string))
		}

		if v, ok := value["value_elicitation_prompt"]; ok && len(v.([]interface{})) != 0 {
			slot.ValueElicitationPrompt = expandPrompt(v)
		}

		slots = append(slots, slot)
	}

	return slots
}

func flattenStatement(statement *lexmodelbuildingservice.Statement) (flattened []map[string]interface{}) {
	flattened = []map[string]interface{}{
		{
			names.AttrMessage: flattenMessages(statement.Messages),
		},
	}

	if statement.ResponseCard != nil {
		flattened[0]["response_card"] = aws.StringValue(statement.ResponseCard)
	}

	return
}

func expandStatement(rawObject interface{}) (statement *lexmodelbuildingservice.Statement) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	statement = &lexmodelbuildingservice.Statement{}
	statement.Messages = expandMessages(m[names.AttrMessage].(*schema.Set).List())

	if v, ok := m["response_card"]; ok && v != "" {
		statement.ResponseCard = aws.String(v.(string))
	}

	return
}
