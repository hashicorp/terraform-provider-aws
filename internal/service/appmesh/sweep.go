// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_appmesh_gateway_route", &resource.Sweeper{
		Name: "aws_appmesh_gateway_route",
		F:    sweepGatewayRoutes,
	})

	resource.AddTestSweepers("aws_appmesh_mesh", &resource.Sweeper{
		Name: "aws_appmesh_mesh",
		F:    sweepMeshes,
		Dependencies: []string{
			"aws_appmesh_virtual_service",
			"aws_appmesh_virtual_router",
			"aws_appmesh_virtual_node",
			"aws_appmesh_virtual_gateway",
		},
	})

	resource.AddTestSweepers("aws_appmesh_route", &resource.Sweeper{
		Name: "aws_appmesh_route",
		F:    sweepRoutes,
	})

	resource.AddTestSweepers("aws_appmesh_virtual_gateway", &resource.Sweeper{
		Name: "aws_appmesh_virtual_gateway",
		F:    sweepVirtualGateways,
		Dependencies: []string{
			"aws_appmesh_gateway_route",
		},
	})

	resource.AddTestSweepers("aws_appmesh_virtual_node", &resource.Sweeper{
		Name: "aws_appmesh_virtual_node",
		F:    sweepVirtualNodes,
	})

	resource.AddTestSweepers("aws_appmesh_virtual_router", &resource.Sweeper{
		Name: "aws_appmesh_virtual_router",
		F:    sweepVirtualRouters,
		Dependencies: []string{
			"aws_appmesh_route",
		},
	})

	resource.AddTestSweepers("aws_appmesh_virtual_service", &resource.Sweeper{
		Name: "aws_appmesh_virtual_service",
		F:    sweepVirtualServices,
	})
}

