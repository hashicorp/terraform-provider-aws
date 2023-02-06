package sweep

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	ThrottlingRetryTimeout = 10 * time.Minute

	ResourcePrefix = "tf-acc-test"
)

const defaultSweeperAssumeRoleDurationSeconds = 3600

// SweeperClients is a shared cache of regional conns.AWSClient
// This prevents client re-initialization for every resource with no benefit.
var SweeperClients map[string]interface{}

// SharedRegionalSweepClient returns a common conns.AWSClient setup needed for the sweeper
// functions for a given region
func SharedRegionalSweepClient(region string) (interface{}, error) {
	return SharedRegionalSweepClientWithContext(Context(region), region)
}

func SharedRegionalSweepClientWithContext(ctx context.Context, region string) (interface{}, error) {
	if client, ok := SweeperClients[region]; ok {
		return client, nil
	}

	_, _, err := envvar.RequireOneOf([]string{envvar.Profile, envvar.AccessKeyId, envvar.ContainerCredentialsFullURI}, "credentials for running sweepers")
	if err != nil {
		return nil, err
	}

	if os.Getenv(envvar.AccessKeyId) != "" {
		_, err := envvar.Require(envvar.SecretAccessKey, "static credentials value when using "+envvar.AccessKeyId)
		if err != nil {
			return nil, err
		}
	}

	conf := &conns.Config{
		MaxRetries:       5,
		Region:           region,
		SuppressDebugLog: true,
	}

	if role := os.Getenv(envvar.AssumeRoleARN); role != "" {
		conf.AssumeRole.RoleARN = role

		conf.AssumeRole.Duration = time.Duration(defaultSweeperAssumeRoleDurationSeconds) * time.Second
		if v := os.Getenv(envvar.AssumeRoleDuration); v != "" {
			d, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("environment variable %s: %w", envvar.AssumeRoleDuration, err)
			}
			conf.AssumeRole.Duration = time.Duration(d) * time.Second
		}

		if v := os.Getenv(envvar.AssumeRoleExternalID); v != "" {
			conf.AssumeRole.ExternalID = v
		}

		if v := os.Getenv(envvar.AssumeRoleSessionName); v != "" {
			conf.AssumeRole.SessionName = v
		}
	}

	// configures a default client for the region, using the above env vars
	client, diags := conf.ConfigureProvider(ctx, &conns.AWSClient{})

	if diags.HasError() {
		return nil, fmt.Errorf("getting AWS client: %#v", diags)
	}

	SweeperClients[region] = client

	return client, nil
}

type Sweepable interface {
	Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error
}

type SweepResource struct {
	d        *schema.ResourceData
	meta     interface{}
	resource *schema.Resource
}

func NewSweepResource(resource *schema.Resource, d *schema.ResourceData, meta interface{}) *SweepResource {
	return &SweepResource{
		d:        d,
		meta:     meta,
		resource: resource,
	}
}

func (sr *SweepResource) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	err := tfresource.Retry(ctx, timeout, func() *resource.RetryError {
		err := DeleteResource(ctx, sr.resource, sr.d, sr.meta)

		if err != nil {
			if strings.Contains(err.Error(), "Throttling") {
				log.Printf("[INFO] While sweeping resource (%s), encountered throttling error (%s). Retrying...", sr.d.Id(), err)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	}, optFns...)

	if tfresource.TimedOut(err) {
		err = DeleteResource(ctx, sr.resource, sr.d, sr.meta)
	}

	return err
}

func SweepOrchestrator(sweepables []Sweepable) error {
	return SweepOrchestratorWithContext(context.Background(), sweepables)
}

func SweepOrchestratorWithContext(ctx context.Context, sweepables []Sweepable, optFns ...tfresource.OptionsFunc) error {
	var g multierror.Group

	for _, sweepable := range sweepables {
		sweepable := sweepable

		g.Go(func() error {
			return sweepable.Delete(ctx, ThrottlingRetryTimeout, optFns...)
		})
	}

	return g.Wait().ErrorOrNil()
}

// Check sweeper API call error for reasons to skip sweeping
// These include missing API endpoints and unsupported API calls
func SkipSweepError(err error) bool {
	// Ignore missing API endpoints for AWS SDK for Go v1
	if tfawserr.ErrMessageContains(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if tfawserr.ErrCodeEquals(err, "UnsupportedOperation") {
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
	if tfawserr.ErrCodeEquals(err, "AccessDeniedException") {
		return true
	}
	// Example: BadRequestException: vpc link not supported for region us-gov-west-1
	if tfawserr.ErrMessageContains(err, "BadRequestException", "not supported") {
		return true
	}
	// Example: InvalidAction: InvalidAction: Operation (ListPlatformApplications) is not supported in this region
	if tfawserr.ErrMessageContains(err, "InvalidAction", "is not supported in this region") {
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
	// For example from us-gov-west-1 EventBridge archive
	if tfawserr.ErrMessageContains(err, "UnknownOperationException", "Operation is disabled in this region") {
		return true
	}
	// For example from us-east-1 SageMaker
	if tfawserr.ErrMessageContains(err, "UnknownOperationException", "The requested operation is not supported in the called region") {
		return true
	}
	// For example from us-west-2 ECR public repository
	if tfawserr.ErrMessageContains(err, "UnsupportedCommandException", "command is only supported in") {
		return true
	}
	// For example from us-west-1 EMR studio
	if tfawserr.ErrMessageContains(err, "ValidationException", "Account is not whitelisted to use this feature") {
		return true
	}

	// Ignore missing API endpoints for AWS SDK for Go v2
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.IsNotFound
	}
	return false
}

func DeleteResource(ctx context.Context, resource *schema.Resource, d *schema.ResourceData, meta any) error {
	if resource.DeleteContext != nil || resource.DeleteWithoutTimeout != nil {
		var diags diag.Diagnostics

		if resource.DeleteContext != nil {
			diags = resource.DeleteContext(ctx, d, meta)
		} else {
			diags = resource.DeleteWithoutTimeout(ctx, d, meta)
		}

		return sdkdiag.DiagnosticsError(diags)
	}

	return resource.Delete(d, meta)
}

func ReadResource(ctx context.Context, resource *schema.Resource, d *schema.ResourceData, meta any) error {
	if resource.ReadContext != nil || resource.ReadWithoutTimeout != nil {
		var diags diag.Diagnostics

		if resource.ReadContext != nil {
			diags = resource.ReadContext(ctx, d, meta)
		} else {
			diags = resource.ReadWithoutTimeout(ctx, d, meta)
		}

		return sdkdiag.DiagnosticsError(diags)
	}

	return resource.Read(d, meta)
}

func Partition(region string) string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		return partition.ID()
	}
	return "aws"
}

func PartitionDNSSuffix(region string) string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		return partition.DNSSuffix()
	}
	return "amazonaws.com"
}
