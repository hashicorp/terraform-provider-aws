//go:build sweep
// +build sweep

package memorydb

import (
	"fmt"
	"log"
	"strings"

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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).MemoryDBConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &memorydb.DescribeACLsInput{}

	for {
		output, err := conn.DescribeACLs(input)

		for _, ACL := range output.ACLs {
			r := ResourceACL()
			d := r.Data(nil)

			id := aws.StringValue(ACL.Name)
			if id == "open-access" {
				continue // The open-access parameter group cannot be deleted.
			}

			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("error reading MemoryDB ACL (%s): %w", id, err)
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
		errs = multierror.Append(errs, fmt.Errorf("error sweeping MemoryDB ACL for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB ACL sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).MemoryDBConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &memorydb.DescribeParameterGroupsInput{}

	for {
		output, err := conn.DescribeParameterGroups(input)

		for _, ParameterGroup := range output.ParameterGroups {
			r := ResourceParameterGroup()
			d := r.Data(nil)

			id := aws.StringValue(ParameterGroup.Name)
			if strings.HasPrefix(id, "default.") {
				continue // Default parameter groups cannot be deleted.
			}

			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("error reading MemoryDB Parameter Group (%s): %w", id, err)
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
		errs = multierror.Append(errs, fmt.Errorf("error sweeping MemoryDB Parameter Group for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB Parameter Group sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).MemoryDBConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &memorydb.DescribeUsersInput{}

	for {
		output, err := conn.DescribeUsers(input)

		for _, User := range output.Users {
			r := ResourceUser()
			d := r.Data(nil)

			id := aws.StringValue(User.Name)
			if id == "default" {
				continue // The default user cannot be deleted.
			}

			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("error reading MemoryDB User (%s): %w", id, err)
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
		errs = multierror.Append(errs, fmt.Errorf("error sweeping MemoryDB User for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MemoryDB User sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
