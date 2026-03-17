// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func Test_GetRoleNameFromARN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arn  string
		want string
	}{
		{"empty", "", ""},
		{
			names.AttrRole,
			"arn:aws:iam::0123456789:role/EcsService", // lintignore:AWSAT005
			"EcsService",
		},
		{
			"role with path",
			"arn:aws:iam::0123456789:role/group/EcsService", // lintignore:AWSAT005
			"/group/EcsService",
		},
		{
			"role with complex path",
			"arn:aws:iam::0123456789:role/group/subgroup/my-role", // lintignore:AWSAT005
			"/group/subgroup/my-role",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tfecs.RoleNameFromARN(tt.arn); got != tt.want {
				t.Errorf("GetRoleNameFromARN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClustereNameFromARN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arn  string
		want string
	}{
		{"empty", "", ""},
		{
			"cluster",
			"arn:aws:ecs:us-west-2:0123456789:cluster/my-cluster", // lintignore:AWSAT003,AWSAT005
			"my-cluster",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tfecs.ClusterNameFromARN(tt.arn); got != tt.want {
				t.Errorf("GetClusterNameFromARN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceNameFromARN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		arn      string
		expected string
	}{
		{
			name:     "empty ARN",
			arn:      "",
			expected: "",
		},
		{
			name:     "invalid ARN",
			arn:      "invalid",
			expected: "",
		},
		{
			name:     "short ARN format",
			arn:      "arn:aws:ecs:us-west-2:123456789:service/service_name", // lintignore:AWSAT003,AWSAT005
			expected: names.AttrServiceName,
		},
		{
			name:     "long ARN format",
			arn:      "arn:aws:ecs:us-west-2:123456789:service/cluster_name/service_name", // lintignore:AWSAT003,AWSAT005
			expected: names.AttrServiceName,
		},
		{
			name:     "ARN with special characters",
			arn:      "arn:aws:ecs:us-west-2:123456789:service/cluster-name/service-name-123", // lintignore:AWSAT003,AWSAT005
			expected: "service-name-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := tfecs.ServiceNameFromARN(tt.arn)
			if actual != tt.expected {
				t.Errorf("Expected: %s, Got: %s", tt.expected, actual)
			}
		})
	}
}

func TestAccECSService_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_basic(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "alarms.#", "0"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ecs", fmt.Sprintf("service/%s/%s", clusterName, rName)),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_rebalancing", "ENABLED"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster", "aws_ecs_cluster.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "REPLICA"),
					resource.TestCheckResourceAttrPair(resourceName, "task_definition", "aws_ecs_task_definition.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "vpc_lattice_configuration.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", clusterName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_steady_state",
					"task_definition", // https://github.com/hashicorp/terraform-provider-aws/issues/42731
				},
			},
			{
				Config: testAccServiceConfig_modified(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "alarms.#", "0"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ecs", fmt.Sprintf("service/%s/%s", clusterName, rName)),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_rebalancing", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "REPLICA"),
					resource.TestCheckResourceAttr(resourceName, "vpc_lattice_configuration.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccECSService_Identity_Basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_basic(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ecs", fmt.Sprintf("service/%s/%s", clusterName, rName))),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", clusterName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_steady_state",
					"task_definition", // https://github.com/hashicorp/terraform-provider-aws/issues/42731
				},
			},
		},
	})
}

func TestAccECSService_Identity_RegionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_regionOverride(rName, clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("ecs", fmt.Sprintf("service/%s/%s", clusterName, rName))),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s@%s", clusterName, rName, acctest.AlternateRegion()),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_steady_state",
					"task_definition", // https://github.com/hashicorp/terraform-provider-aws/issues/42731
				},
			},
		},
	})
}

func TestAccECSService_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_basic(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					acctest.CheckSDKResourceDisappears(ctx, t, tfecs.ResourceService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSService_LatticeConfigurations(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster awstypes.Cluster
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"
	serviceName := "aws_ecs_service.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_vpcLatticeConfiguration_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, t, resourceName, &cluster),
					testAccCheckServiceExists(ctx, t, serviceName, &service),
					resource.TestCheckResourceAttr(serviceName, "vpc_lattice_configurations.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(serviceName, "vpc_lattice_configurations.*", map[string]string{
						"port_name": "testvpclattice",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(serviceName, "vpc_lattice_configurations.*", map[string]string{
						"port_name": "testvpclattice",
					}),
				),
			},
			{
				Config: testAccServiceConfig_vpcLatticeConfiguration_removed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, serviceName, &service),
					resource.TestCheckResourceAttr(serviceName, "vpc_lattice_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccECSService_PlacementStrategy_unnormalized(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_interchangeablePlacementStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
			},
		},
	})
}

func TestAccECSService_CapacityProviderStrategy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_capacityProviderStrategy(rName, 1, 0, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.base", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config:      testAccServiceConfig_capacityProviderStrategy(rName, 10, 1, false),
				ExpectError: regexache.MustCompile(`force_new_deployment should be true when capacity_provider_strategy is being updated`),
			},
			{
				Config: testAccServiceConfig_capacityProviderStrategy(rName, 10, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.weight", "10"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.base", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccECSService_CapacityProviderStrategy_forceNewDeployment(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_capacityProviderStrategy(rName, 1, 0, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.base", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccServiceConfig_capacityProviderStrategy(rName, 10, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.weight", "10"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.base", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccECSService_CapacityProviderStrategy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_updateCapacityProviderStrategyRemove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccServiceConfig_updateCapacityProviderStrategy(rName, 1, "FARGATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.capacity_provider", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.weight", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccServiceConfig_updateCapacityProviderStrategy(rName, 1, "FARGATE_SPOT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.capacity_provider", "FARGATE_SPOT"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.0.weight", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccServiceConfig_updateCapacityProviderStrategyRemove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccECSService_VolumeConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_volumeConfiguration_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
			},
		},
	})
}

func TestAccECSService_VolumeConfiguration_volumeInitializationRate(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_volumeConfiguration_volumeInitializationRate(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.name", "vol1"),
					resource.TestCheckResourceAttrPair(
						resourceName, "volume_configuration.0.managed_ebs_volume.0.snapshot_id",
						"aws_ebs_snapshot.test", names.AttrID,
					),
					resource.TestCheckResourceAttr(
						resourceName, "volume_configuration.0.managed_ebs_volume.0.volume_initialization_rate",
						"100",
					),
				),
			},
		},
	})
}

func TestAccECSService_VolumeConfiguration_tagSpecifications(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_volumeConfiguration_tagSpecifications(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
			},
		},
	})
}

func TestAccECSService_VolumeConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_volumeConfiguration_update(rName, "gp2", 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.size_in_gb", "8"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportStateVerify: true,
				// Resource currently defaults to importing task_definition as family:revision
				// and wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"task_definition", "wait_for_steady_state"},
			},
			{
				Config: testAccServiceConfig_volumeConfiguration_update(rName, "gp3", 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.volume_type", "gp3"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.size_in_gb", "8"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccServiceConfig_volumeConfiguration_update(rName, "gp3", 16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.volume_type", "gp3"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.size_in_gb", "16"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/38475
func TestAccECSService_VolumeConfiguration_throughputTypeChange(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ECSServiceID),
		CheckDestroy: testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.58.0",
					},
				},
				Config: testAccServiceConfig_volumeConfiguration_gp3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.name", "vol1"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.size_in_gb", "10"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.throughput", ""),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.volume_type", "gp3"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccServiceConfig_volumeConfiguration_gp3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.name", "vol1"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.size_in_gb", "10"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.throughput", "0"),
					resource.TestCheckResourceAttr(resourceName, "volume_configuration.0.managed_ebs_volume.0.volume_type", "gp3"),
				),
			},
		},
	})
}

func TestAccECSService_familyAndRevision(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_familyAndRevision(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateId:           fmt.Sprintf("%s/%s", rName, rName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_steady_state"},
			},
			{
				Config: testAccServiceConfig_familyAndRevisionModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
			},
		},
	})
}

func TestAccECSService_healthCheckGracePeriodSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceConfig_healthCheckGracePeriodSeconds(rName, -1),
				ExpectError: regexache.MustCompile(`expected health_check_grace_period_seconds to be in the range`),
			},
			{
				Config:      testAccServiceConfig_healthCheckGracePeriodSeconds(rName, math.MaxInt32+1),
				ExpectError: regexache.MustCompile(`expected health_check_grace_period_seconds to be in the range`),
			},
			{
				Config: testAccServiceConfig_healthCheckGracePeriodSeconds(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "300"),
				),
			},
			{
				Config: testAccServiceConfig_healthCheckGracePeriodSeconds(rName, 600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "600"),
				),
			},
			{
				Config: testAccServiceConfig_healthCheckGracePeriodSeconds(rName, math.MaxInt32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "2147483647"),
				),
			},
		},
	})
}

