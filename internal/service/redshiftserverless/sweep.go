// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_redshiftserverless_namespace", &resource.Sweeper{
		Name: "aws_redshiftserverless_namespace",
		F:    sweepNamespaces,
		Dependencies: []string{
			"aws_redshiftserverless_workgroup",
		},
	})

	resource.AddTestSweepers("aws_redshiftserverless_workgroup", &resource.Sweeper{
		Name: "aws_redshiftserverless_workgroup",
		F:    sweepWorkgroups,
	})

	resource.AddTestSweepers("aws_redshiftserverless_snapshot", &resource.Sweeper{
		Name: "aws_redshiftserverless_snapshot",
		F:    sweepSnapshots,
	})
}

func sweepNamespaces(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.RedshiftServerlessConn(ctx)
	input := &redshiftserverless.ListNamespacesInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.ListNamespacesPagesWithContext(ctx, input, func(page *redshiftserverless.ListNamespacesOutput, lastPage bool) bool {
		if len(page.Namespaces) == 0 {
			log.Print("[DEBUG] No Redshift Serverless Namespaces to sweep")
			return !lastPage
		}

		for _, namespace := range page.Namespaces {
			r := resourceNamespace()
			d := r.Data(nil)
			d.SetId(aws.StringValue(namespace.NamespaceName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Serverless Namespaces: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Serverless Namespaces for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Serverless Namespaces sweep for %s: %s", region, errs)
		return nil
	}

	return nil
}

func sweepWorkgroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.RedshiftServerlessConn(ctx)
	input := &redshiftserverless.ListWorkgroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.ListWorkgroupsPagesWithContext(ctx, input, func(page *redshiftserverless.ListWorkgroupsOutput, lastPage bool) bool {
		if len(page.Workgroups) == 0 {
			log.Print("[DEBUG] No Redshift Serverless Workgroups to sweep")
			return !lastPage
		}

		for _, workgroup := range page.Workgroups {
			r := resourceWorkgroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(workgroup.WorkgroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Serverless Workgroups: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Serverless Workgroups for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Serverless Workgroups sweep for %s: %s", region, errs)
		return nil
	}

	return nil
}

func sweepSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.RedshiftServerlessConn(ctx)
	input := &redshiftserverless.ListSnapshotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.ListSnapshotsPagesWithContext(ctx, input, func(page *redshiftserverless.ListSnapshotsOutput, lastPage bool) bool {
		if len(page.Snapshots) == 0 {
			log.Print("[DEBUG] No Redshift Serverless Snapshots to sweep")
			return !lastPage
		}

		for _, workgroup := range page.Snapshots {
			r := resourceSnapshot()
			d := r.Data(nil)
			d.SetId(aws.StringValue(workgroup.SnapshotName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Serverless Snapshots: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Serverless Snapshots for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Serverless Snapshots sweep for %s: %s", region, errs)
		return nil
	}

	return nil
}
