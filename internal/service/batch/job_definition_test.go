// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "arn_prefix", "batch", "job-definition/{name}"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "container_properties", `{
 						"command": ["echo", "test"],
 						"image": "busybox",
 						"memory": 128,
 						"vcpus": 1,
 						"environment": [],
 						"mountPoints": [],
 						"resourceRequirements": [],
 						"secrets": [],
 						"ulimits": [],
 						"volumes": []
 						}`),
					resource.TestCheckResourceAttr(resourceName, "ecs_properties", ""),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_Identity_ChangeOnUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-la"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("revision"), knownvalue.Int32Exact(1)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName, tfjsonpath.New(names.AttrARN), "batch", "job-definition/{name}:{revision}"),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrARN)),
				},
			},

			// Step 2: Update
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-lah"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("revision"), knownvalue.Int32Exact(2)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName, tfjsonpath.New(names.AttrARN), "batch", "job-definition/{name}:{revision}"),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrARN)),
				},
			},
		},
	})
}

func TestAccBatchJobDefinition_attributes(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_attributes(rName, 2, true, 3, 120, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "arn_prefix", "batch", "job-definition/{name}"),
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
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					testAccCheckJobDefinitionPreviousRegistered(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "arn_prefix", "batch", "job-definition/{name}"),
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
			},
			{
				Config: testAccJobDefinitionConfig_attributes(rName, 1, false, 1, 60, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "arn_prefix", "batch", "job-definition/{name}"),
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
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckSDKResourceDisappears(ctx, t, tfbatch.ResourceJobDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_PlatformCapabilities_ec2(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_capabilitiesEC2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "container_properties", `{
 						"command": ["echo", "test"],
 						"image": "busybox",
 						"memory": 128,
 						"vcpus": 1,
 						"environment": [],
 						"mountPoints": [],
 						"resourceRequirements": [],
 						"secrets": [],
 						"ulimits": [],
 						"volumes": []
 						}`),
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
			},
		},
	})
}

func TestAccBatchJobDefinition_PlatformCapabilitiesFargate_containerPropertiesDefaults(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_capabilitiesFargateContainerPropertiesDefaults(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(command)", "0"),
					acctest.CheckResourceAttrJMESPair(resourceName, "container_properties", "executionRoleArn", "aws_iam_role.ecs_task_execution_role", names.AttrARN),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "fargatePlatformConfiguration.platformVersion", "LATEST"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(resourceRequirements)", "2"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='VCPU'].value | [0]", "0.25"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='MEMORY'].value | [0]", "512"),
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
			},
		},
	})
}

func TestAccBatchJobDefinition_PlatformCapabilities_fargate(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_capabilitiesFargate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrJMESPair(resourceName, "container_properties", "executionRoleArn", "aws_iam_role.ecs_task_execution_role", names.AttrARN),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "fargatePlatformConfiguration.platformVersion", "LATEST"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "networkConfiguration.assignPublicIp", "DISABLED"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(resourceRequirements)", "2"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='VCPU'].value | [0]", "0.25"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='MEMORY'].value | [0]", "512"),
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
			},
		},
	})
}

func TestAccBatchJobDefinition_ContainerProperties_advanced(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val2", 1, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					testAccCheckJobDefinitionAttributes(&jd, &compare),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 1, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 1, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 3, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", "3"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 3, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", "3"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 3, 120),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", "4"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "5"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_ContainerProperties_minorUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-la"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-lah"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-hal"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"), testAccCheckJobDefinitionPreviousDeregistered(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "3"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_propagateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_propagateTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "container_properties", `{
 						"command": ["echo", "test"],
 						"image": "busybox",
 						"memory": 128,
 						"vcpus": 1,
 						"environment": [],
 						"mountPoints": [],
 						"resourceRequirements": [],
 						"secrets": [],
 						"ulimits": [],
 						"volumes": []
 						}`),
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
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_containerProperties_emptyField(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(environment)", "1"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "environment[?name=='VALUE'].value | [0]", names.AttrValue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_NodeProperties_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_nodeProperties(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "node_properties", `{
 						"mainNode": 0,
 						"nodeRangeProperties": [
 							{
 								"container": {
 									"command": ["ls","-la"],
 									"environment": [],
 									"image": "busybox",
 									"memory": 128,
 									"mountPoints": [],
 									"resourceRequirements": [],
 									"secrets": [],
 									"ulimits": [],
 									"vcpus": 1,
 									"volumes": []
 								},
 								"instanceTypes": [],
 								"targetNodes": "0:"
 							},
 							{
 								"container": {
 									"command": ["echo","test"],
 									"environment": [],
 									"image": "busybox",
 									"memory": 128,
 									"mountPoints": [],
 									"resourceRequirements": [],
 									"secrets": [],
 									"ulimits": [],
 									"vcpus": 1,
 									"volumes": []
 								},
 								"instanceTypes": [],
 								"targetNodes": "1:"
 							}
 						],
 						"numNodes": 2
 					}`),
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
			},
		},
	})
}

func TestAccBatchJobDefinition_NodeProperties_advanced(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_nodePropertiesAdvanced(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "node_properties", `{
 						"mainNode": 1,
 						"nodeRangeProperties": [
 							{
 								"container": {
 									"command": ["ls","-la"],
 									"environment": [{"name":"VARNAME","value":"VARVAL"}],
 									"image": "busybox",
 									"memory": 512,
 									"mountPoints": [{"containerPath":"/tmp","readOnly":false,"sourceVolume":"tmp"}],
 									"resourceRequirements": [],
 									"secrets": [],
 									"ulimits": [{"hardLimit":1024,"name":"nofile","softLimit":1024}],
 									"vcpus": 1,
 									"volumes": [{"host":{"sourcePath":"/tmp"},"name":"tmp"}]
 								},
 								"instanceTypes": [],
 								"targetNodes": "0:"
 							},
 							{
 								"container": {
 									"command": ["echo","test"],
 									"environment": [],
 									"image": "busybox",
 									"memory": 128,
 									"mountPoints": [],
 									"resourceRequirements": [],
 									"secrets": [],
 									"ulimits": [],
 									"vcpus":1,
 									"volumes": []
 								},
 								"instanceTypes": [],
 								"targetNodes": "1:"
 							}
 						],
 						"numNodes":4
 					}`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobDefinitionConfig_nodePropertiesAdvancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "node_properties", `{
 						"mainNode": 1,
 						"nodeRangeProperties": [
 							{
 								"container": {
 									"command": ["ls","-la"],
 									"environment": [],
 									"image": "busybox",
 									"memory": 512,
 									"mountPoints": [],
 									"resourceRequirements": [],
 									"secrets": [],
 									"ulimits": [],
 									"vcpus": 1,
 									"volumes": []
 								},
 								"instanceTypes": [],
 								"targetNodes": "0:"
 							},
 							{
 								"container": {
 									"command": ["echo","test"],
 									"environment": [],
 									"image": "busybox",
 									"memory": 128,
 									"mountPoints": [],
 									"resourceRequirements": [],
 									"secrets": [],
 									"ulimits": [],
 									"vcpus": 1,
 									"volumes": []
 								},
 								"instanceTypes": [],
 								"targetNodes": "1:"
 							}
 						],
 						"numNodes":4
 					}`),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_NodeProperties_withEKS(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_nodePropertiesEKS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "node_properties", `{
 						"mainNode": 0,
 						"nodeRangeProperties": [
 							{
 							"eksProperties": {
 								"podProperties": {
 								"containers": [
 									{
 									"args": [],
 									"command": ["sleep", "60"],
 									"env": [],
 									"image": "public.ecr.aws/amazonlinux/amazonlinux = 2",
 									"name": "test-eks-container-1",
 									"resources": { "requests": { "memory": "1024Mi", "cpu": "1" } },
 									"securityContext": {
 										"privileged": true,
 										"readOnlyRootFilesystem": true,
 										"runAsGroup": 3000,
 										"runAsNonRoot": true,
 										"runAsUser": 1000
 									},
 									"volumeMounts": []
 									}
 								],
 								"imagePullSecrets": [],
 								"initContainers": [],
 								"volumes": []
 								}
 							},
 							"instanceTypes": [],
 							"targetNodes": "0:"
 							}
 						],
 						"numNodes": 1
 						}`),
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
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_nodePropertiesECS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_NodeProperties_withECS_update(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_nodePropertiesECS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttrSet(resourceName, "node_properties"),
					acctest.CheckResourceAttrJMES(resourceName, "node_properties", "nodeRangeProperties[0].ecsProperties.taskProperties[0].containers[0].resourceRequirements[?type=='VCPU'].value | [0]", "1"),
					acctest.CheckResourceAttrJMES(resourceName, "node_properties", "nodeRangeProperties[0].ecsProperties.taskProperties[0].containers[0].resourceRequirements[?type=='MEMORY'].value | [0]", "2048"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_nodePropertiesECS_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttrSet(resourceName, "node_properties"),
					acctest.CheckResourceAttrJMES(resourceName, "node_properties", "nodeRangeProperties[0].ecsProperties.taskProperties[0].containers[0].resourceRequirements[?type=='VCPU'].value | [0]", "1"),
					acctest.CheckResourceAttrJMES(resourceName, "node_properties", "nodeRangeProperties[0].ecsProperties.taskProperties[0].containers[0].resourceRequirements[?type=='MEMORY'].value | [0]", "4096"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_EKSProperties_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_EKSProperties_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.init_containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.0.image_pull_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.init_containers.0.image_pull_policy", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_EKSProperties_update(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_EKSProperties_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
				),
			},
			{
				Config: testAccJobDefinitionConfig_EKSProperties_advancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.0.image_pull_policy", "Always"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.volumes.0.name", "tmp"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.0.security_context.0.allow_privilege_escalation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_EKSProperties_imagePullSecrets(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_EKSProperties_imagePullSecrets(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.0.image_pull_policy", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.image_pull_secret.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "eks_properties.*.pod_properties.*.image_pull_secret.*", map[string]string{
						names.AttrName: "chihiro",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "eks_properties.*.pod_properties.*.image_pull_secret.*", map[string]string{
						names.AttrName: "haku",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_EKSProperties_multiContainers(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_EKSProperties_multiContainer(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_createTypeContainerWithNodeProperties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
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
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_schedulingPriority(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_emptyRetryStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_emptyRetryStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-definition/{name}:{revision}"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_ECSProperties_update(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_ECSProperties_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_properties"),
					acctest.CheckResourceAttrJMES(resourceName, "ecs_properties", "length(taskProperties)", "1"),
					acctest.CheckResourceAttrJMES(resourceName, "ecs_properties", "length(taskProperties[0].containers)", "2"),
					acctest.CheckResourceAttrJMES(resourceName, "ecs_properties", "length(taskProperties[0].containers[0].environment)", "1"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_ECSProperties_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_properties"),
					acctest.CheckResourceAttrJMES(resourceName, "ecs_properties", "length(taskProperties)", "1"),
					acctest.CheckResourceAttrJMES(resourceName, "ecs_properties", "length(taskProperties[0].containers)", "2"),
					acctest.CheckResourceAttrJMES(resourceName, "ecs_properties", "length(taskProperties[0].containers[0].environment)", "2"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_updateWithTags(t *testing.T) {
	ctx := acctest.Context(t)
	var jd awstypes.JobDefinition
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx, t),
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
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
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
					testAccCheckJobDefinitionExists(ctx, t, resourceName, &jd),
				),
			},
		},
	})
}

func testAccCheckJobDefinitionExists(ctx context.Context, t *testing.T, n string, v *awstypes.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BatchClient(ctx)

		output, err := tfbatch.FindJobDefinitionByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckJobDefinitionPreviousRegistered(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BatchClient(ctx)

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

func testAccCheckJobDefinitionPreviousDeregistered(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BatchClient(ctx)

		previousARN := parseJobDefinitionPreviousARN(rs.Primary.ID)

		_, err := tfbatch.FindJobDefinitionByARN(ctx, conn, previousARN)

		if retry.NotFound(err) {
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

func testAccCheckJobDefinitionAttributes(jd *awstypes.JobDefinition, compare *awstypes.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_definition" {
				continue
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

func testAccCheckJobDefinitionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_definition" {
				continue
			}

			_, err := tfbatch.FindJobDefinitionByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
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
  container_properties = <<CONTAINER_PROPERTIES
{
    "command": ["ls", "-la"],
    "image": "busybox",
    "memory": 512,
    "vcpus": 1,
    "volumes": [
      {
        "host": {
          "sourcePath": "/tmp"
        },
        "name": "tmp"
      }
    ],
    "environment": [
        {"name": "VARNAME", "value": "VARVAL"}
    ],
    "mountPoints": [
        {
          "sourceVolume": "tmp",
          "containerPath": "/tmp",
          "readOnly": false
        }
    ],
    "ulimits": [
      {
        "hardLimit": 1024,
        "name": "nofile",
        "softLimit": 1024
      }
    ]
}
CONTAINER_PROPERTIES
}
`, rName, param, retries, timeout)
}

