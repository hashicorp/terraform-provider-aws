// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.0", "echo"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.1", "test"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.image", "busybox"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.memory", "128"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.vcpus", "1"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_attributes(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_attributes(rName, 2, true, 3, 120, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:1`, rName))),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.attempts", "3"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "timeout.0.attempt_duration_seconds", "120"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_attributes(rName, 2, true, 4, 120, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:2`, rName))),
					testAccCheckJobDefinitionPreviousRegistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:3`, rName))),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "3"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
			{
				Config: testAccJobDefinitionConfig_attributes(rName, 1, false, 1, 60, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.attempts", "1"),
					resource.TestCheckResourceAttr(resourceName, "revision", "4"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "timeout.0.attempt_duration_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", "1"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbatch.ResourceJobDefinition, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_PlatformCapabilities_ec2(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_capabilitiesEC2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "container_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.0", "echo"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.1", "test"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.image", "busybox"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.memory", "128"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.vcpus", "1"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_PlatformCapabilitiesFargate_containerPropertiesDefaults(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_capabilitiesFargateContainerPropertiesDefaults(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.fargate_platform_configuration.#", "0"), // default block ignored
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.resource_requirements.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "container_properties.0.resource_requirements.*", map[string]string{
						names.AttrType:  "MEMORY",
						names.AttrValue: "512",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "container_properties.0.resource_requirements.*", map[string]string{
						names.AttrType:  "VCPU",
						names.AttrValue: "0.25",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
					"container_properties.0.fargate_platform_configuration",
					// on import, ignoring the default block value isn't necessary.
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_PlatformCapabilities_fargate(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_capabilitiesFargate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "container_properties.#", "1"),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"container_properties.0.execution_role_arn",
						"aws_iam_role.ecs_task_execution_role",
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.fargate_platform_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.fargate_platform_configuration.0.platform_version", "LATEST"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.network_configuration.0.assign_public_ip", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.resource_requirements.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "container_properties.0.resource_requirements.*", map[string]string{
						names.AttrType:  "MEMORY",
						names.AttrValue: "512",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "container_properties.0.resource_requirements.*", map[string]string{
						names.AttrType:  "VCPU",
						names.AttrValue: "0.25",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_ContainerProperties_advanced(t *testing.T) {
	ctx := acctest.Context(t)

	compare := awstypes.JobDefinition{
		Parameters: map[string]string{
			"param1": "val1",
			"param2": "val2",
		},
		RetryStrategy: &awstypes.RetryStrategy{
			Attempts: aws.Int32(1),
			EvaluateOnExit: []awstypes.EvaluateOnExit{
				{Action: awstypes.RetryAction(strings.ToLower(string(awstypes.RetryActionRetry))), OnStatusReason: aws.String("Host EC2*")},
				{Action: awstypes.RetryAction(strings.ToLower(string(awstypes.RetryActionExit))), OnReason: aws.String("*")},
			},
		},
		Timeout: &awstypes.JobTimeout{
			AttemptDurationSeconds: aws.Int32(60),
		},
		ContainerProperties: &awstypes.ContainerProperties{
			Command: []string{"ls", "-la"},
			Environment: []awstypes.KeyValuePair{
				{Name: aws.String("VARNAME"), Value: aws.String("VARVAL")},
			},
			Image:  aws.String("busybox"),
			Memory: aws.Int32(512),
			MountPoints: []awstypes.MountPoint{
				{ContainerPath: aws.String("/tmp"), ReadOnly: aws.Bool(false), SourceVolume: aws.String("tmp")},
			},
			ResourceRequirements: []awstypes.ResourceRequirement{},
			Secrets:              []awstypes.Secret{},
			Ulimits: []awstypes.Ulimit{
				{HardLimit: aws.Int32(1024), Name: aws.String("nofile"), SoftLimit: aws.Int32(1024)},
			},
			Vcpus: aws.Int32(1),
			Volumes: []awstypes.Volume{
				{
					Host: &awstypes.Host{SourcePath: aws.String("/tmp")},
					Name: aws.String("tmp"),
				},
			},
		},
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val2", 1, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					testAccCheckJobDefinitionAttributes(ctx, &compare),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val2", 1, 60),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
					"retry_strategy.0.evaluate_on_exit.0.action",
					// ^ ImportStateVerify ignores semantic equivalence of differently-cased strings
				},
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 1, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 1, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 3, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "3"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 3, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "3"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 3, 120),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "4"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "5"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_ContainerProperties_minorUpdate(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-la"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:1`, rName))),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-lah"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:2`, rName))),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-hal"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:3`, rName))),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "3"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_propagateTags(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_propagateTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "container_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.0", "echo"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.resource_requirements.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.mount_points.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.ulimits.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.command.volumes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_ContainerProperties_EmptyField(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_containerProperties_emptyField(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "container_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.environment.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.environment.1.value", names.AttrValue),
					// Note: the fixEnvVars() functions preserve written order, so it's safe to
					// index directly into the expected env var
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
					"container_properties.0.environment",
					// ^ importing the resource will result in a diff since since the AWS Batch
					// API does not keep track of environment variables
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_NodeProperties_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_nodeProperties(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "node_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.target_nodes", "0:"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.1.target_nodes", "1:"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "multinode"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_NodeProperties_advanced(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_nodePropertiesAdvanced(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
			{
				Config: testAccJobDefinitionConfig_nodePropertiesAdvancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.target_nodes", "0:"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.1.target_nodes", "1:"),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_NodeProperties_withEKS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_nodePropertiesEKS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "node_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.main_node", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.args.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.command.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.command.0", "sleep"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.command.1", "60"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.env.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.image", "public.ecr.aws/amazonlinux/amazonlinux = 2"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.name", "test-eks-container-1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.image_pull_secrets.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.env.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.resources.0.requests.memory", "1024Mi"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.resources.0.requests.cpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.security_context.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.security_context.0.privileged", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.security_context.0.read_only_root_file_system", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.security_context.0.run_as_user", "1000"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.security_context.0.run_as_non_root", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.security_context.0.run_as_group", "3000"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.containers.0.volume_mounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.init_containers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.image_pull_secrets.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.eks_properties.0.pod_properties.0.instance_types.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.node_range_properties.0.target_nodes", "0:"),
					resource.TestCheckResourceAttr(resourceName, "node_properties.0.num_nodes", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
					"node_properties",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_NodeProperties_withECS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_nodePropertiesECS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}
func TestAccBatchJobDefinition_EKSProperties_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_EKSProperties_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.init_containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_EKSProperties_update(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_EKSProperties_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
				),
			},
			{
				Config: testAccJobDefinitionConfig_EKSProperties_advancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.0.image_pull_policy", "Always"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.volumes.0.name", "tmp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_EKSProperties_imagePullSecrets(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_EKSProperties_imagePullSecrets(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.image_pull_secrets.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "eks_properties.0.pod_properties.0.image_pull_secrets.*", map[string]string{
						names.AttrName: "chihiro",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "eks_properties.0.pod_properties.0.image_pull_secrets.*", map[string]string{
						names.AttrName: "haku",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_EKSProperties_multiContainers(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_EKSProperties_multiContainer(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_createTypeContainerWithNodeProperties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobDefinitionConfig_createTypeContainerWithNodeProperties(rName),
				ExpectError: regexache.MustCompile("No `node_properties` can be specified when `type` is \"container\""),
			},
		},
	})
}

func TestAccBatchJobDefinition_createTypeMultiNodeWithContainerProperties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobDefinitionConfig_createTypeMultiNodeWithContainerProperties(rName),
				ExpectError: regexache.MustCompile("No `container_properties` can be specified when `type` is \"multinode\""),
			},
		},
	})
}

func TestAccBatchJobDefinition_schedulingPriority(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_schedulingPriority(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_emptyRetryStrategy(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobDefinitionConfig_emptyRetryStrategy(rName),
				ExpectError: regexache.MustCompile(`ClientException: Error executing request, Exception : RetryAttempts must be provided with retry strategy`),
			},
		},
	})
}

func TestAccBatchJobDefinition_ECSProperties_update(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_ECSProperties_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ecs_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ecs_properties.0.task_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ecs_properties.0.task_properties.0.containers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ecs_properties.0.task_properties.0.containers.0.environment.#", "1"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_ECSProperties_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ecs_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ecs_properties.0.task_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ecs_properties.0.task_properties.0.containers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ecs_properties.0.task_properties.0.containers.0.environment.#", "2"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_updateWithTags(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_simpleWithTags(rName, "echo", "test1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
				),
			},
			// Ensure that tags are put on the new revision.
			{
				Config: testAccJobDefinitionConfig_simpleWithTags(rName, "echo", "test2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("revision")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
				),
			},
		},
	})
}

func testAccCheckJobDefinitionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchClient(ctx)

		_, err = tfbatch.FindJobDefinitionByARN(ctx, conn, rs.Primary.ID)
		return err
	}
}

func testAccCheckJobDefinitionPreviousRegistered(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchClient(ctx)

		previousARN := parseJobDefinitionPreviousARN(rs.Primary.ID)

		jobDefinition, err := tfbatch.FindJobDefinitionByARN(ctx, conn, previousARN)

		if err != nil {
			return err
		}

		if aws.ToString(jobDefinition.Status) != "ACTIVE" {
			return fmt.Errorf("Batch Job Definition %s is a previous revision that is not ACTIVE", previousARN)
		}

		return nil
	}
}

func testAccCheckJobDefinitionPreviousDeregistered(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchClient(ctx)

		previousARN := parseJobDefinitionPreviousARN(rs.Primary.ID)

		_, err := tfbatch.FindJobDefinitionByARN(ctx, conn, previousARN)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Batch Job Definition %s is a previous revision that is still ACTIVE", previousARN)
	}
}

func parseJobDefinitionPreviousARN(currentARN string) (previousARN string) {
	re := regexache.MustCompile(`job-definition/.*?:(.*)`)
	revisionCurrentStr := re.FindStringSubmatch(currentARN)[1]

	revisionCurrent, _ := strconv.Atoi(revisionCurrentStr)
	revisionPrevious := revisionCurrent - 1

	re = regexache.MustCompile(`^(arn:.*:batch:[a-z0-9-]+:[0-9]+:job-definition/[a-z0-9-]+):`)
	arnPrefix := re.FindStringSubmatch(currentARN)[1]
	previousARN = fmt.Sprintf("%s:%d", arnPrefix, revisionPrevious)

	return previousARN
}

func testAccCheckJobDefinitionAttributes(ctx context.Context, compare *awstypes.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_definition" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).BatchClient(ctx)
			jd, err := tfbatch.FindJobDefinitionByARN(ctx, conn, rs.Primary.ID)

			if err != nil {
				return err
			}
			if aws.ToString(jd.JobDefinitionArn) != rs.Primary.Attributes[names.AttrARN] {
				return fmt.Errorf("Bad Job Definition ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes[names.AttrARN], aws.ToString(jd.JobDefinitionArn))
			}
			if compare != nil {
				if compare.Parameters != nil && !reflect.DeepEqual(compare.Parameters, jd.Parameters) {
					return fmt.Errorf("Bad Job Definition Params\n\t expected: %v\n\tgot: %v\n", compare.Parameters, jd.Parameters)
				}
				if compare.RetryStrategy != nil && aws.ToInt32(compare.RetryStrategy.Attempts) != aws.ToInt32(jd.RetryStrategy.Attempts) {
					return fmt.Errorf("Bad Job Definition Retry Strategy\n\t expected: %d\n\tgot: %d\n", aws.ToInt32(compare.RetryStrategy.Attempts), aws.ToInt32(jd.RetryStrategy.Attempts))
				}
				if compare.RetryStrategy != nil && !reflect.DeepEqual(compare.RetryStrategy.EvaluateOnExit, jd.RetryStrategy.EvaluateOnExit) {
					return fmt.Errorf("Bad Job Definition Retry Strategy\n\t expected: %v\n\tgot: %v\n", compare.RetryStrategy.EvaluateOnExit, jd.RetryStrategy.EvaluateOnExit)
				}
				if compare.ContainerProperties != nil && compare.ContainerProperties.Command != nil && !reflect.DeepEqual(compare.ContainerProperties, jd.ContainerProperties) {
					return fmt.Errorf("Bad Job Definition Container Properties\n\t expected: %v\n\tgot: %v\n", compare.ContainerProperties, jd.ContainerProperties)
				}
				if compare.NodeProperties != nil && compare.NodeProperties.NumNodes != nil && !reflect.DeepEqual(compare.NodeProperties, jd.NodeProperties) {
					return fmt.Errorf("Bad Job Definition Node Properties\n\t expected: %v\n\tgot: %v\n", compare.NodeProperties, jd.NodeProperties)
				}
			}
		}
		return nil
	}
}

func testAccCheckJobDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_definition" {
				continue
			}

			_, err := tfbatch.FindJobDefinitionByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Batch Job Definition %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccJobDefinitionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties {
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  }
  name = %[1]q
  type = "container"
}
`, rName)
}

func testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, param string, retries, timeout int) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

  parameters = {
    param1 = "val1"
    param2 = %[2]q
  }

  retry_strategy {
    attempts = %[3]d
    evaluate_on_exit {
      action           = "RETRY"
      on_status_reason = "Host EC2*"
    }
    evaluate_on_exit {
      action    = "exit"
      on_reason = "*"
    }
  }

  timeout {
    attempt_duration_seconds = %[4]d
  }

  container_properties {
    command = ["ls", "-la"]
    image   = "busybox"
    memory  = 512
    vcpus   = 1

    volumes {
      host {
        source_path = "/tmp"
      }
      name = "tmp"
    }

    environment {
      name  = "VARNAME"
      value = "VARVAL"
    }

    mount_points {
      source_volume  = "tmp"
      container_path = "/tmp"
      read_only      = false
    }

    ulimits {
      hard_limit = 1024
      soft_limit = 1024
      name       = "nofile"
    }
  }
}
`, rName, param, retries, timeout)
}

func testAccJobDefinitionConfig_containerPropertiesAdvancedUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  container_properties {
    command = ["ls", "-la"]
    image   = "busybox"
    memory  = 1024
    vcpus   = 1
    volumes {
      host {
        source_path = "/tmp"
      }
      name = "tmp"
    }

    environment {
      name  = "VARNAME"
      value = "VARVAL"
    }
    mount_points {
      source_volume  = "tmp"
      container_path = "/tmp"
      read_only      = false
    }

    ulimits {
      hard_limit = 1024
      name       = "nofile"
      soft_limit = 1024
    }
  }

}
`, rName)
}