func TestAccECSService_iamRole(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
			},
		},
	})
}

func TestAccECSService_DeploymentControllerType_codeDeploy(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentControllerType(rName, string(awstypes.DeploymentControllerTypeCodeDeploy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", string(awstypes.DeploymentControllerTypeCodeDeploy)),
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
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentControllerTypeCodeDeployUpdate(rName, 1, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", string(awstypes.DeploymentControllerTypeCodeDeploy)),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
				),
			},
			{
				Config: testAccServiceConfig_deploymentControllerTypeCodeDeployUpdate(rName, 2, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "2"),
				),
			},
			{
				Config: testAccServiceConfig_deploymentControllerTypeCodeDeployUpdate(rName, 2, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "120"),
				),
			},
		},
	})
}

func TestAccECSService_DeploymentControllerType_external(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentControllerTypeExternal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", string(awstypes.DeploymentControllerTypeExternal)),
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

func TestAccECSService_DeploymentControllerMutability_codeDeployToECS(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentControllerType(rName, string(awstypes.DeploymentControllerTypeCodeDeploy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", string(awstypes.DeploymentControllerTypeCodeDeploy)),
				),
			},
			{
				Config: testAccServiceConfig_deploymentControllerType(rName, string(awstypes.DeploymentControllerTypeEcs)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", string(awstypes.DeploymentControllerTypeEcs)),
				),
			},
		},
	})
}

func TestAccECSService_alarmsAdd(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_noAlarms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "alarms.#", "0"),
				),
			},
			{
				Config: testAccServiceConfig_alarms(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "alarms.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "alarms.0.enable", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccECSService_alarmsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_alarms(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "alarms.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "alarms.0.enable", acctest.CtTrue),
				),
			},
			{
				Config: testAccServiceConfig_alarms(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "alarms.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "alarms.0.enable", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "2"),
					// Lifecycle hooks configuration checks
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.lifecycle_hook.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "deployment_configuration.0.lifecycle_hook.*.hook_target_arn", "aws_lambda_function.hook_success", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "deployment_configuration.0.lifecycle_hook.*.role_arn", "aws_iam_role.global", names.AttrARN),
					resource.TestCheckTypeSetElemAttr(resourceName, "deployment_configuration.0.lifecycle_hook.*.lifecycle_stages.*", "POST_SCALE_UP"),
					resource.TestCheckTypeSetElemAttr(resourceName, "deployment_configuration.0.lifecycle_hook.*.lifecycle_stages.*", "POST_TEST_TRAFFIC_SHIFT"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment_configuration.0.lifecycle_hook.*", map[string]string{
						"hook_details":       "[1,\"2\",true]",
						"lifecycle_stages.0": "POST_SCALE_UP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment_configuration.0.lifecycle_hook.*", map[string]string{
						"hook_details":       "3.14",
						"lifecycle_stages.0": "TEST_TRAFFIC_SHIFT",
					}),
					// Load balancer advanced configuration checks
					resource.TestCheckResourceAttr(resourceName, "load_balancer.0.advanced_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.alternate_target_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.production_listener_rule"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.test_listener_rule"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.role_arn"),
					// Service Connect test traffic rules checks
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.test_traffic_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.test_traffic_rules.0.header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.test_traffic_rules.0.header.0.name", "x-test-header"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.test_traffic_rules.0.header.0.value.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.test_traffic_rules.0.header.0.value.0.exact", "test-value"),
				),
			},
			{
				Config: testAccServiceConfig_blueGreenDeployment_withHookBehavior(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "3"),
					// Lifecycle hooks configuration checks
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.lifecycle_hook.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "deployment_configuration.0.lifecycle_hook.*.lifecycle_stages.*", "PRE_SCALE_UP"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment_configuration.0.lifecycle_hook.*", map[string]string{
						"hook_details":       "{\"bool_key\":true,\"int_key\":10,\"list_key\":[1,\"2\",true],\"object_key\":{\"bool_key\":true,\"int_key\":10,\"list_key\":[1,\"2\",true],\"string_key\":\"string_val\"},\"string_key\":\"string_val\"}",
						"lifecycle_stages.0": "PRE_SCALE_UP",
					}),
					// Service Connect test traffic rules checks
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.test_traffic_rules.0.header.0.name", "x-test-header-2"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.test_traffic_rules.0.header.0.value.0.exact", "test-value-2"),
				),
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.lifecycle_hook.#", "2"),
					testAccCheckServiceRemoveBlueGreenDeploymentConfigurations(ctx, t, &service),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.lifecycle_hook.#", "2"),
				),
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_sigintRollback(t *testing.T) {
	acctest.Skip(t, "SIGINT handling can't reliably be tested in CI")

	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttrPair(resourceName, "task_definition", "aws_ecs_task_definition.test", names.AttrARN),
				),
			},
			{
				Config: testAccServiceConfig_blueGreenDeployment_withHookBehavior(rName, false),
				PreConfig: func() {
					go func() {
						_ = exec.Command("go", "run", "test-fixtures/sigint_helper.go", "30").Start() // lintignore:XR007
					}()
				},
				ExpectError: regexache.MustCompile("execution halted|context canceled"),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttrPair(resourceName, "task_definition", "aws_ecs_task_definition.test", names.AttrARN),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_circuitBreakerRollback(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceConfig_blueGreenDeployment_withCircuitBreaker(rName),
				ExpectError: regexache.MustCompile(`No rollback candidate was found to run the rollback`),
			},
			{
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
				),
			},
			{
				Config:      testAccServiceConfig_blueGreenDeployment_withCircuitBreaker(rName),
				ExpectError: regexache.MustCompile(`Service deployment rolled back because the circuit breaker threshold was exceeded.`),
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_createFailure(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceConfig_blueGreenDeployment_withHookBehavior(rName, true),
				ExpectError: regexache.MustCompile(`No rollback candidate was found`),
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_changeStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_blueGreenDeployment_zeroBakeTime(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment_configuration.0.lifecycle_hook.*", map[string]string{
						"hook_details":       acctest.CtTrue,
						"lifecycle_stages.0": "POST_SCALE_UP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment_configuration.0.lifecycle_hook.*", map[string]string{
						"hook_details":       "\"Test string\"",
						"lifecycle_stages.0": "TEST_TRAFFIC_SHIFT",
					}),
				),
			},
			{
				Config: testAccServiceConfig_blueGreenDeployment_switchToRolling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "ROLLING"),
				),
			},
			{
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
				),
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_updateFailure(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.lifecycle_hook.#", "2"),
				),
			},
			{
				Config:      testAccServiceConfig_blueGreenDeployment_withHookBehavior(rName, true),
				ExpectError: regexache.MustCompile(`Service deployment rolled back`),
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_updateInPlace(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
				),
			},
			{
				Config: testAccServiceConfig_blueGreenDeployment_zeroBakeTime(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "2"),
				),
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_waitServiceActive(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
				),
			},
		},
	})
}

func TestAccECSService_BlueGreenDeployment_withoutTestListenerRule(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16] // Use shorter name to avoid target group name length issues
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_blueGreenDeployment_withoutTestListenerRule(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.0.advanced_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.alternate_target_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.production_listener_rule"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.0.advanced_configuration.0.test_listener_rule", ""),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.role_arn"),
				),
			},
			{
				// Set test_listener_rule
				Config: testAccServiceConfig_blueGreenDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.0.advanced_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.alternate_target_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.production_listener_rule"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.test_listener_rule"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.role_arn"),
				),
			},
			{
				// Remove test_listener_rule again
				Config: testAccServiceConfig_blueGreenDeployment_withoutTestListenerRule(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.0.advanced_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.alternate_target_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.production_listener_rule"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.0.advanced_configuration.0.test_listener_rule", ""),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer.0.advanced_configuration.0.role_arn"),
				),
			},
		},
	})
}

func TestAccECSService_DeploymentConfiguration_strategy(t *testing.T) {
	// Test for deployment configuration strategy
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentConfiguration_strategy(rName, "ROLLING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "ROLLING"),
				),
			},
			{
				Config: testAccServiceConfig_deploymentConfiguration_strategy(rName, "BLUE_GREEN"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "BLUE_GREEN"),
				),
			},
			{
				Config: testAccServiceConfig_deploymentConfiguration_strategy(rName, "ROLLING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "ROLLING"),
				),
			},
		},
	})
}

