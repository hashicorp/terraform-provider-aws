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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
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
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", acctest.Ct0),
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
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:1`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.attempts", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "timeout.0.attempt_duration_seconds", "120"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", acctest.Ct2),
				),
			},
			{
				Config: testAccJobDefinitionConfig_attributes(rName, 2, true, 4, 120, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:2`, rName))),
					testAccCheckJobDefinitionPreviousRegistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct2),
				),
			},
			{
				Config: testAccJobDefinitionConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:3`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", acctest.Ct0),
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.attempts", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "timeout.0.attempt_duration_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfbatch.ResourceJobDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBatchJobDefinition_PlatformCapabilities_ec2(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", acctest.Ct0),
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
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(command)", acctest.Ct0),
					acctest.CheckResourceAttrJMESPair(resourceName, "container_properties", "executionRoleArn", "aws_iam_role.ecs_task_execution_role", names.AttrARN),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "fargatePlatformConfiguration.platformVersion", "LATEST"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(resourceRequirements)", acctest.Ct2),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='VCPU'].value | [0]", "0.25"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='MEMORY'].value | [0]", "512"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", acctest.Ct0),
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

func TestAccBatchJobDefinition_PlatformCapabilities_fargate(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.CheckResourceAttrJMESPair(resourceName, "container_properties", "executionRoleArn", "aws_iam_role.ecs_task_execution_role", names.AttrARN),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "fargatePlatformConfiguration.platformVersion", "LATEST"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "networkConfiguration.assignPublicIp", "DISABLED"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(resourceRequirements)", acctest.Ct2),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='VCPU'].value | [0]", "0.25"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='MEMORY'].value | [0]", "512"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", acctest.Ct0),
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
	var jd batch.JobDefinition
	compare := batch.JobDefinition{
		Parameters: map[string]*string{
			"param1": aws.String("val1"),
			"param2": aws.String("val2"),
		},
		RetryStrategy: &batch.RetryStrategy{
			Attempts: aws.Int64(1),
			EvaluateOnExit: []*batch.EvaluateOnExit{
				{Action: aws.String(strings.ToLower(batch.RetryActionRetry)), OnStatusReason: aws.String("Host EC2*")},
				{Action: aws.String(strings.ToLower(batch.RetryActionExit)), OnReason: aws.String("*")},
			},
		},
		Timeout: &batch.JobTimeout{
			AttemptDurationSeconds: aws.Int64(60),
		},
		ContainerProperties: &batch.ContainerProperties{
			Command: []*string{aws.String("ls"), aws.String("-la")},
			Environment: []*batch.KeyValuePair{
				{Name: aws.String("VARNAME"), Value: aws.String("VARVAL")},
			},
			Image:  aws.String("busybox"),
			Memory: aws.Int64(512),
			MountPoints: []*batch.MountPoint{
				{ContainerPath: aws.String("/tmp"), ReadOnly: aws.Bool(false), SourceVolume: aws.String("tmp")},
			},
			ResourceRequirements: []*batch.ResourceRequirement{},
			Secrets:              []*batch.Secret{},
			Ulimits: []*batch.Ulimit{
				{HardLimit: aws.Int64(1024), Name: aws.String("nofile"), SoftLimit: aws.Int64(1024)},
			},
			Vcpus: aws.Int64(1),
			Volumes: []*batch.Volume{
				{
					Host: &batch.Host{SourcePath: aws.String("/tmp")},
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					testAccCheckJobDefinitionAttributes(&jd, &compare),
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
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 1, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct2),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 1, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct2),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 3, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct3),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 3, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct3),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName, "val3", 3, 120),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct4),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "5"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_ContainerProperties_minorUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:1`, rName))),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-lah"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:2`, rName))),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct2),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerProperties(rName, "-hal"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:3`, rName))),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_propagateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_ContainerProperties_EmptyField(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(environment)", acctest.Ct1),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "environment[?name=='VALUE'].value | [0]", names.AttrValue),
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

