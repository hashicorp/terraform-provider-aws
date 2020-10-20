package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	tfevents "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents"
)

func Rule(conn *events.CloudWatchEvents, eventBusName, ruleName string) (*events.DescribeRuleOutput, error) {
	input := events.DescribeRuleInput{
		Name: aws.String(ruleName),
	}
	if eventBusName != "" {
		input.EventBusName = aws.String(eventBusName)
	}

	return conn.DescribeRule(&input)

}

func RuleByID(conn *events.CloudWatchEvents, ruleID string) (*events.DescribeRuleOutput, error) {
	busName, ruleName, err := tfevents.RuleParseID(ruleID)
	if err != nil {
		return nil, err
	}

	return Rule(conn, busName, ruleName)
}
