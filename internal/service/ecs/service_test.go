package ecs_test

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccECSService_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "alarms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "REPLICA"),
				),
			},

			{
				Config: testAccServiceConfig_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "alarms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "REPLICA"),
				),
			},
		},
	})
}

func TestAccECSService_basicImport(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"
	importInput := fmt.Sprintf("%s/%s", rName, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_familyAndRevision(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
				),
			},
			// Test existent resource import
			{
				ResourceName:      resourceName,
				ImportStateId:     importInput,
				ImportState:       true,
				ImportStateVerify: true,
				// wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"wait_for_steady_state"},
			},
			// Test non-existent resource import
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/nonexistent", rName),
				ImportState:       true,
				ImportStateVerify: false,
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccECSService_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecs.ResourceService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSService_PlacementStrategy_unnormalized(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_interchangeablePlacementStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
				),
			},
		},
	})
}

func TestAccECSService_CapacityProviderStrategy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_capacityProviderStrategy(rName, 1, 0, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
				),
			},
			{
				Config: testAccServiceConfig_capacityProviderStrategy(rName, 10, 1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
				),
			},
		},
	})
}

func TestAccECSService_CapacityProviderStrategy_forceNewDeployment(t *testing.T) {
	ctx := acctest.Context(t)
	var service1, service2 ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_capacityProviderStrategy(rName, 1, 0, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service1),
				),
			},
			{
				Config: testAccServiceConfig_capacityProviderStrategy(rName, 10, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service2),
					testAccCheckServiceNotRecreated(&service1, &service2),
				),
			},
		},
	})
}

func TestAccECSService_CapacityProviderStrategy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var service1, service2 ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_updateCapacityProviderStrategyRemove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service1),
				),
			},
			{
				Config: testAccServiceConfig_updateCapacityProviderStrategy(rName, 1, "FARGATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service2),
				),
			},
			{
				Config: testAccServiceConfig_updateCapacityProviderStrategy(rName, 1, "FARGATE_SPOT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service1),
					testAccCheckServiceNotRecreated(&service1, &service2),
				),
			},
			{
				Config: testAccServiceConfig_updateCapacityProviderStrategyRemove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service1),
				),
			},
		},
	})
}

func TestAccECSService_CapacityProviderStrategy_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_multipleCapacityProviderStrategies(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "2"),
				),
			},
		},
	})
}

func TestAccECSService_familyAndRevision(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_familyAndRevision(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
				),
			},

			{
				Config: testAccServiceConfig_familyAndRevisionModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/2427
func TestAccECSService_renamedCluster(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_renamedCluster(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttrPair(resourceName, "cluster", "aws_ecs_cluster.default", "arn"),
				),
			},

			{
				Config: testAccServiceConfig_renamedCluster(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttrPair(resourceName, "cluster", "aws_ecs_cluster.default", "arn"),
				),
			},
		},
	})
}

func TestAccECSService_healthCheckGracePeriodSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceConfig_healthCheckGracePeriodSeconds(rName, -1),
				ExpectError: regexp.MustCompile(`expected health_check_grace_period_seconds to be in the range`),
			},
			{
				Config:      testAccServiceConfig_healthCheckGracePeriodSeconds(rName, math.MaxInt32+1),
				ExpectError: regexp.MustCompile(`expected health_check_grace_period_seconds to be in the range`),
			},
			{
				Config: testAccServiceConfig_healthCheckGracePeriodSeconds(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "300"),
				),
			},
			{
				Config: testAccServiceConfig_healthCheckGracePeriodSeconds(rName, 600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "600"),
				),
			},
			{
				Config: testAccServiceConfig_healthCheckGracePeriodSeconds(rName, math.MaxInt32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "2147483647"),
				),
			},
		},
	})
}

func TestAccECSService_iamRole(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
				),
			},
		},
	})
}

func TestAccECSService_DeploymentControllerType_codeDeploy(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentControllerTypeCodeDeploy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", ecs.DeploymentControllerTypeCodeDeploy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				// Resource currently defaults to importing task_definition as family:revision
				// and wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"task_definition", "wait_for_steady_state"},
			},
		},
	})
}

