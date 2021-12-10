package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
)

func FindPullThroughCacheRuleByRepositoryPrefix(
	ctx context.Context,
	conn *ecr.ECR,
	repositoryPrefix string,
) (*ecr.PullThroughCacheRule, error) {
	input := ecr.DescribePullThroughCacheRulesInput{
		EcrRepositoryPrefixes: aws.StringSlice([]string{repositoryPrefix}),
	}

	output, err := conn.DescribePullThroughCacheRulesWithContext(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodePullThroughCacheRuleNotFoundException) {
			return nil, nil
		}

		return nil, err
	}

	if output == nil || len(output.PullThroughCacheRules) != 1 {
		return nil, nil
	}

	return output.PullThroughCacheRules[0], nil
}
