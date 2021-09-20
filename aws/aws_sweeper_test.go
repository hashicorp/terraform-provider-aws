package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

const (
	SweepThrottlingRetryTimeout = 10 * time.Minute
)

const defaultSweeperAssumeRoleDurationSeconds = 3600

// sweeperAwsClients is a shared cache of regional AWSClient
// This prevents client re-initialization for every resource with no benefit.
var sweeperAwsClients map[string]interface{}

func TestMain(m *testing.M) {
	sweeperAwsClients = make(map[string]interface{})
	resource.TestMain(m)
}

// sharedClientForRegion returns a common AWSClient setup needed for the sweeper
// functions for a given region
func sharedClientForRegion(region string) (interface{}, error) {
	if client, ok := sweeperAwsClients[region]; ok {
		return client, nil
	}

	_, _, err := envvar.RequireOneOf([]string{envvar.AwsProfile, envvar.AwsAccessKeyId, envvar.AwsContainerCredentialsFullUri}, "credentials for running sweepers")
	if err != nil {
		return nil, err
	}

	if os.Getenv(envvar.AwsAccessKeyId) != "" {
		_, err := envvar.Require(envvar.AwsSecretAccessKey, "static credentials value when using "+envvar.AwsAccessKeyId)
		if err != nil {
			return nil, err
		}
	}

	conf := &Config{
		MaxRetries: 5,
		Region:     region,
	}

	if role := os.Getenv(envvar.TfAwsAssumeRoleARN); role != "" {
		conf.AssumeRoleARN = role

		conf.AssumeRoleDurationSeconds = defaultSweeperAssumeRoleDurationSeconds
		if v := os.Getenv(envvar.TfAwsAssumeRoleDuration); v != "" {
			d, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("environment variable %s: %w", envvar.TfAwsAssumeRoleDuration, err)
			}
			conf.AssumeRoleDurationSeconds = d
		}

		if v := os.Getenv(envvar.TfAwsAssumeRoleExternalID); v != "" {
			conf.AssumeRoleExternalID = v
		}

		if v := os.Getenv(envvar.TfAwsAssumeRoleSessionName); v != "" {
			conf.AssumeRoleSessionName = v
		}
	}

	// configures a default client for the region, using the above env vars
	client, err := conf.Client()
	if err != nil {
		return nil, fmt.Errorf("error getting AWS client: %w", err)
	}

	sweeperAwsClients[region] = client

	return client, nil
}

type testSweepResource struct {
	d        *schema.ResourceData
	meta     interface{}
	resource *schema.Resource
}

func NewTestSweepResource(resource *schema.Resource, d *schema.ResourceData, meta interface{}) *testSweepResource {
	return &testSweepResource{
		d:        d,
		meta:     meta,
		resource: resource,
	}
}

func testSweepResourceOrchestrator(sweepResources []*testSweepResource) error {
	return testSweepResourceOrchestratorContext(context.Background(), sweepResources, 0*time.Millisecond, 0*time.Millisecond, 0*time.Millisecond, 0*time.Millisecond, SweepThrottlingRetryTimeout)
}

func testSweepResourceOrchestratorContext(ctx context.Context, sweepResources []*testSweepResource, delay time.Duration, delayRand time.Duration, minTimeout time.Duration, pollInterval time.Duration, timeout time.Duration) error {
	var g multierror.Group

	for _, sweepResource := range sweepResources {
		sweepResource := sweepResource

		g.Go(func() error {
			err := tfresource.RetryConfigContext(ctx, delay, delayRand, minTimeout, pollInterval, timeout, func() *resource.RetryError {
				err := testAccDeleteResource(sweepResource.resource, sweepResource.d, sweepResource.meta)

				if err != nil {
					if strings.Contains(err.Error(), "Throttling") {
						log.Printf("[INFO] While sweeping resource (%s), encountered throttling error (%s). Retrying...", sweepResource.d.Id(), err)
						return resource.RetryableError(err)
					}

					return resource.NonRetryableError(err)
				}

				return nil
			})

			if tfresource.TimedOut(err) {
				err = testAccDeleteResource(sweepResource.resource, sweepResource.d, sweepResource.meta)
			}

			return err
		})
	}

	return g.Wait().ErrorOrNil()
}

// Check sweeper API call error for reasons to skip sweeping
// These include missing API endpoints and unsupported API calls
func testSweepSkipSweepError(err error) bool {
	// Ignore missing API endpoints
	if tfawserr.ErrMessageContains(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if tfawserr.ErrMessageContains(err, "UnsupportedOperation", "") {
		return true
	}
	// Ignore more unsupported API calls
	// InvalidParameterValue: Use of cache security groups is not permitted in this API version for your account.
	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "not permitted in this API version for your account") {
		return true
	}
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version") {
		return true
	}
	// GovCloud has endpoints that respond with (no message provided):
	// AccessDeniedException:
	// Since acceptance test sweepers are best effort and this response is very common,
	// we allow bypassing this error globally instead of individual test sweeper fixes.
	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "") {
		return true
	}
	// Example: BadRequestException: vpc link not supported for region us-gov-west-1
	if tfawserr.ErrMessageContains(err, "BadRequestException", "not supported") {
		return true
	}
	// Example: InvalidAction: The action DescribeTransitGatewayAttachments is not valid for this web service
	if tfawserr.ErrMessageContains(err, "InvalidAction", "is not valid") {
		return true
	}
	// For example from GovCloud SES.SetActiveReceiptRuleSet.
	if tfawserr.ErrMessageContains(err, "InvalidAction", "Unavailable Operation") {
		return true
	}
	// For example from us-west-2 Route53 key signing key
	if tfawserr.ErrMessageContains(err, "InvalidKeySigningKeyStatus", "cannot be deleted because") {
		return true
	}
	// For example from us-west-2 Route53 zone
	if tfawserr.ErrMessageContains(err, "KeySigningKeyInParentDSRecord", "Due to DNS lookup failure") {
		return true
	}
	return false
}