func TestAccECSService_DeploymentControllerType_codeDeployUpdateDesiredCountAndHealthCheckGracePeriod(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentControllerTypeCodeDeployUpdate(rName, 1, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", ecs.DeploymentControllerTypeCodeDeploy),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
				),
			},
			{
				Config: testAccServiceConfig_deploymentControllerTypeCodeDeployUpdate(rName, 2, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "2"),
				),
			},
			{
				Config: testAccServiceConfig_deploymentControllerTypeCodeDeployUpdate(rName, 2, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "120"),
				),
			},
		},
	})
}

func TestAccECSService_DeploymentControllerType_external(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentControllerTypeExternal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", ecs.DeploymentControllerTypeExternal),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				// wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"wait_for_steady_state"},
			},
		},
	})
}

func TestAccECSService_Alarms(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_alarms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "alarms.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_DeploymentValues_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentValues(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_maximum_percent", "200"),
					resource.TestCheckResourceAttr(resourceName, "deployment_minimum_healthy_percent", "100"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform-provider-aws/issues/6315
func TestAccECSService_DeploymentValues_minZeroMaxOneHundred(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentPercents(rName, 0, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_maximum_percent", "100"),
					resource.TestCheckResourceAttr(resourceName, "deployment_minimum_healthy_percent", "0"),
				),
			},
		},
	})
}

func TestAccECSService_deploymentCircuitBreaker(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentCircuitBreaker(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_circuit_breaker.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_circuit_breaker.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "deployment_circuit_breaker.0.rollback", "true"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3444
func TestAccECSService_loadBalancerChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var s1, s2 ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_lbChanges(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &s1),
				),
			},
			{
				Config: testAccServiceConfig_lbChangesModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &s2),
					testAccCheckServiceNotRecreated(&s2, &s1),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3361
func TestAccECSService_clusterName(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_clusterName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "cluster", rName),
				),
			},
		},
	})
}

func TestAccECSService_alb(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_alb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_multipleTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_multipleTargetGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", "2"),
				),
			},
		},
	})
}

func TestAccECSService_forceNewDeployment(t *testing.T) {
	ctx := acctest.Context(t)
	var service1, service2 ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service1),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "0"),
				),
			},
			{
				Config: testAccServiceConfig_forceNewDeployment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service2),
					testAccCheckServiceNotRecreated(&service1, &service2),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
			},
		},
	})
}

func TestAccECSService_forceNewDeploymentTriggers(t *testing.T) {
	ctx := acctest.Context(t)
	var service1, service2 ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_forceNewDeployment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service1),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
			},
			{
				Config: testAccServiceConfig_forceNewDeploymentTriggers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service2),
					testAccCheckServiceNotRecreated(&service1, &service2),
					resource.TestCheckResourceAttr(resourceName, "force_new_deployment", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "triggers.update"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
			},
		},
	})
}

func TestAccECSService_PlacementStrategy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service1, service2, service3, service4 ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service1),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "0"),
				),
			},
			{
				Config: testAccServiceConfig_placementStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service2),
					testAccCheckServiceNotRecreated(&service1, &service2),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
			},
			{
				Config: testAccServiceConfig_randomPlacementStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service3),
					testAccCheckServiceNotRecreated(&service2, &service3),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "random"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", ""),
				),
			},
			{
				Config: testAccServiceConfig_multiplacementStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service4),
					testAccCheckServiceNotRecreated(&service3, &service4),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.1.type", "spread"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.1.field", "instanceId"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13146
func TestAccECSService_PlacementStrategy_missing(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceConfig_placementStrategyType(rName, ""),
				ExpectError: regexp.MustCompile(`expected ordered_placement_strategy.0.type to be one of`),
			},
		},
	})
}

func TestAccECSService_PlacementConstraints_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service1, service2 ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_placementConstraint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service1),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
				),
			},
			{
				Config: testAccServiceConfig_placementConstraintEmptyExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service2),
					testAccCheckServiceNotRecreated(&service1, &service2),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_PlacementConstraints_emptyExpression(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_placementConstraintEmptyExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_LaunchTypeFargate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_launchTypeFargate(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "LATEST"),
				),
			},
			{
				Config: testAccServiceConfig_launchTypeFargate(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "true"),
				),
			},
			{
				Config: testAccServiceConfig_launchTypeFargate(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
				),
			},
		},
	})
}

