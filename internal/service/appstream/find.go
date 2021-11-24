package appstream

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// FindStackByName Retrieve a appstream stack by name
func FindStackByName(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Stack, error) {
	input := &appstream.DescribeStacksInput{
		Names: []*string{aws.String(name)},
	}

	var stack *appstream.Stack
	resp, err := conn.DescribeStacksWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(resp.Stacks) > 1 {
		return nil, fmt.Errorf("got more than one stack with the name %s", name)
	}

	if len(resp.Stacks) == 1 {
		stack = resp.Stacks[0]
	}

	return stack, nil
}

// FindFleetByName Retrieve a appstream fleet by name
func FindFleetByName(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Fleet, error) {
	input := &appstream.DescribeFleetsInput{
		Names: []*string{aws.String(name)},
	}

	var fleet *appstream.Fleet
	resp, err := conn.DescribeFleetsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(resp.Fleets) > 1 {
		return nil, fmt.Errorf("got more than one fleet with the name %s", name)
	}

	if len(resp.Fleets) == 1 {
		fleet = resp.Fleets[0]
	}

	return fleet, nil
}

// FindImageBuilderByName Retrieve a appstream ImageBuilder by name
func FindImageBuilderByName(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.ImageBuilder, error) {
	input := &appstream.DescribeImageBuildersInput{
		Names: []*string{aws.String(name)},
	}

	var result *appstream.ImageBuilder

	err := describeImageBuildersPagesWithContext(ctx, conn, input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, imageBuilder := range page.ImageBuilders {
			if imageBuilder == nil {
				continue
			}
			if aws.StringValue(imageBuilder.Name) == name {
				result = imageBuilder
				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return result, nil
}

// FindUserByUserNameAndAuthType Retrieve a appstream fleet by Username and authentication type
func FindUserByUserNameAndAuthType(ctx context.Context, conn *appstream.AppStream, username, authType string) (*appstream.User, error) {
	input := &appstream.DescribeUsersInput{
		AuthenticationType: aws.String(authType),
	}

	var result *appstream.User

	err := describeUsersPagesWithContext(ctx, conn, input, func(page *appstream.DescribeUsersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, user := range page.Users {
			if user == nil {
				continue
			}
			if aws.StringValue(user.UserName) == username {
				result = user
				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return result, nil
}

// FindFleetStackAssociation Validates that a fleet has the named associated stack
func FindFleetStackAssociation(ctx context.Context, conn *appstream.AppStream, fleetName, stackName string) error {
	input := &appstream.ListAssociatedStacksInput{
		FleetName: aws.String(fleetName),
	}

	found := false
	err := listAssociatedStacksPagesWithContext(ctx, conn, input, func(page *appstream.ListAssociatedStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, name := range page.Names {
			if stackName == aws.StringValue(name) {
				found = true
				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		return &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return err
	}

	if !found {
		return &resource.NotFoundError{
			Message:     fmt.Sprintf("No stack %q associated with fleet %q", stackName, fleetName),
			LastRequest: input,
		}
	}

	return nil
}