func TestAccECSService_DeploymentValues_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentValues(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
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
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentPercents(rName, 0, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_maximum_percent", "100"),
					resource.TestCheckResourceAttr(resourceName, "deployment_minimum_healthy_percent", "0"),
				),
			},
		},
	})
}

func TestAccECSService_deploymentCircuitBreaker(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_deploymentCircuitBreaker(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_circuit_breaker.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_circuit_breaker.0.enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "deployment_circuit_breaker.0.rollback", acctest.CtTrue),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3444
func TestAccECSService_loadBalancerChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_lbChanges(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccServiceConfig_lbChangesModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3361
func TestAccECSService_clusterName(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_clusterName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "cluster", rName),
				),
			},
		},
	})
}

func TestAccECSService_alb(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_alb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_multipleTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_multipleTargetGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", "2"),
				),
			},
		},
	})
}

func TestAccECSService_forceNewDeployment(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_basic(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccServiceConfig_forceNewDeployment(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccECSService_forceNewDeploymentTriggers(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_forceNewDeployment(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccServiceConfig_forceNewDeploymentTriggers(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "force_new_deployment", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "triggers.update"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccECSService_PlacementStrategy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_basic(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccServiceConfig_placementStrategy(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccServiceConfig_randomPlacementStrategy(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "random"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", ""),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccServiceConfig_multiplacementStrategy(rName, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.1.type", "spread"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.1.field", "instanceId"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13146
func TestAccECSService_PlacementStrategy_missing(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceConfig_placementStrategyType(rName, ""),
				ExpectError: regexache.MustCompile(`expected type to be one of`),
			},
		},
	})
}

func TestAccECSService_PlacementConstraints_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_placementConstraint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccServiceConfig_placementConstraintEmptyExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccECSService_PlacementConstraints_emptyExpression(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_placementConstraintEmptyExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_LaunchTypeFargate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_launchTypeFargate(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "LATEST"),
				),
			},
			{
				Config: testAccServiceConfig_launchTypeFargate(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", acctest.CtTrue),
				),
			},
			{
				Config: testAccServiceConfig_launchTypeFargate(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccECSService_LaunchTypeFargate_platformVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_launchTypeFargateAndPlatformVersion(rName, "1.3.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.3.0"),
				),
			},
			{
				Config: testAccServiceConfig_launchTypeFargateAndPlatformVersion(rName, "LATEST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "LATEST"),
				),
			},
			{
				Config: testAccServiceConfig_launchTypeFargateAndPlatformVersion(rName, "1.4.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.4.0"),
				),
			},
		},
	})
}

func TestAccECSService_LaunchTypeFargate_waitForSteadyState(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Wait for the ECS Cluster to reach a steady state w/specified count
				Config: testAccServiceConfig_launchTypeFargateAndWait(rName, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", acctest.CtTrue),
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
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_launchTypeFargateNoWait(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", acctest.CtFalse),
				),
			},
			{
				// Modify desired count and wait for the ECS Cluster to reach steady state
				Config: testAccServiceConfig_launchTypeFargateAndWait(rName, 2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", acctest.CtTrue),
				),
			},
			{
				// Modify desired count without wait
				Config: testAccServiceConfig_launchTypeFargateAndWait(rName, 1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccECSService_LaunchTypeEC2_network(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_networkConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
				),
			},
			{
				Config: testAccServiceConfig_networkConfigurationModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
				),
			},
		},
	})
}

func TestAccECSService_DaemonSchedulingStrategy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_daemonSchedulingStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "DAEMON"),
				),
			},
		},
	})
}

func TestAccECSService_DaemonSchedulingStrategy_setDeploymentMinimum(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_daemonSchedulingStrategySetDeploymentMinimum(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "DAEMON"),
				),
			},
		},
	})
}

func TestAccECSService_replicaSchedulingStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_replicaSchedulingStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "REPLICA"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceRegistries_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_registries(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceRegistries_container(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_registriesContainer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceRegistries_changes(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	serviceDiscoveryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	updatedServiceDiscoveryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_registriesChanges(rName, serviceDiscoveryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
			{
				Config: testAccServiceConfig_registriesChanges(rName, updatedServiceDiscoveryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceRegistries_removal(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	serviceDiscoveryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_registriesRemoval(rName, serviceDiscoveryName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
			{
				Config: testAccServiceConfig_registriesRemoval(rName, serviceDiscoveryName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceConnect_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportStateVerify: true,
				// Resource currently defaults to importing task_definition as family:revision
				// and wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"task_definition", "wait_for_steady_state"},
			},
		},
	})
}

func TestAccECSService_ServiceConnect_full(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectAllAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceConnect_tls_with_empty_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnect_tls_with_empty_timeout_block(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "1"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceConnect_ingressPortOverride(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectIngressPortOverride(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.log_configuration.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "service_connect_configuration.0.namespace"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.dns_name", "nginx-http."+rName),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.client_alias.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.discovery_name", "nginx-http"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.ingress_port_override", "0"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.service.0.port_name", "nginx-http"),
				),
			},
		},
	})
}

func TestAccECSService_ServiceConnect_remove(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccServiceConfig_serviceConnectRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.#", "0"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform-provider-aws/issues/42818
func TestAccECSService_ServiceConnect_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					testAccCheckServiceDisableServiceConnect(ctx, t, &service),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccServiceConfig_serviceConnectBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccECSService_ServiceConnect_accessLogConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_serviceConnectAccessLogConfiguration(rName, "TEXT", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.access_log_configuration.0.format", "TEXT"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.access_log_configuration.0.include_query_parameters", "ENABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s/%s", rName, rName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"task_definition", "wait_for_steady_state"},
			},
			{
				Config: testAccServiceConfig_serviceConnectAccessLogConfiguration(rName, "JSON", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.access_log_configuration.0.format", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "service_connect_configuration.0.access_log_configuration.0.include_query_parameters", "DISABLED"),
				),
			},
		},
	})
}

func TestAccECSService_Tags_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
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
				Config: testAccServiceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccServiceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccECSService_Tags_managed(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_managedTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "enable_ecs_managed_tags", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccECSService_Tags_UpgradeFromV5_100_0(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ECSServiceID),
		CheckDestroy: testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccServiceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				// Just only upgrading to the latest provider version
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccServiceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccServiceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccServiceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccECSService_Tags_UpgradeFromV5_100_0ThroughV6_08_0(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ECSServiceID),
		CheckDestroy: testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccServiceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.8.0",
					},
				},
				Config: testAccServiceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				// Just only upgrading to the latest provider version
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccServiceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccServiceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccServiceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccECSService_Tags_propagate(t *testing.T) {
	ctx := acctest.Context(t)
	var first, second, third awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_propagateTags(rName, "SERVICE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, string(awstypes.PropagateTagsService)),
				),
			},
			{
				Config: testAccServiceConfig_propagateTags(rName, "TASK_DEFINITION"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, string(awstypes.PropagateTagsTaskDefinition)),
				),
			},
			{
				Config: testAccServiceConfig_propagateTags(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &third),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, string(awstypes.PropagateTagsNone)),
				),
			},
		},
	})
}

func TestAccECSService_executeCommand(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_executeCommand(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "enable_execute_command", acctest.CtTrue),
				),
			},
			{
				Config: testAccServiceConfig_executeCommand(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "enable_execute_command", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccECSService_AvailabilityZoneRebalancing(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_availabilityZoneRebalancing(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_rebalancing"), tfknownvalue.StringExact(awstypes.AvailabilityZoneRebalancingEnabled)),
				},
			},
			{
				Config: testAccServiceConfig_availabilityZoneRebalancing(rName, awstypes.AvailabilityZoneRebalancingEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_rebalancing"), tfknownvalue.StringExact(awstypes.AvailabilityZoneRebalancingEnabled)),
				},
			},
			{
				Config: testAccServiceConfig_availabilityZoneRebalancing(rName, awstypes.AvailabilityZoneRebalancingDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_rebalancing"), tfknownvalue.StringExact(awstypes.AvailabilityZoneRebalancingDisabled)),
				},
			},
			{
				Config: testAccServiceConfig_availabilityZoneRebalancing(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_rebalancing"), tfknownvalue.StringExact(awstypes.AvailabilityZoneRebalancingDisabled)),
				},
			},
			{
				Config: testAccServiceConfig_availabilityZoneRebalancing(rName, awstypes.AvailabilityZoneRebalancingEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_rebalancing"), tfknownvalue.StringExact(awstypes.AvailabilityZoneRebalancingEnabled)),
				},
			},
		},
	})
}

