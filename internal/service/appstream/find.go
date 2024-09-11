// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindFleetByName Retrieve a appstream fleet by name
func FindFleetByName(ctx context.Context, conn *appstream.Client, name string) (*awstypes.Fleet, error) {
	input := &appstream.DescribeFleetsInput{
		Names: []string{name},
	}

	var fleet awstypes.Fleet
	resp, err := conn.DescribeFleets(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(resp.Fleets) > 1 {
		return nil, fmt.Errorf("got more than one fleet with the name %s", name)
	}

	if len(resp.Fleets) == 1 {
		fleet = resp.Fleets[0]
	}

	return &fleet, nil
}

func FindImageBuilderByName(ctx context.Context, conn *appstream.Client, name string) (*awstypes.ImageBuilder, error) {
	input := &appstream.DescribeImageBuildersInput{
		Names: []string{name},
	}

	return findImageBuilder(ctx, conn, input)
}

func findImageBuilders(ctx context.Context, conn *appstream.Client, input *appstream.DescribeImageBuildersInput) ([]awstypes.ImageBuilder, error) {
	var output []awstypes.ImageBuilder

	err := describeImageBuildersPages(ctx, conn, input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.ImageBuilders...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findImageBuilder(ctx context.Context, conn *appstream.Client, input *appstream.DescribeImageBuildersInput) (*awstypes.ImageBuilder, error) {
	output, err := findImageBuilders(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindUserByTwoPartKey(ctx context.Context, conn *appstream.Client, username, authType string) (*awstypes.User, error) {
	input := &appstream.DescribeUsersInput{
		AuthenticationType: awstypes.AuthenticationType(authType),
	}

	return findUser(ctx, conn, input, func(v *awstypes.User) bool {
		return aws.ToString(v.UserName) == username
	})
}

func findUsers(ctx context.Context, conn *appstream.Client, input *appstream.DescribeUsersInput, filter tfslices.Predicate[*awstypes.User]) ([]awstypes.User, error) {
	var output []awstypes.User

	err := describeUsersPages(ctx, conn, input, func(page *appstream.DescribeUsersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Users {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findUser(ctx context.Context, conn *appstream.Client, input *appstream.DescribeUsersInput, filter tfslices.Predicate[*awstypes.User]) (*awstypes.User, error) {
	output, err := findUsers(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

// FindFleetStackAssociation Validates that a fleet has the named associated stack
func FindFleetStackAssociation(ctx context.Context, conn *appstream.Client, fleetName, stackName string) error {
	input := &appstream.ListAssociatedStacksInput{
		FleetName: aws.String(fleetName),
	}

	found := false
	err := listAssociatedStacksPages(ctx, conn, input, func(page *appstream.ListAssociatedStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, name := range page.Names {
			if stackName == name {
				found = true
				return false
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

// findImages finds all images from a describe images input
func findImages(ctx context.Context, conn *appstream.Client, input *appstream.DescribeImagesInput) ([]awstypes.Image, error) {
	var output []awstypes.Image

	pages := appstream.NewDescribeImagesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Images...)
	}

	return output, nil
}