func testAccJobDefinitionConfig_containerPropertiesAdvancedUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name                 = %[1]q
  type                 = "container"
  container_properties = <<CONTAINER_PROPERTIES
{
    "command": ["ls", "-la"],
    "image": "busybox",
    "memory": 1024,
    "vcpus": 1,
    "volumes": [
      {
        "host": {
          "sourcePath": "/tmp"
        },
        "name": "tmp"
      }
    ],
    "environment": [
        {"name": "VARNAME", "value": "VARVAL"}
    ],
    "mountPoints": [
        {
          "sourceVolume": "tmp",
          "containerPath": "/tmp",
          "readOnly": false
        }
    ],
    "ulimits": [
      {
        "hardLimit": 1024,
        "name": "nofile",
        "softLimit": 1024
      }
    ]
}
CONTAINER_PROPERTIES
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

  node_properties = jsonencode({
    mainNode = 0
    numNodes = 1
    nodeRangeProperties = [{
      targetNodes = "0:"
      eksProperties = {
        podProperties = {
          containers = [
            {
              name  = "test-eks-container-1"
              image = "public.ecr.aws/amazonlinux/amazonlinux = 2"
              command = [
                "sleep",
                "60"
              ]
              resources = {
                requests = {
                  memory = "1024Mi"
                  cpu    = "1"
                }
              }
              securityContext = {
                "runAsUser"              = 1000
                "runAsGroup"             = 3000
                "privileged"             = true
                "readOnlyRootFilesystem" = true
                "runAsNonRoot"           = true
              }
            }
          ]
        }
      }
    }]
  })
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

  node_properties = jsonencode({
    mainNode = 0
    numNodes = 1
    nodeRangeProperties = [{
      targetNodes = "0:"
      ecsProperties = {
        taskProperties = [{
          containers = [{
            image      = "public.ecr.aws/amazonlinux/amazonlinux:1"
            command    = ["sleep", "60"]
            name       = "container_a"
            privileged = false
            resourceRequirements = [{
              value = "1"
              type  = "VCPU"
              },
              {
                value = "2048"
                type  = "MEMORY"
            }]
            },
            {
              image   = "public.ecr.aws/amazonlinux/amazonlinux:1"
              command = ["sleep", "360"]
              name    = "container_b"
              resourceRequirements = [{
                value = "1"
                type  = "VCPU"
                },
                {
                  value = "2048"
                  type  = "MEMORY"
              }]
          }]
        }]
      }
    }]
  })
}
`, rName)
}

func testAccJobDefinitionConfig_nodePropertiesECS_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "multinode"
  retry_strategy {
    attempts = 1
  }

  node_properties = jsonencode({
    mainNode = 0
    numNodes = 1
    nodeRangeProperties = [{
      targetNodes = "0:"
      ecsProperties = {
        taskProperties = [{
          containers = [{
            image      = "public.ecr.aws/amazonlinux/amazonlinux:1"
            command    = ["sleep", "60"]
            name       = "container_a"
            privileged = false
            resourceRequirements = [{
              value = "1"
              type  = "VCPU"
              },
              {
                value = "4096"
                type  = "MEMORY"
            }]
            },
            {
              image   = "public.ecr.aws/amazonlinux/amazonlinux:1"
              command = ["sleep", "360"]
              name    = "container_b"
              resourceRequirements = [{
                value = "1"
                type  = "VCPU"
                },
                {
                  value = "2048"
                  type  = "MEMORY"
              }]
          }]
        }]
      }
    }]
  })
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
  container_properties = jsonencode({
    command = ["ls", "%[2]s"],
    image   = "busybox"

    resourceRequirements = [
      {
        type  = "VCPU"
        value = "1"
      },
      {
        type  = "MEMORY"
        value = "512"
      }
    ]

    volumes = [
      {
        host = {
          sourcePath = "/tmp"
        }
        name = "tmp"
      }
    ]

    environment = [
      {
        name  = "VARNAME"
        value = "VARVAL"
      }
    ]

    mountPoints = [
      {
        sourceVolume  = "tmp"
        containerPath = "/tmp"
        readOnly      = false
      }
    ]

    ulimits = [
      {
        hardLimit = 1024
        name      = "nofile"
        softLimit = 1024
      }
    ]
  })
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

  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
}
`, rName)
}