func testAccJobDefinitionConfig_nodePropertiesEKS(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "multinode"
  retry_strategy {
    attempts = 1
  }

  node_properties {
    main_node = 0
    num_nodes = 1
    node_range_properties {
      target_nodes = "0:"
      eks_properties {
        pod_properties {
          containers {
            name  = "test-eks-container-1"
            image = "public.ecr.aws/amazonlinux/amazonlinux = 2"
            command = [
              "sleep",
              "60"
            ]
            resources {
              requests = {
                memory = "1024Mi"
                cpu    = "1"
              }
            }
            security_context {
              run_as_user                = 1000
              run_as_group               = 3000
              privileged                 = true
              read_only_root_file_system = true
              run_as_non_root            = true
            }
          }
        }
      }
    }
  }
}
  `, rName)
}

func testAccJobDefinitionConfig_nodePropertiesECS(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "multinode"
  retry_strategy {
    attempts = 1
  }

  node_properties {
    main_node = 0
    num_nodes = 1
    node_range_properties {
      target_nodes = "0:"
      ecs_properties {
        task_properties {
          containers {
            image      = "public.ecr.aws/amazonlinux/amazonlinux:1"
            command    = ["sleep", "60"]
            name       = "container_a"
            privileged = false
            resource_requirements {
              type  = "VCPU"
              value = "1"
            }
            resource_requirements {
              type  = "MEMORY"
              value = "2048"
            }
          }
          containers {
            image   = "public.ecr.aws/amazonlinux/amazonlinux:1"
            command = ["sleep", "360"]
            name    = "container_b"
            resource_requirements {
              type  = "VCPU"
              value = "1"
            }
            resource_requirements {
              type  = "MEMORY"
              value = "2048"
            }
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccJobDefinitionConfig_containerProperties(rName, subcommand string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  environment {
    variables = {
      JobDefinition = aws_batch_job_definition.test.arn
    }
  }

  depends_on = [aws_batch_job_definition.test]
}

resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  container_properties {
    command = ["ls", "%[2]s"]
    image   = "busybox"
    memory  = 512
    vcpus   = 1

    volumes {
      name = "tmp"
      host {
        source_path = "/tmp"
      }
    }

    environment {
      name  = "VARNAME"
      value = "VARVAL"
    }

    mount_points {
      source_volume  = "tmp"
      container_path = "/tmp"
      read_only      = false
    }

    ulimits {
      hard_limit = 1024
      name       = "nofile"
      soft_limit = 1024
    }
  }
}
`, rName, subcommand))
}

