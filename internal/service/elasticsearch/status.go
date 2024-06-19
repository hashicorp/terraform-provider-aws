// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elasticsearch "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	UpgradeStatusUnknown = "Unknown"
	ConfigStatusNotFound = "NotFound"
	ConfigStatusUnknown  = "Unknown"
	ConfigStatusExists   = "Exists"
)

func statusUpgradeStatus(ctx context.Context, conn *elasticsearch.ElasticsearchService, name string) retry.StateRefreshFunc {
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
		if aws.ToString(out.StepStatus) == elasticsearch.UpgradeStatusSucceeded && string(out.UpgradeStep) != elasticsearch.UpgradeStepUpgrade {
			return out, elasticsearch.UpgradeStatusInProgress, nil
		}

		return out, aws.ToString(out.StepStatus), nil
	}
}

func domainConfigStatus(ctx context.Context, conn *elasticsearch.ElasticsearchService, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.DescribeElasticsearchDomainConfig(ctx, &elasticsearch.DescribeElasticsearchDomainConfigInput{
			DomainName: aws.String(name),
		})

		if tfawserr.ErrCodeEquals(err, elasticsearch.ErrCodeResourceNotFoundException) {
			// if first return value is nil, WaitForState treats as not found - here not found is treated differently
			return "not nil", ConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, ConfigStatusUnknown, err
		}

		return out, ConfigStatusExists, nil
	}
}