func testAccJobDefinitionConfig_capabilitiesFargateContainerPropertiesDefaults(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_service_principal" "ecs_tasks" {
  service_name = "ecs-tasks"
}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = [data.aws_service_principal.ecs_tasks.name]
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

  container_properties = <<CONTAINER_PROPERTIES
{
  "image": "busybox",
  "resourceRequirements": [
    {"type": "MEMORY", "value": "512"},
    {"type": "VCPU", "value": "0.25"}
  ],
  "executionRoleArn": "${aws_iam_role.ecs_task_execution_role.arn}"
}
CONTAINER_PROPERTIES
}
`, rName)
}

func testAccJobDefinitionConfig_capabilitiesFargate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_service_principal" "ecs_tasks" {
  service_name = "ecs-tasks"
}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = [data.aws_service_principal.ecs_tasks.name]
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

  container_properties = <<CONTAINER_PROPERTIES
{
  "command": ["echo", "test"],
  "image": "busybox",
  "fargatePlatformConfiguration": {
    "platformVersion": "LATEST"
  },
  "networkConfiguration": {
    "assignPublicIp": "DISABLED"
  },
  "resourceRequirements": [
    {"type": "VCPU", "value": "0.25"},
    {"type": "MEMORY", "value": "512"}
  ],
  "executionRoleArn": "${aws_iam_role.ecs_task_execution_role.arn}"
}
CONTAINER_PROPERTIES
}
`, rName)
}

