// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_ecr_repository", sweepRepositories)
}

func sweepRepositories(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ECRClient(ctx)
	var input ecr.DescribeRepositoriesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ecr.NewDescribeRepositoriesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Repositories {
			r := resourceRepository()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.RepositoryName))
			d.Set(names.AttrForceDelete, true)
			d.Set("registry_id", v.RegistryId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
