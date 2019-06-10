package aws

// Amazon Lex Resource Constants. Data models are documented here
// https://docs.aws.amazon.com/lex/latest/dg/API_Types_Amazon_Lex_Model_Building_Service.html

const (

	// General

	lexNameMinLength = 1
	lexNameMaxLength = 100
	lexNameRegex     = "^([A-Za-z]_?)+$"

	lexVersionMinLength = 1
	lexVersionMaxLength = 64
	lexVersionRegex     = "\\$LATEST|[0-9]+"
	lexVersionLatest    = "$LATEST"
	lexVersionDefault   = "$LATEST"

	lexDescriptionMinLength = 0
	lexDescriptionMaxLength = 200
	lexDescriptionDefault   = ""

	// Message

	lexMessageContentMinLength = 1
	lexMessageContentMaxLength = 1000
	lexMessageGroupNumberMin   = 1
	lexMessageGroupNumberMax   = 5

	// Statement

	lexResponseCardMinLength = 1
	lexResponseCardMaxLength = 50000
	lexStatementMessagesMin  = 1
	lexStatementMessagesMax  = 15

	// Prompt

	lexPromptMaxAttemptsMin = 1
	lexPromptMaxAttemptsMax = 5

	// Code Hook

	lexCodeHookMessageVersionMinLength = 1
	lexCodeHookMessageVersionMaxLength = 5

	// Slot

	lexSlotsMin                = 0
	lexSlotsMax                = 100
	lexSlotPriorityMin         = 0
	lexSlotPriorityMax         = 100
	lexSlotPriorityDefault     = 0
	lexSlotSampleUtterancesMin = 1
	lexSlotSampleUtterancesMax = 10

	// Slot Type

	lexSlotTypeMinLength = 1
	lexSlotTypeMaxLength = 100
	lexSlotTypeRegex     = "^((AMAZON\\.)_?|[A-Za-z]_?)+"

	// Utterance

	lexUtterancesMin      = 0
	lexUtterancesMax      = 1500
	lexUtteranceMinLength = 1
	lexUtteranceMaxLength = 200
)
