// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	domainUpgradeSuccessMinTimeout = 10 * time.Second
	domainUpgradeSuccessDelay      = 30 * time.Second
)

// UpgradeSucceeded waits for an Upgrade to return Success
func waitUpgradeSucceeded(ctx context.Context, conn *opensearchservice.OpenSearchService, name string, timeout time.Duration) (*opensearchservice.GetUpgradeStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{opensearchservice.UpgradeStatusInProgress},
		Target:     []string{opensearchservice.UpgradeStatusSucceeded},
		Refresh:    statusUpgradeStatus(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: domainUpgradeSuccessMinTimeout,
		Delay:      domainUpgradeSuccessDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*opensearchservice.GetUpgradeStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitForDomainCreation(ctx context.Context, conn *opensearchservice.OpenSearchService, domainName string, timeout time.Duration) error {
	var out *opensearchservice.DomainStatus
	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		var err error
		out, err = FindDomainByName(ctx, conn, domainName)
		if tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if !aws.BoolValue(out.Processing) && (out.Endpoint != nil || out.Endpoints != nil) {
			return nil
		}

		return retry.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for OpenSearch Domain to be created", domainName))
	}, tfresource.WithDelay(10*time.Minute), tfresource.WithPollInterval(10*time.Second))

	if tfresource.TimedOut(err) {
		out, err = FindDomainByName(ctx, conn, domainName)
		if err != nil {
			return fmt.Errorf("describing OpenSearch Domain: %w", err)
		}
		if !aws.BoolValue(out.Processing) && (out.Endpoint != nil || out.Endpoints != nil) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("waiting for OpenSearch Domain to be created: %w", err)
	}
	return nil
}

func waitForDomainUpdate(ctx context.Context, conn *opensearchservice.OpenSearchService, domainName string, timeout time.Duration) error {
	var out *opensearchservice.DomainStatus
	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
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
	}, tfresource.WithDelay(1*time.Minute), tfresource.WithPollInterval(10*time.Second))

	if tfresource.TimedOut(err) {
		out, err = FindDomainByName(ctx, conn, domainName)
		if err != nil {
			return fmt.Errorf("describing OpenSearch Domain: %w", err)
		}
		if !aws.BoolValue(out.Processing) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("waiting for OpenSearch Domain changes to be processed: %w", err)
	}
	return nil
}

func waitForDomainDelete(ctx context.Context, conn *opensearchservice.OpenSearchService, domainName string, timeout time.Duration) error {
	var out *opensearchservice.DomainStatus
	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
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

		return retry.RetryableError(fmt.Errorf("timeout while waiting for the OpenSearch Domain %q to be deleted", domainName))
	}, tfresource.WithDelay(10*time.Minute), tfresource.WithPollInterval(10*time.Second))

	if tfresource.TimedOut(err) {
		out, err = FindDomainByName(ctx, conn, domainName)
		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("describing OpenSearch Domain: %s", err)
		}
		if out != nil && !aws.BoolValue(out.Processing) {
			return nil
		}
	}

	if err != nil {
		return fmt.Errorf("waiting for OpenSearch Domain to be deleted: %s", err)
	}

	// opensearch maintains information about the domain in multiple (at least 2) places that need
	// to clear before it is really deleted - otherwise, requesting information about domain immediately
	// after delete will return info about just deleted domain
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