func TestAccECSService_LaunchTypeFargate_platformVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_launchTypeFargateAndPlatformVersion(rName, "1.3.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.3.0"),
				),
			},
			{
				Config: testAccServiceConfig_launchTypeFargateAndPlatformVersion(rName, "LATEST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "LATEST"),
				),
			},
			{
				Config: testAccServiceConfig_launchTypeFargateAndPlatformVersion(rName, "1.4.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.4.0"),
				),
			},
		},
	})
}

func TestAccECSService_LaunchTypeFargate_waitForSteadyState(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Wait for the ECS Cluster to reach a steady state w/specified count
				Config: testAccServiceConfig_launchTypeFargateAndWait(rName, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				// Resource currently defaults to importing task_definition as family:revision
				// and wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"task_definition", "wait_for_steady_state"},
			},
		},
	})
}

func TestAccECSService_LaunchTypeFargate_updateWaitForSteadyState(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_launchTypeFargateNoWait(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", "false"),
				),
			},
			{
				// Modify desired count and wait for the ECS Cluster to reach steady state
				Config: testAccServiceConfig_launchTypeFargateAndWait(rName, 2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", "true"),
				),
			},
			{
				// Modify desired count without wait
				Config: testAccServiceConfig_launchTypeFargateAndWait(rName, 1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", "false"),
				),
			},
		},
	})
}

func TestAccECSService_LaunchTypeEC2_network(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_networkConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
				),
			},
			{
				Config: testAccServiceConfig_networkConfigurationModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
				),
			},
		},
	})
}

func TestAccECSService_DaemonSchedulingStrategy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_daemonSchedulingStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "DAEMON"),
				),
			},
		},
	})
}

func TestAccECSService_DaemonSchedulingStrategy_setDeploymentMinimum(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_daemonSchedulingStrategySetDeploymentMinimum(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "DAEMON"),
				),
			},
		},
	})
}

func TestAccECSService_replicaSchedulingStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_replicaSchedulingStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "REPLICA"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceRegistries_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_registries(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceRegistries_container(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_registriesContainer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceRegistries_changes(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	serviceDiscoveryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedServiceDiscoveryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_registriesChanges(rName, serviceDiscoveryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
			{
				Config: testAccServiceConfig_registriesChanges(rName, updatedServiceDiscoveryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceConnect_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceConnect_full(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectAllAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceConnect_ingressPortOverride(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectIngressPortOverride(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.log_configuration.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "service_connect_configuration.0.namespace"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.discovery_name", ""),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.ingress_port_override", "0"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.port_name", "nginx-http"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceConnect_remove(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.enabled", "true"),
				),
			},
			{
				Config: testAccServiceConfig_serviceConnectRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccECSService_Tags_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				// Resource currently defaults to importing task_definition as family:revision
				// and wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"task_definition", "wait_for_steady_state"},
			},
			{
				Config: testAccServiceConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServiceConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccECSService_Tags_managed(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_managedTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "enable_ecs_managed_tags", "true"),
				),
			},
		},
	})
}

func TestAccECSService_Tags_propagate(t *testing.T) {
	ctx := acctest.Context(t)
	var first, second, third ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_propagateTags(rName, "SERVICE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", ecs.PropagateTagsService),
				),
			},
			{
				Config: testAccServiceConfig_propagateTags(rName, "TASK_DEFINITION"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", ecs.PropagateTagsTaskDefinition),
				),
			},
			{
				Config: testAccServiceConfig_propagateTags(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &third),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", ecs.PropagateTagsNone),
				),
			},
		},
	})
}

func TestAccECSService_executeCommand(t *testing.T) {
	ctx := acctest.Context(t)
	var service ecs.Service
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_executeCommand(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "enable_execute_command", "true"),
				),
			},
			{
				Config: testAccServiceConfig_executeCommand(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "enable_execute_command", "false"),
				),
			},
		},
	})
}

func testAccCheckServiceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_service" {
				continue
			}

			service, err := tfecs.FindServiceNoTagsByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["cluster"])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			if aws.StringValue(service.Status) == "INACTIVE" {
				return nil
			}

			return fmt.Errorf("ECS service still exists:\n%#v", service)
		}

		return nil
	}
}

func testAccCheckServiceExists(ctx context.Context, name string, service *ecs.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn()

		err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
			var err error
			service, err = tfecs.FindServiceNoTagsByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["cluster"])
			if tfresource.NotFound(err) {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
				return resource.RetryableError(err)
			}
			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		return err
	}
}

func testAccCheckServiceNotRecreated(i, j *ecs.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreatedAt).Equal(aws.TimeValue(j.CreatedAt)) {
			return fmt.Errorf("ECS Service (%s) unexpectedly recreated", aws.StringValue(j.ServiceArn))
		}

		return nil
	}
}

func testAccServiceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}
`, rName)
}

func testAccServiceConfig_modified(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 2
}
`, rName)
}

func testAccServiceConfig_launchTypeFargateBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count          = 2
  subnet_id      = element(aws_subnet.test[*].id, count.index)
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  count = 2

  name        = "%[1]s-${count.index}"
  description = "Allow all traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"

    cidr_blocks = [
      "0.0.0.0/0",
    ]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}
`, rName))
}

func testAccServiceConfig_launchTypeFargate(rName string, assignPublicIP bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_launchTypeFargateBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = aws_security_group.test[*].id
    subnets          = aws_subnet.test[*].id
    assign_public_ip = %[2]t
  }
}
`, rName, assignPublicIP))
}

func testAccServiceConfig_launchTypeFargateAndPlatformVersion(rName, platformVersion string) string {
	return acctest.ConfigCompose(testAccServiceConfig_launchTypeFargateBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name             = %[1]q
  cluster          = aws_ecs_cluster.test.id
  task_definition  = aws_ecs_task_definition.test.arn
  desired_count    = 1
  launch_type      = "FARGATE"
  platform_version = %[2]q

  network_configuration {
    security_groups  = aws_security_group.test[*].id
    subnets          = aws_subnet.test[*].id
    assign_public_ip = false
  }
}
`, rName, platformVersion))
}

func testAccServiceConfig_launchTypeFargateNoWait(rName string) string {
	return acctest.ConfigCompose(testAccServiceConfig_launchTypeFargateBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.test[0].id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }
}
`, rName))
}

func testAccServiceConfig_launchTypeFargateAndWait(rName string, desiredCount int, waitForSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_launchTypeFargateBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = %[2]d
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.test[0].id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  wait_for_steady_state = %[3]t
}

`, rName, desiredCount, waitForSteadyState))
}

func testAccServiceConfig_interchangeablePlacementStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  ordered_placement_strategy {
    field = "host"
    type  = "spread"
  }
}
`, rName)
}

func testAccServiceConfig_capacityProviderStrategy(rName string, weight, base int, forceNewDeployment bool) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}

resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                 = %[1]q
  cluster              = aws_ecs_cluster.default.id
  task_definition      = aws_ecs_task_definition.test.arn
  desired_count        = 1
  force_new_deployment = %[4]t

  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.test.name
    weight            = %[2]d
    base              = %[3]d
  }
}
`, rName, weight, base, forceNewDeployment))
}

func testAccServiceConfig_updateCapacityProviderStrategy(rName string, weight int, capacityProvider string) string {
	return acctest.ConfigCompose(
		testAccCapacityProviderConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_service" "test" {
  name                 = %[1]q
  cluster              = aws_ecs_cluster.test.id
  task_definition      = aws_ecs_task_definition.test.arn
  desired_count        = 1
  force_new_deployment = true

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = [aws_subnet.test.id]
    assign_public_ip = false
  }

  capacity_provider_strategy {
    capacity_provider = %[3]q
    weight            = %[2]d
  }
}
`, rName, weight, capacityProvider))
}