func testAccJobDefinitionConfig_capabilitiesEC2(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

  platform_capabilities = [
    "EC2",
  ]

  container_properties {
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  }
}
`, rName)
}

func testAccJobDefinitionConfig_capabilitiesFargateContainerPropertiesDefaults(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

  platform_capabilities = [
    "FARGATE",
  ]

  container_properties {
    image = "busybox"
    resource_requirements {
      type  = "MEMORY"
      value = "512"
    }
    resource_requirements {
      type  = "VCPU"
      value = "0.25"
    }
    execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  }
}
`, rName)
}

func testAccJobDefinitionConfig_capabilitiesFargate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

  platform_capabilities = [
    "FARGATE",
  ]

  container_properties {
    command = ["echo", "test"]
    image   = "busybox"
    fargate_platform_configuration {
      platform_version = "LATEST"
    }
    network_configuration {
      assign_public_ip = "DISABLED"
    }
    resource_requirements {
      type  = "MEMORY"
      value = "512"
    }
    resource_requirements {
      type  = "VCPU"
      value = "0.25"
    }
    execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  }

}
`, rName)
}

func testAccJobDefinitionConfig_propagateTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties {
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  }
  name = %[1]q
  type = "container"

  propagate_tags = true
}
`, rName)
}

func testAccJobDefinitionConfig_containerProperties_emptyField(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties {
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
    environment {
      name  = "EMPTY"
      value = ""
    }

    environment {
      name  = "VALUE"
      value = "value"
    }
  }
  name = %[1]q
  type = "container"
}
`, rName)
}

func testAccJobDefinitionConfig_nodeProperties(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "multinode"

  node_properties {
    main_node = 0
    node_range_properties {
      container {
        command = ["ls", "-la"]
        image   = "busybox"
        memory  = 128
        vcpus   = 1
      }
      target_nodes = "0:"
    }
    node_range_properties {
      container {
        command = ["echo", "test"]
        image   = "busybox"
        memory  = 128
        vcpus   = 1
      }
      target_nodes = "1:"
    }
    num_nodes = 2
  }
}
`, rName)
}

