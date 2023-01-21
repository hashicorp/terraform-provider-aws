//go:build sweep
// +build sweep

package ecs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_ecs_capacity_provider", &resource.Sweeper{
		Name: "aws_ecs_capacity_provider",
		F:    sweepCapacityProviders,
		Dependencies: []string{
			"aws_ecs_cluster",
			"aws_ecs_service",
		},
	})

	resource.AddTestSweepers("aws_ecs_cluster", &resource.Sweeper{
		Name: "aws_ecs_cluster",
		F:    sweepClusters,
		Dependencies: []string{
			"aws_ecs_service",
		},
	})

	resource.AddTestSweepers("aws_ecs_service", &resource.Sweeper{
		Name: "aws_ecs_service",
		F:    sweepServices,
	})

	resource.AddTestSweepers("aws_ecs_task_definition", &resource.Sweeper{
		Name: "aws_ecs_task_definition",
		F:    sweepTaskDefinitions,
		Dependencies: []string{
			"aws_ecs_service",
		},
	})
}

func sweepCapacityProviders(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ECSConn()
	input := &ecs.DescribeCapacityProvidersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeCapacityProvidersPages(ctx, conn, input, func(page *ecs.DescribeCapacityProvidersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CapacityProviders {
			arn := aws.StringValue(v.CapacityProviderArn)

			if name := aws.StringValue(v.Name); name == "FARGATE" || name == "FARGATE_SPOT" {
				log.Printf("[INFO] Skipping AWS managed ECS Capacity Provider: %s", arn)
				continue
			}

			r := ResourceCapacityProvider()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ECS Capacity Provider sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing ECS Capacity Providers (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ECS Capacity Providers (%s): %w", region, err)
	}

	return nil
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ECSConn()
	input := &ecs.ListClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListClustersPagesWithContext(ctx, input, func(page *ecs.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClusterArns {
			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ECS Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing ECS Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ECS Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepServices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ECSConn()
	input := &ecs.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListClustersPagesWithContext(ctx, input, func(page *ecs.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClusterArns {
			clusterARN := aws.StringValue(v)
			input := &ecs.ListServicesInput{
				Cluster: aws.String(clusterARN),
			}

			err := conn.ListServicesPagesWithContext(ctx, input, func(page *ecs.ListServicesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.ServiceArns {
					r := ResourceService()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v))
					d.Set("cluster", clusterARN)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing ECS Services (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ECS Service sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing ECS Clusters (%s): %w", region, err))
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping ECS Services (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTaskDefinitions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ECSConn()
	input := &ecs.ListTaskDefinitionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListTaskDefinitionsPagesWithContext(ctx, input, func(page *ecs.ListTaskDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TaskDefinitionArns {
			r := ResourceTaskDefinition()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v))
			d.Set("arn", v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ECS Task Definition sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing ECS Task Definitions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ECS Task Definitions (%s): %w", region, err)
	}

	return nil
}