func testAccServiceConfig_updateCapacityProviderStrategyRemove(rName string) string {
	return acctest.ConfigCompose(
		testAccCapacityProviderConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test2" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                 = %[1]q
  cluster              = aws_ecs_cluster.test.id
  task_definition      = aws_ecs_task_definition.test2.arn
  desired_count        = 1
  force_new_deployment = true
}
`, rName))
}

func testAccServiceConfig_multipleCapacityProviderStrategies(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_capacityProviders(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  network_configuration {
    security_groups  = [aws_security_group.allow_all.id]
    subnets          = [aws_subnet.test.id]
    assign_public_ip = false
  }

  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_security_group" "allow_all" {
  name        = %[1]q
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_subnet" "test" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServiceConfig_forceNewDeployment(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster              = aws_ecs_cluster.default.id
  desired_count        = 1
  force_new_deployment = true
  name                 = %[1]q
  task_definition      = aws_ecs_task_definition.test.arn

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }
}
`, rName)
}

func testAccServiceConfig_forceNewDeploymentTriggers(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster              = aws_ecs_cluster.default.id
  desired_count        = 1
  force_new_deployment = true
  name                 = %[1]q
  task_definition      = aws_ecs_task_definition.test.arn

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }

  triggers = {
    update = timestamp()
  }
}
`, rName)
}

func testAccServiceConfig_placementStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }
}
`, rName)
}

func testAccServiceConfig_placementStrategyType(rName string, placementStrategyType string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 1
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  ordered_placement_strategy {
    type = %[1]q
  }
}
`, rName, placementStrategyType)
}

func testAccServiceConfig_randomPlacementStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  ordered_placement_strategy {
    type = "random"
  }
}
`, rName)
}

func testAccServiceConfig_multiplacementStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }

  ordered_placement_strategy {
    field = "host"
    type  = "spread"
  }
}
`, rName)
}

func testAccServiceConfig_placementConstraint(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  placement_constraints {
    type       = "memberOf"
    expression = "attribute:ecs.availability-zone in [${data.aws_availability_zones.available.names[0]}]"
  }
}
`, rName)
}

func testAccServiceConfig_placementConstraintEmptyExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  placement_constraints {
    type = "distinctInstance"
  }
}
`, rName)
}

func testAccServiceConfig_healthCheckGracePeriodSeconds(rName string, healthCheckGracePeriodSeconds int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "ghost:latest",
    "memory": 512,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}

resource "aws_iam_role" "ecs_service" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_service" {
  name = %[1]q
  role = aws_iam_role.ecs_service.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
        "elasticloadbalancing:DeregisterTargets",
        "elasticloadbalancing:Describe*",
        "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
        "elasticloadbalancing:RegisterTargets"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lb_target_group" "test" {
  name     = aws_lb.test.name
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_ecs_service" "test" {
  name                              = %[1]q
  cluster                           = aws_ecs_cluster.test.id
  task_definition                   = aws_ecs_task_definition.test.arn
  desired_count                     = 1
  health_check_grace_period_seconds = %d
  iam_role                          = aws_iam_role.ecs_service.name

  load_balancer {
    target_group_arn = aws_lb_target_group.test.id
    container_name   = "ghost"
    container_port   = "2368"
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName, healthCheckGracePeriodSeconds)
}

func testAccServiceConfig_iamRole(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}

resource "aws_iam_role" "ecs_service" {
  name = %[1]q

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {"AWS": "*"},
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_service" {
  name = %[1]q
  role = aws_iam_role.ecs_service.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:*",
        "ec2:*",
        "ecs:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_elb" "test" {
  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 8080
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  iam_role        = aws_iam_role.ecs_service.name

  load_balancer {
    elb_name       = aws_elb.test.id
    container_name = "ghost"
    container_port = "2368"
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName)
}

func testAccServiceConfig_alarms(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  alarms {
    enable   = true
    rollback = true
    alarm_names = [
      aws_cloudwatch_metric_alarm.test.alarm_name
    ]
  }
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUReservation"
  namespace                 = "AWS/ECS"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  insufficient_data_actions = []
}
`, rName)
}

