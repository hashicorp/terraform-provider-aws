---
subcategory: "Greengrassv2"
layout: "aws"
page_title: "AWS: aws_greengrassv2_component"
description: |-
  Creates and manages an AWS IoT Greengrassv2 Component Definition
---

# Resource: aws_greengrassv2_component

Provides a Greengrassv2 component Job resource.

## Example Usage

## Example Usage for Json inline_recipe

```terraform
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
```

## Example Usage for Yaml inline_recipe

```terraform
resource "aws_greengrassv2_component" "test" {
  tags = {
    Name = "tagValue"
  }
  inline_recipe          = <<EOF
---
RecipeFormatVersion: '2020-01-25'
ComponentName: "com.example.test.yaml"
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
```

## Example Usage for Lambda function

```terraform
resource "aws_greengrassv2_component" "default" {
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

    component_name    = "com.example.test.lambda"
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
  name = "greengrassv2component-test"

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
  function_name = "test-lambda"
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
```
## Argument Reference

The following arguments are supported:

* `inlineRecipe` – (Optional) The recipe to use to create the component. The recipe defines the component's metadata, parameters, dependencies, lifecycle, artifacts, and platform compatibility.You must specify either inlineRecipe or lambdaFunction.
* `lambdaFunction` – (Optional) The parameters to create a component from a Lambda function.You must specify either inlineRecipe or lambdaFunction.
* `tags` - (Optional) Key-value map of resource tags

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the Greengrassv2 Component.
* `id` - The ID of the Greengrassv2 Component.

## Import

IoT Greengrassv2 Component can be imported using the `arn`, e.g.

```
$ terraform import aws_greengrassv2_component.default <arn>
```
