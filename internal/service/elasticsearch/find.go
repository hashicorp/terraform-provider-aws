// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elasticsearch "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findDomainByName(ctx context.Context, conn *elasticsearch.Client, name string) (*awstypes.ElasticsearchDomainStatus, error) {
	input := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(name),
	}

	return findDomain(ctx, conn, input)
}

func findDomain(ctx context.Context, conn *elasticsearch.Client, input *elasticsearch.DescribeElasticsearchDomainInput) (*awstypes.ElasticsearchDomainStatus, error) {
	output, err := conn.DescribeElasticsearchDomain(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DomainStatus, nil
}

func findDomainConfigByName(ctx context.Context, conn *elasticsearch.Client, name string) (*awstypes.ElasticsearchDomainConfig, error) {
	input := &elasticsearch.DescribeElasticsearchDomainConfigInput{
		DomainName: aws.String(name),
	}

	return findDomainConfig(ctx, conn, input)
}

func findDomainConfig(ctx context.Context, conn *elasticsearch.Client, input *elasticsearch.DescribeElasticsearchDomainConfigInput) (*awstypes.ElasticsearchDomainConfig, error) {
	output, err := conn.DescribeElasticsearchDomainConfig(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DomainConfig, nil
}
