package lexmodels

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	LexIntentCreateTimeout = 1 * time.Minute
	LexIntentUpdateTimeout = 1 * time.Minute
	LexIntentDeleteTimeout = 5 * time.Minute
)

func ResourceIntent() *schema.Resource {
	return &schema.Resource{
		Create: resourceIntentCreate,
		Read:   resourceIntentRead,
		Update: resourceIntentUpdate,
		Delete: resourceIntentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(LexIntentCreateTimeout),
			Update: schema.DefaultTimeout(LexIntentUpdateTimeout),
			Delete: schema.DefaultTimeout(LexIntentDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
				Elem:          lexStatementResource,
			},
			"confirmation_prompt": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				MaxItems:     1,
				RequiredWith: []string{"rejection_statement"},
				Elem:         lexPromptResource,
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
			"dialog_code_hook": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexCodeHookResource,
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
							Elem:     lexPromptResource,
						},
						"rejection_statement": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     lexStatementResource,
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
							Elem:     lexCodeHookResource,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.FulfillmentActivityType_Values(), false),
						},
					},
				},
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
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
				Elem:         lexStatementResource,
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
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "",
							ValidateFunc: validation.StringLenBetween(0, 200),
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
							),
						},
						"priority": {
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
								validation.StringMatch(regexp.MustCompile(`^((AMAZON\.)_?|[A-Za-z]_?)+`), ""),
							),
						},
						"slot_type_version": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringMatch(regexp.MustCompile(`\$LATEST|[0-9]+`), ""),
							),
						},
						"value_elicitation_prompt": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     lexPromptResource,
						},
					},
				},
			},
			"version": {
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
		d.SetNewComputed("version")
	}
	return nil
}

func hasIntentConfigChanges(d verify.ResourceDiffer) bool {
	for _, key := range []string{
		"description",
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

func resourceIntentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelsConn
	name := d.Get("name").(string)

	input := &lexmodelbuildingservice.PutIntentInput{
		CreateVersion: aws.Bool(d.Get("create_version").(bool)),
		Description:   aws.String(d.Get("description").(string)),
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

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err := conn.PutIntent(input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			input.Checksum = output.Checksum
			return resource.RetryableError(fmt.Errorf("%q intent still creating, another operation is pending: %s", d.Id(), err))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) { // nosemgrep: helper-schema-TimeoutError-check-doesnt-return-output
		_, err = conn.PutIntent(input)
	}

	if err != nil {
		return fmt.Errorf("error creating intent %s: %w", name, err)
	}

	d.SetId(name)

	return resourceIntentRead(d, meta)
}

func resourceIntentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelsConn

	resp, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
		Name:    aws.String(d.Id()),
		Version: aws.String(IntentVersionLatest),
	})
	if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
		log.Printf("[WARN] Intent (%s) not found, removing from state", d.Id())
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
		Resource:  fmt.Sprintf("intent:%s", d.Id()),
	}
	d.Set("arn", arn.String())

	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set("name", resp.Name)

	version, err := FindLatestIntentVersionByName(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading Lex Intent (%s) latest version: %w", d.Id(), err)
	}

	d.Set("version", version)

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

	if resp.ParentIntentSignature != nil {
		d.Set("parent_intent_signature", resp.ParentIntentSignature)
	}

	if resp.RejectionStatement != nil {
		d.Set("rejection_statement", flattenStatement(resp.RejectionStatement))
	}

	if resp.SampleUtterances != nil {
		d.Set("sample_utterances", resp.SampleUtterances)
	}

	if resp.Slots != nil {
		d.Set("slot", flattenSlots(resp.Slots))
	}

	return nil
}

func resourceIntentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelsConn

	input := &lexmodelbuildingservice.PutIntentInput{
		Checksum:      aws.String(d.Get("checksum").(string)),
		CreateVersion: aws.Bool(d.Get("create_version").(bool)),
		Description:   aws.String(d.Get("description").(string)),
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

	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := conn.PutIntent(input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return resource.RetryableError(fmt.Errorf("%q: intent still updating", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutIntent(input)
	}

	if err != nil {
		return fmt.Errorf("error updating intent %s: %w", d.Id(), err)
	}

	return resourceIntentRead(d, meta)
}

func resourceIntentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelsConn

	input := &lexmodelbuildingservice.DeleteIntentInput{
		Name: aws.String(d.Id()),
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteIntent(input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return resource.RetryableError(fmt.Errorf("%q: there is a pending operation, intent still deleting", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteIntent(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting intent %s: %w", d.Id(), err)
	}

	_, err = waitIntentDeleted(conn, d.Id())

	return err
}

var lexCodeHookResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"message_version": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 5),
		},
		"uri": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: verify.ValidARN,
		},
	},
}

var lexMessageResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"content": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 1000),
		},
		"content_type": {
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

var lexPromptResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"max_attempts": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(1, 5),
		},
		"message": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			MaxItems: 15,
			Elem:     lexMessageResource,
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 50000),
		},
	},
}

var lexStatementResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"message": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			MaxItems: 15,
			Elem:     lexMessageResource,
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
			"uri":             aws.StringValue(hook.Uri),
		},
	}
}

func expandCodeHook(rawObject interface{}) (hook *lexmodelbuildingservice.CodeHook) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	return &lexmodelbuildingservice.CodeHook{
		MessageVersion: aws.String(m["message_version"].(string)),
		Uri:            aws.String(m["uri"].(string)),
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
			"type": aws.StringValue(activity.Type),
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
	activity.Type = aws.String(m["type"].(string))

	if v, ok := m["code_hook"]; ok && len(v.([]interface{})) != 0 {
		activity.CodeHook = expandCodeHook(v)
	}

	return
}

func flattenMessages(messages []*lexmodelbuildingservice.Message) (flattenedMessages []map[string]interface{}) {
	for _, message := range messages {
		flattenedMessages = append(flattenedMessages, map[string]interface{}{
			"content":      aws.StringValue(message.Content),
			"content_type": aws.StringValue(message.ContentType),
			"group_number": aws.Int64Value(message.GroupNumber),
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
			Content:     aws.String(value["content"].(string)),
			ContentType: aws.String(value["content_type"].(string)),
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
			"max_attempts": aws.Int64Value(prompt.MaxAttempts),
			"message":      flattenMessages(prompt.Messages),
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
	prompt.Messages = expandMessages(m["message"].(*schema.Set).List())

	if v, ok := m["response_card"]; ok && v != "" {
		prompt.ResponseCard = aws.String(v.(string))
	}

	return
}

func flattenSlots(slots []*lexmodelbuildingservice.Slot) (flattenedSlots []map[string]interface{}) {
	for _, slot := range slots {
		flattenedSlot := map[string]interface{}{
			"name":            aws.StringValue(slot.Name),
			"priority":        aws.Int64Value(slot.Priority),
			"slot_constraint": aws.StringValue(slot.SlotConstraint),
			"slot_type":       aws.StringValue(slot.SlotType),
		}

		if slot.Description != nil {
			flattenedSlot["description"] = aws.StringValue(slot.Description)
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
			Name:           aws.String(value["name"].(string)),
			Priority:       aws.Int64(int64(value["priority"].(int))),
			SlotConstraint: aws.String(value["slot_constraint"].(string)),
			SlotType:       aws.String(value["slot_type"].(string)),
		}

		if v, ok := value["description"]; ok && v != "" {
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
			"message": flattenMessages(statement.Messages),
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
	statement.Messages = expandMessages(m["message"].(*schema.Set).List())

	if v, ok := m["response_card"]; ok && v != "" {
		statement.ResponseCard = aws.String(v.(string))
	}

	return
}
