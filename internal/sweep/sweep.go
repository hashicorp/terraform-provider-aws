// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sweep

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ThrottlingRetryTimeout = 10 * time.Minute

	ResourcePrefix = "tf-acc-test"
)

const defaultSweeperAssumeRoleDurationSeconds = 3600

// ServicePackages is set in TestMain in order to break an import cycle.
var ServicePackages []conns.ServicePackage

// sweeperClients is a shared cache of regional conns.AWSClient
// This prevents client re-initialization for every resource with no benefit.
var sweeperClients map[string]*conns.AWSClient = make(map[string]*conns.AWSClient)

// SharedRegionalSweepClient returns a common conns.AWSClient setup needed for the sweeper functions for a given Region.
func SharedRegionalSweepClient(ctx context.Context, region string) (*conns.AWSClient, error) {
	if client, ok := sweeperClients[region]; ok {
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

	meta := new(conns.AWSClient)
	servicePackageMap := make(map[string]conns.ServicePackage)
	for _, sp := range ServicePackages {
		servicePackageName := sp.ServicePackageName()
		servicePackageMap[servicePackageName] = sp
	}
	meta.ServicePackages = servicePackageMap

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
	client, diags := conf.ConfigureProvider(ctx, meta)

	if diags.HasError() {
		return nil, fmt.Errorf("getting AWS client: %#v", diags)
	}

	sweeperClients[region] = client

	return client, nil
}

type Sweepable interface {
	Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error
}

func SweepOrchestrator(ctx context.Context, sweepables []Sweepable, optFns ...tfresource.OptionsFunc) error {
	if len(sweepables) == 0 {
		tflog.Info(ctx, "No resources to sweep")
	}

	var g multierror.Group

	for _, sweepable := range sweepables {
		sweepable := sweepable

		g.Go(func() error {
			return sweepable.Delete(ctx, ThrottlingRetryTimeout, optFns...)
		})
	}

	return g.Wait().ErrorOrNil()
}

// Deprecated: Use awsv1.SkipSweepError
var SkipSweepError = awsv1.SkipSweepError

func Partition(region string) string {
	return names.PartitionForRegion(region)
}

func PartitionDNSSuffix(region string) string {
	return names.DNSSuffixForPartition(Partition(region))
}

type SweeperFn func(ctx context.Context, client *conns.AWSClient) ([]Sweepable, error)

func Register(name string, f SweeperFn, dependencies ...string) {
	resource.AddTestSweepers(name, &resource.Sweeper{
		Name: name,
		F: func(region string) error {
			ctx := Context(region)
			ctx = logWithResourceType(ctx, name)

			client, err := SharedRegionalSweepClient(ctx, region)
			if err != nil {
				return fmt.Errorf("getting client: %w", err)
			}
			tflog.Info(ctx, "listing resources")
			sweepResources, err := f(ctx, client)

			if SkipSweepError(err) {
				tflog.Warn(ctx, "Skipping sweeper", map[string]any{
					"error": err.Error(),
				})
				return nil
			}
			if err != nil {
				return fmt.Errorf("listing %q (%s): %w", name, region, err)
			}

			err = SweepOrchestrator(ctx, sweepResources)
			if err != nil {
				return fmt.Errorf("sweeping %q (%s): %w", name, region, err)
			}

			return nil
		},
	})
}
