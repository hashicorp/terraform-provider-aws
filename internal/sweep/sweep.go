// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sweep

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
	var g multierror.Group

	for _, sweepable := range sweepables {
		sweepable := sweepable

		g.Go(func() error {
			return sweepable.Delete(ctx, ThrottlingRetryTimeout, optFns...)
		})
	}

	return g.Wait().ErrorOrNil()
}

// Deprecated: Usse awsv1.SkipSweepError
//
//nolint:stylecheck // It's not required for functions, so why for variables?
var SkipSweepError = awsv1.SkipSweepError

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
