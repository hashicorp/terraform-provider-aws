//go:build sweep
// +build sweep

package elasticache

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

// These timeouts are lower to fail faster during sweepers
const (
	sweeperGlobalReplicationGroupDisassociationReadyTimeout = 10 * time.Minute
	sweeperGlobalReplicationGroupDefaultUpdatedTimeout      = 10 * time.Minute
)

func init() {
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

	resource.AddTestSweepers("aws_elasticache_security_group", &resource.Sweeper{
		Name: "aws_elasticache_security_group",
		F:    sweepCacheSecurityGroups,
		Dependencies: []string{
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
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
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ElastiCacheConn

	var sweeperErrs *multierror.Error

	input := &elasticache.DescribeCacheClustersInput{
		ShowCacheClustersNotInReplicationGroups: aws.Bool(true),
	}
	err = conn.DescribeCacheClustersPages(input, func(page *elasticache.DescribeCacheClustersOutput, lastPage bool) bool {
		if len(page.CacheClusters) == 0 {
			log.Print("[DEBUG] No ElastiCache Replication Groups to sweep")
			return false
		}

		for _, cluster := range page.CacheClusters {
			id := aws.StringValue(cluster.CacheClusterId)

			log.Printf("[INFO] Deleting ElastiCache Cluster: %s", id)
			err := DeleteCacheCluster(conn, id, "")
			if err != nil {
				log.Printf("[ERROR] Failed to delete ElastiCache Cache Cluster (%s): %s", id, err)
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting ElastiCache Cache Cluster (%s): %w", id, err))
			}
			_, err = WaitCacheClusterDeleted(conn, id, CacheClusterDeletedTimeout)
			if err != nil {
				log.Printf("[ERROR] Failed waiting for ElastiCache Cache Cluster (%s) to be deleted: %s", id, err)
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting ElastiCache Cache Cluster (%s): waiting for completion: %w", id, err))
			}
		}
		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ElastiCache Cluster sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error retrieving ElastiCache Clusters: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepGlobalReplicationGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ElastiCacheConn

	var grgGroup multierror.Group

	input := &elasticache.DescribeGlobalReplicationGroupsInput{
		ShowMemberInfo: aws.Bool(true),
	}
	err = conn.DescribeGlobalReplicationGroupsPages(input, func(page *elasticache.DescribeGlobalReplicationGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, globalReplicationGroup := range page.GlobalReplicationGroups {
			globalReplicationGroup := globalReplicationGroup

			grgGroup.Go(func() error {
				id := aws.StringValue(globalReplicationGroup.GlobalReplicationGroupId)

				disassociationErrors := DisassociateMembers(conn, globalReplicationGroup)
				if disassociationErrors != nil {
					sweeperErr := fmt.Errorf("failed to disassociate ElastiCache Global Replication Group (%s) members: %w", id, disassociationErrors)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				log.Printf("[INFO] Deleting ElastiCache Global Replication Group: %s", id)
				err := DeleteGlobalReplicationGroup(conn, id, sweeperGlobalReplicationGroupDefaultUpdatedTimeout)
				if err != nil {
					sweeperErr := fmt.Errorf("error deleting ElastiCache Global Replication Group (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}
				return nil
			})
		}

		return !lastPage
	})

	grgErrs := grgGroup.Wait()

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ElastiCache Global Replication Group sweep for %q: %s", region, err)
		return grgErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		grgErrs = multierror.Append(grgErrs, fmt.Errorf("error listing ElastiCache Global Replication Groups: %w", err))
	}

	return grgErrs.ErrorOrNil()
}

func sweepParameterGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ElastiCacheConn

	err = conn.DescribeCacheParameterGroupsPages(&elasticache.DescribeCacheParameterGroupsInput{}, func(page *elasticache.DescribeCacheParameterGroupsOutput, lastPage bool) bool {
		if len(page.CacheParameterGroups) == 0 {
			log.Print("[DEBUG] No Elasticache Parameter Groups to sweep")
			return false
		}

		for _, parameterGroup := range page.CacheParameterGroups {
			name := aws.StringValue(parameterGroup.CacheParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping Elasticache Cache Parameter Group: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting Elasticache Parameter Group: %s", name)
			_, err := conn.DeleteCacheParameterGroup(&elasticache.DeleteCacheParameterGroupInput{
				CacheParameterGroupName: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete Elasticache Parameter Group (%s): %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Elasticache Parameter Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Elasticache Parameter Group: %w", err)
	}
	return nil
}

func sweepReplicationGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).ElastiCacheConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeReplicationGroupsPages(&elasticache.DescribeReplicationGroupsInput{}, func(page *elasticache.DescribeReplicationGroupsOutput, lastPage bool) bool {
		if len(page.ReplicationGroups) == 0 {
			log.Print("[DEBUG] No ElastiCache Replication Groups to sweep")
			return !lastPage // in rare cases across API, one page may have empty results but not be last page
		}

		for _, replicationGroup := range page.ReplicationGroups {
			r := ResourceReplicationGroup()
			d := r.Data(nil)

			if replicationGroup.GlobalReplicationGroupInfo != nil {
				d.Set("global_replication_group_id", replicationGroup.GlobalReplicationGroupInfo.GlobalReplicationGroupId)
			}

			d.SetId(aws.StringValue(replicationGroup.ReplicationGroupId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Elasticache Replication Groups: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Elasticache Replication Groups for %s: %w", region, err))
	}

	// waiting for deletion is not necessary in the sweeper since the resource's delete waits

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Elasticache Replication Group sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepCacheSecurityGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ElastiCacheConn

	err = conn.DescribeCacheSecurityGroupsPages(&elasticache.DescribeCacheSecurityGroupsInput{}, func(page *elasticache.DescribeCacheSecurityGroupsOutput, lastPage bool) bool {
		if len(page.CacheSecurityGroups) == 0 {
			log.Print("[DEBUG] No Elasticache Cache Security Groups to sweep")
			return false
		}

		for _, securityGroup := range page.CacheSecurityGroups {
			name := aws.StringValue(securityGroup.CacheSecurityGroupName)

			if name == "default" {
				log.Printf("[INFO] Skipping Elasticache Cache Security Group: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting Elasticache Cache Security Group: %s", name)
			_, err := conn.DeleteCacheSecurityGroup(&elasticache.DeleteCacheSecurityGroupInput{
				CacheSecurityGroupName: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete Elasticache Cache Security Group (%s): %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Elasticache Cache Security Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Elasticache Cache Security Groups: %s", err)
	}
	return nil
}

func sweepSubnetGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ElastiCacheConn

	err = conn.DescribeCacheSubnetGroupsPages(&elasticache.DescribeCacheSubnetGroupsInput{}, func(page *elasticache.DescribeCacheSubnetGroupsOutput, lastPage bool) bool {
		if len(page.CacheSubnetGroups) == 0 {
			log.Print("[DEBUG] No Elasticache Subnet Groups to sweep")
			return false
		}

		for _, subnetGroup := range page.CacheSubnetGroups {
			name := aws.StringValue(subnetGroup.CacheSubnetGroupName)

			log.Printf("[INFO] Deleting Elasticache Subnet Group: %s", name)
			_, err := conn.DeleteCacheSubnetGroup(&elasticache.DeleteCacheSubnetGroupInput{
				CacheSubnetGroupName: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete Elasticache Subnet Group (%s): %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Elasticache Subnet Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Elasticache Subnet Groups: %w", err)
	}
	return nil
}

func DisassociateMembers(conn *elasticache.ElastiCache, globalReplicationGroup *elasticache.GlobalReplicationGroup) error {
	var membersGroup multierror.Group

	for _, member := range globalReplicationGroup.Members {
		member := member

		if aws.StringValue(member.Role) == GlobalReplicationGroupMemberRolePrimary {
			continue
		}

		id := aws.StringValue(globalReplicationGroup.GlobalReplicationGroupId)

		membersGroup.Go(func() error {
			if err := DisassociateReplicationGroup(conn, id, aws.StringValue(member.ReplicationGroupId), aws.StringValue(member.ReplicationGroupRegion), sweeperGlobalReplicationGroupDisassociationReadyTimeout); err != nil {
				sweeperErr := fmt.Errorf(
					"error disassociating ElastiCache Replication Group (%s) in %s from Global Group (%s): %w",
					aws.StringValue(member.ReplicationGroupId), aws.StringValue(member.ReplicationGroupRegion), id, err,
				)
				log.Printf("[ERROR] %s", sweeperErr)
				return sweeperErr
			}
			return nil
		})
	}

	return membersGroup.Wait().ErrorOrNil()
}
