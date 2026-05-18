// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSDaemonTaskDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, rName),
					resource.TestCheckResourceAttr(resourceName, "cpu", "512"),
					resource.TestCheckResourceAttr(resourceName, "memory", "1024"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ecs", regexache.MustCompile(`daemon-task-definition/`+rName+`:\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.name", "nginx"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.image", "nginx:latest"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.cpu", "256"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.memory", "512"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.essential", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecs.ResourceDaemonTaskDefinition, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccDaemonTaskDefinitionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDaemonTaskDefinitionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_executionRole(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_executionRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.execution", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_taskRole(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_taskRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "task_role_arn", "aws_iam_role.task", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_volume(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_volume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccDaemonTaskDefinitionConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "0"),
				),
			},
			{
				Config: testAccDaemonTaskDefinitionConfig_volume(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "2"),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_multipleContainers(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_multipleContainers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.#", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccDaemonTaskDefinitionConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.#", "1"),
				),
			},
			{
				Config: testAccDaemonTaskDefinitionConfig_multipleContainers(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.#", "2"),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_containerDefinitionsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
				),
			},
			{
				Config: testAccDaemonTaskDefinitionConfig_containerDefinitionsUpdated(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_containerDefinitionHealthCheck(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_healthCheck(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.health_check.0.command.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.health_check.0.retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.health_check.0.start_period", "10"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.health_check.0.timeout", "5"),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_containerDefinitionLogConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_logConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.log_configuration.0.log_driver", "awslogs"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.log_configuration.0.options.awslogs-group", "/ecs/daemon"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.log_configuration.0.options.awslogs-region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.log_configuration.0.options.awslogs-stream-prefix", "ecs"),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_containerDefinitionLogConfigurationSecretOption(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_logConfigurationSecretOption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.log_configuration.0.secret_option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.log_configuration.0.secret_option.0.name", "LOG_SECRET"),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_containerDefinitionEnvironment(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_environment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.environment.#", "2"),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_containerDefinitionMountPoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_mountPoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.mount_point.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.mount_point.0.source_volume", "data"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.mount_point.0.container_path", "/usr/share/nginx/html"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.mount_point.0.read_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_containerDefinitionAllNestedBlocks(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_allNestedBlocks(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.log_configuration.0.log_driver", "awsfirelens"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.environment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.mount_point.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.mount_point.0.source_volume", "data"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.depends_on.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.depends_on.0.container_name", "sidecar"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.depends_on.0.condition", "START"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.linux_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.linux_parameters.0.init_process_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.restart_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.restart_policy.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.restart_policy.0.restart_attempt_period", "120"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.secret.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.secret.0.name", "MY_SECRET"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.system_control.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.system_control.0.namespace", "net.core.somaxconn"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.system_control.0.value", "1024"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.ulimit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.ulimit.0.name", "nofile"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.ulimit.0.hard_limit", "65536"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.ulimit.0.soft_limit", "65536"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.1.firelens_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.1.firelens_configuration.0.type", "fluentbit"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_containerDefinitionOptionalFields(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_optionalFields(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.working_directory", "/app"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.user", "nginx"),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.privileged", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "container_definition.0.readonly_root_filesystem", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_volumeWithoutHostPath(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_volumeWithoutHostPath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "volume.*", map[string]string{
						names.AttrName: "scratch",
					}),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinition_noCPUMemory(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon_task_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonTaskDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionConfig_noCPUMemory(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonTaskDefinitionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, rName),
					resource.TestCheckNoResourceAttr(resourceName, "cpu"),
					resource.TestCheckNoResourceAttr(resourceName, "memory"),
				),
			},
		},
	})
}

// Helper functions

func testAccCheckDaemonTaskDefinitionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		_, err := tfecs.FindDaemonTaskDefinitionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccCheckDaemonTaskDefinitionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_daemon_task_definition" {
				continue
			}

			dtd, err := tfecs.FindDaemonTaskDefinitionByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECS Daemon Task Definition %s still exists with status %s", rs.Primary.ID, dtd.Status)
		}

		return nil
	}
}

// Config generators

func testAccDaemonTaskDefinitionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDaemonTaskDefinitionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDaemonTaskDefinitionConfig_executionRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "execution" {
  name = "%[1]s-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_ecs_daemon_task_definition" "test" {
  family             = %[1]q
  cpu                = "512"
  memory             = "1024"
  execution_role_arn = aws_iam_role.execution.arn

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_taskRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "task" {
  name = "%[1]s-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_ecs_daemon_task_definition" "test" {
  family        = %[1]q
  cpu           = "512"
  memory        = "1024"
  task_role_arn = aws_iam_role.task.arn

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_volume(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }

  volume {
    name      = "data"
    host_path = "/mnt/data"
  }

  volume {
    name      = "logs"
    host_path = "/var/log"
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_multipleContainers(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }

  container_definition {
    name      = "sidecar"
    image     = "busybox:latest"
    cpu       = 128
    memory    = 256
    essential = false
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_containerDefinitionsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "nginx"
    image     = "nginx:alpine"
    cpu       = 256
    memory    = 512
    essential = true
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_healthCheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true

    health_check {
      command      = ["CMD-SHELL", "curl -f http://localhost/ || exit 1"]
      interval     = 30
      retries      = 3
      start_period = 10
      timeout      = 5
    }
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_logConfiguration(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_iam_role" "execution" {
  name = "%[1]s-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_ecs_daemon_task_definition" "test" {
  family             = %[1]q
  cpu                = "512"
  memory             = "1024"
  execution_role_arn = aws_iam_role.execution.arn

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true

    log_configuration {
      log_driver = "awslogs"
      options = {
        "awslogs-group"         = "/ecs/daemon"
        "awslogs-region"        = data.aws_region.current.name
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_logConfigurationSecretOption(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_iam_role" "execution" {
  name = "%[1]s-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "SecureString"
  value = "secret-value"
}

resource "aws_ecs_daemon_task_definition" "test" {
  family             = %[1]q
  cpu                = "512"
  memory             = "1024"
  execution_role_arn = aws_iam_role.execution.arn

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true

    log_configuration {
      log_driver = "awslogs"
      options = {
        "awslogs-group"         = "/ecs/daemon"
        "awslogs-region"        = data.aws_region.current.name
        "awslogs-stream-prefix" = "ecs"
      }
      secret_option {
        name       = "LOG_SECRET"
        value_from = aws_ssm_parameter.test.arn
      }
    }
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_environment(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true

    environment {
      name  = "ENV_VAR_1"
      value = "value1"
    }

    environment {
      name  = "ENV_VAR_2"
      value = "value2"
    }
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_mountPoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true

    mount_point {
      source_volume  = "data"
      container_path = "/usr/share/nginx/html"
      read_only      = true
    }
  }

  volume {
    name      = "data"
    host_path = "/mnt/data"
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_allNestedBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "execution" {
  name = "%[1]s-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-secret-value"
}

resource "aws_ecs_daemon_task_definition" "test" {
  family             = %[1]q
  cpu                = "512"
  memory             = "1024"
  execution_role_arn = aws_iam_role.execution.arn

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true

    depends_on {
      container_name = "sidecar"
      condition      = "START"
    }

    health_check {
      command  = ["CMD-SHELL", "curl -f http://localhost/ || exit 1"]
      interval = 30
      retries  = 3
      timeout  = 5
    }

    log_configuration {
      log_driver = "awsfirelens"
      options = {
        "Name" = "stdout"
      }
    }

    environment {
      name  = "APP_ENV"
      value = "production"
    }

    mount_point {
      source_volume  = "data"
      container_path = "/usr/share/nginx/html"
      read_only      = true
    }

    linux_parameters {
      init_process_enabled = true
    }

    restart_policy {
      enabled                = true
      restart_attempt_period = 120
    }

    secret {
      name       = "MY_SECRET"
      value_from = aws_secretsmanager_secret.test.arn
    }

    system_control {
      namespace = "net.core.somaxconn"
      value     = "1024"
    }

    ulimit {
      name       = "nofile"
      hard_limit = 65536
      soft_limit = 65536
    }
  }

  container_definition {
    name      = "sidecar"
    image     = "amazon/aws-for-fluent-bit:latest"
    cpu       = 128
    memory    = 256
    essential = false

    firelens_configuration {
      type = "fluentbit"
    }
  }

  volume {
    name      = "data"
    host_path = "/mnt/data"
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_optionalFields(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name                     = "nginx"
    image                    = "nginx:latest"
    cpu                      = 256
    memory                   = 512
    essential                = true
    working_directory        = "/app"
    user                     = "nginx"
    privileged               = false
    readonly_root_filesystem = true
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_volumeWithoutHostPath(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true

    mount_point {
      source_volume  = "scratch"
      container_path = "/tmp"
    }
  }

  volume {
    name = "scratch"
  }
}
`, rName)
}

func testAccDaemonTaskDefinitionConfig_noCPUMemory(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q

  container_definition {
    name      = "nginx"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}
`, rName)
}