func TestAccBatchJobDefinition_NodeProperties_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_NodeProperties(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", acctest.Ct0),
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
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
				ImportStateVerifyIgnore: []string{
					"deregister_on_new_revision",
				},
			},
			{
				Config: testAccJobDefinitionConfig_nodePropertiesAdvancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_EKSProperties_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.0.image_pull_policy", ""),
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
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
				),
			},
			{
				Config: testAccJobDefinitionConfig_EKSProperties_advancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", acctest.Ct1),
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
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.containers.0.image_pull_policy", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "container"),
					resource.TestCheckResourceAttr(resourceName, "eks_properties.0.pod_properties.0.image_pull_secret.#", acctest.Ct2),
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
	var jd batch.JobDefinition
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
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", acctest.Ct2),
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

func testAccCheckJobDefinitionExists(ctx context.Context, n string, jd *batch.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Job Queue ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn(ctx)

		jobDefinition, err := tfbatch.FindJobDefinitionByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*jd = *jobDefinition

		return nil
	}
}

func testAccCheckJobDefinitionPreviousRegistered(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Job Queue ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn(ctx)

		previousARN := parseJobDefinitionPreviousARN(rs.Primary.ID)

		jobDefinition, err := tfbatch.FindJobDefinitionByARN(ctx, conn, previousARN)

		if err != nil {
			return err
		}

		if aws.StringValue(jobDefinition.Status) != "ACTIVE" {
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

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Job Queue ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn(ctx)

		previousARN := parseJobDefinitionPreviousARN(rs.Primary.ID)

		_, err := tfbatch.FindJobDefinitionByARN(ctx, conn, previousARN)

		// FindJobDefinitionByARN returns an error if the job is INACTIVE (deregistered)
		if err != nil {
			if err.Error() == "INACTIVE" {
				return nil
			}
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

func testAccCheckJobDefinitionAttributes(jd *batch.JobDefinition, compare *batch.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_definition" {
				continue
			}
			if aws.StringValue(jd.JobDefinitionArn) != rs.Primary.Attributes[names.AttrARN] {
				return fmt.Errorf("Bad Job Definition ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes[names.AttrARN], aws.StringValue(jd.JobDefinitionArn))
			}
			if compare != nil {
				if compare.Parameters != nil && !reflect.DeepEqual(compare.Parameters, jd.Parameters) {
					return fmt.Errorf("Bad Job Definition Params\n\t expected: %v\n\tgot: %v\n", compare.Parameters, jd.Parameters)
				}
				if compare.RetryStrategy != nil && aws.Int64Value(compare.RetryStrategy.Attempts) != aws.Int64Value(jd.RetryStrategy.Attempts) {
					return fmt.Errorf("Bad Job Definition Retry Strategy\n\t expected: %d\n\tgot: %d\n", aws.Int64Value(compare.RetryStrategy.Attempts), aws.Int64Value(jd.RetryStrategy.Attempts))
				}
				if compare.RetryStrategy != nil && !reflect.DeepEqual(compare.RetryStrategy.EvaluateOnExit, jd.RetryStrategy.EvaluateOnExit) {
					return fmt.Errorf("Bad Job Definition Retry Strategy\n\t expected: %v\n\tgot: %v\n", compare.RetryStrategy.EvaluateOnExit, jd.RetryStrategy.EvaluateOnExit)
				}
				if compare.ContainerProperties != nil && compare.ContainerProperties.Command != nil && !reflect.DeepEqual(compare.ContainerProperties, jd.ContainerProperties) {
					return fmt.Errorf("Bad Job Definition Container Properties\n\t expected: %s\n\tgot: %s\n", compare.ContainerProperties, jd.ContainerProperties)
				}
				if compare.NodeProperties != nil && compare.NodeProperties.NumNodes != nil && !reflect.DeepEqual(compare.NodeProperties, jd.NodeProperties) {
					return fmt.Errorf("Bad Job Definition Node Properties\n\t expected: %s\n\tgot: %s\n", compare.NodeProperties, jd.NodeProperties)
				}
			}
		}
		return nil
	}
}

func testAccCheckJobDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_definition" {
				continue
			}

			re := regexache.MustCompile(`job-definition/(.*?):`)
			m := re.FindStringSubmatch(rs.Primary.ID)
			if len(m) < 1 {
				continue
			}

			name := m[1]

			jds, err := tfbatch.ListActiveJobDefinitionByName(ctx, conn, name)

			if count := len(jds); count == 0 {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Batch Job Definition %s has revisions that still exist", name)
		}
		return nil
	}
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

func testAccJobDefinitionConfig_containerProperties(rName, subcommand string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

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

func testAccJobDefinitionConfig_name(rName string) string {
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

func testAccJobDefinitionConfig_NodeProperties(rName string) string {
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