// Linear and Canary Deployment Strategy Tests

func TestAccECSService_LinearDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_linearDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.linear_configuration.0.step_percent", "50"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.linear_configuration.0.step_bake_time_in_minutes", "1"),
				),
			},
			{
				Config: testAccServiceConfig_linearDeployment_updated(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.linear_configuration.0.step_percent", "25"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.linear_configuration.0.step_bake_time_in_minutes", "2"),
				),
			},
		},
	})
}

func TestAccECSService_LinearDeployment_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_linearDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
					testAccCheckServiceRemoveLinearDeploymentConfigurations(ctx, t, &service),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccServiceConfig_linearDeployment_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
				),
			},
		},
	})
}

func TestAccECSService_LinearDeployment_createFailure(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceConfig_linearDeployment_withFailure(rName, true),
				ExpectError: regexache.MustCompile(`No rollback candidate was found`),
			},
		},
	})
}

func TestAccECSService_LinearDeployment_changeStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_linearDeployment_zeroBakeTime(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
				),
			},
			{
				Config: testAccServiceConfig_linearDeployment_switchToRolling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "ROLLING"),
				),
			},
			{
				Config: testAccServiceConfig_linearDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
				),
			},
		},
	})
}

func TestAccECSService_LinearDeployment_updateFailure(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_linearDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
				),
			},
			{
				Config:      testAccServiceConfig_linearDeployment_withFailure(rName, true),
				ExpectError: regexache.MustCompile(`Service deployment rolled back`),
			},
		},
	})
}

func TestAccECSService_LinearDeployment_updateInPlace(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_linearDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
				),
			},
			{
				Config: testAccServiceConfig_linearDeployment_zeroBakeTime(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "2"),
				),
			},
		},
	})
}

func TestAccECSService_LinearDeployment_waitServiceActive(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_linearDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "LINEAR"),
				),
			},
		},
	})
}

func TestAccECSService_CanaryDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_canaryDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.canary_configuration.0.canary_percent", "20"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.canary_configuration.0.canary_bake_time_in_minutes", "1"),
				),
			},
			{
				Config: testAccServiceConfig_canaryDeployment_updated(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.canary_configuration.0.canary_percent", "10"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.canary_configuration.0.canary_bake_time_in_minutes", "2"),
				),
			},
		},
	})
}

func TestAccECSService_CanaryDeployment_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_canaryDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
					testAccCheckServiceRemoveCanaryDeploymentConfigurations(ctx, t, &service),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccServiceConfig_canaryDeployment_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
				),
			},
		},
	})
}

func TestAccECSService_CanaryDeployment_createFailure(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceConfig_canaryDeployment_withFailure(rName, true),
				ExpectError: regexache.MustCompile(`No rollback candidate was found`),
			},
		},
	})
}

func TestAccECSService_CanaryDeployment_changeStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_canaryDeployment_zeroBakeTime(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
				),
			},
			{
				Config: testAccServiceConfig_canaryDeployment_switchToRolling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "ROLLING"),
				),
			},
			{
				Config: testAccServiceConfig_canaryDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
				),
			},
		},
	})
}

func TestAccECSService_CanaryDeployment_updateFailure(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_canaryDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
				),
			},
			{
				Config:      testAccServiceConfig_canaryDeployment_withFailure(rName, true),
				ExpectError: regexache.MustCompile(`Service deployment rolled back`),
			},
		},
	})
}

func TestAccECSService_CanaryDeployment_updateInPlace(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_canaryDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
				),
			},
			{
				Config: testAccServiceConfig_canaryDeployment_zeroBakeTime(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "2"),
				),
			},
		},
	})
}

func TestAccECSService_CanaryDeployment_waitServiceActive(t *testing.T) {
	ctx := acctest.Context(t)
	var service awstypes.Service
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]
	resourceName := "aws_ecs_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_canaryDeployment_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.strategy", "CANARY"),
				),
			},
		},
	})
}

func testAccServiceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		cluster := rs.Primary.Attributes["cluster"]
		if arn.IsARN(cluster) {
			clusterArn, err := arn.Parse(cluster)
			if err != nil {
				return "", err
			}

			clusterParts := strings.Split(clusterArn.Resource, "/")
			if len(clusterParts) != 2 {
				return "", fmt.Errorf("wrong format of resource: %s, expecting 'cluster-name/service-name'", clusterArn)
			}

			cluster = clusterParts[1]
		}

		return fmt.Sprintf("%s/%s", cluster, rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccCheckServiceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_service" {
				continue
			}

			output, err := tfecs.FindServiceNoTagsByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["cluster"])

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			if aws.ToString(output.Status) == "INACTIVE" {
				return nil
			}

			return fmt.Errorf("ECS Service %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceExists(ctx context.Context, t *testing.T, name string, service *awstypes.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		var output *awstypes.Service
		err := tfresource.Retry(ctx, 1*time.Minute, func(ctx context.Context) *tfresource.RetryError {
			var err error
			output, err = tfecs.FindServiceNoTagsByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["cluster"])

			if retry.NotFound(err) {
				return tfresource.RetryableError(err)
			}

			if err != nil {
				return tfresource.NonRetryableError(err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		*service = *output

		return nil
	}
}

func testAccCheckServiceDisableServiceConnect(ctx context.Context, t *testing.T, service *awstypes.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		input := &ecs.UpdateServiceInput{
			Cluster: service.ClusterArn,
			Service: service.ServiceName,
			ServiceConnectConfiguration: &awstypes.ServiceConnectConfiguration{
				Enabled: false,
			},
		}

		_, err := conn.UpdateService(ctx, input)
		return err
	}
}

func testAccCheckServiceRemoveBlueGreenDeploymentConfigurations(ctx context.Context, t *testing.T, service *awstypes.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		input := &ecs.UpdateServiceInput{
			Cluster: service.ClusterArn,
			Service: service.ServiceName,
			DeploymentConfiguration: &awstypes.DeploymentConfiguration{
				Strategy:          awstypes.DeploymentStrategyRolling,
				BakeTimeInMinutes: aws.Int32(0),
				LifecycleHooks:    []awstypes.DeploymentLifecycleHook{},
			},
		}

		_, err := conn.UpdateService(ctx, input)
		return err
	}
}

func testAccCheckServiceRemoveLinearDeploymentConfigurations(ctx context.Context, t *testing.T, service *awstypes.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		input := &ecs.UpdateServiceInput{
			Cluster: service.ClusterArn,
			Service: service.ServiceArn,
			DeploymentConfiguration: &awstypes.DeploymentConfiguration{
				Strategy: awstypes.DeploymentStrategyRolling,
			},
		}

		_, err := conn.UpdateService(ctx, input)
		return err
	}
}

func testAccCheckServiceRemoveCanaryDeploymentConfigurations(ctx context.Context, t *testing.T, service *awstypes.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		input := &ecs.UpdateServiceInput{
			Cluster: service.ClusterArn,
			Service: service.ServiceArn,
			DeploymentConfiguration: &awstypes.DeploymentConfiguration{
				Strategy: awstypes.DeploymentStrategyRolling,
			},
		}

		_, err := conn.UpdateService(ctx, input)
		return err
	}
}

func testAccServiceConfig_basic(rName, clusterName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[2]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[2]q

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
}
`, rName, clusterName)
}

func testAccServiceConfig_regionOverride(rName, clusterName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  region = %[3]q

  name = %[2]q
}

resource "aws_ecs_task_definition" "test" {
  region = %[3]q

  family = %[2]q

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
  region = %[3]q

  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}
`, rName, clusterName, acctest.AlternateRegion())
}

func testAccServiceConfig_modified(rName, clusterName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[2]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[2]q

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
  desired_count   = 2
}
`, rName, clusterName)
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
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  ordered_placement_strategy {
    field = "host"
    type  = "spread"
  }
}
`, rName)
}

func testAccServiceConfig_blueGreenDeploymentBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_partition" "current" {}

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

resource "aws_security_group" "alb" {
  name        = "%[1]s-alb"
  description = "Security group for ALB"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lb" "main" {
  name               = %[1]q
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false

  tags = {
    Name = "%[1]s-alb"
  }
}

resource "aws_lb_target_group" "primary" {
  name        = "%[1]s-prim"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.test.id
  target_type = "ip"

  health_check {
    path                = "/"
    healthy_threshold   = 2
    unhealthy_threshold = 5
    timeout             = 5
    interval            = 30
  }
}

