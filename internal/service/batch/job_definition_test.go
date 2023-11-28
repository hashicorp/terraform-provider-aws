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
)

func TestAccBatchJobDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
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
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_attributes(rName, 2, true, 3, 120, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "true"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.attempts", "3"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "timeout.0.attempt_duration_seconds", "120"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_attributes(rName, 2, true, 4, 120, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					testAccCheckJobDefinitionPreviousRegistered(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					testAccCheckJobDefinitionPreviousDeregistered(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
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
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.attempts", "1"),
					resource.TestCheckResourceAttr(resourceName, "revision", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "timeout.0.attempt_duration_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_priority", "1"),
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
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
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

func TestAccBatchJobDefinition_ContainerProperties_fargateDefaults(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_ContainerProperties_defaultsFargate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(command)", "0"),
					acctest.CheckResourceAttrJMESPair(resourceName, "container_properties", "executionRoleArn", "aws_iam_role.ecs_task_execution_role", "arn"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "fargatePlatformConfiguration.platformVersion", "LATEST"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(resourceRequirements)", "2"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='VCPU'].value | [0]", "0.25"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='MEMORY'].value | [0]", "512"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
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

func TestAccBatchJobDefinition_ContainerProperties_fargate(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_capabilitiesFargate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn_prefix", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s`, rName))),
					acctest.CheckResourceAttrJMESPair(resourceName, "container_properties", "executionRoleArn", "aws_iam_role.ecs_task_execution_role", "arn"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "fargatePlatformConfiguration.platformVersion", "LATEST"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "networkConfiguration.assignPublicIp", "DISABLED"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(resourceRequirements)", "2"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='VCPU'].value | [0]", "0.25"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "resourceRequirements[?type=='MEMORY'].value | [0]", "512"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
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

func TestAccBatchJobDefinition_ContainerProperties_EC2(t *testing.T) {
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
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_ContainerProperties_advanced(rName),
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
		},
	})
}

func TestAccBatchJobDefinition_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
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
				Config: testAccJobDefinitionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccJobDefinitionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_ContainerProperties_EC2EmptyField(t *testing.T) {
	ctx := acctest.Context(t)
	var jd batch.JobDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_ContainerProperties_emptyField(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "length(environment)", "1"),
					acctest.CheckResourceAttrJMES(resourceName, "container_properties", "environment[?name=='VALUE'].value | [0]", "value"),
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
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_NodeProperties(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
								"targetNodes": "1:"
							}
						],
						"numNodes": 2
					}`),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "multinode"),
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
				Config: testAccJobDefinitionConfig_NodeProperties_advancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
								"targetNodes": "1:"
							}
						],
						"numNodes": 4
					}`),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "type", "multinode"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinition_NodeProperties_createTypeContainerErr(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobDefinitionConfig_NodeProperties_createTypeContainerErr(rName),
				ExpectError: regexache.MustCompile("No `node_properties` can be specified when `type` is \"container\""),
			},
		},
	})
}

func TestAccBatchJobDefinition_ContainerProperties_createTypeMultiNodeErr(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobDefinitionConfig_ContainerProperties_createTypeMultiNodeErr(rName),
				ExpectError: regexache.MustCompile("No `container_properties` can be specified when `type` is \"multinode\""),
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

func testAccCheckJobDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_definition" {
				continue
			}

			re := regexache.MustCompile(`job-definition/(.*?):`)
			name := re.FindStringSubmatch(rs.Primary.ID)[1]

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
			if aws.StringValue(jd.JobDefinitionArn) != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Job Definition ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["arn"], aws.StringValue(jd.JobDefinitionArn))
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

func testAccJobDefinitionConfig_ContainerProperties_advanced(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  parameters = {
    param1 = "val1"
    param2 = "val2"
  }
  retry_strategy {
    attempts = 1
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
    attempt_duration_seconds = 60
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
`, rName)
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

func testAccJobDefinitionConfig_attributes(rName string, sp int, pt bool, rsa int, timeout int, dereg bool) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

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

func testAccJobDefinitionConfig_ContainerProperties_defaultsFargate(rName string) string {
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

func testAccJobDefinitionConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccJobDefinitionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccJobDefinitionConfig_ContainerProperties_emptyField(rName string) string {
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

func testAccJobDefinitionConfig_NodeProperties_advancedUpdate(rName string) string {
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

func testAccJobDefinitionConfig_NodeProperties_createTypeContainerErr(rName string) string {
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

func testAccJobDefinitionConfig_ContainerProperties_createTypeMultiNodeErr(rName string) string {
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