func testAccJobDefinitionConfig_nodePropertiesAdvanced(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "multinode"
  parameters = {
    param1 = "val1"
    param2 = "val2"
  }
  timeout {
    attempt_duration_seconds = 60
  }

  node_properties {
    main_node = 1
    num_nodes = 4
    node_range_properties {
      target_nodes = "0:"
      container {
        command = ["ls", "-la"]
        image   = "busybox"
        memory  = 512
        vcpus   = 1
        volumes {
          host {
            source_path = "/tmp"
          }
          name = "tmp"
        }

        environment {
          name  = "VARNAME"
          value = "VARVAL"
        }
        mount_points {
          source_volume  = "tmp"
          container_path = "/tmp"
          read_only      = false
        }

        ulimits {
          hard_limit = 1024
          name       = "nofile"
          soft_limit = 1024
        }
      }
    }

    node_range_properties {
      target_nodes = "1:"
      container {
        command = ["echo", "test"]
        image   = "busybox"
        memory  = 128
        vcpus   = 1
      }
    }
  }
}
`, rName)
}

func testAccJobDefinitionConfig_nodePropertiesAdvancedUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "multinode"
  parameters = {
    param1 = "val1"
    param2 = "val2"
  }
  timeout {
    attempt_duration_seconds = 60
  }

  node_properties {
    main_node = 1
    node_range_properties {
      container {
        command = ["ls", "-la"]
        image   = "busybox"
        memory  = 512
        vcpus   = 1
      }
      target_nodes = "0:"
    }


    node_range_properties {
      container {
        command = ["echo", "test"]
        image   = "busybox"
        memory  = 128
        vcpus   = 1
      }
      target_nodes = "1:"
    }

    num_nodes = 4
  }
}
`, rName)
}
func testAccJobDefinitionConfig_EKSProperties_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  eks_properties {
    pod_properties {
      host_network            = true
      share_process_namespace = false
      init_containers {
        name  = "init0"
        image = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = [
          "sleep",
          "60"
        ]
        resources {
          limits = {
            cpu    = "1"
            memory = "1024Mi"
          }
        }
      }
      containers {
        image = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = [
          "sleep",
          "60"
        ]
        resources {
          limits = {
            cpu    = "1"
            memory = "1024Mi"
          }
        }
      }
      metadata {
        labels = {
          environment = "test"
          name        = %[1]q
        }
      }
    }
  }
}
`, rName)
}

func testAccJobDefinitionConfig_EKSProperties_imagePullSecrets(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  eks_properties {
    pod_properties {
      host_network = true
      containers {
        image = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = [
          "sleep",
          "60"
        ]
        resources {
          limits = {
            cpu    = "1"
            memory = "1024Mi"
          }
        }
      }
      image_pull_secrets {
        name = "chihiro"
      }
      image_pull_secrets {
        name = "haku"
      }
      metadata {
        labels = {
          environment = "test"
          name        = %[1]q
        }
      }
    }
  }
}
`, rName)
}

