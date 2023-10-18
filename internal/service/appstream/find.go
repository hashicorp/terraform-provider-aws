// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

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

func FindImageBuilderByName(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.ImageBuilder, error) {
	input := &appstream.DescribeImageBuildersInput{
		Names: aws.StringSlice([]string{name}),
	}

	output, err := findImageBuilder(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.Name) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findImageBuilders(ctx context.Context, conn *appstream.AppStream, input *appstream.DescribeImageBuildersInput) ([]*appstream.ImageBuilder, error) {
	var output []*appstream.ImageBuilder

	err := describeImageBuildersPages(ctx, conn, input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ImageBuilders {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findImageBuilder(ctx context.Context, conn *appstream.AppStream, input *appstream.DescribeImageBuildersInput) (*appstream.ImageBuilder, error) {
	output, err := findImageBuilders(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

// FindUserByUserNameAndAuthType Retrieve a appstream fleet by Username and authentication type
func FindUserByUserNameAndAuthType(ctx context.Context, conn *appstream.AppStream, username, authType string) (*appstream.User, error) {
	input := &appstream.DescribeUsersInput{
		AuthenticationType: aws.String(authType),
	}

	var result *appstream.User

	err := describeUsersPages(ctx, conn, input, func(page *appstream.DescribeUsersOutput, lastPage bool) bool {
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
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &retry.NotFoundError{
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
	err := listAssociatedStacksPages(ctx, conn, input, func(page *appstream.ListAssociatedStacksOutput, lastPage bool) bool {
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
		return &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return err
	}

	if !found {
		return &retry.NotFoundError{
			Message:     fmt.Sprintf("No stack %q associated with fleet %q", stackName, fleetName),
			LastRequest: input,
		}
	}

	return nil
}
