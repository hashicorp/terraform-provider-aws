// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
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

	pages := redshiftserverless.NewListNamespacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Serverless Namespace sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Redshift Serverless Namespaces (%s): %w", region, err)
		}

		for _, v := range page.Namespaces {
			r := resourceNamespace()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.NamespaceName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift Serverless Namespaces (%s): %w", region, err)
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

	pages := redshiftserverless.NewListWorkgroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Serverless Workgroup sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Redshift Serverless Workgroups (%s): %w", region, err)
		}

		for _, v := range page.Workgroups {
			r := resourceWorkgroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.WorkgroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift Serverless Workgroups (%s): %w", region, err)
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

	pages := redshiftserverless.NewListSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Serverless Snapshot sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Redshift Serverless Snapshots (%s): %w", region, err)
		}

		for _, v := range page.Snapshots {
			r := resourceSnapshot()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.SnapshotName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift Serverless Snapshots (%s): %w", region, err)
	}

	return nil
}
