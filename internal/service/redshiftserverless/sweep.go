// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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
	conn := client.RedshiftServerlessClient(ctx)
	input := &redshiftserverless.ListNamespacesInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := redshiftserverless.NewListNamespacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Serverless Namespaces: %w", err))
		}

		for _, namespace := range page.Namespaces {
			r := resourceNamespace()
			d := r.Data(nil)
			d.SetId(aws.ToString(namespace.NamespaceName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Serverless Namespaces for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
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
	conn := client.RedshiftServerlessClient(ctx)
	input := &redshiftserverless.ListWorkgroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := redshiftserverless.NewListWorkgroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Serverless Workgroups: %w", err))
		}

		for _, workgroup := range page.Workgroups {
			r := resourceWorkgroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(workgroup.WorkgroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Serverless Workgroups for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
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
	conn := client.RedshiftServerlessClient(ctx)
	input := &redshiftserverless.ListSnapshotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := redshiftserverless.NewListSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Serverless Snapshots: %w", err))
		}

		for _, workgroup := range page.Snapshots {
			r := resourceSnapshot()
			d := r.Data(nil)
			d.SetId(aws.ToString(workgroup.SnapshotName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Serverless Snapshots for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Serverless Snapshots sweep for %s: %s", region, errs)
		return nil
	}

	return nil
}
