//go:build sweep
// +build sweep

package ecs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/go-multierror"
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ECSConn
	input := &ecs.DescribeCapacityProvidersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeCapacityProvidersPages(conn, input, func(page *ecs.DescribeCapacityProvidersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, capacityProvider := range page.CapacityProviders {
			arn := aws.StringValue(capacityProvider.CapacityProviderArn)

			if name := aws.StringValue(capacityProvider.Name); name == "FARGATE" || name == "FARGATE_SPOT" {
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
		log.Printf("[WARN] Skipping ECS Capacity Providers sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing ECS Capacity Providers for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping ECS Capacity Providers for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ECSConn

	err = conn.ListClustersPages(&ecs.ListClustersInput{}, func(page *ecs.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, clusterARNPtr := range page.ClusterArns {
			clusterARN := aws.StringValue(clusterARNPtr)

			log.Printf("[INFO] Deleting ECS Cluster: %s", clusterARN)
			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(clusterARN)
			err = r.Delete(d, client)
			if err != nil {
				log.Printf("[ERROR] Error deleting ECS Cluster (%s): %s", clusterARN, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ECS Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving ECS Clusters: %w", err)
	}

	return nil
}

func sweepServices(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ECSConn

	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListClustersPages(&ecs.ListClustersInput{}, func(page *ecs.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, clusterARN := range page.ClusterArns {
			input := &ecs.ListServicesInput{
				Cluster: clusterARN,
			}
			err := conn.ListServicesPages(input, func(page *ecs.ListServicesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, serviceARN := range page.ServiceArns {
					r := ResourceService()
					d := r.Data(nil)
					d.SetId(aws.StringValue(serviceARN))
					d.Set("cluster", clusterARN)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("listing ECS Services for Cluster (%s): %w", aws.StringValue(clusterARN), err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ECS Service sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving ECS Services: %w", err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping ECS Services for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTaskDefinitions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ECSConn
	var sweeperErrs *multierror.Error

	err = conn.ListTaskDefinitionsPages(&ecs.ListTaskDefinitionsInput{}, func(page *ecs.ListTaskDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, taskDefinitionArn := range page.TaskDefinitionArns {
			arn := aws.StringValue(taskDefinitionArn)

			log.Printf("[INFO] Deleting ECS Task Definition: %s", arn)
			_, err := conn.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
				TaskDefinition: aws.String(arn),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting ECS Task Definition (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ECS Task Definitions sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving ECS Task Definitions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
