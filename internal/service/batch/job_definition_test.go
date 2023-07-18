// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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

func TestAccBatchJobDefinition_PlatformCapabilities_ec2(t *testing.T) {
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
				Config: testAccJobDefinitionConfig_capabilitiesEC2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
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
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_capabilitiesFargateContainerPropertiesDefaults(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_capabilitiesFargate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					testAccCheckJobDefinitionAttributes(&jd, &compare),
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

func TestAccBatchJobDefinition_updateForcesNewResource(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after batch.JobDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvanced(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &before),
					testAccCheckJobDefinitionAttributes(&before, nil),
				),
			},
			{
				Config: testAccJobDefinitionConfig_containerPropertiesAdvancedUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &after),
					testAccCheckJobDefinitionRecreated(t, &before, &after),
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

func TestAccBatchJobDefinition_propagateTags(t *testing.T) {
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
				Config: testAccJobDefinitionConfig_propagateTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName, &jd),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
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
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "true"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
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
		ErrorCheck:               acctest.ErrorCheck(t, batch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionConfig_containerProperties_emptyField(rName),
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
			}
		}
		return nil
	}
}

func testAccCheckJobDefinitionRecreated(t *testing.T, before, after *batch.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.Int64Value(before.Revision) == aws.Int64Value(after.Revision) {
			t.Fatalf("Expected change of JobDefinition Revisions, but both were %d", aws.Int64Value(before.Revision))
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

func testAccJobDefinitionConfig_containerPropertiesAdvanced(rName string) string {
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