resource "aws_lb_target_group" "alternate" {
  name        = "%[1]s-alt"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.test.id
  target_type = "ip"

  health_check {
    path                = "/"
    healthy_threshold   = 2
    unhealthy_threshold = 5
    timeout             = 5
    interval            = 30
  }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "404: Page not found"
      status_code  = 404
    }
  }
}

resource "aws_lb_listener_rule" "production" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 1

  action {
    type = "forward"
    forward {
      dynamic "target_group" {
        for_each = {
          primary   = aws_lb_target_group.primary.arn
          alternate = aws_lb_target_group.alternate.arn
        }
        content {
          arn    = target_group.value
          weight = target_group.key == "primary" ? 100 : 0
        }
      }
    }
  }

  condition {
    path_pattern {
      values = ["/*"]
    }
  }

  lifecycle {
    ignore_changes = [
      action[0].forward[0].target_group
    ]
  }
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 2 # Make sure this is different from the production rule priority

  action {
    type = "forward"
    forward {
      target_group {
        arn    = aws_lb_target_group.alternate.arn
        weight = 100
      }
    }
  }

  condition {
    path_pattern {
      values = ["/*"]
    }
  }

  lifecycle {
    ignore_changes = [
      action[0].forward[0].target_group
    ]
  }
}

resource "aws_iam_role" "global" {
  name = "%[1]s-global"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = [
            "lambda.amazonaws.com",
            "ecs-tasks.amazonaws.com",
            "elasticloadbalancing.amazonaws.com",
            "ecs.amazonaws.com",
          ]
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "global_admin_attach" {
  role       = aws_iam_role.global.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AdministratorAccess"
}

resource "aws_iam_role_policy" "ecs_elb_permissions" {
  name = "${aws_iam_role.global.name}-elb-permissions"
  role = aws_iam_role.global.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "elasticloadbalancing:DeregisterTargets",
          "elasticloadbalancing:RegisterTargets"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_service_role" {
  role       = aws_iam_role.global.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceRole"
}

resource "aws_iam_role" "lambda_role" {
  name = "%[1]s-hook-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.lambda_role.name
}

resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster" "main" {
  name = %[1]q

  service_connect_defaults {
    namespace = aws_service_discovery_http_namespace.test.arn
  }
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  lifecycle {
    create_before_destroy = true
  }

  container_definitions = jsonencode([
    {
      name      = "test"
      image     = "nginx:latest"
      cpu       = 256
      memory    = 512
      essential = true
      environment = [
        {
          name  = "test_name"
          value = "test_val"
        }
      ]
      portMappings = [
        {
          containerPort = 80
          hostPort      = 80
          protocol      = "tcp"
          name          = "http"
          appProtocol   = "http"
        }
      ]
    }
  ])
}

resource "aws_lambda_function" "hook_success" {
  filename         = "test-fixtures/success_lambda_func.zip"
  function_name    = "%[1]s-hook-success"
  role             = aws_iam_role.lambda_role.arn
  handler          = "index.handler"
  runtime          = "nodejs20.x"
  source_code_hash = filebase64sha256("test-fixtures/success_lambda_func.zip")
}

resource "aws_lambda_function" "hook_failure" {
  filename         = "test-fixtures/failure_lambda_func.zip"
  function_name    = "%[1]s-hook-failure"
  role             = aws_iam_role.lambda_role.arn
  handler          = "index.handler"
  runtime          = "nodejs20.x"
  source_code_hash = filebase64sha256("test-fixtures/failure_lambda_func.zip")
}
`, rName))
}

func testAccServiceConfig_blueGreenDeployment_basic(rName string, waitSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "BLUE_GREEN"
    bake_time_in_minutes = 2

    lifecycle_hook {
      hook_target_arn  = aws_lambda_function.hook_success.arn
      role_arn         = aws_iam_role.global.arn
      lifecycle_stages = ["POST_SCALE_UP", "POST_TEST_TRAFFIC_SHIFT"]
      hook_details     = jsonencode([1, "2", true])
    }

    lifecycle_hook {
      hook_target_arn  = aws_lambda_function.hook_success.arn
      role_arn         = aws_iam_role.global.arn
      lifecycle_stages = ["TEST_TRAFFIC_SHIFT", "POST_PRODUCTION_TRAFFIC_SHIFT"]
      hook_details     = "3.14"
    }
  }

  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_http_namespace.test.arn

    service {
      client_alias {
        dns_name = "test-service.local"
        port     = 8080

        test_traffic_rules {
          header {
            name = "x-test-header"
            value {
              exact = "test-value"
            }
          }
        }
      }
      discovery_name = "test-service"
      port_name      = "http"
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.primary.arn
    container_name   = "test"
    container_port   = 80

    advanced_configuration {
      alternate_target_group_arn = aws_lb_target_group.alternate.arn
      production_listener_rule   = aws_lb_listener_rule.production.arn
      test_listener_rule         = aws_lb_listener_rule.test.arn
      role_arn                   = aws_iam_role.global.arn
    }
  }

  wait_for_steady_state = %[2]t

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, waitSteadyState))
}

func testAccServiceConfig_blueGreenDeployment_withCircuitBreaker(rName string) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_task_definition" "should_fail" {
  family                   = "%[1]s-should-fail"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  lifecycle {
    create_before_destroy = true
  }

  container_definitions = jsonencode([
    {
      name      = "should_fail"
      image     = "nginx:latest"
      cpu       = 256
      memory    = 512
      essential = true,
      command = [
        "/bin/sh -c \"while true; do /bin/date > /var/www/my-vol/date; sleep 1; done\""
      ]
      environment = [
        {
          name  = "test_name"
          value = "test_val"
        }
      ]
      portMappings = [
        {
          containerPort = 80
          hostPort      = 80
          protocol      = "tcp"
          name          = "http"
          appProtocol   = "http"
        }
      ]
    }
  ])
}


resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.should_fail.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "BLUE_GREEN"
    bake_time_in_minutes = 1

    lifecycle_hook {
      hook_target_arn  = aws_lambda_function.hook_success.arn
      role_arn         = aws_iam_role.global.arn
      lifecycle_stages = ["PRE_SCALE_UP"]
    }
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_http_namespace.test.arn

    service {
      client_alias {
        dns_name = "test-service.local"
        port     = 8080

        test_traffic_rules {
          header {
            name = "x-test-header-2"
            value {
              exact = "test-value-2"
            }
          }
        }
      }
      discovery_name = "test-service"
      port_name      = "http"
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.primary.arn
    container_name   = "should_fail"
    container_port   = 80

    advanced_configuration {
      alternate_target_group_arn = aws_lb_target_group.alternate.arn
      production_listener_rule   = aws_lb_listener_rule.production.arn
      test_listener_rule         = aws_lb_listener_rule.test.arn
      role_arn                   = aws_iam_role.global.arn
    }
  }

  wait_for_steady_state = true

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName))
}

func testAccServiceConfig_blueGreenDeployment_withHookBehavior(rName string, shouldFail bool) string {
	var hookTargetArn string
	if shouldFail {
		hookTargetArn = "aws_lambda_function.hook_failure.arn"
	} else {
		hookTargetArn = "aws_lambda_function.hook_success.arn"
	}

	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`

resource "aws_ecs_task_definition" "test2" {
  family                   = "%[1]s-test2"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  lifecycle {
    create_before_destroy = true
  }

  container_definitions = jsonencode([
    {
      name      = "test"
      image     = "nginx:latest"
      cpu       = 256
      memory    = 512
      essential = true
      environment = [
        {
          name  = "test_name_2"
          value = "test_val_2"
        }
      ]
      portMappings = [
        {
          containerPort = 80
          hostPort      = 80
          protocol      = "tcp"
          name          = "http"
          appProtocol   = "http"
        }
      ]
    }
  ])
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test2.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "BLUE_GREEN"
    bake_time_in_minutes = 3

    lifecycle_hook {
      hook_target_arn  = %[2]s
      role_arn         = aws_iam_role.global.arn
      lifecycle_stages = ["PRE_SCALE_UP"]
      hook_details = jsonencode({ "bool_key" : true, "string_key" : "string_val", "int_key" : 10, "list_key" : [1, "2", true], "object_key" : {
        "bool_key" : true,
        "string_key" : "string_val",
        "int_key" : 10,
        "list_key" : [1, "2", true]
      } })
    }
  }

  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_http_namespace.test.arn

    service {
      client_alias {
        dns_name = "test-service.local"
        port     = 8080

        test_traffic_rules {
          header {
            name = "x-test-header-2"
            value {
              exact = "test-value-2"
            }
          }
        }
      }
      discovery_name = "test-service"
      port_name      = "http"
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.primary.arn
    container_name   = "test"
    container_port   = 80

    advanced_configuration {
      alternate_target_group_arn = aws_lb_target_group.alternate.arn
      production_listener_rule   = aws_lb_listener_rule.production.arn
      test_listener_rule         = aws_lb_listener_rule.test.arn
      role_arn                   = aws_iam_role.global.arn
    }
  }

  sigint_rollback       = true
  wait_for_steady_state = true

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, hookTargetArn))
}

