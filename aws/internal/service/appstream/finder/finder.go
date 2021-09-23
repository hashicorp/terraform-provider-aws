package finder

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
)

// StackByName Retrieve a appstream stack by name
func StackByName(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Stack, error) {
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

// FleetByName Retrieve a appstream fleet by name
func FleetByName(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Fleet, error) {
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

// ImageBuilderByName Retrieve a appstream ImageBuilder by name
func ImageBuilderByName(conn *appstream.AppStream, name string) (*appstream.ImageBuilder, error) {
	input := &appstream.DescribeImageBuildersInput{
		Names: []*string{aws.String(name)},
	}

	var imageBuilder *appstream.ImageBuilder
	resp, err := conn.DescribeImageBuilders(input)
	if err != nil {
		return nil, err
	}

	if len(resp.ImageBuilders) > 1 {
		return nil, fmt.Errorf("got more than one image builder with the name %s", name)
	}

	if len(resp.ImageBuilders) == 1 {
		imageBuilder = resp.ImageBuilders[0]
	}

	return imageBuilder, nil
}
