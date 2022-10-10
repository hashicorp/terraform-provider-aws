//go:build sweep
// +build sweep

package appmesh

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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

func sweepGatewayRoutes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).AppMeshConn

	var sweeperErrs *multierror.Error

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mesh := range page.Meshes {
			meshName := aws.StringValue(mesh.MeshName)

			err = conn.ListVirtualGatewaysPages(&appmesh.ListVirtualGatewaysInput{MeshName: mesh.MeshName}, func(page *appmesh.ListVirtualGatewaysOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, virtualGateway := range page.VirtualGateways {
					virtualGatewayName := aws.StringValue(virtualGateway.VirtualGatewayName)

					err = conn.ListGatewayRoutesPages(&appmesh.ListGatewayRoutesInput{MeshName: mesh.MeshName, VirtualGatewayName: virtualGateway.VirtualGatewayName}, func(page *appmesh.ListGatewayRoutesOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, gatewayRoute := range page.GatewayRoutes {
							gatewayRouteName := aws.StringValue(gatewayRoute.GatewayRouteName)

							log.Printf("[INFO] Deleting App Mesh service mesh (%s) virtual gateway (%s) gateway route: %s", meshName, virtualGatewayName, gatewayRouteName)
							r := ResourceGatewayRoute()
							d := r.Data(nil)
							d.SetId("????????????????") // ID not used in Delete.
							d.Set("mesh_name", meshName)
							d.Set("name", gatewayRouteName)
							d.Set("virtual_gateway_name", virtualGatewayName)
							err := r.Delete(d, client)

							if err != nil {
								log.Printf("[ERROR] %s", err)
								sweeperErrs = multierror.Append(sweeperErrs, err)
								continue
							}
						}

						return !lastPage
					})

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving App Mesh service mesh (%s) virtual gateway (%s) gateway routes: %w", meshName, virtualGatewayName, err))
					}
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving App Mesh service mesh (%s) virtual gateways: %w", meshName, err))
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Appmesh virtual gateway sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving App Mesh virtual gateways: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepMeshes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AppMeshConn

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mesh := range page.Meshes {
			name := aws.StringValue(mesh.MeshName)

			input := &appmesh.DeleteMeshInput{
				MeshName: aws.String(name),
			}

			log.Printf("[INFO] Deleting Appmesh Mesh: %s", name)
			_, err := conn.DeleteMesh(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting Appmesh Mesh (%s): %s", name, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Appmesh Mesh sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Meshes: %s", err)
	}

	return nil
}

func sweepRoutes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AppMeshConn

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mesh := range page.Meshes {
			listVirtualRoutersInput := &appmesh.ListVirtualRoutersInput{
				MeshName: mesh.MeshName,
			}
			meshName := aws.StringValue(mesh.MeshName)

			err := conn.ListVirtualRoutersPages(listVirtualRoutersInput, func(page *appmesh.ListVirtualRoutersOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, virtualRouter := range page.VirtualRouters {
					listRoutesInput := &appmesh.ListRoutesInput{
						MeshName:          mesh.MeshName,
						VirtualRouterName: virtualRouter.VirtualRouterName,
					}
					virtualRouterName := aws.StringValue(virtualRouter.VirtualRouterName)

					err := conn.ListRoutesPages(listRoutesInput, func(page *appmesh.ListRoutesOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, route := range page.Routes {
							input := &appmesh.DeleteRouteInput{
								MeshName:          mesh.MeshName,
								RouteName:         route.RouteName,
								VirtualRouterName: virtualRouter.VirtualRouterName,
							}
							routeName := aws.StringValue(route.RouteName)

							log.Printf("[INFO] Deleting Appmesh Mesh (%s) Virtual Router (%s) Route: %s", meshName, virtualRouterName, routeName)
							_, err := conn.DeleteRoute(input)

							if err != nil {
								log.Printf("[ERROR] Error deleting Appmesh Mesh (%s) Virtual Router (%s) Route (%s): %s", meshName, virtualRouterName, routeName, err)
							}
						}

						return !lastPage
					})

					if err != nil {
						log.Printf("[ERROR] Error retrieving Appmesh Mesh (%s) Virtual Router (%s) Routes: %s", meshName, virtualRouterName, err)
					}
				}

				return !lastPage
			})

			if err != nil {
				log.Printf("[ERROR] Error retrieving Appmesh Mesh (%s) Virtual Routers: %s", meshName, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Appmesh Mesh sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Meshes: %s", err)
	}

	return nil
}

func sweepVirtualGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).AppMeshConn

	var sweeperErrs *multierror.Error

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mesh := range page.Meshes {
			meshName := aws.StringValue(mesh.MeshName)

			err = conn.ListVirtualGatewaysPages(&appmesh.ListVirtualGatewaysInput{MeshName: mesh.MeshName}, func(page *appmesh.ListVirtualGatewaysOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, virtualGateway := range page.VirtualGateways {
					virtualGatewayName := aws.StringValue(virtualGateway.VirtualGatewayName)

					log.Printf("[INFO] Deleting App Mesh service mesh (%s) virtual gateway: %s", meshName, virtualGatewayName)
					r := ResourceVirtualGateway()
					d := r.Data(nil)
					d.SetId("????????????????") // ID not used in Delete.
					d.Set("mesh_name", meshName)
					d.Set("name", virtualGatewayName)
					err := r.Delete(d, client)

					if err != nil {
						log.Printf("[ERROR] %s", err)
						sweeperErrs = multierror.Append(sweeperErrs, err)
						continue
					}
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving App Mesh service mesh (%s) virtual gateways: %w", meshName, err))
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Appmesh virtual gateway sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving App Mesh virtual gateways: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVirtualNodes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).AppMeshConn

	var sweeperErrs *multierror.Error

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mesh := range page.Meshes {
			listVirtualNodesInput := &appmesh.ListVirtualNodesInput{
				MeshName: mesh.MeshName,
			}
			meshName := aws.StringValue(mesh.MeshName)

			err := conn.ListVirtualNodesPages(listVirtualNodesInput, func(page *appmesh.ListVirtualNodesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, virtualNode := range page.VirtualNodes {
					input := &appmesh.DeleteVirtualNodeInput{
						MeshName:        mesh.MeshName,
						VirtualNodeName: virtualNode.VirtualNodeName,
					}
					virtualNodeName := aws.StringValue(virtualNode.VirtualNodeName)

					log.Printf("[INFO] Deleting Appmesh Mesh (%s) Virtual Node: %s", meshName, virtualNodeName)
					_, err := conn.DeleteVirtualNode(input)

					if err != nil {
						sweeperErr := fmt.Errorf("error deleting Appmesh Mesh (%s) Virtual Node (%s): %w", meshName, virtualNodeName, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
						continue
					}
				}

				return !lastPage
			})

			if err != nil {
				log.Printf("[ERROR] Error retrieving Appmesh Mesh (%s) Virtual Nodes: %s", meshName, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Appmesh Virtual Node sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Virtual Nodes: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVirtualRouters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AppMeshConn

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mesh := range page.Meshes {
			listVirtualRoutersInput := &appmesh.ListVirtualRoutersInput{
				MeshName: mesh.MeshName,
			}
			meshName := aws.StringValue(mesh.MeshName)

			err := conn.ListVirtualRoutersPages(listVirtualRoutersInput, func(page *appmesh.ListVirtualRoutersOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, virtualRouter := range page.VirtualRouters {
					input := &appmesh.DeleteVirtualRouterInput{
						MeshName:          mesh.MeshName,
						VirtualRouterName: virtualRouter.VirtualRouterName,
					}
					virtualRouterName := aws.StringValue(virtualRouter.VirtualRouterName)

					log.Printf("[INFO] Deleting Appmesh Mesh (%s) Virtual Router: %s", meshName, virtualRouterName)
					_, err := conn.DeleteVirtualRouter(input)

					if err != nil {
						log.Printf("[ERROR] Error deleting Appmesh Mesh (%s) Virtual Router (%s): %s", meshName, virtualRouterName, err)
					}
				}

				return !lastPage
			})

			if err != nil {
				log.Printf("[ERROR] Error retrieving Appmesh Mesh (%s) Virtual Routers: %s", meshName, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Appmesh Virtual Router sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Virtual Routers: %s", err)
	}

	return nil
}

func sweepVirtualServices(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AppMeshConn

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mesh := range page.Meshes {
			listVirtualServicesInput := &appmesh.ListVirtualServicesInput{
				MeshName: mesh.MeshName,
			}
			meshName := aws.StringValue(mesh.MeshName)

			err := conn.ListVirtualServicesPages(listVirtualServicesInput, func(page *appmesh.ListVirtualServicesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, virtualService := range page.VirtualServices {
					input := &appmesh.DeleteVirtualServiceInput{
						MeshName:           mesh.MeshName,
						VirtualServiceName: virtualService.VirtualServiceName,
					}
					virtualServiceName := aws.StringValue(virtualService.VirtualServiceName)

					log.Printf("[INFO] Deleting Appmesh Mesh (%s) Virtual Service: %s", meshName, virtualServiceName)
					_, err := conn.DeleteVirtualService(input)

					if err != nil {
						log.Printf("[ERROR] Error deleting Appmesh Mesh (%s) Virtual Service (%s): %s", meshName, virtualServiceName, err)
					}
				}

				return !lastPage
			})

			if err != nil {
				log.Printf("[ERROR] Error retrieving Appmesh Mesh (%s) Virtual Services: %s", meshName, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Appmesh Virtual Service sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Virtual Services: %s", err)
	}

	return nil
}