func testAccJobDefinitionConfig_propagateTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
  name = %[1]q
  type = "container"

  propagate_tags = true
}
`, rName)
}

func testAccJobDefinitionConfig_containerProperties_emptyField(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
    environment = [
      {
        name  = "EMPTY"
        value = ""
      },
      {
        name  = "VALUE"
        value = "value"
      }
    ]
  })
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

  node_properties = jsonencode({
    mainNode = 0
    nodeRangeProperties = [
      {
        container = {
          command = ["ls", "-la"]
          image   = "busybox"
          memory  = 128
          vcpus   = 1
        }
        targetNodes = "0:"
      },
      {
        container = {
          command = ["echo", "test"]
          image   = "busybox"
          memory  = 128
          vcpus   = 1
        }
        targetNodes = "1:"
      }
    ]
    numNodes = 2
  })
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

  node_properties = jsonencode({
    mainNode = 1
    nodeRangeProperties = [
      {
        container = {
          "command" : ["ls", "-la"],
          "image" : "busybox",
          "memory" : 512,
          "vcpus" : 1,
          "volumes" : [
            {
              "host" : {
                "sourcePath" : "/tmp"
              },
              "name" : "tmp"
            }
          ],
          "environment" : [
            { "name" : "VARNAME", "value" : "VARVAL" }
          ],
          "mountPoints" : [
            {
              "sourceVolume" : "tmp",
              "containerPath" : "/tmp",
              "readOnly" : false
            }
          ],
          "ulimits" : [
            {
              "hardLimit" : 1024,
              "name" : "nofile",
              "softLimit" : 1024
            }
          ]
        }
        targetNodes = "0:"
      },
      {
        container = {
          command              = ["echo", "test"]
          environment          = []
          image                = "busybox"
          memory               = 128
          mountPoints          = []
          resourceRequirements = []
          secrets              = []
          ulimits              = []
          vcpus                = 1
          volumes              = []
        }
        targetNodes = "1:"
      }
    ]
    numNodes = 4
  })
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

  node_properties = jsonencode({
    mainNode = 1
    nodeRangeProperties = [
      {
        container = {
          "command" : ["ls", "-la"],
          "image" : "busybox",
          "memory" : 512,
          "vcpus" : 1
        }
        targetNodes = "0:"
      },
      {
        container = {
          command     = ["echo", "test"]
          environment = []
          image       = "busybox"
          memory      = 128
          mountPoints = []
          ulimits     = []
          vcpus       = 1
          volumes     = []
        }
        targetNodes = "1:"
      }
    ]
    numNodes = 4
  })
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
      image_pull_secret {
        name = "chihiro"
      }
      image_pull_secret {
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
      image_pull_secret {
        name = "chihiro"
      }
      image_pull_secret {
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
          allow_privilege_escalation = true
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

  node_properties = jsonencode({
    mainNode = 1
    nodeRangeProperties = [
      {
        container = {
          "command" : ["ls", "-la"],
          "image" : "busybox",
          "memory" : 512,
          "vcpus" : 1
        }
        targetNodes = "0:"
      },
      {
        container = {
          command     = ["echo", "test"]
          environment = []
          image       = "busybox"
          memory      = 128
          mountPoints = []
          ulimits     = []
          vcpus       = 1
          volumes     = []
        }
        targetNodes = "1:"
      }
    ]
    numNodes = 4
  })
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

  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
}
`, rName)
}

