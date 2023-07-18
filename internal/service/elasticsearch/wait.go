// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	domainUpgradeSuccessMinTimeout = 10 * time.Second
	domainUpgradeSuccessDelay      = 30 * time.Second
)

// UpgradeSucceeded waits for an Upgrade to return Success
func waitUpgradeSucceeded(ctx context.Context, conn *elasticsearch.ElasticsearchService, name string, timeout time.Duration) (*elasticsearch.GetUpgradeStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{elasticsearch.UpgradeStatusInProgress},
		Target:     []string{elasticsearch.UpgradeStatusSucceeded},
		Refresh:    statusUpgradeStatus(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: domainUpgradeSuccessMinTimeout,
		Delay:      domainUpgradeSuccessDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elasticsearch.GetUpgradeStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitForDomainCreation(ctx context.Context, conn *elasticsearch.ElasticsearchService, domainName string, timeout time.Duration) error {
	var out *elasticsearch.ElasticsearchDomainStatus
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		var err error
		out, err = FindDomainByName(ctx, conn, domainName)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if !aws.BoolValue(out.Processing) && (out.Endpoint != nil || out.Endpoints != nil) {
			return nil
		}

		return retry.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for the domain to be created", domainName))
	})
	if tfresource.TimedOut(err) {
		out, err = FindDomainByName(ctx, conn, domainName)
		if err != nil {
			return fmt.Errorf("describing Elasticsearch domain: %w", err)
		}
		if !aws.BoolValue(out.Processing) && (out.Endpoint != nil || out.Endpoints != nil) {
			return nil
		}
	}

	return err
}

func waitForDomainUpdate(ctx context.Context, conn *elasticsearch.ElasticsearchService, domainName string, timeout time.Duration) error {
	var out *elasticsearch.ElasticsearchDomainStatus
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		var err error
		out, err = FindDomainByName(ctx, conn, domainName)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if !aws.BoolValue(out.Processing) {
			return nil
		}

		return retry.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for changes to be processed", domainName))
	})
	if tfresource.TimedOut(err) {
		out, err = FindDomainByName(ctx, conn, domainName)
		if err != nil {
			return fmt.Errorf("describing Elasticsearch domain: %w", err)
		}
		if !aws.BoolValue(out.Processing) {
			return nil
		}
	}

	return err
}

func waitForDomainDelete(ctx context.Context, conn *elasticsearch.ElasticsearchService, domainName string, timeout time.Duration) error {
	var out *elasticsearch.ElasticsearchDomainStatus
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		var err error
		out, err = FindDomainByName(ctx, conn, domainName)

		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return retry.NonRetryableError(err)
		}

		if out != nil && !aws.BoolValue(out.Processing) {
			return nil
		}

		return retry.RetryableError(fmt.Errorf("timeout while waiting for the domain %q to be deleted", domainName))
	})
	if tfresource.TimedOut(err) {
		out, err = FindDomainByName(ctx, conn, domainName)
		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("describing Elasticsearch domain: %s", err)
		}
		if out != nil && !aws.BoolValue(out.Processing) {
			return nil
		}
	}

	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		Pending:                   []string{ConfigStatusUnknown, ConfigStatusExists},
		Target:                    []string{ConfigStatusNotFound},
		Refresh:                   domainConfigStatus(ctx, conn, domainName),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	return err
}
