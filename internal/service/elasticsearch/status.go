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
)

const (
	UpgradeStatusUnknown = "Unknown"
	ConfigStatusNotFound = "NotFound"
	ConfigStatusUnknown  = "Unknown"
	ConfigStatusExists   = "Exists"
)

func statusUpgradeStatus(ctx context.Context, conn *elasticsearch.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.GetUpgradeStatus(ctx, &elasticsearch.GetUpgradeStatusInput{
			DomainName: aws.String(name),
		})
		if err != nil {
			return nil, UpgradeStatusUnknown, err
		}

		// Elasticsearch upgrades consist of multiple steps:
		// https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-version-migration.html
		// Prevent false positive completion where the UpgradeStep is not the final UPGRADE step.
		if out.StepStatus == awstypes.UpgradeStatusSucceeded && out.UpgradeStep != awstypes.UpgradeStepUpgrade {
			return out, string(awstypes.UpgradeStatusInProgress), nil
		}

		return out, string(out.StepStatus), nil
	}
}

func domainConfigStatus(ctx context.Context, conn *elasticsearch.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.DescribeElasticsearchDomainConfig(ctx, &elasticsearch.DescribeElasticsearchDomainConfigInput{
			DomainName: aws.String(name),
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			// if first return value is nil, WaitForState treats as not found - here not found is treated differently
			return "not nil", ConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, ConfigStatusUnknown, err
		}

		return out, ConfigStatusExists, nil
	}
}
