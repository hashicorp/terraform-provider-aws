package appstream

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// findStackByName Retrieve a appstream stack by name
func findStackByName(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Stack, error) {
	input := &appstream.DescribeStacksInput{
		Names: []*string{aws.String(name)},
	}

	var stack *appstream.Stack
	resp, err := conn.DescribeStacksWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(resp.Stacks) > 1 {
		return nil, fmt.Errorf("[ERROR] got more than one stack with the name %s", name)
	}

	if len(resp.Stacks) == 1 {
		stack = resp.Stacks[0]
	}

	return stack, nil
}
