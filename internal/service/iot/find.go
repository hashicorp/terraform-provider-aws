package iot

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAuthorizerByName(conn *iot.IoT, name string) (*iot.AuthorizerDescription, error) {
	input := &iot.DescribeAuthorizerInput{
		AuthorizerName: aws.String(name),
	}

	output, err := conn.DescribeAuthorizer(input)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AuthorizerDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AuthorizerDescription, nil
}

func FindThingByName(conn *iot.IoT, name string) (*iot.DescribeThingOutput, error) {
	input := &iot.DescribeThingInput{
		ThingName: aws.String(name),
	}

	output, err := conn.DescribeThing(input)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
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

func FindThingGroupByName(conn *iot.IoT, name string) (*iot.DescribeThingGroupOutput, error) {
	input := &iot.DescribeThingGroupInput{
		ThingGroupName: aws.String(name),
	}

	output, err := conn.DescribeThingGroup(input)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
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

func FindThingGroupMembership(conn *iot.IoT, thingGroupName, thingName string) error {
	input := &iot.ListThingGroupsForThingInput{
		ThingName: aws.String(thingName),
	}

	var v *iot.GroupNameAndArn

	err := conn.ListThingGroupsForThingPages(input, func(page *iot.ListThingGroupsForThingOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, group := range page.ThingGroups {
			if aws.StringValue(group.GroupName) == thingGroupName {
				v = group

				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if v == nil {
		return tfresource.NewEmptyResultError(input)
	}

	return nil
}

func FindTopicRuleByName(conn *iot.IoT, name string) (*iot.GetTopicRuleOutput, error) {
	// GetTopicRule returns unhelpful errors such as
	//	"An error occurred (UnauthorizedException) when calling the GetTopicRule operation: Access to topic rule 'xxxxxxxx' was denied"
	// when querying for a rule that doesn't exist.
	var rule *iot.TopicRuleListItem

	err := conn.ListTopicRulesPages(&iot.ListTopicRulesInput{}, func(page *iot.ListTopicRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Rules {
			if v == nil {
				continue
			}

			if aws.StringValue(v.RuleName) == name {
				rule = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if rule == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	input := &iot.GetTopicRuleInput{
		RuleName: aws.String(name),
	}

	output, err := conn.GetTopicRule(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindTopicRuleDestinationByARN(ctx context.Context, conn *iot.IoT, arn string) (*iot.TopicRuleDestination, error) {
	// GetTopicRuleDestination returns unhelpful errors such as
	//	"UnauthorizedException: Access to TopicRuleDestination 'arn:aws:iot:us-west-2:123456789012:ruledestination/vpc/f267138a-7383-4670-9e44-a7fe2f48af5e' was denied"
	// when querying for a rule destination that doesn't exist.
	var destination *iot.TopicRuleDestinationSummary

	err := conn.ListTopicRuleDestinationsPages(&iot.ListTopicRuleDestinationsInput{}, func(page *iot.ListTopicRuleDestinationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DestinationSummaries {
			if v == nil {
				continue
			}

			if aws.StringValue(v.Arn) == arn {
				destination = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if destination == nil {
		return nil, tfresource.NewEmptyResultError(destination)
	}

	input := &iot.GetTopicRuleDestinationInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetTopicRuleDestinationWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.TopicRuleDestination == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TopicRuleDestination, nil
}