func testAccServiceConfig_blueGreenDeployment_switchToRolling(rName string) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy = "ROLLING"

    lifecycle_hook {
      hook_target_arn  = aws_lambda_function.hook_success.arn
      role_arn         = aws_iam_role.global.arn
      lifecycle_stages = ["PRE_SCALE_UP"]
    }
  }

  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_http_namespace.test.arn

    service {
      client_alias {
        dns_name = "test-service.local"
        port     = 8080

        test_traffic_rules {
          header {
            name = "x-test-header"
            value {
              exact = "test-value"
            }
          }
        }
      }
      discovery_name = "test-service"
      port_name      = "http"
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.alternate.arn
    container_name   = "test"
    container_port   = 80
  }

  wait_for_steady_state = true

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName))
}

func testAccServiceConfig_blueGreenDeployment_zeroBakeTime(rName string, waitSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "BLUE_GREEN"
    bake_time_in_minutes = 0

    lifecycle_hook {
      hook_target_arn  = aws_lambda_function.hook_success.arn
      role_arn         = aws_iam_role.global.arn
      lifecycle_stages = ["POST_SCALE_UP", "POST_TEST_TRAFFIC_SHIFT"]
      hook_details     = "true"
    }

    lifecycle_hook {
      hook_target_arn  = aws_lambda_function.hook_success.arn
      role_arn         = aws_iam_role.global.arn
      lifecycle_stages = ["TEST_TRAFFIC_SHIFT", "POST_PRODUCTION_TRAFFIC_SHIFT"]
      hook_details     = jsonencode("Test string")
    }
  }

  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_http_namespace.test.arn

    service {
      client_alias {
        dns_name = "test-service.local"
        port     = 8080

        test_traffic_rules {
          header {
            name = "x-test-header"
            value {
              exact = "test-value"
            }
          }
        }
      }
      discovery_name = "test-service"
      port_name      = "http"
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.primary.arn
    container_name   = "test"
    container_port   = 80

    advanced_configuration {
      alternate_target_group_arn = aws_lb_target_group.alternate.arn
      production_listener_rule   = aws_lb_listener_rule.production.arn
      test_listener_rule         = aws_lb_listener_rule.test.arn
      role_arn                   = aws_iam_role.global.arn
    }
  }

  wait_for_steady_state = %[2]t

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, waitSteadyState))
}

func testAccServiceConfig_blueGreenDeployment_withoutTestListenerRule(rName string, waitSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "BLUE_GREEN"
    bake_time_in_minutes = 2

    lifecycle_hook {
      hook_target_arn  = aws_lambda_function.hook_success.arn
      role_arn         = aws_iam_role.global.arn
      lifecycle_stages = ["POST_SCALE_UP", "POST_TEST_TRAFFIC_SHIFT"]
    }

    lifecycle_hook {
      hook_target_arn  = aws_lambda_function.hook_success.arn
      role_arn         = aws_iam_role.global.arn
      lifecycle_stages = ["TEST_TRAFFIC_SHIFT", "POST_PRODUCTION_TRAFFIC_SHIFT"]
    }
  }

  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_http_namespace.test.arn

    service {
      client_alias {
        dns_name = "test-service.local"
        port     = 8080

        test_traffic_rules {
          header {
            name = "x-test-header"
            value {
              exact = "test-value"
            }
          }
        }
      }
      discovery_name = "test-service"
      port_name      = "http"
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.primary.arn
    container_name   = "test"
    container_port   = 80

    advanced_configuration {
      alternate_target_group_arn = aws_lb_target_group.alternate.arn
      production_listener_rule   = aws_lb_listener_rule.production.arn
      role_arn                   = aws_iam_role.global.arn
    }
  }

  wait_for_steady_state = %[2]t

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, waitSteadyState))
}

func testAccServiceConfig_linearDeployment_basic(rName string, waitSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "LINEAR"
    bake_time_in_minutes = 2

    linear_configuration {
      step_percent              = 50
      step_bake_time_in_minutes = 1
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  wait_for_steady_state = %[2]t

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, waitSteadyState))
}

func testAccServiceConfig_linearDeployment_updated(rName string, waitSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "LINEAR"
    bake_time_in_minutes = 3

    linear_configuration {
      step_percent              = 25
      step_bake_time_in_minutes = 2
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  wait_for_steady_state = %[2]t

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, waitSteadyState))
}

func testAccServiceConfig_linearDeployment_zeroBakeTime(rName string, waitSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "LINEAR"
    bake_time_in_minutes = 0

    linear_configuration {
      step_percent              = 100
      step_bake_time_in_minutes = 0
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  wait_for_steady_state = %[2]t

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, waitSteadyState))
}

func testAccServiceConfig_linearDeployment_switchToRolling(rName string) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy = "ROLLING"
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName))
}

func testAccServiceConfig_linearDeployment_withFailure(rName string, shouldFail bool) string {
	var taskDef string
	if shouldFail {
		taskDef = "aws_ecs_task_definition.should_fail.arn"
	} else {
		taskDef = "aws_ecs_task_definition.test.arn"
	}

	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_task_definition" "should_fail" {
  family                   = "%[1]s-should-fail"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.global.arn

  container_definitions = jsonencode([
    {
      name      = "test"
      image     = "nginx:invalid-tag"
      essential = true
      portMappings = [
        {
          containerPort = 80
          protocol      = "tcp"
        }
      ]
    }
  ])
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = %[2]s
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "LINEAR"
    bake_time_in_minutes = 1

    linear_configuration {
      step_percent              = 50
      step_bake_time_in_minutes = 1
    }
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.primary.arn
    container_name   = "test"
    container_port   = 80

    advanced_configuration {
      alternate_target_group_arn = aws_lb_target_group.alternate.arn
      production_listener_rule   = aws_lb_listener_rule.production.arn
      test_listener_rule         = aws_lb_listener_rule.test.arn
      role_arn                   = aws_iam_role.global.arn
    }
  }

  wait_for_steady_state = true

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, taskDef))
}

func testAccServiceConfig_canaryDeployment_basic(rName string, waitSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "CANARY"
    bake_time_in_minutes = 2

    canary_configuration {
      canary_percent              = 20
      canary_bake_time_in_minutes = 1
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  wait_for_steady_state = %[2]t

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, waitSteadyState))
}

func testAccServiceConfig_canaryDeployment_updated(rName string, waitSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "CANARY"
    bake_time_in_minutes = 3

    canary_configuration {
      canary_percent              = 10
      canary_bake_time_in_minutes = 2
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  wait_for_steady_state = %[2]t

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, waitSteadyState))
}

func testAccServiceConfig_canaryDeployment_zeroBakeTime(rName string, waitSteadyState bool) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "CANARY"
    bake_time_in_minutes = 0

    canary_configuration {
      canary_percent              = 100
      canary_bake_time_in_minutes = 0
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  wait_for_steady_state = %[2]t

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, waitSteadyState))
}

func testAccServiceConfig_canaryDeployment_switchToRolling(rName string) string {
	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy = "ROLLING"
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName))
}

func testAccServiceConfig_canaryDeployment_withFailure(rName string, shouldFail bool) string {
	var taskDef string
	if shouldFail {
		taskDef = "aws_ecs_task_definition.should_fail.arn"
	} else {
		taskDef = "aws_ecs_task_definition.test.arn"
	}

	return acctest.ConfigCompose(testAccServiceConfig_blueGreenDeploymentBase(rName), fmt.Sprintf(`
resource "aws_ecs_task_definition" "should_fail" {
  family                   = "%[1]s-should-fail"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.global.arn

  container_definitions = jsonencode([
    {
      name      = "test"
      image     = "nginx:invalid-tag"
      essential = true
      portMappings = [
        {
          containerPort = 80
          protocol      = "tcp"
        }
      ]
    }
  ])
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.main.id
  task_definition = %[2]s
  desired_count   = 1
  launch_type     = "FARGATE"

  deployment_configuration {
    strategy             = "CANARY"
    bake_time_in_minutes = 1

    canary_configuration {
      canary_percent              = 20
      canary_bake_time_in_minutes = 1
    }
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.primary.arn
    container_name   = "test"
    container_port   = 80

    advanced_configuration {
      alternate_target_group_arn = aws_lb_target_group.alternate.arn
      production_listener_rule   = aws_lb_listener_rule.production.arn
      test_listener_rule         = aws_lb_listener_rule.test.arn
      role_arn                   = aws_iam_role.global.arn
    }
  }

  wait_for_steady_state = true

  depends_on = [
    aws_iam_role_policy_attachment.global_admin_attach,
    aws_iam_role_policy.ecs_elb_permissions,
    aws_iam_role_policy_attachment.ecs_service_role
  ]
}
`, rName, taskDef))
}

