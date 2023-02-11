//go:build sweep
// +build sweep

package memorydb

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_memorydb_acl", &resource.Sweeper{
		Name: "aws_memorydb_acl",
		F:    sweepACLs,
		Dependencies: []string{
			"aws_memorydb_cluster",
		},
	})

	resource.AddTestSweepers("aws_memorydb_cluster", &resource.Sweeper{
		Name: "aws_memorydb_cluster",
		F:    sweepClusters,
	})

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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).MemoryDBConn()
	input := &memorydb.DescribeACLsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeACLsPages(ctx, conn, input, func(page *memorydb.DescribeACLsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ACLs {
			id := aws.StringValue(v.Name)

			if id == "open-access" {
				continue // The open-access ACL cannot be deleted.
			}

			r := ResourceACL()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB ACL sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MemoryDB ACLs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB ACLs (%s): %w", region, err)
	}

	return nil
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).MemoryDBConn()
	input := &memorydb.DescribeClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeClustersPages(ctx, conn, input, func(page *memorydb.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Clusters {
			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MemoryDB Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepParameterGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).MemoryDBConn()
	input := &memorydb.DescribeParameterGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeParameterGroupsPages(ctx, conn, input, func(page *memorydb.DescribeParameterGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ParameterGroups {
			id := aws.StringValue(v.Name)

			if strings.HasPrefix(id, "default.") {
				continue // Default parameter groups cannot be deleted.
			}

			r := ResourceParameterGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB Parameter Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MemoryDB Parameter Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB Parameter Groups (%s): %w", region, err)
	}

	return nil
}

func sweepSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).MemoryDBConn()
	input := &memorydb.DescribeSnapshotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeSnapshotsPages(ctx, conn, input, func(page *memorydb.DescribeSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Snapshots {
			r := ResourceSnapshot()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB Snapshot sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MemoryDB Snapshots (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB Snapshots (%s): %w", region, err)
	}

	return nil
}

func sweepSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).MemoryDBConn()
	input := &memorydb.DescribeSubnetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeSubnetGroupsPages(ctx, conn, input, func(page *memorydb.DescribeSubnetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SubnetGroups {
			id := aws.StringValue(v.Name)

			if id == "default" {
				continue // The default subnet group cannot be deleted.
			}

			r := ResourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB Subnet Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MemoryDB Subnet Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB Subnet Groups (%s): %w", region, err)
	}

	return nil
}

func sweepUsers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).MemoryDBConn()
	input := &memorydb.DescribeUsersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeUsersPages(ctx, conn, input, func(page *memorydb.DescribeUsersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Users {
			id := aws.StringValue(v.Name)

			if id == "default" {
				continue // The default user cannot be deleted.
			}

			r := ResourceUser()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB User sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MemoryDB Users (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MemoryDB Users (%s): %w", region, err)
	}

	return nil
}
