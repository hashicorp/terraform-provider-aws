//go:build sweep
// +build sweep

package memorydb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/go-multierror"
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
	return nil
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).MemoryDBConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &memorydb.DescribeClustersInput{}

	for {
		output, err := conn.DescribeClusters(input)

		for _, Cluster := range output.Clusters {
			r := ResourceCluster()
			d := r.Data(nil)

			id := aws.StringValue(Cluster.Name)
			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("error reading MemoryDB Cluster (%s): %w", id, err)
				log.Printf("[ERROR] %s", err)
				errs = multierror.Append(errs, err)
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping MemoryDB Cluster for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB Cluster sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepParameterGroups(region string) error {
	return nil
}

func sweepSnapshots(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).MemoryDBConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &memorydb.DescribeSnapshotsInput{}

	for {
		output, err := conn.DescribeSnapshots(input)

		for _, Snapshot := range output.Snapshots {
			r := ResourceSnapshot()
			d := r.Data(nil)

			id := aws.StringValue(Snapshot.Name)
			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("error reading MemoryDB Snapshot (%s): %w", id, err)
				log.Printf("[ERROR] %s", err)
				errs = multierror.Append(errs, err)
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping MemoryDB Snapshot for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB Snapshot sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepSubnetGroups(region string) error {
	return nil
}

func sweepUsers(region string) error {
	return nil
}