func testAccServiceConfig_deploymentConfiguration_strategy(rName string, strategy string) string {
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
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  deployment_configuration {
    strategy = %[2]q
  }
}
`, rName, strategy)
}

func testAccServiceConfig_capacityProviderStrategy(rName string, weight, base int, forceNewDeployment bool) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
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

func testAccServiceConfig_baseVolumeConfiguration(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

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
    "name": "mongodb",
    "mountPoints": [
      {"sourceVolume": "vol1", "containerPath": "/vol1"}
    ]
  }
]
DEFINITION

  volume {
    name                = "vol1"
    configure_at_launch = true
  }
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
`, rName))
}

func testAccServiceConfig_volumeConfiguration_basic(rName string) string {
	return acctest.ConfigCompose(testAccServiceConfig_baseVolumeConfiguration(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  volume_configuration {
    name = "vol1"
    managed_ebs_volume {
      role_arn   = aws_iam_role.ecs_service.arn
      size_in_gb = "8"
    }
  }

  network_configuration {
    subnets = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName))
}

func testAccServiceConfig_volumeConfiguration_volumeInitializationRate(rName string, volumeInitializationRate int) string {
	return acctest.ConfigCompose(testAccServiceConfig_baseVolumeConfiguration(rName), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  volume_configuration {
    name = "vol1"
    managed_ebs_volume {
      role_arn                   = aws_iam_role.ecs_service.arn
      snapshot_id                = aws_ebs_snapshot.test.id
      volume_initialization_rate = %[2]d
    }
  }

  network_configuration {
    subnets = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName, volumeInitializationRate))
}

func testAccServiceConfig_volumeConfiguration_tagSpecifications(rName string) string {
	return acctest.ConfigCompose(testAccServiceConfig_baseVolumeConfiguration(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  volume_configuration {
    name = "vol1"
    managed_ebs_volume {
      role_arn   = aws_iam_role.ecs_service.arn
      size_in_gb = "8"
      tag_specifications {
        resource_type  = "volume"
        propagate_tags = "SERVICE"
        tags = {
          Name = %[1]q
        }
      }
    }
  }

  network_configuration {
    subnets = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName))
}

func testAccServiceConfig_volumeConfiguration_update(rName, volumeType string, size int) string {
	return acctest.ConfigCompose(testAccServiceConfig_baseVolumeConfiguration(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  volume_configuration {
    name = "vol1"
    managed_ebs_volume {
      role_arn    = aws_iam_role.ecs_service.arn
      size_in_gb  = %[3]d
      volume_type = %[2]q
    }
  }

  network_configuration {
    subnets = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName, volumeType, size))
}

func testAccServiceConfig_volumeConfiguration_gp3(rName string) string {
	return acctest.ConfigCompose(testAccServiceConfig_baseVolumeConfiguration(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  volume_configuration {
    name = "vol1"
    managed_ebs_volume {
      role_arn    = aws_iam_role.ecs_service.arn
      size_in_gb  = 10
      volume_type = "gp3"
    }
  }

  network_configuration {
    subnets = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName))
}

func testAccServiceConfig_forceNewDeployment(rName, clusterName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[2]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[2]q

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
  cluster              = aws_ecs_cluster.test.id
  desired_count        = 1
  force_new_deployment = true
  name                 = %[1]q
  task_definition      = aws_ecs_task_definition.test.arn

  availability_zone_rebalancing = "DISABLED"

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }
}
`, rName, clusterName)
}

func testAccServiceConfig_forceNewDeploymentTriggers(rName, clusterName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[2]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[2]q

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
  cluster              = aws_ecs_cluster.test.id
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
`, rName, clusterName)
}

func testAccServiceConfig_placementStrategy(rName, clusterName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[2]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[2]q

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

  availability_zone_rebalancing = "DISABLED"

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }
}
`, rName, clusterName)
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
    type = %[2]q
  }
}
`, rName, placementStrategyType)
}

func testAccServiceConfig_randomPlacementStrategy(rName, clusterName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[2]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[2]q

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

  availability_zone_rebalancing = "DISABLED"

  ordered_placement_strategy {
    type = "random"
  }
}
`, rName, clusterName)
}

func testAccServiceConfig_multiplacementStrategy(rName, clusterName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[2]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[2]q

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

  availability_zone_rebalancing = "DISABLED"

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }

  ordered_placement_strategy {
    field = "host"
    type  = "spread"
  }
}
`, rName, clusterName)
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
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
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
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  placement_constraints {
    type = "distinctInstance"
  }
}
`, rName)
}

func testAccServiceConfig_healthCheckGracePeriodSeconds(rName string, healthCheckGracePeriodSeconds int) string {
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
`, rName, healthCheckGracePeriodSeconds))
}

func testAccServiceConfig_iamRole(rName string) string {
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
`, rName))
}

func testAccServiceConfig_alarms(rName string, enable bool) string {
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
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  alarms {
    enable   = %[2]t
    rollback = %[2]t
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
`, rName, enable)
}

func testAccServiceConfig_noAlarms(rName string) string {
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
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
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
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
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
    "image": "jenkins:latest",
    "memory": 128,
    "name": "jenkins"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  desired_count   = 1
}
`, rName)
}

func testAccServiceConfig_familyAndRevisionModified(rName string) string {
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
    "image": "jenkins:latest",
    "memory": 128,
    "name": "jenkins"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  desired_count   = 1
}
`, rName)
}

func testAccServiceConfig_clusterName(rName string) string {
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
    "image": "jenkins:latest",
    "memory": 128,
    "name": "jenkins"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.name
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

func testAccServiceConfig_registriesRemoval(rName, discoveryName string, removed bool) string {
	if removed {
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

  network_configuration {
    security_groups = [aws_security_group.test.id]
    subnets         = aws_subnet.test[*].id
  }
}
`, rName, discoveryName))
	}
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
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                = %[1]q
  cluster             = aws_ecs_cluster.test.id
  task_definition     = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  scheduling_strategy = "DAEMON"
}
`, rName)
}

func testAccServiceConfig_daemonSchedulingStrategySetDeploymentMinimum(rName string) string {
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
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                               = %[1]q
  cluster                            = aws_ecs_cluster.test.id
  task_definition                    = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  scheduling_strategy                = "DAEMON"
  deployment_minimum_healthy_percent = "50"
}
`, rName)
}

func testAccServiceConfig_deploymentControllerType(rName string, deploymentController string) string {
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
  desired_count   = 1
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  deployment_controller {
    type = %[2]q
  }

  load_balancer {
    container_name   = "test"
    container_port   = "80"
    target_group_arn = aws_lb_target_group.test.id
  }
}
`, rName, deploymentController))
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
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                = %[1]q
  cluster             = aws_ecs_cluster.test.id
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

func testAccServiceConfig_baseVPCLattice(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block      = "10.10.0.0/16"
  ipv6_cidr_block = aws_vpc.test.ipv6_cidr_block

  vpc_id                          = aws_vpc.test.id
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch         = true

  tags = {
    Name = %[1]q
  }
}

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

  route {
    ipv6_cidr_block = "::/0"
    gateway_id      = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}


data "aws_ec2_managed_prefix_list" "test" {
  filter {
    name   = "prefix-list-name"
    values = ["com.amazonaws.${data.aws_region.current.region}.vpc-lattice"]
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["10.0.0.0/16"]
    ipv6_cidr_blocks = [aws_vpc.test.ipv6_cidr_block]
  }

  ingress {
    prefix_list_ids = [data.aws_ec2_managed_prefix_list.test.id]
    from_port       = 0
    to_port         = 0
    protocol        = -1

  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name = %[1]q
  }

  # ECS Infrastructure role must exist long enough for the VPC Lattice targets to be deregistered.
  # Security group removal is a good indicator of this.
  depends_on = [aws_iam_role.vpc_lattice_infrastructure]
}


