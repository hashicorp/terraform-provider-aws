package aws

import (
	"regexp"

	"github.com/hashicorp/terraform/helper/validation"
)

const (
	// General constants

	// The documented bot name regex isn't the same regex the AWS API validates against.
	// http://docs.aws.amazon.com/lex/latest/dg/API_PutBot.html
	//
	// It appears the AWS API is validating against the botName regex documented for bot aliases.
	// https://docs.aws.amazon.com/lex/latest/dg/API_PutBotAlias.html
	//
	// Intent names are using the same regex for validation
	// https://docs.aws.amazon.com/lex/latest/dg/API_PutIntent.html

	lexNameMinLength = 1
	lexNameMaxLength = 100
	lexNameRegex     = "^([A-Za-z]_?)+$"

	lexVersionMinLength = 1
	lexVersionMaxLength = 64
	lexVersionRegex     = "\\$LATEST|[0-9]+"

	lexDescriptionMaxLength = 200

	// Bot constants

	lexBotIdleSessionTtlMin = 60
	lexBotIdleSessionTtlMax = 86400
	lexBotMaxIntents        = 100

	// Message constants

	lexMessageContentMinLength = 1
	lexMessageContentMaxLength = 1000

	// Statement constants

	lexResponseCardMinLength = 1
	lexResponseCardMaxLength = 50000

	// Prompt constants

	lexPromptMaxAttemptsMin = 1
	lexPromptMaxAttemptsMax = 5

	// Slot type constants

	lexSlotTypeMinEnumerationValues = 1
	lexSlotTypeMaxEnumerationValues = 10000

	// Enumeration value constants

	lexEnumerationValueSynonymMinLength = 1
	lexEnumerationValueSynonymMaxLength = 140
	lexEnumerationValueMinLength        = 1
	lexEnumerationValueMaxLength        = 140
)

func validateLexName(v interface{}, k string) (ws []string, errors []error) {
	ws, errors = validation.StringLenBetween(lexNameMinLength, lexNameMaxLength)(v, k)
	if len(errors) > 0 {
		return ws, errors
	}

	return validation.StringMatch(regexp.MustCompile(lexNameRegex), "")(v, k)
}

func validateLexVersion(v interface{}, k string) (ws []string, errors []error) {
	ws, errors = validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength)(v, k)
	if len(errors) > 0 {
		return ws, errors
	}

	return validation.StringMatch(regexp.MustCompile(lexVersionRegex), "")(v, k)
}

func validateLexMessageContentType(v interface{}, k string) (ws []string, errors []error) {
	return validation.StringInSlice([]string{
		"PlainText",
		"SSML",
		"CustomPayload",
	}, false)(v, k)
}

func validateLexSlotSelectionStrategy(v interface{}, k string) (ws []string, errors []error) {
	return validation.StringInSlice([]string{
		"ORIGINAL_VALUE",
		"TOP_RESOLUTION",
	}, false)(v, k)
}