func testAccServiceConfig_deploymentValues(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}
`, rName)
}

func testAccServiceLBChangesConfig_base(rName, image, containerName string, containerPort, hostPort int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": %[2]q,
    "memory": 128,
    "name": %[3]q,
    "portMappings": [
      {
        "containerPort": %[4]d,
        "hostPort": %[5]d
      }
    ]
  }
]
DEFINITION
}

resource "aws_iam_role" "ecs_service" {
  name = %[1]q

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {"AWS": "*"},
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_service" {
  name = %[1]q
  role = aws_iam_role.ecs_service.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:*",
        "ec2:*",
        "ecs:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_lb_target_group" "test" {
  name     = aws_lb.test.name
  port     = %[5]d
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  iam_role        = aws_iam_role.ecs_service.name

  load_balancer {
    target_group_arn = aws_lb_target_group.test.id
    container_name   = %[3]q
    container_port   = %[4]d
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName, image, containerName, containerPort, hostPort))
}

func testAccServiceConfig_lbChanges(rName string) string {
	return testAccServiceLBChangesConfig_base(rName, "ghost:latest", "ghost", 2368, 8080)
}

func testAccServiceConfig_lbChangesModified(rName string) string {
	return testAccServiceLBChangesConfig_base(rName, "nginx:latest", "nginx", 80, 8080)
}

func testAccServiceConfig_familyAndRevision(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "jenkins:latest",
    "memory": 128,
    "name": "jenkins"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  desired_count   = 1
}
`, rName)
}

func testAccServiceConfig_familyAndRevisionModified(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "jenkins:latest",
    "memory": 128,
    "name": "jenkins"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  desired_count   = 1
}
`, rName)
}

func testAccServiceConfig_renamedCluster(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  desired_count   = 1
}
`, rName)
}

func testAccServiceConfig_clusterName(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "jenkins:latest",
    "memory": 128,
    "name": "jenkins"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.name
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}
`, rName)
}

func testAccServiceConfig_alb(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "ghost:latest",
    "memory": 512,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}

resource "aws_iam_role" "ecs_service" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_service" {
  name = %[1]q
  role = aws_iam_role.ecs_service.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
        "elasticloadbalancing:DeregisterTargets",
        "elasticloadbalancing:Describe*",
        "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
        "elasticloadbalancing:RegisterTargets"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lb_target_group" "test" {
  name     = aws_lb.test.name
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  iam_role        = aws_iam_role.ecs_service.name

  load_balancer {
    target_group_arn = aws_lb_target_group.test.id
    container_name   = "ghost"
    container_port   = "2368"
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName))
}

func testAccServiceConfig_multipleTargetGroups(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "ghost:latest",
    "memory": 512,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 2368
      },
      {
        "containerPort": 4501,
        "hostPort": 4501
      }
    ]
  }
]
DEFINITION
}

resource "aws_lb_target_group" "test" {
  name     = "${aws_lb.test.name}1"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_target_group" "static" {
  name     = "${aws_lb.test.name}2"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.static.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  load_balancer {
    target_group_arn = aws_lb_target_group.test.id
    container_name   = "ghost"
    container_port   = "2368"
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.static.id
    container_name   = "ghost"
    container_port   = "4501"
  }
}
`, rName))
}

func testAccServiceNetworkConfigurationConfig_base(rName, securityGroups string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                = %[1]q
  network_mode          = "awsvpc"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  network_configuration {
    security_groups = [%[2]s]
    subnets         = aws_subnet.test[*].id
  }
}
`, rName, securityGroups))
}

func testAccServiceConfig_networkConfiguration(rName string) string {
	return testAccServiceNetworkConfigurationConfig_base(rName, "aws_security_group.test[0].id, aws_security_group.test[1].id")
}

func testAccServiceConfig_networkConfigurationModified(rName string) string {
	return testAccServiceNetworkConfigurationConfig_base(rName, "aws_security_group.test[0].id")
}

func testAccServiceConfig_registries(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = "%[1]s.terraform.local"
  description = "test"
  vpc         = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "SRV"
    }
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "awsvpc"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  service_registries {
    port         = 34567
    registry_arn = aws_service_discovery_service.test.arn
  }

  network_configuration {
    security_groups = [aws_security_group.test.id]
    subnets         = aws_subnet.test[*].id
  }
}
`, rName))
}

