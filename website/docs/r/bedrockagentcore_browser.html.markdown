---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_browser"
description: |-
  Manages an AWS Bedrock AgentCore Browser.
---

# Resource: aws_bedrockagentcore_browser

Manages an AWS Bedrock AgentCore Browser. Browser provides AI agents with web browsing capabilities, allowing them to navigate websites, extract information, and interact with web content in a controlled environment.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_browser" "example" {
  name        = "example-browser"
  description = "Browser for web data extraction"

  network_configuration {
    network_mode = "PUBLIC"
  }
}
```

### Browser with VPC Configuration

```terraform
resource "aws_bedrockagentcore_browser" "vpc_example" {
  name        = "vpc-browser"
  description = "Browser with VPC configuration"

  network_configuration {
    network_mode = "VPC"
    vpc_config {
      security_groups = ["sg-12345678"]
      subnets         = ["subnet-12345678", "subnet-87654321"]
    }
  }
}
```

### Browser with Execution Role and Recording

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "bedrock-agentcore-browser-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_s3_bucket" "recording" {
  bucket = "browser-recording-bucket"
}

resource "aws_bedrockagentcore_browser" "example" {
  name               = "example-browser"
  description        = "Browser with recording enabled"
  execution_role_arn = aws_iam_role.example.arn

  network_configuration {
    network_mode = "PUBLIC"
  }

  recording {
    enabled = true
    s3_location {
      bucket = aws_s3_bucket.recording.bucket
      prefix = "browser-sessions/"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the browser.
* `network_configuration` - (Required) Network configuration for the browser. See [`network_configuration`](#network_configuration) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the browser.
* `execution_role_arn` - (Optional) ARN of the IAM role that the browser assumes for execution.
* `recording` - (Optional) Recording configuration for browser sessions. See [`recording`](#recording) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `network_configuration`

The `network_configuration` object supports the following:

* `network_mode` - (Required) Network mode for the browser. Valid values: `PUBLIC`, `VPC`.
* `vpc_config` - (Optional) VPC configuration when `network_mode` is `VPC`. See [`vpc_config`](#vpc_config) below.

### `vpc_config`

The `vpc_config` object supports the following:

* `security_groups` - (Required) Set of security group IDs for the VPC configuration.
* `subnets` - (Required) Set of subnet IDs for the VPC configuration.

### `recording`

The `recording` object supports the following:

* `enabled` - (Optional) Whether to enable recording for browser sessions. Defaults to `false`.
* `s3_location` - (Optional) S3 location where browser session recordings are stored. See [`s3_location`](#s3_location) below.

### `s3_location`

The `s3_location` object supports the following:

* `bucket` - (Required) Name of the S3 bucket where recordings are stored.
* `prefix` - (Required) S3 key prefix for recording files.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `browser_arn` - ARN of the Browser.
* `browser_id` - Unique identifier of the Browser.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Browser using the browser ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_browser.example
  id = "BROWSER1234567890"
}
```

Using `terraform import`, import Bedrock AgentCore Browser using the browser ID. For example:

```console
% terraform import aws_bedrockagentcore_browser.example BROWSER1234567890
```
