// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
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
			"aws_ecs_express_gateway_service",
			"aws_ecs_gateway_cluster_association",
			"aws_ecs_service",
		},
	})

	awsv2.Register("aws_ecs_express_gateway_service", sweepExpressGatewayServices)

	awsv2.Register("aws_ecs_gateway", sweepGateways, "aws_ecs_gateway_cluster_association")

	awsv2.Register("aws_ecs_gateway_cluster_association", sweepGatewayClusterAssociations, "aws_ecs_service")

	resource.AddTestSweepers("aws_ecs_service", &resource.Sweeper{
		Name: "aws_ecs_service",
		F:    sweepServices,
	})

	resource.AddTestSweepers("aws_ecs_task_definition", &resource.Sweeper{
		Name: "aws_ecs_task_definition",
		F:    sweepTaskDefinitions,
		Dependencies: []string{
			"aws_ecs_express_gateway_service",
			"aws_ecs_service",
		},
	})
}

func sweepCapacityProviders(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.ECSClient(ctx)
	input := &ecs.DescribeCapacityProvidersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeCapacityProvidersPages(ctx, conn, input, func(page *ecs.DescribeCapacityProvidersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CapacityProviders {
			arn := aws.ToString(v.CapacityProviderArn)

			if name := aws.ToString(v.Name); name == "FARGATE" || name == "FARGATE_SPOT" {
				log.Printf("[INFO] Skipping AWS managed ECS Capacity Provider: %s", arn)
				continue
			}

			r := resourceCapacityProvider()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ECS Capacity Provider sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing ECS Capacity Providers (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ECS Capacity Providers (%s): %w", region, err)
	}

	return nil
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.ECSClient(ctx)
	input := &ecs.ListClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ecs.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ECS Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ECS Clusters (%s): %w", region, err)
		}

		for _, v := range page.ClusterArns {
			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ECS Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepExpressGatewayServices(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ECSClient(ctx)
	var sweepResources []sweep.Sweepable

	var input ecs.ListClustersInput
	pages := ecs.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, clusterARN := range page.ClusterArns {
			input := ecs.ListServicesInput{
				Cluster: aws.String(clusterARN),
			}
			pages := newListExpressGatewayServicesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, v := range page.ServiceArns {
					sweepResources = append(sweepResources, framework.NewSweepResource(newExpressGatewayServiceResource, client,
						framework.NewAttribute("service_arn", v)),
					)
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepGatewayClusterAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ECSClient(ctx)
	var sweepResources []sweep.Sweepable

	var input ecs.ListGatewayClusterAssociationsInput
	for page, err := range listGatewayClusterAssociationPages(ctx, conn, &input) {
		if errs.IsA[*awstypes.UnsupportedFeatureException](err) {
			return nil, nil
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page {
			sweepResources = append(sweepResources, framework.NewSweepResource(newGatewayClusterAssociationResource, client,
				framework.NewAttribute("cluster_arn", aws.ToString(v.ClusterArn)),
				framework.NewAttribute("gateway_arn", aws.ToString(v.GatewayArn))))
		}
	}

	return sweepResources, nil
}

func sweepGateways(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ECSClient(ctx)
	var sweepResources []sweep.Sweepable

	var input ecs.ListGatewaysInput
	pages := ecs.NewListGatewaysPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if errs.IsA[*awstypes.UnsupportedFeatureException](err) {
			return nil, nil
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Gateways {
			arn := aws.ToString(v.GatewayArn)

			gateway, err := findGatewayByARN(ctx, conn, arn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				tflog.Warn(ctx, "Skipping resource", map[string]any{
					"error":       err.Error(),
					names.AttrARN: arn,
				})
				continue
			}

			if aws.ToBool(gateway.DeletionProtection) {
				if err := disableGatewayDeletionProtection(ctx, conn, gateway); err != nil {
					tflog.Warn(ctx, "Skipping resource", map[string]any{
						"error":       err.Error(),
						names.AttrARN: arn,
					})
					continue
				}
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newGatewayResource, client,
				framework.NewAttribute(names.AttrARN, arn)))
		}
	}

	return sweepResources, nil
}

// UpdateGateway rejects a request without the role or the configuration, so both are sent back as described.
func disableGatewayDeletionProtection(ctx context.Context, conn *ecs.Client, gateway *awstypes.Gateway) error {
	const (
		timeout = 20 * time.Minute
	)
	arn := aws.ToString(gateway.GatewayArn)

	input := ecs.UpdateGatewayInput{
		DeletionProtection:    aws.Bool(false),
		Gateway:               aws.String(arn),
		InfrastructureRoleArn: gateway.InfrastructureRoleArn,
		PropagateTags:         gateway.PropagateTags,
	}
	if v, ok := gateway.Configuration.(*awstypes.DescribeGatewayConfigurationMemberHostSharedAlb); ok {
		alb := awstypes.UpdateGatewayAlbConfiguration{
			AccessLogs:       v.Value.AccessLogs,
			HttpDesync:       v.Value.HttpDesync,
			MaxAlbCount:      v.Value.MaxAlbCount,
			TestTrafficRules: v.Value.TestTrafficRules,
			WebACLArn:        v.Value.WebACLArn,
		}
		if v := v.Value.NetworkConfiguration; v != nil {
			alb.NetworkConfiguration = &awstypes.UpdateGatewayNetworkConfiguration{
				InboundSecurityGroup: v.InboundSecurityGroup,
			}
		}
		input.Configuration = &awstypes.UpdateGatewayConfigurationMemberHostSharedAlb{
			Value: alb,
		}
	}

	if _, err := conn.UpdateGateway(ctx, &input); err != nil {
		return smarterr.NewError(err)
	}

	if _, err := waitGatewayStable(ctx, conn, arn, timeout); err != nil {
		return smarterr.NewError(err)
	}

	return nil
}

func sweepServices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.ECSClient(ctx)
	input := &ecs.ListClustersInput{}
	var sweepResources []sweep.Sweepable

	pages := ecs.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ECS Service sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ECS Clusters (%s): %w", region, err)
		}

		for _, clusterARN := range page.ClusterArns {
			input := &ecs.ListServicesInput{
				Cluster: aws.String(clusterARN),
			}

			pages := newListRegularServicesPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.ServiceArns {
					r := resourceService()
					d := r.Data(nil)
					d.SetId(v)
					d.Set("cluster", clusterARN)
					d.Set(names.AttrForceDelete, true)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ECS Services (%s): %w", region, err)
	}

	return nil
}

func sweepTaskDefinitions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.ECSClient(ctx)
	input := &ecs.ListTaskDefinitionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ecs.NewListTaskDefinitionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ECS Task Definition sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ECS Task Definitions (%s): %w", region, err)
		}

		for _, v := range page.TaskDefinitionArns {
			r := resourceTaskDefinition()
			d := r.Data(nil)
			d.SetId(v)
			d.Set(names.AttrARN, v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ECS Task Definitions (%s): %w", region, err)
	}

	return nil
}
