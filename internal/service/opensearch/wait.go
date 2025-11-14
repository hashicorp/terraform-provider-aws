// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfretry "github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	domainUpgradeSuccessMinTimeout = 10 * time.Second
	domainUpgradeSuccessDelay      = 30 * time.Second
)

func checkAndReturnChangeProgressError(ctx context.Context, conn *opensearch.Client, domainName string, details *awstypes.ChangeProgressDetails) error {
	if details != nil && (details.ConfigChangeStatus == awstypes.ConfigChangeStatusValidationFailed || details.ConfigChangeStatus == awstypes.ConfigChangeStatusCancelled) {
		return getChangeProgressError(ctx, conn, domainName, details.ChangeId)
	}
	return nil
}

func getChangeProgressError(ctx context.Context, conn *opensearch.Client, domainName string, changeId *string) error {
	input := &opensearch.DescribeDomainChangeProgressInput{
		DomainName: aws.String(domainName),
	}
	if changeId != nil {
		input.ChangeId = changeId
	}

	output, err := conn.DescribeDomainChangeProgress(ctx, input)
	if err != nil {
		return fmt.Errorf("domain %s: unable to get change progress details: %w", domainName, err)
	}

	if output.ChangeProgressStatus == nil {
		return fmt.Errorf("domain %s: unexpected API response, ChangeProgressStatus is nil", domainName)
	}

	status := output.ChangeProgressStatus
	var failedStages []string
	for _, stage := range status.ChangeProgressStages {
		if stage.Status != nil && *stage.Status == "FAILED" {
			desc := aws.ToString(stage.Description)
			if desc == "" {
				desc = aws.ToString(stage.Name)
			}
			failedStages = append(failedStages, desc)
		}
	}

	if len(failedStages) > 0 {
		return fmt.Errorf("domain %s: %s: %s", domainName, status.ConfigChangeStatus, strings.Join(failedStages, "; "))
	}

	return fmt.Errorf("domain %s: %s", domainName, status.ConfigChangeStatus)
}

func waitUpgradeSucceeded(ctx context.Context, conn *opensearch.Client, name string, timeout time.Duration) (*opensearch.GetUpgradeStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.UpgradeStatusInProgress),
		Target:     enum.Slice(awstypes.UpgradeStatusSucceeded),
		Refresh:    statusUpgradeStatus(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: domainUpgradeSuccessMinTimeout,
		Delay:      domainUpgradeSuccessDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*opensearch.GetUpgradeStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func waitForDomainCreation(ctx context.Context, conn *opensearch.Client, domainName string, timeout time.Duration) error {
	var out *awstypes.DomainStatus
	err := tfresource.Retry(ctx, timeout, func(ctx context.Context) *tfresource.RetryError {
		var err error
		out, err = findDomainByName(ctx, conn, domainName)
		if tfresource.NotFound(err) {
			return tfresource.RetryableError(err)
		}
		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		if err := checkAndReturnChangeProgressError(ctx, conn, domainName, out.ChangeProgressDetails); err != nil {
			return tfresource.NonRetryableError(err)
		}

		if !aws.ToBool(out.Processing) && (out.Endpoint != nil || out.Endpoints != nil) {
			return nil
		}

		return tfresource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for OpenSearch Domain to be created", domainName))
	}, tfresource.WithDelay(10*time.Minute), tfresource.WithPollInterval(10*time.Second))

	if err != nil {
		return fmt.Errorf("waiting for OpenSearch Domain to be created: %w", err)
	}

	return nil
}

func waitForDomainUpdate(ctx context.Context, conn *opensearch.Client, domainName string, timeout time.Duration) error {
	var out *awstypes.DomainStatus
	err := tfresource.Retry(ctx, timeout, func(ctx context.Context) *tfresource.RetryError {
		var err error
		out, err = findDomainByName(ctx, conn, domainName)
		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		if err := checkAndReturnChangeProgressError(ctx, conn, domainName, out.ChangeProgressDetails); err != nil {
			return tfresource.NonRetryableError(err)
		}

		if !aws.ToBool(out.Processing) {
			return nil
		}

		return tfresource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for changes to be processed", domainName))
	}, tfresource.WithDelay(1*time.Minute), tfresource.WithPollInterval(10*time.Second))

	if err != nil {
		return fmt.Errorf("waiting for OpenSearch Domain changes to be processed: %w", err)
	}

	return nil
}

func waitForDomainDelete(ctx context.Context, conn *opensearch.Client, domainName string, timeout time.Duration) error {
	var out *awstypes.DomainStatus
	err := tfresource.Retry(ctx, timeout, func(ctx context.Context) *tfresource.RetryError {
		var err error
		out, err = findDomainByName(ctx, conn, domainName)

		if err != nil {
			if tfretry.NotFound(err) {
				return nil
			}
			return tfresource.NonRetryableError(err)
		}

		if err := checkAndReturnChangeProgressError(ctx, conn, domainName, out.ChangeProgressDetails); err != nil {
			return tfresource.NonRetryableError(err)
		}

		if out != nil && !aws.ToBool(out.Processing) {
			return nil
		}

		return tfresource.RetryableError(fmt.Errorf("timeout while waiting for the OpenSearch Domain %q to be deleted", domainName))
	}, tfresource.WithDelay(10*time.Minute), tfresource.WithPollInterval(10*time.Second))

	if err != nil {
		return fmt.Errorf("waiting for OpenSearch Domain to be deleted: %w", err)
	}

	// opensearch maintains information about the domain in multiple (at least 2) places that need
	// to clear before it is really deleted - otherwise, requesting information about domain immediately
	// after delete will return info about just deleted domain
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{configStatusUnknown, configStatusExists},
		Target:                    []string{configStatusNotFound},
		Refresh:                   domainConfigStatus(ctx, conn, domainName),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	return err
}
