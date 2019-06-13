package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

// Many of the Lex resources require complex nested objects. Terraform maps only support simple key
// value pairs and not complex or mixed types. That is why these resources are defined using the
// schema.TypeList and a max of 1 item instead of the schema.TypeMap.

// Convert a slice of items to a map[string]interface{}
// Expects input as a single item slice.
// Required because we use TypeList instead of TypeMap due to TypeMap not supporting nested and mixed complex values.
func expandLexObject(v interface{}) map[string]interface{} {
	return v.([]interface{})[0].(map[string]interface{})
}

// Covert a map[string]interface{} to a slice of items
// Expects a single map[string]interface{}
// Required because we use TypeList instead of TypeMap due to TypeMap not supporting nested and mixed complex values.
func flattenLexObject(m map[string]interface{}) []map[string]interface{} {
	return []map[string]interface{}{m}
}

func expandLexSet(s *schema.Set) (items []map[string]interface{}) {
	for _, rawItem := range s.List() {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}

		items = append(items, item)
	}

	return
}

var lexMessageResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"content": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 1000),
		},
		"content_type": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				lexmodelbuildingservice.ContentTypeCustomPayload,
				lexmodelbuildingservice.ContentTypePlainText,
				lexmodelbuildingservice.ContentTypeSsml,
			}, false),
		},
		"group_number": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(1, 5),
		},
	},
}

func flattenLexMessages(messages []*lexmodelbuildingservice.Message) (flattenedMessages []map[string]interface{}) {
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
func expandLexMessages(rawValues []map[string]interface{}) (messages []*lexmodelbuildingservice.Message) {
	for _, rawValue := range rawValues {
		message := &lexmodelbuildingservice.Message{
			Content:     aws.String(rawValue["content"].(string)),
			ContentType: aws.String(rawValue["content_type"].(string)),
		}

		if v, ok := rawValue["group_number"]; ok && v != 0 {
			message.GroupNumber = aws.Int64(int64(v.(int)))
		}

		messages = append(messages, message)
	}

	return
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

func flattenLexStatement(statement *lexmodelbuildingservice.Statement) (flattened map[string]interface{}) {
	flattened = map[string]interface{}{}
	flattened["message"] = flattenLexMessages(statement.Messages)

	if statement.ResponseCard != nil {
		flattened["response_card"] = aws.StringValue(statement.ResponseCard)
	}

	return
}

func expandLexStatement(m map[string]interface{}) (statement *lexmodelbuildingservice.Statement) {
	statement = &lexmodelbuildingservice.Statement{}
	statement.Messages = expandLexMessages(expandLexSet(m["message"].(*schema.Set)))

	if v, ok := m["response_card"]; ok && v != "" {
		statement.ResponseCard = aws.String(v.(string))
	}

	return
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

func flattenLexPrompt(prompt *lexmodelbuildingservice.Prompt) (flattened map[string]interface{}) {
	flattened = map[string]interface{}{}
	flattened["max_attempts"] = aws.Int64Value(prompt.MaxAttempts)
	flattened["message"] = flattenLexMessages(prompt.Messages)

	if prompt.ResponseCard != nil {
		flattened["response_card"] = aws.StringValue(prompt.ResponseCard)
	}

	return
}

func expandLexPrompt(m map[string]interface{}) (prompt *lexmodelbuildingservice.Prompt) {
	prompt = &lexmodelbuildingservice.Prompt{}
	prompt.MaxAttempts = aws.Int64(int64(m["max_attempts"].(int)))
	prompt.Messages = expandLexMessages(expandLexSet(m["message"].(*schema.Set)))

	if v, ok := m["response_card"]; ok && v != "" {
		prompt.ResponseCard = aws.String(v.(string))
	}

	return
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
			ValidateFunc: validateArn,
		},
	},
}

func flattenLexCodeHook(hook *lexmodelbuildingservice.CodeHook) (flattened map[string]interface{}) {
	return map[string]interface{}{
		"message_version": aws.StringValue(hook.MessageVersion),
		"uri":             aws.StringValue(hook.Uri),
	}
}

func expandLexCodeHook(m map[string]interface{}) (hook *lexmodelbuildingservice.CodeHook) {
	return &lexmodelbuildingservice.CodeHook{
		MessageVersion: aws.String(m["message_version"].(string)),
		Uri:            aws.String(m["uri"].(string)),
	}
}

func flattenLexFollowUpPrompt(followUp *lexmodelbuildingservice.FollowUpPrompt) (flattened map[string]interface{}) {
	return map[string]interface{}{
		"prompt":              flattenLexObject(flattenLexPrompt(followUp.Prompt)),
		"rejection_statement": flattenLexObject(flattenLexStatement(followUp.RejectionStatement)),
	}
}

func expandLexFollowUpPrompt(m map[string]interface{}) (followUp *lexmodelbuildingservice.FollowUpPrompt) {
	return &lexmodelbuildingservice.FollowUpPrompt{
		Prompt:             expandLexPrompt(expandLexObject(m["prompt"])),
		RejectionStatement: expandLexStatement(expandLexObject(m["rejection_statement"])),
	}
}

func flattenLexFulfilmentActivity(activity *lexmodelbuildingservice.FulfillmentActivity) (flattened map[string]interface{}) {
	flattened = map[string]interface{}{}
	flattened["type"] = aws.StringValue(activity.Type)

	if activity.CodeHook != nil {
		flattened["code_hook"] = flattenLexObject(flattenLexCodeHook(activity.CodeHook))
	}

	return
}

func expandLexFulfilmentActivity(m map[string]interface{}) (activity *lexmodelbuildingservice.FulfillmentActivity) {
	activity = &lexmodelbuildingservice.FulfillmentActivity{}
	activity.Type = aws.String(m["type"].(string))

	if v, ok := m["code_hook"]; ok && len(v.([]interface{})) != 0 {
		activity.CodeHook = expandLexCodeHook(expandLexObject(v))
	}

	return
}

func flattenLexSlots(slots []*lexmodelbuildingservice.Slot) (flattenedSlots []map[string]interface{}) {
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
			flattenedSlot["sample_utterances"] = flattenStringList(slot.SampleUtterances)
		}

		if slot.SlotTypeVersion != nil {
			flattenedSlot["slot_type_version"] = aws.StringValue(slot.SlotTypeVersion)
		}

		if slot.ValueElicitationPrompt != nil {
			flattenedSlot["value_elicitation_prompt"] = flattenLexPrompt(slot.ValueElicitationPrompt)
		}

		flattenedSlots = append(flattenedSlots, flattenedSlot)
	}

	return
}

