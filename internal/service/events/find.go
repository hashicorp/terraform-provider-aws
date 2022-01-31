package events

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindConnectionByName(conn *eventbridge.EventBridge, name string) (*eventbridge.DescribeConnectionOutput, error) {
	input := &eventbridge.DescribeConnectionInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeConnection(input)

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}

func FindRuleByEventBusAndRuleNames(conn *eventbridge.EventBridge, eventBusName, ruleName string) (*eventbridge.DescribeRuleOutput, error) {
	input := eventbridge.DescribeRuleInput{
		Name: aws.String(ruleName),
	}

	if eventBusName != "" {
		input.EventBusName = aws.String(eventBusName)
	}

	output, err := conn.DescribeRule(&input)

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}

func FindRuleByResourceID(conn *eventbridge.EventBridge, id string) (*eventbridge.DescribeRuleOutput, error) {
	eventBusName, ruleName, err := RuleParseResourceID(id)

	if err != nil {
		return nil, err
	}

	return FindRuleByEventBusAndRuleNames(conn, eventBusName, ruleName)
}

func FindTarget(conn *eventbridge.EventBridge, busName, ruleName, targetId string) (*eventbridge.Target, error) {
	var result *eventbridge.Target
	err := ListAllTargetsForRulePages(conn, busName, ruleName, func(page *eventbridge.ListTargetsByRuleOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, t := range page.Targets {
			if targetId == aws.StringValue(t.Id) {
				result = t
				return false
			}
		}

		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("EventBridge FindTarget %q (\"%s/%s\") not found", targetId, busName, ruleName)
	}
	return result, nil
}
