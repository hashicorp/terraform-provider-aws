package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrassv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsGreengrassv2Component_JsonFormat(t *testing.T) {

	var component greengrassv2.DescribeComponentOutput
	resourceName := "aws_greengrassv2_component.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	recipe := fmt.Sprintf(`{"ComponentConfiguration":{"DefaultConfiguration":{"Message":"sample"}},"ComponentDescription":"sample","ComponentName":"com.example.test.json.%s","ComponentType":"aws.greengrass.generic","ComponentVersion":"1.0.0","Manifests":[{"Lifecycle":{"Install":{"Script":"ls"},"Run":{"Script":"ls -l"}},"Name":"Linux","Platform":{"os":"linux"}}],"RecipeFormatVersion":"2020-01-25"}`, rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, greengrassv2.EndpointsID),
		CheckDestroy: testAccCheckComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGreengrassv2ComponentConfigJson(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGreengrassv2ComponentExists(resourceName, &component),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "tags.Name", "tagValue"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "inline_recipe", recipe),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"inline_recipe"},
			},
		},
	})
}

func TestAccAwsGreengrassv2Component_YamlFormat(t *testing.T) {

	var component greengrassv2.DescribeComponentOutput
	resourceName := "aws_greengrassv2_component.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	recipe := fmt.Sprintf(`---
RecipeFormatVersion: '2020-01-25'
ComponentName: "com.example.test.yaml.%s"
ComponentVersion: 1.0.0
ComponentType: aws.greengrass.generic
ComponentDescription: sample
ComponentConfiguration:
  DefaultConfiguration:
    Message: sample
Manifests:
- Platform:
    os: linux
  Name: Linux
  Lifecycle:
    Install:
      Script: ls
    Run:
      Script: ls -l
Lifecycle: {}
`, rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, greengrassv2.EndpointsID),
		CheckDestroy: testAccCheckComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGreengrassv2ComponentConfigYaml(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGreengrassv2ComponentExists(resourceName, &component),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "tags.Name", "tagValue"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "inline_recipe", recipe),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"inline_recipe"},
			},
		},
	})
}

func TestAccAwsGreengrassv2Component_Lambda(t *testing.T) {

	var component greengrassv2.DescribeComponentOutput
	resourceName := "aws_greengrassv2_component.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	roleName := acctest.RandomWithPrefix("tf-acc-test-iam")
	functionName := acctest.RandomWithPrefix("tf-acc-test-lambda")
	componentName := fmt.Sprintf(`com.example.test.lambda.%s`, rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, greengrassv2.EndpointsID),
		CheckDestroy: testAccCheckComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGreengrassv2ComponentConfigLambda(rName, roleName, functionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGreengrassv2ComponentExists(resourceName, &component),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "tags.Name", "tagValue"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_dependencies.0.component_name", "aws.greengrass.test"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_dependencies.0.dependency_type", "SOFT"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_dependencies.0.version_requirement", "1.0.0"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_dependencies.1.component_name", "aws.greengrass.test2"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_dependencies.1.dependency_type", "HARD"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_dependencies.1.version_requirement", ">1.0.0"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_platforms.0.attributes.0.os", "Linux"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_platforms.0.attributes.0.architecture", "arm"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_platforms.0.name", "test Linux platform"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_platforms.1.attributes.0.os", "Windows"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_platforms.1.attributes.0.architecture", "x86"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_platforms.1.name", "test Windows platform"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.max_idle_time_in_seconds", "2147483647"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.max_instances_count", "1"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.max_queue_size", "1"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.status_timeout_in_seconds", "2147483647"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.timeout_in_seconds", "10"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.input_payload_encoding_type", "binary"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.exec_args.0", "hoge"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.exec_args.1", "fuga"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.environment_variables.hoge", "hoge"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.environment_variables.number", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_function.0.component_lambda_parameters.0.event_sources.0.topic"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.event_sources.0.type", "IOT_CORE"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_function.0.component_lambda_parameters.0.event_sources.1.topic"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.event_sources.1.type", "PUB_SUB"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.container_params.0.devices.0.add_group_owner", "true"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.container_params.0.devices.0.path", "/dev/stdout"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.container_params.0.devices.0.permission", "ro"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.container_params.0.memory_size_in_kb", "2048"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.container_params.0.mount_ro_sysfs", "true"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.container_params.0.volumes.0.add_group_owner", "true"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.container_params.0.volumes.0.destination_path", "/tmp"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.container_params.0.volumes.0.permission", "ro"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.container_params.0.volumes.0.source_path", "/tmp"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_lambda_parameters.0.linux_process_params.0.isolation_mode", "GreengrassContainer"),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_name", componentName),
					resource.TestCheckResourceAttr("aws_greengrassv2_component.test", "lambda_function.0.component_version", "1.0.0"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_function.0.lambda_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"lambda_function"},
			},
		},
	})
}

func testAccCheckGreengrassv2ComponentExists(n string, component *greengrassv2.DescribeComponentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).greengrassv2conn
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		describeComponentOpts := &greengrassv2.DescribeComponentInput{
			Arn: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeComponent(describeComponentOpts)
		if err != nil {
			return err
		}

		*component = *resp

		return nil
	}
}

func testAccCheckComponentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_greengrassv2_component" {
			continue
		}

		describeComponentOpts := &greengrassv2.DescribeComponentInput{
			Arn: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeComponent(describeComponentOpts)
		if err == nil {
			if len(aws.StringValue(resp.Arn)) > 0 {
				return fmt.Errorf("Greengrassv2 component still exists.")
			}
			return nil
		}
	}

	return nil
}

func testAccGreengrassv2ComponentConfigJson(rName string) string {
	return fmt.Sprintf(`
resource "aws_greengrassv2_component" "test" {
  tags = {
    Name  = "tagValue"
  }
  inline_recipe = jsonencode(
    {
      "RecipeFormatVersion" : "2020-01-25",
      "ComponentName" : "com.example.test.json.%s",
      "ComponentVersion" : "1.0.0",
      "ComponentType" : "aws.greengrass.generic",
      "ComponentDescription" : "sample",
      "ComponentConfiguration" : {
        "DefaultConfiguration" : {
          "Message" : "sample"
        }
      },
      "Manifests" : [
        {
          "Platform" : {
            "os" : "linux"
          },
          "Name" : "Linux",
          "Lifecycle" : {
            "Install" : {
              "Script" : "ls"
            },
            "Run" : {
              "Script" : "ls -l"
            }
          }
        }
      ],
    }
	)
}
`, rName)
}

func testAccGreengrassv2ComponentConfigYaml(rName string) string {
	return fmt.Sprintf(`
resource "aws_greengrassv2_component" "test" {
  tags = {
    Name = "tagValue"
  }
  inline_recipe          = <<EOF
---
RecipeFormatVersion: '2020-01-25'
ComponentName: "com.example.test.yaml.%s"
ComponentVersion: 1.0.0
ComponentType: aws.greengrass.generic
ComponentDescription: sample
ComponentConfiguration:
  DefaultConfiguration:
    Message: sample
Manifests:
- Platform:
    os: linux
  Name: Linux
  Lifecycle:
    Install:
      Script: ls
    Run:
      Script: ls -l
Lifecycle: {}
EOF
}
`, rName)
}

func testAccGreengrassv2ComponentConfigLambda(rName string, roleName string, functionName string) string {
	return fmt.Sprintf(`
resource "aws_greengrassv2_component" "test" {
  tags = {
    Name = "tagValue"
  }
  lambda_function {
    component_dependencies {
      component_name      = "aws.greengrass.test"
      dependency_type     = "SOFT"
      version_requirement = "1.0.0"
    }
		component_dependencies {
      component_name      = "aws.greengrass.test2"
      dependency_type     = "HARD"
      version_requirement = ">1.0.0"
    }
		component_platforms {
      attributes {
        os           = "Linux"
        architecture = "arm"
      }
      name = "test Linux platform"
    }
    component_platforms {
      attributes {
        os           = "Windows"
        architecture = "x86"
      }
      name = "test Windows platform"
    }
		component_lambda_parameters {
      max_idle_time_in_seconds    = 2147483647
      max_instances_count         = 1
      max_queue_size              = 1
      status_timeout_in_seconds   = 2147483647
      timeout_in_seconds          = 10
      input_payload_encoding_type = "binary"
      exec_args                   = ["hoge", "fuga"]
      environment_variables = {
        hoge   = "hoge"
        number = 1
      }
      event_sources {
        topic = aws_sns_topic.test.arn
        type  = "IOT_CORE"
      }
      event_sources {
        topic = aws_sns_topic.test2.arn
        type  = "PUB_SUB"
      }
			linux_process_params {
        container_params {
          devices {
            add_group_owner = true
            path            = "/dev/stdout"
            permission      = "ro"
          }
          memory_size_in_kb = 2048
          mount_ro_sysfs    = true
          volumes {
            add_group_owner  = true
            destination_path = "/tmp"
            permission       = "ro"
            source_path      = "/tmp"
          }
        }
        isolation_mode = "GreengrassContainer"
      }
    }

    component_name    = "com.example.test.lambda.%s"
    component_version = "1.0.0"
    lambda_arn        = aws_lambda_function.test.qualified_arn
  }
}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "hoge"
  role = aws_iam_role.iam_for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "greengrassv2component-test-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "test-lambda-%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
	publish       = "true"
}

resource "aws_sns_topic" "test" {
  name = "lambda-test-topic"
}

resource "aws_sns_topic" "test2" {
  name = "lambda-test-topic2"
}
`, rName, roleName, functionName)
}
