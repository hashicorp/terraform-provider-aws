package elbv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func FindListenerByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) (*elbv2.Listener, error) {
	input := &elbv2.DescribeListenersInput{
		ListenerArns: aws.StringSlice([]string{arn}),
	}

	var result *elbv2.Listener

	err := conn.DescribeListenersPagesWithContext(ctx, input, func(page *elbv2.DescribeListenersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, l := range page.Listeners {
			if l == nil {
				continue
			}

			if aws.StringValue(l.ListenerArn) == arn {
				result = l
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