func testAccServiceConfig_registriesContainer(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = "%[1]s.terraform.local"
  description = "test"
  vpc         = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "SRV"
    }
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "bridge"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb",
    "portMappings": [
    {
      "hostPort": 0,
      "protocol": "tcp",
      "containerPort": 27017
    }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  service_registries {
    container_name = "mongodb"
    container_port = 27017
    registry_arn   = aws_service_discovery_service.test.arn
  }
}
`, rName))
}

func testAccServiceConfig_registriesChanges(rName, discoveryName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = "%[2]s.terraform.local"
  description = "test"
  vpc         = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[2]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "SRV"
    }
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "awsvpc"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  service_registries {
    port         = 34567
    registry_arn = aws_service_discovery_service.test.arn
  }

  network_configuration {
    security_groups = [aws_security_group.test.id]
    subnets         = aws_subnet.test[*].id
  }
}
`, rName, discoveryName))
}

func testAccServiceConfig_daemonSchedulingStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                = %[1]q
  cluster             = aws_ecs_cluster.default.id
  task_definition     = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  scheduling_strategy = "DAEMON"
}
`, rName)
}

func testAccServiceConfig_daemonSchedulingStrategySetDeploymentMinimum(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                               = %[1]q
  cluster                            = aws_ecs_cluster.default.id
  task_definition                    = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  scheduling_strategy                = "DAEMON"
  deployment_minimum_healthy_percent = "50"
}
`, rName)
}

func testAccServiceConfig_deploymentControllerTypeCodeDeploy(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  name     = aws_lb.test.name
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "test",
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 0
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  deployment_controller {
    type = "CODE_DEPLOY"
  }

  load_balancer {
    container_name   = "test"
    container_port   = "80"
    target_group_arn = aws_lb_target_group.test.id
  }
}
`, rName))
}

func testAccServiceConfig_deploymentControllerTypeCodeDeployUpdate(rName string, desiredReplicas, healthCheckGracePeriodSeconds int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"

    cidr_blocks = [
      "0.0.0.0/0",
    ]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count          = 2
  subnet_id      = element(aws_subnet.test[*].id, count.index)
  route_table_id = aws_route_table.test.id
}

resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  target_type = "ip"
  name        = aws_lb.test.name
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.test.id
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<DEFINITION
[
  {
    "essential": true,
    "image": "nginx:latest",
    "name": "test",
    "portMappings": [
      {
        "containerPort": 80,
        "protocol": "tcp"
      }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster                           = aws_ecs_cluster.test.id
  desired_count                     = %[2]d
  name                              = %[1]q
  task_definition                   = aws_ecs_task_definition.test.arn
  health_check_grace_period_seconds = %[3]d

  deployment_controller {
    type = "CODE_DEPLOY"
  }

  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    base              = 1
    weight            = 1
  }

  load_balancer {
    container_name   = "test"
    container_port   = "80"
    target_group_arn = aws_lb_target_group.test.id
  }

  network_configuration {
    subnets          = aws_subnet.test[*].id
    security_groups  = [aws_security_group.test.id]
    assign_public_ip = true
  }
}
`, rName, desiredReplicas, healthCheckGracePeriodSeconds))
}

func testAccServiceConfig_deploymentControllerTypeExternal(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_service" "test" {
  cluster       = aws_ecs_cluster.test.id
  desired_count = 0
  name          = %[1]q

  deployment_controller {
    type = "EXTERNAL"
  }
}
`, rName)
}

func testAccServiceConfig_deploymentPercents(rName string, deploymentMinimumHealthyPercent, deploymentMaximumPercent int) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster                            = aws_ecs_cluster.test.id
  deployment_maximum_percent         = %[2]d
  deployment_minimum_healthy_percent = %[3]d
  desired_count                      = 1
  name                               = %[1]q
  task_definition                    = aws_ecs_task_definition.test.arn
}
`, rName, deploymentMaximumPercent, deploymentMinimumHealthyPercent)
}