func testAccJobDefinitionConfig_schedulingPriority(rName string, priority int) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
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
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
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
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
  retry_strategy {
  }
}
`, rName)
}

func testAccJobDefinitionConfig_ECSProperties_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_service_principal" "ecs_tasks" {
  service_name = "ecs-tasks"
}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = [data.aws_service_principal.ecs_tasks.name]
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

  ecs_properties = jsonencode({
    taskProperties = [
      {
        executionRoleArn = aws_iam_role.ecs_task_execution_role.arn
        containers = [
          {
            image   = "public.ecr.aws/amazonlinux/amazonlinux:1"
            command = ["sleep", "60"]
            dependsOn = [
              {
                containerName = "container_b"
                condition     = "COMPLETE"
              }
            ]
            secrets = [
              {
                name      = "TEST"
                valueFrom = "DUMMY"
              }
            ]
            environment = [
              {
                name  = "test"
                value = "Environment Variable"
              }
            ]
            essential = true
            logConfiguration = {
              logDriver = "awslogs"
              options = {
                "awslogs-group"         = %[1]q
                "awslogs-region"        = %[2]q
                "awslogs-stream-prefix" = "ecs"
              }
            }
            name                   = "container_a"
            privileged             = false
            readonlyRootFilesystem = false
            resourceRequirements = [
              {
                value = "1.0"
                type  = "VCPU"
              },
              {
                value = "2048"
                type  = "MEMORY"
              }
            ]
          },
          {
            image     = "public.ecr.aws/amazonlinux/amazonlinux:1"
            command   = ["sleep", "360"]
            name      = "container_b"
            essential = false
            resourceRequirements = [
              {
                value = "1.0"
                type  = "VCPU"
              },
              {
                value = "2048"
                type  = "MEMORY"
              }
            ]
          }
        ]
      }
    ]
  })
}
`, rName, acctest.Region())
}