// Expects a slice of maps representing the Lex objects.
// The value passed into this function should have been run through the expandLexSet function.
// Example: []map[name: test priority: 0 ...]
func expandLexSlots(rawValues []map[string]interface{}) (slots []*lexmodelbuildingservice.Slot) {
	for _, rawValue := range rawValues {
		slot := &lexmodelbuildingservice.Slot{
			Name:           aws.String(rawValue["name"].(string)),
			Priority:       aws.Int64(int64(rawValue["priority"].(int))),
			SlotConstraint: aws.String(rawValue["slot_constraint"].(string)),
			SlotType:       aws.String(rawValue["slot_type"].(string)),
		}

		if v, ok := rawValue["description"]; ok && v != "" {
			slot.Description = aws.String(v.(string))
		}

		if v, ok := rawValue["response_card"]; ok && v != "" {
			slot.ResponseCard = aws.String(v.(string))
		}

		if v, ok := rawValue["response_card"]; ok && v != "" {
			slot.ResponseCard = aws.String(v.(string))
		}

		if v, ok := rawValue["sample_utterances"]; ok && len(v.([]interface{})) != 0 {
			slot.SampleUtterances = expandStringList(v.([]interface{}))
		}

		if v, ok := rawValue["slot_type_version"]; ok && v != "" {
			slot.SlotTypeVersion = aws.String(v.(string))
		}

		if v, ok := rawValue["value_elicitation_prompt"]; ok && len(v.([]interface{})) != 0 {
			slot.ValueElicitationPrompt = expandLexPrompt(expandLexObject(v))
		}

		slots = append(slots, slot)
	}

	return
}

func resourceAwsLexIntent() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLexIntentCreate,
		Read:   resourceAwsLexIntentRead,
		Update: resourceAwsLexIntentUpdate,
		Delete: resourceAwsLexIntentDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				// The version is not required for import but it is required for the get request.
				d.Set("version", "$LATEST")
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"conclusion_statement": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexStatementResource,
			},
			"confirmation_prompt": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexPromptResource,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
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
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
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
			// Must be required because required by updates even though optional for creates
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
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								lexmodelbuildingservice.FulfillmentActivityTypeCodeHook,
								lexmodelbuildingservice.FulfillmentActivityTypeReturnIntent,
							}, false),
						},
					},
				},
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
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexStatementResource,
			},
			"sample_utterances": {
				Type:     schema.TypeList,
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
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								lexmodelbuildingservice.SlotConstraintOptional,
								lexmodelbuildingservice.SlotConstraintRequired,
							}, false),
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
				Optional: true,
				Default:  "$LATEST",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`\$LATEST|[0-9]+`), ""),
				),
			},
		},
	}
}

func resourceAwsLexIntentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	name := d.Get("name").(string)

	input := &lexmodelbuildingservice.PutIntentInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("conclusion_statement"); ok {
		input.ConclusionStatement = expandLexStatement(expandLexObject(v))
	}

	if v, ok := d.GetOk("confirmation_prompt"); ok {
		input.ConfirmationPrompt = expandLexPrompt(expandLexObject(v))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dialog_code_hook"); ok {
		input.DialogCodeHook = expandLexCodeHook(expandLexObject(v))
	}

	if v, ok := d.GetOk("follow_up_prompt"); ok {
		input.FollowUpPrompt = expandLexFollowUpPrompt(expandLexObject(v))
	}

	if v, ok := d.GetOk("fulfillment_activity"); ok {
		input.FulfillmentActivity = expandLexFulfilmentActivity(expandLexObject(v))
	}

	if v, ok := d.GetOk("parent_intent_signature"); ok {
		input.ParentIntentSignature = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rejection_statement"); ok {
		input.RejectionStatement = expandLexStatement(expandLexObject(v))
	}

	if v, ok := d.GetOk("sample_utterances"); ok {
		input.SampleUtterances = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("slot"); ok {
		input.Slots = expandLexSlots(expandLexSet(v.(*schema.Set)))
	}

	if _, err := conn.PutIntent(input); err != nil {
		return fmt.Errorf("error creating Lex Intent %s: %s", name, err)
	}

	d.SetId(name)

	return resourceAwsLexIntentRead(d, meta)
}

func resourceAwsLexIntentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
		Name:    aws.String(d.Id()),
		Version: aws.String(d.Get("version").(string)),
	})
	if isAWSErr(err, lexmodelbuildingservice.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Intent (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting intent %s: %s", d.Id(), err)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("version", resp.Version)

	if resp.ConclusionStatement != nil {
		d.Set("conclusion_statement", flattenLexObject(flattenLexStatement(resp.ConclusionStatement)))
	}

	if resp.ConfirmationPrompt != nil {
		d.Set("confirmation_prompt", flattenLexObject(flattenLexPrompt(resp.ConfirmationPrompt)))
	}

	if resp.DialogCodeHook != nil {
		d.Set("dialog_code_hook", flattenLexObject(flattenLexCodeHook(resp.DialogCodeHook)))
	}

	if resp.FollowUpPrompt != nil {
		d.Set("follow_up_prompt", flattenLexObject(flattenLexFollowUpPrompt(resp.FollowUpPrompt)))
	}

	if resp.FulfillmentActivity != nil {
		d.Set("fulfillment_activity", flattenLexObject(flattenLexFulfilmentActivity(resp.FulfillmentActivity)))
	}

	if resp.ParentIntentSignature != nil {
		d.Set("parent_intent_signature", resp.ParentIntentSignature)
	}

	if resp.RejectionStatement != nil {
		d.Set("rejection_statement", flattenLexObject(flattenLexStatement(resp.RejectionStatement)))
	}

	if resp.SampleUtterances != nil {
		d.Set("sample_utterances", resp.SampleUtterances)
	}

	if resp.Slots != nil {
		d.Set("slot", flattenLexSlots(resp.Slots))
	}

	return nil
}

func resourceAwsLexIntentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.PutIntentInput{
		Checksum: aws.String(d.Get("checksum").(string)),
		Name:     aws.String(d.Id()),
	}

	if v, ok := d.GetOk("conclusion_statement"); ok {
		input.ConclusionStatement = expandLexStatement(expandLexObject(v))
	}

	if v, ok := d.GetOk("confirmation_prompt"); ok {
		input.ConfirmationPrompt = expandLexPrompt(expandLexObject(v))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dialog_code_hook"); ok {
		input.DialogCodeHook = expandLexCodeHook(expandLexObject(v))
	}

	if v, ok := d.GetOk("follow_up_prompt"); ok {
		input.FollowUpPrompt = expandLexFollowUpPrompt(expandLexObject(v))
	}

	if v, ok := d.GetOk("fulfillment_activity"); ok {
		input.FulfillmentActivity = expandLexFulfilmentActivity(expandLexObject(v))
	}

	if v, ok := d.GetOk("parent_intent_signature"); ok {
		input.ParentIntentSignature = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rejection_statement"); ok {
		input.RejectionStatement = expandLexStatement(expandLexObject(v))
	}

	if v, ok := d.GetOk("sample_utterances"); ok {
		input.SampleUtterances = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("slot"); ok {
		input.Slots = expandLexSlots(expandLexSet(v.(*schema.Set)))
	}

	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := conn.PutIntent(input)

		if isAWSErr(err, lexmodelbuildingservice.ErrCodeConflictException, "") {
			return resource.RetryableError(fmt.Errorf("%q: intent still updating", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error updating intent %s: %s", d.Id(), err)
	}

	return resourceAwsLexIntentRead(d, meta)
}

func resourceAwsLexIntentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteIntent(&lexmodelbuildingservice.DeleteIntentInput{
			Name: aws.String(d.Id()),
		})

		if isAWSErr(err, lexmodelbuildingservice.ErrCodeConflictException, "") {
			return resource.RetryableError(fmt.Errorf("%q: intent still deleting", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error deleting intent %s: %s", d.Id(), err)
	}

	return nil
}