func testAccJobDefinitionConfig_EKSProperties_multiContainer(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  eks_properties {
    pod_properties {
      host_network = true
      containers {
        name  = "container1"
        image = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = [
          "sleep",
          "60"
        ]
        resources {
          limits = {
            cpu    = "1"
            memory = "1024Mi"
          }
        }
      }
      containers {
        name  = "container2"
        image = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = [
          "sleep",
          "60"
        ]
        resources {
          limits = {
            cpu    = "1"
            memory = "1024Mi"
          }
        }
      }
      image_pull_secrets {
        name = "chihiro"
      }
      image_pull_secrets {
        name = "haku"
      }
      metadata {
        labels = {
          environment = "test"
          name        = %[1]q
        }
      }
    }
  }
}
`, rName)
}

func testAccJobDefinitionConfig_EKSProperties_advancedUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  eks_properties {
    pod_properties {
      host_network = true
      containers {
        args              = ["60"]
        image             = "public.ecr.aws/amazonlinux/amazonlinux:2"
        image_pull_policy = "Always"
        name              = "sleep"
        command = [
          "sleep",
        ]
        resources {
          requests = {
            cpu    = "1"
            memory = "1024Mi"
          }
          limits = {
            cpu    = "1"
            memory = "1024Mi"
          }
        }
        security_context {
          privileged                 = true
          read_only_root_file_system = true
          run_as_group               = 1000
          run_as_user                = 1000
          run_as_non_root            = true
        }
        volume_mounts {
          mount_path = "/tmp"
          read_only  = true
          name       = "tmp"
        }
        env {
          name  = "Test"
          value = "Environment Variable"
        }
      }
      metadata {
        labels = {
          environment = "test"
          name        = %[1]q
        }
      }
      volumes {
        name = "tmp"
        empty_dir {
          medium     = "Memory"
          size_limit = "128Mi"
        }
      }
      service_account_name = "test-service-account"
      dns_policy           = "ClusterFirst"
    }
  }
  parameters = {
    param1 = "val1"
    param2 = "val2"
  }

  timeout {
    attempt_duration_seconds = 60
  }
}
`, rName)
}

func testAccJobDefinitionConfig_createTypeContainerWithNodeProperties(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  parameters = {
    param1 = "val1"
    param2 = "val2"
  }
  timeout {
    attempt_duration_seconds = 60
  }

  node_properties {
    main_node = 1
    node_range_properties {
      container {
        command = ["ls", "-la"]
        image   = "busybox"
        memory  = 512
        vcpus   = 1
      }
      target_nodes = "0:"
    }
    node_range_properties {
      container {
        command = ["echo", "test"]
        image   = "busybox"
        memory  = 128
        vcpus   = 1
      }
      target_nodes = "1:"
    }
    num_nodes = 4
  }
}
`, rName)
}

func testAccJobDefinitionConfig_createTypeMultiNodeWithContainerProperties(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "multinode"
  parameters = {
    param1 = "val1"
    param2 = "val2"
  }
  timeout {
    attempt_duration_seconds = 60
  }

  container_properties {
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  }
}
`, rName)
}

