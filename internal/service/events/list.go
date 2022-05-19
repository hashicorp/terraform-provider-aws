package events

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

func ListAllTargetsForRulePages(conn *eventbridge.EventBridge, busName, ruleName string, fn func(*eventbridge.ListTargetsByRuleOutput, bool) bool) error {
	input := &eventbridge.ListTargetsByRuleInput{
		Rule:  aws.String(ruleName),
		Limit: aws.Int64(100), // Set limit to allowed maximum to prevent API throttling
	}

	if busName != "" {
		input.EventBusName = aws.String(busName)
	}

	return listTargetsByRulePages(conn, input, fn)
}