func testAccJobDefinitionConfig_ECSProperties_updated(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_service_principal" "ecs_tasks" {
  service_name = "ecs-tasks"
}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = [data.aws_service_principal.ecs_tasks.name]
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

  ecs_properties = jsonencode({
    taskProperties = [
      {
        executionRoleArn = aws_iam_role.ecs_task_execution_role.arn
        containers = [
          {
            image   = "public.ecr.aws/amazonlinux/amazonlinux:1"
            command = ["sleep", "60"]
            dependsOn = [
              {
                containerName = "container_b"
                condition     = "COMPLETE"
              }
            ]
            secrets = [
              {
                name      = "TEST"
                valueFrom = "DUMMY"
              }
            ]
            environment = [
              {
                name  = "test 1"
                value = "Environment Variable 1"
              },
              {
                name  = "test 2"
                value = "Environment Variable 2"
              }
            ]
            essential = true
            logConfiguration = {
              logDriver = "awslogs"
              options = {
                "awslogs-group"         = %[1]q
                "awslogs-region"        = %[2]q
                "awslogs-stream-prefix" = "ecs"
              }
            }
            name                   = "container_a"
            privileged             = false
            readonlyRootFilesystem = false
            resourceRequirements = [
              {
                value = "1.0"
                type  = "VCPU"
              },
              {
                value = "2048"
                type  = "MEMORY"
              }
            ]
          },
          {
            image     = "public.ecr.aws/amazonlinux/amazonlinux:1"
            command   = ["sleep", "360"]
            name      = "container_b"
            essential = false
            resourceRequirements = [
              {
                value = "1.0"
                type  = "VCPU"
              },
              {
                value = "2048"
                type  = "MEMORY"
              }
            ]
          }
        ]
      }
    ]
  })
}
`, rName, acctest.Region())
}

func testAccJobDefinitionConfig_simpleWithTags(rName, command1, command2 string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties = jsonencode({
    command = [%[2]q, %[3]q]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
  name = %[1]q
  type = "container"

  tags = {
    Name = %[1]q
  }
}
`, rName, command1, command2)
}
