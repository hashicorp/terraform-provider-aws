---
page_title: "aws_bedrock_knowledge_base_logging_configuration Resource"
category: "Resources"
---

# aws_bedrock_knowledge_base_logging_configuration

Creates and manages a Bedrock Knowledge Base Logging Configuration in AWS.

## Example Usage

```hcl
resource "aws_bedrock_knowledge_base_logging_configuration" "example" {
  knowledge_base_id = "kb-example-id"

  logging_config {
    embedding_data_delivery_enabled = true

    cloudwatch_config {
      log_group_name = aws_cloudwatch_log_group.example.name
      role_arn       = aws_iam_role.example.arn
    }
  }

  tags = {
    Environment = "Production"
    Project     = "BedrockKnowledgeBase"
  }
}

resource "aws_cloudwatch_log_group" "example" {
  name = "/aws/vendedlogs/bedrock/knowledge-base/example"
}

resource "aws_iam_role" "example" {
  name = "bedrock_logging_role_example"

  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["bedrock.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_policy" "example" {
  name   = "bedrock_logging_policy_example"
  policy = data.aws_iam_policy_document.logging_policy.json
}

data "aws_iam_policy_document" "logging_policy" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      "${aws_cloudwatch_log_group.example.arn}:*",
    ]
  }
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = aws_iam_policy.example.arn
}
```

## Arguments

| Argument                                          | Description                                                                                                | Required | Type   |
| ------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | -------- | ------ |
| `knowledge_base_id`                               | The identifier of the Bedrock Knowledge Base to which the logging configuration will be applied.           | Yes      | String |
| `logging_config`                                  | A block that defines the logging configuration settings.                                                   | Yes      | Block  |
| `logging_config.embedding_data_delivery_enabled`  | Enables or disables the delivery of embedding data.                                                        | Yes      | Bool   |
| `logging_config.cloudwatch_config`                | A block that specifies the CloudWatch Logs configuration for logging.                                      | Yes      | Block  |
| `logging_config.cloudwatch_config.log_group_name` | The name of the CloudWatch Logs log group where Bedrock will send logs.                                    | Yes      | String |
| `logging_config.cloudwatch_config.role_arn`       | The ARN of the IAM role that Bedrock will assume to write logs to the specified CloudWatch Logs log group. | Yes      | String |
| `tags`                                            | A mapping of tags to assign to the resource.                                                               | No       | Map    |

## Attributes

| Attribute                                         | Description                                                 |
| ------------------------------------------------- | ----------------------------------------------------------- |
| `id`                                              | The ID of the Bedrock Knowledge Base Logging Configuration. |
| `knowledge_base_id`                               | The ID of the associated Bedrock Knowledge Base.            |
| `tags`                                            | The tags assigned to the resource.                          |
| `logging_config`                                  | Detailed logging configuration settings.                    |
| `logging_config.embedding_data_delivery_enabled`  | Indicates if embedding data delivery is enabled.            |
| `logging_config.cloudwatch_config.log_group_name` | The name of the CloudWatch Logs log group.                  |
| `logging_config.cloudwatch_config.role_arn`       | The ARN of the IAM role used for logging.                   |

## Import

Bedrock Knowledge Base Logging Configuration resources can be imported using their `knowledge_base_id`.

```sh
terraform import aws_bedrock_knowledge_base_logging_configuration.example kb-example-id
```
