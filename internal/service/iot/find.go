package iot

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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