func testAccServiceConfig_deploymentCircuitBreaker(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 1
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }
}
`, rName)
}

func testAccServiceConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 0
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccServiceConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 0
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccServiceConfig_managedTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster                 = aws_ecs_cluster.test.id
  desired_count           = 0
  name                    = %[1]q
  task_definition         = aws_ecs_task_definition.test.arn
  enable_ecs_managed_tags = true

  tags = {
    tag-key = "tag-value"
  }
}
`, rName)
}

func testAccServiceConfig_propagateTags(rName, propagate string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION

  tags = {
    tag-key = "task-def"
  }
}

resource "aws_ecs_service" "test" {
  cluster                 = aws_ecs_cluster.test.id
  desired_count           = 0
  name                    = %[1]q
  task_definition         = aws_ecs_task_definition.test.arn
  enable_ecs_managed_tags = true
  propagate_tags          = %[2]q

  tags = {
    tag-key = "service"
  }
}
`, rName, propagate)
}

func testAccServiceConfig_replicaSchedulingStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                = %[1]q
  cluster             = aws_ecs_cluster.default.id
  task_definition     = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  scheduling_strategy = "REPLICA"
  desired_count       = 1
}
`, rName)
}

func testAccServiceConfig_executeCommand(rName string, enable bool) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

  inline_policy {
    name = "exec_policy"

    policy = <<EOF
{
   "Version": "2012-10-17",
   "Statement": [
       {
       "Effect": "Allow",
       "Action": [
            "ssmmessages:CreateControlChannel",
            "ssmmessages:CreateDataChannel",
            "ssmmessages:OpenControlChannel",
            "ssmmessages:OpenDataChannel"
       ],
      "Resource": "*"
      }
   ]
}
EOF
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  task_role_arn = aws_iam_role.test.arn

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster                = aws_ecs_cluster.test.id
  desired_count          = 0
  name                   = %[1]q
  task_definition        = aws_ecs_task_definition.test.arn
  enable_execute_command = %[2]t
}
`, rName, enable)
}

func testAccServiceConfig_serviceConnectBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q

  service_connect_defaults {
    namespace = aws_service_discovery_http_namespace.test.arn
  }
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "bridge"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb",
    "portMappings": [
    {
      "hostPort": 0,
      "protocol": "tcp",
      "containerPort": 27017
    }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  service_connect_configuration {
    enabled = true
  }
}
`, rName)
}

func testAccServiceConfig_serviceConnectAllAttributes(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "bridge"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb",
    "portMappings": [
    {
      "hostPort": 0,
      "protocol": "tcp",
      "containerPort": 27017,
      "name": "tf-test"
    }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_http_namespace.test.arn

    log_configuration {
      log_driver = "json-file"

      options = {
        key = "value"
      }
    }

    service {
      client_alias {
        dns_name = "example.com"
        port     = 8080
      }

      discovery_name        = "test"
      ingress_port_override = 8443
      port_name             = "tf-test"
    }
  }
}
`, rName)
}

func testAccServiceConfig_serviceConnectIngressPortOverride(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"

    cidr_blocks = [
      "0.0.0.0/0",
    ]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "awsvpc"

  container_definitions = jsonencode([{
    name      = "test-nginx"
    image     = "nginx"
    cpu       = 10
    memory    = 512
    essential = true
    portMappings = [{
      name          = "nginx-http"
      containerPort = 8080
      protocol      = "tcp"
      appProtocol   = "http"
    }]
  }])
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  network_configuration {
    subnets          = aws_subnet.test[*].id
    security_groups  = [aws_security_group.test.id]
    assign_public_ip = false
  }

  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_http_namespace.test.arn

    service {
      client_alias {
        port = 8080
      }

      port_name = "nginx-http"
    }
  }
}
`, rName))
}

func testAccServiceConfig_serviceConnectRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q

  service_connect_defaults {
    namespace = aws_service_discovery_http_namespace.test.arn
  }
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "bridge"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb",
    "portMappings": [
    {
      "hostPort": 0,
      "protocol": "tcp",
      "containerPort": 27017
    }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}
`, rName)
}
