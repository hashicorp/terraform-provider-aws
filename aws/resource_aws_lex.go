package aws

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
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
			ValidateFunc: validation.StringLenBetween(lexMessageContentMinLength, lexMessageContentMaxLength),
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
			ValidateFunc: validation.IntBetween(lexMessageGroupNumberMin, lexMessageGroupNumberMax),
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
			MinItems: lexStatementMessagesMin,
			MaxItems: lexStatementMessagesMax,
			Elem:     lexMessageResource,
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(lexResponseCardMinLength, lexResponseCardMaxLength),
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
			ValidateFunc: validation.IntBetween(lexPromptMaxAttemptsMin, lexPromptMaxAttemptsMax),
		},
		"message": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: lexStatementMessagesMin,
			MaxItems: lexStatementMessagesMax,
			Elem:     lexMessageResource,
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(lexResponseCardMinLength, lexResponseCardMaxLength),
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

var lexIntentResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"intent_name": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(lexNameMinLength, lexNameMaxLength),
				validation.StringMatch(regexp.MustCompile(lexNameRegex), ""),
			),
		},
		"intent_version": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength),
				validation.StringMatch(regexp.MustCompile(lexVersionRegex), ""),
			),
		},
	},
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
func expandLexIntents(rawValues []map[string]interface{}) (intents []*lexmodelbuildingservice.Intent) {
	for _, rawValue := range rawValues {
		intents = append(intents, &lexmodelbuildingservice.Intent{
			IntentName:    aws.String(rawValue["intent_name"].(string)),
			IntentVersion: aws.String(rawValue["intent_version"].(string)),
		})
	}

	return
}

var lexCodeHookResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"message_version": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(lexCodeHookMessageVersionMinLength, lexCodeHookMessageVersionMaxLength),
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

var lexFollowUpPromptResource = &schema.Resource{
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

var lexFulfilmentActivityResource = &schema.Resource{
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

var lexSlotResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"description": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      lexDescriptionDefault,
			ValidateFunc: validation.StringLenBetween(lexDescriptionMinLength, lexDescriptionMaxLength),
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(lexNameMinLength, lexNameMaxLength),
				validation.StringMatch(regexp.MustCompile(lexNameRegex), ""),
			),
		},
		"priority": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      lexSlotPriorityDefault,
			ValidateFunc: validation.IntBetween(lexSlotPriorityMin, lexSlotPriorityMax),
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(lexResponseCardMinLength, lexResponseCardMaxLength),
		},
		"sample_utterances": {
			Type:     schema.TypeList,
			Optional: true,
			MinItems: lexSlotSampleUtterancesMin,
			MaxItems: lexSlotSampleUtterancesMax,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(lexUtteranceMinLength, lexUtteranceMaxLength),
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
				validation.StringLenBetween(lexSlotTypeMinLength, lexSlotTypeMaxLength),
				validation.StringMatch(regexp.MustCompile(lexSlotTypeRegex), ""),
			),
		},
		"slot_type_version": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength),
				validation.StringMatch(regexp.MustCompile(lexVersionRegex), ""),
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

var lexEnumerationValueResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"synonyms": {
			Type:     schema.TypeList,
			Optional: true,
			MinItems: 1,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(lexEnumerationValueSynonymMinLength, lexEnumerationValueSynonymMaxLength),
			},
		},
		"value": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(lexEnumerationValueMinLength, lexEnumerationValueMaxLength),
		},
	},
}

func flattenLexEnumerationValues(values []*lexmodelbuildingservice.EnumerationValue) (flattened []map[string]interface{}) {
	for _, value := range values {
		flattened = append(flattened, map[string]interface{}{
			"synonyms": flattenStringList(value.Synonyms),
			"value":    aws.StringValue(value.Value),
		})
	}

	return
}

// Expects a slice of maps representing the Lex objects.
// The value passed into this function should have been run through the expandLexSet function.
// Example: []map[value: lilies synonyms:[]lirium]]
func expandLexEnumerationValues(rawValues []map[string]interface{}) (enums []*lexmodelbuildingservice.EnumerationValue) {
	for _, rawValue := range rawValues {
		enums = append(enums, &lexmodelbuildingservice.EnumerationValue{
			Synonyms: expandStringList(rawValue["synonyms"].([]interface{})),
			Value:    aws.String(rawValue["value"].(string)),
		})
	}

	return
}
