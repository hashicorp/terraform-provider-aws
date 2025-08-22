// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_memorydb_acl", &resource.Sweeper{
		Name: "aws_memorydb_acl",
		F:    sweepACLs,
		Dependencies: []string{
			"aws_memorydb_cluster",
		},
	})

	awsv2.Register("aws_memorydb_cluster", sweepClusters)

	awsv2.Register("aws_memorydb_multi_region_cluster", sweepMultiRegionClusters,
		"aws_memorydb_cluster",
	)

	resource.AddTestSweepers("aws_memorydb_parameter_group", &resource.Sweeper{
		Name: "aws_memorydb_parameter_group",
		F:    sweepParameterGroups,
		Dependencies: []string{
			"aws_memorydb_cluster",
		},
	})

	resource.AddTestSweepers("aws_memorydb_snapshot", &resource.Sweeper{
		Name: "aws_memorydb_snapshot",
		F:    sweepSnapshots,
		Dependencies: []string{
			"aws_memorydb_cluster",
		},
	})

	resource.AddTestSweepers("aws_memorydb_subnet_group", &resource.Sweeper{
		Name: "aws_memorydb_subnet_group",
		F:    sweepSubnetGroups,
		Dependencies: []string{
			"aws_memorydb_cluster",
		},
	})

	resource.AddTestSweepers("aws_memorydb_user", &resource.Sweeper{
		Name: "aws_memorydb_user",
		F:    sweepUsers,
		Dependencies: []string{
			"aws_memorydb_acl",
		},
	})
}

func sweepACLs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.MemoryDBClient(ctx)
	input := memorydb.DescribeACLsInput{}
	var sweepResources []sweep.Sweepable

	pages := memorydb.NewDescribeACLsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MemoryDB ACL sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MemoryDB ACLs (%s): %w", region, err)
		}

		for _, v := range page.ACLs {
			id := aws.ToString(v.Name)

			if id == "open-access" {
				log.Printf("[INFO] Skipping MemoryDB ACL %s", id)
				continue // The open-access ACL cannot be deleted.
			}

			r := resourceACL()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB ACLs (%s): %w", region, err)
	}

	return nil
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.MemoryDBClient(ctx)

	var sweepResources []sweep.Sweepable
	r := resourceCluster()

	input := memorydb.DescribeClustersInput{}
	pages := memorydb.NewDescribeClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Clusters {
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))
			d.Set(names.AttrName, v.Name)
			d.Set("multi_region_cluster_name", v.MultiRegionClusterName)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepMultiRegionClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.MemoryDBClient(ctx)

	var sweepResources []sweep.Sweepable

	input := memorydb.DescribeMultiRegionClustersInput{}
	pages := memorydb.NewDescribeMultiRegionClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, clusters := range page.MultiRegionClusters {
			sweepResources = append(sweepResources, framework.NewSweepResource(newMultiRegionClusterResource, client,
				framework.NewAttribute("multi_region_cluster_name", clusters.MultiRegionClusterName),
			))
		}
	}

	return sweepResources, nil
}

func sweepParameterGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.MemoryDBClient(ctx)
	input := memorydb.DescribeParameterGroupsInput{}
	var sweepResources []sweep.Sweepable

	pages := memorydb.NewDescribeParameterGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MemoryDB Parameter Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MemoryDB Parameter Groups (%s): %w", region, err)
		}

		for _, v := range page.ParameterGroups {
			id := aws.ToString(v.Name)

			if strings.HasPrefix(id, "default.") {
				log.Printf("[INFO] Skipping MemoryDB Parameter Group %s", id)
				continue // Default parameter groups cannot be deleted.
			}

			r := resourceParameterGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB Parameter Groups (%s): %w", region, err)
	}

	return nil
}

func sweepSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.MemoryDBClient(ctx)
	input := memorydb.DescribeSnapshotsInput{}
	var sweepResources []sweep.Sweepable

	pages := memorydb.NewDescribeSnapshotsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MemoryDB Snapshot sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MemoryDB Snapshots (%s): %w", region, err)
		}

		for _, v := range page.Snapshots {
			r := resourceSnapshot()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB Snapshots (%s): %w", region, err)
	}

	return nil
}

func sweepSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.MemoryDBClient(ctx)
	input := memorydb.DescribeSubnetGroupsInput{}
	var sweepResources []sweep.Sweepable

	pages := memorydb.NewDescribeSubnetGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MemoryDB Subnet Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MemoryDB Subnet Groups (%s): %w", region, err)
		}

		for _, v := range page.SubnetGroups {
			id := aws.ToString(v.Name)

			if id == "default" {
				log.Printf("[INFO] Skipping MemoryDB Subnet Group %s", id)
				continue // The default subnet group cannot be deleted.
			}

			r := resourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB Subnet Groups (%s): %w", region, err)
	}

	return nil
}

func sweepUsers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.MemoryDBClient(ctx)
	input := memorydb.DescribeUsersInput{}
	var sweepResources []sweep.Sweepable

	pages := memorydb.NewDescribeUsersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MemoryDB User sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MemoryDB Users (%s): %w", region, err)
		}

		for _, v := range page.Users {
			id := aws.ToString(v.Name)

			if id == "default" {
				log.Printf("[INFO] Skipping MemoryDB User %s", id)
				continue // The default user cannot be deleted.
			}

			r := resourceUser()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB Users (%s): %w", region, err)
	}

	return nil
}
