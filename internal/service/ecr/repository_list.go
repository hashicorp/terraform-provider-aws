// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_ecr_repository")
func repositoryResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceRepository(), framework.SDKListSpec[awstypes.Repository, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.Repository, error], diag.Diagnostics) {
			var input ecr.DescribeRepositoriesInput

			return listRepositories(ctx, c.ECRClient(ctx), &input), nil
		},
		ID: func(r awstypes.Repository, _ framework.RegionQuery) string {
			return aws.ToString(r.RepositoryName)
		},
		Flatten: resourceRepositoryFlatten,
	})
}

func listRepositories(ctx context.Context, conn *ecr.Client, input *ecr.DescribeRepositoriesInput) iter.Seq2[awstypes.Repository, error] {
	return tfiter.PageItems[ecr.Options](ctx, "listing ECR Repository resources", ecr.NewDescribeRepositoriesPaginator(conn, input),
		func(o *ecr.DescribeRepositoriesOutput) []awstypes.Repository { return o.Repositories })
}
