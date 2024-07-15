// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

// These timeouts are lower to fail faster during sweepers
const (
	sweeperGlobalReplicationGroupDisassociationReadyTimeout = 10 * time.Minute
	sweeperGlobalReplicationGroupDefaultUpdatedTimeout      = 10 * time.Minute
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_elasticache_cluster", &resource.Sweeper{
		Name: "aws_elasticache_cluster",
		F:    sweepClusters,
		Dependencies: []string{
			"aws_elasticache_replication_group",
		},
	})

	resource.AddTestSweepers("aws_elasticache_global_replication_group", &resource.Sweeper{
		Name: "aws_elasticache_global_replication_group",
		F:    sweepGlobalReplicationGroups,
	})

	resource.AddTestSweepers("aws_elasticache_parameter_group", &resource.Sweeper{
		Name: "aws_elasticache_parameter_group",
		F:    sweepParameterGroups,
		Dependencies: []string{
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
		},
	})

	resource.AddTestSweepers("aws_elasticache_replication_group", &resource.Sweeper{
		Name: "aws_elasticache_replication_group",
		F:    sweepReplicationGroups,
		Dependencies: []string{
			"aws_elasticache_global_replication_group",
		},
	})

	resource.AddTestSweepers("aws_elasticache_subnet_group", &resource.Sweeper{
		Name: "aws_elasticache_subnet_group",
		F:    sweepSubnetGroups,
		Dependencies: []string{
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
		},
	})

	resource.AddTestSweepers("aws_elasticache_user", &resource.Sweeper{
		Name: "aws_elasticache_user",
		F:    sweepUsers,
		Dependencies: []string{
			"aws_elasticache_user_group",
		},
	})

	resource.AddTestSweepers("aws_elasticache_user_group", &resource.Sweeper{
		Name: "aws_elasticache_user_group",
		F:    sweepUserGroups,
	})
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &elasticache.DescribeCacheClustersInput{
		ShowCacheClustersNotInReplicationGroups: aws.Bool(true),
	}
	conn := client.ElastiCacheClient(ctx)
	var sweeperErrs *multierror.Error

	pages := elasticache.NewDescribeCacheClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ElastiCache Cluster sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error retrieving ElastiCache Clusters: %w", err))
		}

		for _, v := range page.CacheClusters {
			id := aws.ToString(v.CacheClusterId)

			log.Printf("[INFO] Deleting ElastiCache Cluster: %s", id)
			err := deleteCacheCluster(ctx, conn, id, "")

			if err != nil {
				log.Printf("[ERROR] Failed to delete ElastiCache Cache Cluster (%s): %s", id, err)
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting ElastiCache Cache Cluster (%s): %w", id, err))
			}

			const (
				timeout = 40 * time.Minute
			)
			if _, err := waitCacheClusterDeleted(ctx, conn, id, timeout); err != nil {
				log.Printf("[ERROR] Failed waiting for ElastiCache Cache Cluster (%s) to be deleted: %s", id, err)
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting ElastiCache Cache Cluster (%s): waiting for completion: %w", id, err))
			}
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepGlobalReplicationGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &elasticache.DescribeGlobalReplicationGroupsInput{
		ShowMemberInfo: aws.Bool(true),
	}
	conn := client.ElastiCacheClient(ctx)

	var grgGroup multierror.Group
	var grgErrs *multierror.Error

	pages := elasticache.NewDescribeGlobalReplicationGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ElastiCache Global Replication Group sweep for %q: %s", region, err)
			return grgErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			grgErrs = multierror.Append(grgErrs, fmt.Errorf("listing ElastiCache Global Replication Groups: %w", err))
		}

		for _, v := range page.GlobalReplicationGroups {
			globalReplicationGroup := v

			grgGroup.Go(func() error {
				id := aws.ToString(globalReplicationGroup.GlobalReplicationGroupId)

				disassociationErrors := disassociateMembers(ctx, conn, globalReplicationGroup)
				if disassociationErrors != nil {
					return fmt.Errorf("disassociating ElastiCache Global Replication Group (%s) members: %w", id, disassociationErrors)
				}

				log.Printf("[INFO] Deleting ElastiCache Global Replication Group: %s", id)
				err := deleteGlobalReplicationGroup(ctx, conn, id, sweeperGlobalReplicationGroupDefaultUpdatedTimeout, globalReplicationGroupDefaultDeletedTimeout)

				return err
			})
		}
	}

	grgErrs = multierror.Append(grgErrs, grgGroup.Wait())

	return grgErrs.ErrorOrNil()
}

func sweepParameterGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &elasticache.DescribeCacheParameterGroupsInput{}
	conn := client.ElastiCacheClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := elasticache.NewDescribeCacheParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ElastiCache Parameter Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ElastiCache Parameter Groups (%s): %w", region, err)
		}

		for _, v := range page.CacheParameterGroups {
			name := aws.ToString(v.CacheParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping ElastiCache Cache Parameter Group: %s", name)
				continue
			}

			r := resourceParameterGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ElastiCache Parameter Groups (%s): %w", region, err)
	}

	return nil
}

func sweepReplicationGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &elasticache.DescribeReplicationGroupsInput{}
	conn := client.ElastiCacheClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := elasticache.NewDescribeReplicationGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ElastiCache Replication Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ElastiCache Replication Groups (%s): %w", region, err)
		}

		for _, v := range page.ReplicationGroups {
			r := resourceReplicationGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ReplicationGroupId))
			if v.GlobalReplicationGroupInfo != nil {
				d.Set("global_replication_group_id", v.GlobalReplicationGroupInfo.GlobalReplicationGroupId)
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ElastiCache Replication Groups (%s): %w", region, err)
	}

	return nil
}

func sweepSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ElastiCacheClient(ctx)
	input := &elasticache.DescribeCacheSubnetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := elasticache.NewDescribeCacheSubnetGroupsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ElastiCache Subnet Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ElastiCache Subnet Groups (%s): %w", region, err)
		}

		for _, v := range page.CacheSubnetGroups {
			name := aws.ToString(v.CacheSubnetGroupName)

			if name == "default" {
				log.Printf("[INFO] Skipping ElastiCache Subnet Group: %s", name)
				continue
			}

			r := resourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ElastiCache Subnet Groups (%s): %w", region, err)
	}

	return nil
}

func sweepUsers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ElastiCacheClient(ctx)
	input := &elasticache.DescribeUsersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := elasticache.NewDescribeUsersPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ElastiCache User sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing ElastiCache Users (%s): %w", region, err)
		}

		for _, v := range page.Users {
			id := aws.ToString(v.UserId)

			if id == "default" {
				log.Printf("[INFO] Skipping ElastiCache User: %s", id)
				continue
			}

			r := resourceUser()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping ElastiCache Users (%s): %w", region, err)
	}

	return nil
}

func sweepUserGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ElastiCacheClient(ctx)
	input := &elasticache.DescribeUserGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := elasticache.NewDescribeUserGroupsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ElastiCache User Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing ElastiCache User Groups (%s): %w", region, err)
		}

		for _, v := range page.UserGroups {
			r := resourceUserGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.UserGroupId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping ElastiCache User Groups (%s): %w", region, err)
	}

	return nil
}

func disassociateMembers(ctx context.Context, conn *elasticache.Client, globalReplicationGroup awstypes.GlobalReplicationGroup) error {
	var membersGroup multierror.Group

	for _, member := range globalReplicationGroup.Members {
		member := member

		if aws.ToString(member.Role) == globalReplicationGroupMemberRolePrimary {
			continue
		}

		id := aws.ToString(globalReplicationGroup.GlobalReplicationGroupId)

		membersGroup.Go(func() error {
			if err := disassociateReplicationGroup(ctx, conn, id, aws.ToString(member.ReplicationGroupId), aws.ToString(member.ReplicationGroupRegion), sweeperGlobalReplicationGroupDisassociationReadyTimeout); err != nil {
				log.Printf("[ERROR] %s", err)
				return err
			}
			return nil
		})
	}

	return membersGroup.Wait().ErrorOrNil()
}