func testAccJobDefinitionConfig_schedulingPriority(rName string, priority int) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties {
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  }
  name                = %[1]q
  type                = "container"
  scheduling_priority = %[2]d
}
`, rName, priority)
}

func testAccJobDefinitionConfig_attributes(rName string, sp int, pt bool, rsa int, timeout int, dereg bool) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name                = %[1]q
  type                = "container"
  scheduling_priority = %[2]d
  propagate_tags      = %[3]t
  container_properties {
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  }
  retry_strategy {
    attempts = %[4]d
  }
  timeout {
    attempt_duration_seconds = %[5]d
  }
  deregister_on_new_revision = %[6]t
  platform_capabilities = [
    "EC2",
  ]
}
`, rName, sp, pt, rsa, timeout, dereg)
}

func testAccJobDefinitionConfig_emptyRetryStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  container_properties {
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  }
  retry_strategy {}
}
`, rName)
}

func testAccJobDefinitionConfig_ECSProperties_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

  platform_capabilities = ["FARGATE"]

  ecs_properties {
    task_properties {
      execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
      containers {
        image   = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = ["sleep", "60"]
        depends_on {
          container_name = "container_b"
          condition      = "COMPLETE"
        }

        secrets {
          name       = "TEST"
          value_from = "DUMMY"
        }

        environment {
          name  = "test 1"
          value = "Environment Variable 1"
        }

        essential = true
        log_configuration {
          log_driver = "awslogs"
          options = {
            "awslogs-group"         = %[1]q
            "awslogs-region"        = %[2]q
            "awslogs-stream-prefix" = "ecs"
          }
        }
        name                     = "container_a"
        privileged               = false
        readonly_root_filesystem = false
        resource_requirements {
          value = "1.0"
          type  = "VCPU"
        }
        resource_requirements {
          value = "2048"
          type  = "MEMORY"
        }
      }
      containers {
        image     = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command   = ["sleep", "360"]
        name      = "container_b"
        essential = false
        resource_requirements {
          value = "1.0"
          type  = "VCPU"
        }
        resource_requirements {
          value = "2048"
          type  = "MEMORY"
        }
      }
    }
  }
}
`, rName, acctest.Region())
}

func testAccJobDefinitionConfig_ECSProperties_updated(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

  platform_capabilities = ["FARGATE"]

  ecs_properties {
    task_properties {
      execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
      containers {
        image   = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = ["sleep", "60"]
        depends_on {
          container_name = "container_b"
          condition      = "COMPLETE"
        }
        secrets {
          name       = "TEST"
          value_from = "DUMMY"
        }
        environment {
          name  = "test 1"
          value = "Environment Variable 1"
        }
        environment {
          name  = "test 2"
          value = "Environment Variable 2"
        }
        essential = true
        log_configuration {
          log_driver = "awslogs"
          options = {
            "awslogs-group"         = %[1]q
            "awslogs-region"        = %[2]q
            "awslogs-stream-prefix" = "ecs"
          }
        }
        name                     = "container_a"
        privileged               = false
        readonly_root_filesystem = false
        resource_requirements {
          value = "1.0"
          type  = "VCPU"
        }
        resource_requirements {
          value = "2048"
          type  = "MEMORY"
        }
      }
      containers {
        image     = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command   = ["sleep", "360"]
        name      = "container_b"
        essential = false
        resource_requirements {
          value = "1.0"
          type  = "VCPU"
        }
        resource_requirements {
          value = "2048"
          type  = "MEMORY"
        }
      }
    }
  }
}
`, rName, acctest.Region())
}

func testAccJobDefinitionConfig_simpleWithTags(rName, command1, command2 string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties {
    command = [%[2]q, %[3]q]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  }
  name = %[1]q
  type = "container"

  tags = {
    Name = %[1]q
  }
}
`, rName, command1, command2)
}
