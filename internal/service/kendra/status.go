package kendra

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kendra"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func statusIndex(ctx context.Context, conn *kendra.Kendra, Id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &kendra.DescribeIndexInput{
			Id: aws.String(Id),
		}

		output, err := conn.DescribeIndexWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, kendra.ErrCodeResourceNotFoundException) {
			return output, kendra.ErrCodeResourceNotFoundException, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
