package opensearch

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	domainUpgradeSuccessMinTimeout = 10 * time.Second
	domainUpgradeSuccessDelay      = 30 * time.Second
)

// UpgradeSucceeded waits for an Upgrade to return Success
func waitUpgradeSucceeded(conn *opensearchservice.OpenSearchService, name string, timeout time.Duration) (*opensearchservice.GetUpgradeStatusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{opensearchservice.UpgradeStatusInProgress},
		Target:     []string{opensearchservice.UpgradeStatusSucceeded},
		Refresh:    statusUpgradeStatus(conn, name),
		Timeout:    timeout,
		MinTimeout: domainUpgradeSuccessMinTimeout,
		Delay:      domainUpgradeSuccessDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*opensearchservice.GetUpgradeStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func WaitForDomainCreation(conn *opensearchservice.OpenSearchService, domainName string, timeout time.Duration) error {
	var out *opensearchservice.DomainStatus
	err := resource.Retry(timeout, func() *resource.RetryError {
		var err error
		out, err = FindDomainByName(conn, domainName)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if !aws.BoolValue(out.Processing) && (out.Endpoint != nil || out.Endpoints != nil) {
			return nil
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for the domain to be created", domainName))
	})
	if tfresource.TimedOut(err) {
		out, err = FindDomainByName(conn, domainName)
		if err != nil {
			return fmt.Errorf("Error describing OpenSearch domain: %w", err)
		}
		if !aws.BoolValue(out.Processing) && (out.Endpoint != nil || out.Endpoints != nil) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error waiting for OpenSearch domain to be created: %w", err)
	}
	return nil
}

func waitForDomainUpdate(conn *opensearchservice.OpenSearchService, domainName string, timeout time.Duration) error {
	var out *opensearchservice.DomainStatus
	err := resource.Retry(timeout, func() *resource.RetryError {
		var err error
		out, err = FindDomainByName(conn, domainName)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if !aws.BoolValue(out.Processing) {
			return nil
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for changes to be processed", domainName))
	})
	if tfresource.TimedOut(err) {
		out, err = FindDomainByName(conn, domainName)
		if err != nil {
			return fmt.Errorf("Error describing OpenSearch domain: %w", err)
		}
		if !aws.BoolValue(out.Processing) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error waiting for OpenSearch domain changes to be processed: %w", err)
	}
	return nil
}

func waitForDomainDelete(conn *opensearchservice.OpenSearchService, domainName string, timeout time.Duration) error {
	var out *opensearchservice.DomainStatus
	err := resource.Retry(timeout, func() *resource.RetryError {
		var err error
		out, err = FindDomainByName(conn, domainName)

		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		if out != nil && !aws.BoolValue(out.Processing) {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("timeout while waiting for the OpenSearch domain %q to be deleted", domainName))
	})

	if tfresource.TimedOut(err) {
		out, err = FindDomainByName(conn, domainName)
		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("Error describing OpenSearch domain: %s", err)
		}
		if out != nil && !aws.BoolValue(out.Processing) {
			return nil
		}
	}

	if err != nil {
		return fmt.Errorf("Error waiting for OpenSearch domain to be deleted: %s", err)
	}

	// opensearch maintains information about the domain in multiple (at least 2) places that need
	// to clear before it is really deleted - otherwise, requesting information about domain immediately
	// after delete will return info about just deleted domain
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{ConfigStatusUnknown, ConfigStatusExists},
		Target:                    []string{ConfigStatusNotFound},
		Refresh:                   domainConfigStatus(conn, domainName),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return nil
}
