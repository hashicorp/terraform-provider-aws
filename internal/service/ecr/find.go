package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindPullThroughCacheRuleByRepositoryPrefix(ctx context.Context, conn *ecr.ECR, repositoryPrefix string) (*ecr.PullThroughCacheRule, error) {
	input := ecr.DescribePullThroughCacheRulesInput{
		EcrRepositoryPrefixes: aws.StringSlice([]string{repositoryPrefix}),
	}

	output, err := conn.DescribePullThroughCacheRulesWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, ecr.ErrCodePullThroughCacheRuleNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PullThroughCacheRules) == 0 || output.PullThroughCacheRules[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PullThroughCacheRules); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.PullThroughCacheRules[0], nil
}

func FindImageDetails(conn *ecr.ECR, input *ecr.DescribeImagesInput) ([]*ecr.ImageDetail, error) {
	var details []*ecr.ImageDetail
	err := conn.DescribeImagesPages(input, func(output *ecr.DescribeImagesOutput, lastPage bool) bool {
		if output == nil {
			return !lastPage
		}

		details = append(details, output.ImageDetails...)

		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	return details, nil
}