resource "aws_vpclattice_service" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "IP"

  config {
    port             = 80
    protocol         = "HTTP"
    protocol_version = "HTTP1"
    vpc_identifier   = aws_vpc.test.id
    health_check {
      enabled = false
    }
  }

}

resource "aws_vpclattice_target_group" "test_ipv6" {
  name = "%[1]s-ipv6"
  type = "IP"

  config {
    port             = 80
    protocol         = "HTTP"
    protocol_version = "HTTP1"
    vpc_identifier   = aws_vpc.test.id
    ip_address_type  = "IPV6"
    health_check {
      enabled = false
    }
  }
}


resource "aws_vpclattice_listener" "test" {
  name               = "%[1]s-listener"
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id

  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 80
      }
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test_ipv6.id
        weight                  = 20
      }
    }
  }
}

resource "aws_vpclattice_service_network_service_association" "test" {
  service_network_identifier = aws_vpclattice_service_network.test.arn
  service_identifier         = aws_vpclattice_service.test.arn
}

resource "aws_vpclattice_service_network_vpc_association" "test" {
  service_network_identifier = aws_vpclattice_service_network.test.arn
  vpc_identifier             = aws_vpc.test.id
}

resource "aws_iam_role" "vpc_lattice_infrastructure" {
  name                = "%[1]s-vpcl-infrastructure"
  managed_policy_arns = ["arn:${data.aws_partition.current.partition}:iam::aws:policy/VPCLatticeFullAccess"]
  assume_role_policy  = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAccessToECSForInfrastructureManagement",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role" "ecs_task_role" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ecs-tasks.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "ArnLike": {
          "aws:SourceArn": "arn:${data.aws_partition.current.partition}:ecs:*:${data.aws_caller_identity.current.account_id}:*"
        },
        "StringEquals": {
          "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "ecs_task_policy" {
  name = %[1]q
  role = aws_iam_role.ecs_task_role.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject"
      ],
      "Resource": "*",
      "Condition": {
        "ArnLike": {
          "aws:SourceArn": "arn:${data.aws_partition.current.partition}:ecs:*:${data.aws_caller_identity.current.account_id}:*"
        },
        "StringEquals": {
          "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_iam_role" "task_execution_role" {
  name = "%[1]s-task_execution_role"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ecs-tasks.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "ArnLike": {
          "aws:SourceArn": "arn:${data.aws_partition.current.partition}:ecs:*:${data.aws_caller_identity.current.account_id}:*"
        },
        "StringEquals": {
          "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.task_execution_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  execution_role_arn       = aws_iam_role.task_execution_role.arn

  container_definitions = <<DEFINITION
[
  {
    "essential": true,
    "image": "nginx:latest",
    "name": "%[1]s-container",
    "portMappings": [
      {
        "containerPort": 80,
        "name": "testvpclattice"
      }
    ]
  }
]
DEFINITION
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
`, rName)
}

func testAccServiceConfig_vpcLatticeConfiguration_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		testAccServiceConfig_baseVPCLattice(rName),
		fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name                   = %[1]q
  cluster                = aws_ecs_cluster.test.name
  task_definition        = aws_ecs_task_definition.test.arn
  desired_count          = 1
  launch_type            = "FARGATE"
  enable_execute_command = true
  network_configuration {
    subnets          = [aws_subnet.test.id]
    security_groups  = [aws_security_group.test.id]
    assign_public_ip = true
  }
  vpc_lattice_configurations {
    role_arn         = aws_iam_role.vpc_lattice_infrastructure.arn
    target_group_arn = aws_vpclattice_target_group.test.arn
    port_name        = "testvpclattice"
  }

  vpc_lattice_configurations {
    role_arn         = aws_iam_role.vpc_lattice_infrastructure.arn
    target_group_arn = aws_vpclattice_target_group.test_ipv6.arn
    port_name        = "testvpclattice"
  }
}
`, rName))
}

func testAccServiceConfig_vpcLatticeConfiguration_removed(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		testAccServiceConfig_baseVPCLattice(rName),
		fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name                   = %[1]q
  cluster                = aws_ecs_cluster.test.name
  task_definition        = aws_ecs_task_definition.test.arn
  desired_count          = 1
  launch_type            = "FARGATE"
  enable_execute_command = true
  wait_for_steady_state  = true
  network_configuration {
    subnets          = [aws_subnet.test.id]
    security_groups  = [aws_security_group.test.id]
    assign_public_ip = true
  }
}
`, rName))
}

func testAccServiceConfig_serviceConnectAllAttributes(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
  policy                  = data.aws_iam_policy_document.test.json
}


data "aws_iam_policy_document" "test" {
  policy_id = "KMSPolicy"

  statement {
    sid    = "Root User Permissions"
    effect = "Allow"
    principals {
      type = "AWS"
      identifiers = [
      "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    actions = [
    "kms:*"]
    resources = ["*"]
  }

  statement {
    sid    = "EC2 kms permissions"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.test.arn]
    }
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
    "kms:GenerateDataKeyPair"]
    resources = ["*"]
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy  = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ecs.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
  managed_policy_arns = ["arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSInfrastructureRolePolicyForServiceConnectTransportLayerSecurity"]
}

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
      "appProtocol": "http",
      "containerPort": 27017,
      "name": "tf-test",
      "protocol": "tcp"
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
      tls {
        issuer_cert_authority {
          aws_pca_authority_arn = aws_acmpca_certificate_authority.test.arn
        }
        kms_key  = aws_kms_key.test.arn
        role_arn = aws_iam_role.test.arn
      }
      timeout {
        idle_timeout_seconds        = 120
        per_request_timeout_seconds = 60
      }
    }
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
  tags = {
    AmazonECSManaged = "true"
  }
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}
`, rName)
}

func testAccServiceConfig_serviceConnect_tls_with_empty_timeout_block(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
  policy                  = data.aws_iam_policy_document.test.json
}


data "aws_iam_policy_document" "test" {
  policy_id = "KMSPolicy"

  statement {
    sid    = "Root User Permissions"
    effect = "Allow"
    principals {
      type = "AWS"
      identifiers = [
      "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    actions = [
    "kms:*"]
    resources = ["*"]
  }

  statement {
    sid    = "EC2 kms permissions"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.test.arn]
    }
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
    "kms:GenerateDataKeyPair"]
    resources = ["*"]
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy  = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ecs.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
  managed_policy_arns = ["arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSInfrastructureRolePolicyForServiceConnectTransportLayerSecurity"]
}

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
      "appProtocol": "http",
      "containerPort": 27017,
      "name": "tf-test",
      "protocol": "tcp"
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
      tls {
        issuer_cert_authority {
          aws_pca_authority_arn = aws_acmpca_certificate_authority.test.arn
        }
        kms_key  = aws_kms_key.test.arn
        role_arn = aws_iam_role.test.arn
      }
      timeout {
      }
    }
  }

  lifecycle {
    ignore_changes = [service_connect_configuration[0].service[0].timeout]
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
  tags = {
    AmazonECSManaged = "true"
  }
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}
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

func testAccServiceConfig_availabilityZoneRebalancing(rName string, azRebalancing awstypes.AvailabilityZoneRebalancing) string {
	val := string(azRebalancing)
	if val != "null" {
		val = strconv.Quote(val)
	}

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
      "image": "ghost:latest",
      "memory": 128,
      "name": "ghost"
    }
  ]
  DEFINITION
}

resource "aws_ecs_service" "test" {
  name                          = %[1]q
  cluster                       = aws_ecs_cluster.test.id
  task_definition               = aws_ecs_task_definition.test.arn
  availability_zone_rebalancing = %[2]s
}
  `, rName, val)
}

func testAccServiceConfig_serviceConnectAccessLogConfiguration(rName, format, includeQueryParams string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = "/ecs/%[1]s"
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = jsonencode([
    {
      name      = "test"
      image     = "public.ecr.aws/docker/library/nginx:latest"
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 80
          hostPort      = 80
          protocol      = "tcp"
          name          = "http"
          appProtocol   = "http"
        }
      ]
    }
  ])
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.test[*].id
    security_groups  = [aws_security_group.test.id]
    assign_public_ip = true
  }

  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_http_namespace.test.arn

    log_configuration {
      log_driver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.test.name
        "awslogs-region"        = data.aws_region.current.name
        "awslogs-stream-prefix" = "sc"
      }
    }

    access_log_configuration {
      format                   = %[2]q
      include_query_parameters = %[3]q
    }

    service {
      port_name      = "http"
      discovery_name = "test"

      client_alias {
        dns_name = "test"
        port     = 8080
      }
    }
  }
}

data "aws_region" "current" {}
`, rName, format, includeQueryParams))
}
