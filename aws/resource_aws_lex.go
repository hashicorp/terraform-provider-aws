package aws

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func lexHash(v interface{}) int {
	return hashcode.String(fmt.Sprintf("%#v", v.(map[string]interface{})))
}

var lexMessageResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"content": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(lexMessageContentMinLength, lexMessageContentMaxLength),
		},
		"content_type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validateLexMessageContentType,
		},
	},
}

func flattenLexMessages(messages []*lexmodelbuildingservice.Message) *schema.Set {
	flattenedMessages := []interface{}{}

	for _, message := range messages {
		flattenedMessages = append(flattenedMessages, map[string]interface{}{
			"content":      aws.StringValue(message.Content),
			"content_type": aws.StringValue(message.ContentType),
		})
	}

	return schema.NewSet(lexHash, flattenedMessages)
}

var lexStatementResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"message": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			Elem:     lexMessageResource,
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(lexResponseCardMinLength, lexResponseCardMaxLength),
		},
	},
}

func flattenLexStatement(statement *lexmodelbuildingservice.Statement) []map[string]interface{} {
	return []map[string]interface{}{{
		"message":       flattenLexMessages(statement.Messages),
		"response_card": aws.StringValue(statement.ResponseCard),
	}}
}

func expandLexStatement(rawValue interface{}) *lexmodelbuildingservice.Statement {
	rawStatement := rawValue.([]interface{})[0].(map[string]interface{})

	statement := &lexmodelbuildingservice.Statement{
		Messages: []*lexmodelbuildingservice.Message{},
	}

	responseCard := rawStatement["response_card"].(string)
	if responseCard != "" {
		statement.ResponseCard = aws.String(responseCard)
	}

	messages := rawStatement["message"].(*schema.Set)

	for _, rawMessage := range messages.List() {
		message := rawMessage.(map[string]interface{})

		statement.Messages = append(statement.Messages, &lexmodelbuildingservice.Message{
			Content:     aws.String(message["content"].(string)),
			ContentType: aws.String(message["content_type"].(string)),
		})
	}

	return statement
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
			MinItems: 1,
			Elem:     lexMessageResource,
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(lexResponseCardMinLength, lexResponseCardMaxLength),
		},
	},
}

func flattenLexPrompt(prompt *lexmodelbuildingservice.Prompt) []map[string]interface{} {
	return []map[string]interface{}{{
		"max_attempts":  aws.Int64Value(prompt.MaxAttempts),
		"message":       flattenLexMessages(prompt.Messages),
		"response_card": aws.StringValue(prompt.ResponseCard),
	}}
}

func expandLexPrompt(rawValue interface{}) *lexmodelbuildingservice.Prompt {
	rawPrompt := rawValue.([]interface{})[0].(map[string]interface{})

	prompt := &lexmodelbuildingservice.Prompt{
		MaxAttempts: aws.Int64(int64(rawPrompt["max_attempts"].(int))),
		Messages:    []*lexmodelbuildingservice.Message{},
	}

	responseCard := rawPrompt["response_card"].(string)
	if responseCard != "" {
		prompt.ResponseCard = aws.String(responseCard)
	}

	messages := rawPrompt["message"].(*schema.Set)

	for _, rawMessage := range messages.List() {
		message := rawMessage.(map[string]interface{})

		prompt.Messages = append(prompt.Messages, &lexmodelbuildingservice.Message{
			Content:     aws.String(message["content"].(string)),
			ContentType: aws.String(message["content_type"].(string)),
		})
	}

	return prompt
}

var lexIntentResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"intent_name": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validateLexName,
		},
		"intent_version": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validateLexVersion,
		},
	},
}

func flattenLexIntents(intents []*lexmodelbuildingservice.Intent) []map[string]interface{} {
	flattenedIntents := []map[string]interface{}{}

	for _, intent := range intents {
		flattenedIntents = append(flattenedIntents, map[string]interface{}{
			"intent_name":    aws.StringValue(intent.IntentName),
			"intent_version": aws.StringValue(intent.IntentVersion),
		})
	}

	return flattenedIntents
}

func expandLexIntents(rawValue interface{}) []*lexmodelbuildingservice.Intent {
	rawIntents := rawValue.([]interface{})
	intents := []*lexmodelbuildingservice.Intent{}

	for _, rawIntent := range rawIntents {
		intents = append(intents, &lexmodelbuildingservice.Intent{
			IntentName:    aws.String(rawIntent.(map[string]interface{})["intent_name"].(string)),
			IntentVersion: aws.String(rawIntent.(map[string]interface{})["intent_version"].(string)),
		})
	}

	return intents
}

var lexEnumerationValueResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"synonyms": {
			Type:     schema.TypeList,
			Optional: true,
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

func flattenLexEnumerationValues(values []*lexmodelbuildingservice.EnumerationValue) []map[string]interface{} {
	flattenedValues := []map[string]interface{}{}

	for _, value := range values {
		flattenedValues = append(flattenedValues, map[string]interface{}{
			"synonyms": aws.StringValueSlice(value.Synonyms),
			"value":    aws.StringValue(value.Value),
		})
	}

	// Prevent inconsistent diffs by sorting the enumeration values by value
	sort.Slice(flattenedValues, func(i, j int) bool {
		return flattenedValues[i]["value"].(string) < flattenedValues[j]["value"].(string)
	})

	return flattenedValues
}

func expandLexEnumerationValues(rawValue interface{}) []*lexmodelbuildingservice.EnumerationValue {
	rawValues := rawValue.([]interface{})
	values := []*lexmodelbuildingservice.EnumerationValue{}

	for _, rawValue := range rawValues {
		synonyms := []string{}
		rawSynonyms := rawValue.(map[string]interface{})["synonyms"]
		for _, rawSynonym := range rawSynonyms.([]interface{}) {
			synonyms = append(synonyms, rawSynonym.(string))
		}

		values = append(values, &lexmodelbuildingservice.EnumerationValue{
			Synonyms: aws.StringSlice(synonyms),
			Value:    aws.String(rawValue.(map[string]interface{})["value"].(string)),
		})
	}

	return values
}