func sweepMeshes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AppMeshConn(ctx)
	input := &appmesh.ListMeshesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListMeshesPagesWithContext(ctx, input, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Meshes {
			r := resourceMesh()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.MeshName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Mesh Service Mesh sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing App Mesh Service Meshes (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping App Mesh Service Meshes (%s): %w", region, err)
	}

	return nil
}

func sweepVirtualGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppMeshConn(ctx)
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	input := &appmesh.ListMeshesInput{}
	err = conn.ListMeshesPagesWithContext(ctx, input, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Meshes {
			meshName := aws.StringValue(v.MeshName)
			input := &appmesh.ListVirtualGatewaysInput{
				MeshName: aws.String(meshName),
			}

			err := conn.ListVirtualGatewaysPagesWithContext(ctx, input, func(page *appmesh.ListVirtualGatewaysOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.VirtualGateways {
					virtualGatewayName := aws.StringValue(v.VirtualGatewayName)
					r := resourceVirtualGateway()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s/%s", meshName, virtualGatewayName)) // Logged in Delete handler, not used in API call.
					d.Set("mesh_name", meshName)
					d.Set(names.AttrName, virtualGatewayName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Virtual Gateways (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Mesh Virtual Gateway sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Service Meshes (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping App Mesh Virtual Gateways (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVirtualNodes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppMeshConn(ctx)
	input := &appmesh.ListMeshesInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListMeshesPagesWithContext(ctx, input, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Meshes {
			meshName := aws.StringValue(v.MeshName)
			input := &appmesh.ListVirtualNodesInput{
				MeshName: aws.String(meshName),
			}

			err := conn.ListVirtualNodesPagesWithContext(ctx, input, func(page *appmesh.ListVirtualNodesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.VirtualNodes {
					virtualNodeName := aws.StringValue(v.VirtualNodeName)
					r := resourceVirtualNode()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s/%s", meshName, virtualNodeName)) // Logged in Delete handler, not used in API call.
					d.Set("mesh_name", meshName)
					d.Set(names.AttrName, virtualNodeName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Virtual Nodes (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Mesh Virtual Node sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Service Meshes (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping App Mesh Virtual Nodes (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVirtualRouters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AppMeshConn(ctx)
	input := &appmesh.ListMeshesInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListMeshesPagesWithContext(ctx, input, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Meshes {
			meshName := aws.StringValue(v.MeshName)
			input := &appmesh.ListVirtualRoutersInput{
				MeshName: aws.String(meshName),
			}

			err := conn.ListVirtualRoutersPagesWithContext(ctx, input, func(page *appmesh.ListVirtualRoutersOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.VirtualRouters {
					virtualRouterName := aws.StringValue(v.VirtualRouterName)
					r := resourceVirtualRouter()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s/%s", meshName, virtualRouterName)) // Logged in Delete handler, not used in API call.
					d.Set("mesh_name", meshName)
					d.Set(names.AttrName, virtualRouterName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Virtual Routers (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Mesh Virtual Router sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Service Meshes (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping App Mesh Virtual Routers (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVirtualServices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AppMeshConn(ctx)
	input := &appmesh.ListMeshesInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListMeshesPagesWithContext(ctx, input, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Meshes {
			meshName := aws.StringValue(v.MeshName)
			input := &appmesh.ListVirtualServicesInput{
				MeshName: aws.String(meshName),
			}

			err := conn.ListVirtualServicesPagesWithContext(ctx, input, func(page *appmesh.ListVirtualServicesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.VirtualServices {
					virtualServiceName := aws.StringValue(v.VirtualServiceName)
					r := resourceVirtualService()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s/%s", meshName, virtualServiceName)) // Logged in Delete handler, not used in API call.
					d.Set("mesh_name", meshName)
					d.Set(names.AttrName, virtualServiceName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Virtual Services (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Mesh Virtual Service sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Service Meshes (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping App Mesh Virtual Services (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepGatewayRoutes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppMeshConn(ctx)
	input := &appmesh.ListMeshesInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListMeshesPagesWithContext(ctx, input, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Meshes {
			meshName := aws.StringValue(v.MeshName)
			input := &appmesh.ListVirtualGatewaysInput{
				MeshName: aws.String(meshName),
			}

			err := conn.ListVirtualGatewaysPagesWithContext(ctx, input, func(page *appmesh.ListVirtualGatewaysOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.VirtualGateways {
					virtualGatewayName := aws.StringValue(v.VirtualGatewayName)
					input := &appmesh.ListGatewayRoutesInput{
						MeshName:           aws.String(meshName),
						VirtualGatewayName: aws.String(virtualGatewayName),
					}

					err := conn.ListGatewayRoutesPagesWithContext(ctx, input, func(page *appmesh.ListGatewayRoutesOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, v := range page.GatewayRoutes {
							gatewayRouteName := aws.StringValue(v.GatewayRouteName)
							r := resourceGatewayRoute()
							d := r.Data(nil)
							d.SetId(fmt.Sprintf("%s/%s/%s", meshName, virtualGatewayName, gatewayRouteName)) // Logged in Delete handler, not used in API call.
							d.Set("mesh_name", meshName)
							d.Set(names.AttrName, gatewayRouteName)
							d.Set("virtual_gateway_name", virtualGatewayName)

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}

						return !lastPage
					})

					if awsv1.SkipSweepError(err) {
						continue
					}

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Gateway Routes (%s): %w", region, err))
					}
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Virtual Gateways (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Mesh Gateway Route sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Service Meshes (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping App Mesh Gateway Routes (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRoutes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AppMeshConn(ctx)
	input := &appmesh.ListMeshesInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListMeshesPagesWithContext(ctx, input, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Meshes {
			meshName := aws.StringValue(v.MeshName)
			input := &appmesh.ListVirtualRoutersInput{
				MeshName: aws.String(meshName),
			}

			err := conn.ListVirtualRoutersPagesWithContext(ctx, input, func(page *appmesh.ListVirtualRoutersOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.VirtualRouters {
					virtualRouterName := aws.StringValue(v.VirtualRouterName)
					input := &appmesh.ListRoutesInput{
						MeshName:          aws.String(meshName),
						VirtualRouterName: aws.String(virtualRouterName),
					}

					err := conn.ListRoutesPagesWithContext(ctx, input, func(page *appmesh.ListRoutesOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, v := range page.Routes {
							routeName := aws.StringValue(v.RouteName)
							r := resourceRoute()
							d := r.Data(nil)
							d.SetId(fmt.Sprintf("%s/%s/%s", meshName, virtualRouterName, routeName)) // Logged in Delete handler, not used in API call.
							d.Set("mesh_name", meshName)
							d.Set(names.AttrName, routeName)
							d.Set("virtual_router_name", virtualRouterName)

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}

						return !lastPage
					})

					if awsv1.SkipSweepError(err) {
						continue
					}

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Routes (%s): %w", region, err))
					}
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Virtual Routers (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Mesh Route sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing App Mesh Service Meshes (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping App Mesh Routes (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
