package events

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindConnectionByName(ctx context.Context, conn *eventbridge.EventBridge, name string) (*eventbridge.DescribeConnectionOutput, error) {
	input := &eventbridge.DescribeConnectionInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeConnectionWithContext(ctx, input)

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
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindRuleByEventBusAndRuleNames(ctx context.Context, conn *eventbridge.EventBridge, eventBusName, ruleName string) (*eventbridge.DescribeRuleOutput, error) {
	input := eventbridge.DescribeRuleInput{
		Name: aws.String(ruleName),
	}

	if eventBusName != "" {
		input.EventBusName = aws.String(eventBusName)
	}

	output, err := conn.DescribeRuleWithContext(ctx, &input)

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
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindRuleByResourceID(ctx context.Context, conn *eventbridge.EventBridge, id string) (*eventbridge.DescribeRuleOutput, error) {
	eventBusName, ruleName, err := RuleParseResourceID(id)

	if err != nil {
		return nil, err
	}

	return FindRuleByEventBusAndRuleNames(ctx, conn, eventBusName, ruleName)
}

func FindTargetByThreePartKey(ctx context.Context, conn *eventbridge.EventBridge, busName, ruleName, targetID string) (*eventbridge.Target, error) {
	input := &eventbridge.ListTargetsByRuleInput{
		Rule:  aws.String(ruleName),
		Limit: aws.Int64(100), // Set limit to allowed maximum to prevent API throttling
	}

	if busName != "" {
		input.EventBusName = aws.String(busName)
	}

	var output *eventbridge.Target

	err := listTargetsByRulePages(ctx, conn, input, func(page *eventbridge.ListTargetsByRuleOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Targets {
			if targetID == aws.StringValue(v.Id) {
				output = v
				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, "ValidationException", eventbridge.ErrCodeResourceNotFoundException) || (err != nil && regexp.MustCompile(" not found$").MatchString(err.Error())) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{}
	}

	return output, nil
}
