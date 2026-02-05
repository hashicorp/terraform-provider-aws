---
subcategory: ""
layout: "aws"
page_title: "AWS: iam_statements_distribute"
description: |-
  Distributes statements from large IAM policy documents across multiple smaller policies that comply with AWS size limits.
---

# Function: iam_statements_distribute

Distributes statements from large IAM policy documents across multiple smaller policies that comply with AWS size limits.

This function helps manage policies that exceed AWS policy-specific size constraints by intelligently distributing statements across multiple valid policy documents. Each output policy is a complete, standalone IAM policy document that can be used independently.

See the [AWS IAM documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements.html) for additional information on IAM policy structure and [AWS service quotas](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-quotas.html) for policy size limits.

## Example Usage

### Basic Statement Distribution

```terraform
# Distribute statements from a large policy for inline user policy usage (2048 byte limit)
locals {
  large_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject"
        ]
        Resource = "arn:aws:s3:::my-bucket/*"
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem"
        ]
        Resource = "arn:aws:dynamodb:us-east-1:123456789012:table/MyTable"
      }
      # ... many more statements that exceed 2048 bytes
    ]
  })
}

output "distributed_policies" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "inline-user")
}

# Result:
# {
#   "policies": [
#     "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"s3:GetObject\",\"s3:PutObject\",\"s3:DeleteObject\"],\"Resource\":\"arn:aws:s3:::my-bucket/*\"}]}",
#     "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"dynamodb:GetItem\",\"dynamodb:PutItem\",\"dynamodb:UpdateItem\",\"dynamodb:DeleteItem\"],\"Resource\":\"arn:aws:dynamodb:us-east-1:123456789012:table/MyTable\"}]}"
#   ],
#   "metadata": {
#     "original_size": 2150,
#     "average_size": 1052,
#     "largest_policy": 1100,
#     "smallest_policy": 1005,
#     "total_size_reduction": 45
#   }
# }
```

### Complete Example: Automatic Statement Distribution for IAM Role

This example shows a large policy and automatically distribute its statements across multiple policies attached to an IAM role, without worrying about manual policy management:

```terraform
# Define a comprehensive policy that exceeds inline policy size limits
locals {
  comprehensive_policy = {
    Version = "2012-10-17"
    Statement = [
      # S3 permissions
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:GetObjectVersion",
          "s3:PutObjectAcl",
          "s3:GetObjectAcl",
          "s3:RestoreObject",
          "s3:ListBucket",
          "s3:ListBucketVersions"
        ]
        Resource = [
          "arn:aws:s3:::my-app-bucket",
          "arn:aws:s3:::my-app-bucket/*",
          "arn:aws:s3:::my-backup-bucket",
          "arn:aws:s3:::my-backup-bucket/*"
        ]
      },
      # DynamoDB permissions
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:BatchGetItem",
          "dynamodb:BatchWriteItem",
          "dynamodb:DescribeTable"
        ]
        Resource = [
          "arn:aws:dynamodb:us-east-1:123456789012:table/Users",
          "arn:aws:dynamodb:us-east-1:123456789012:table/Orders",
          "arn:aws:dynamodb:us-east-1:123456789012:table/Products",
          "arn:aws:dynamodb:us-east-1:123456789012:table/Sessions"
        ]
      },
      # Lambda permissions
      {
        Effect = "Allow"
        Action = [
          "lambda:InvokeFunction",
          "lambda:InvokeAsync",
          "lambda:GetFunction",
          "lambda:ListFunctions"
        ]
        Resource = [
          "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
          "arn:aws:lambda:us-east-1:123456789012:function:SendNotification",
          "arn:aws:lambda:us-east-1:123456789012:function:GenerateReport"
        ]
      },
      # SQS permissions
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl"
        ]
        Resource = [
          "arn:aws:sqs:us-east-1:123456789012:order-queue",
          "arn:aws:sqs:us-east-1:123456789012:notification-queue",
          "arn:aws:sqs:us-east-1:123456789012:deadletter-queue"
        ]
      },
      # CloudWatch permissions
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogGroups",
          "logs:DescribeLogStreams"
        ]
        Resource = "arn:aws:logs:us-east-1:123456789012:*"
      },
      # Additional permissions that would make this policy exceed 2048 bytes
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          "arn:aws:secretsmanager:us-east-1:123456789012:secret:app/database-*",
          "arn:aws:secretsmanager:us-east-1:123456789012:secret:app/api-keys-*"
        ]
      }
    ]
  }

  # Convert to JSON and automatically distribute statements for inline user policies
  policy_json        = jsonencode(local.comprehensive_policy)
  distribution_result = provider::aws::iam_statements_distribute(local.policy_json, "inline-user")
}

# Create the IAM role
resource "aws_iam_role" "example" {
  name = "my-application-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Environment = "production"
    Application = "my-app"
  }
}

# Automatically create multiple inline policies from the distribution result
resource "aws_iam_role_policy" "example" {
  count = length(local.distribution_result.policies)

  name   = "app-policy-${count.index + 1}"
  role   = aws_iam_role.example.id
  policy = local.distribution_result.policies[count.index]
}

# Optional: Create an instance profile if needed for EC2
resource "aws_iam_instance_profile" "example" {
  name = "my-application-profile"
  role = aws_iam_role.example.name
}
```

This example demonstrates the key benefit: **the a comprehensive policy can be written and the function automatically handles the statement distribution and attachment process**.

### Different Policy Types

```terraform
# For customer-managed policies (6144 byte limit)
output "customer_managed_distribution" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "customer-managed")
}

# For inline policies attached to roles (10240 byte limit)
output "inline_role_distribution" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "inline-role")
}

# For inline policies attached to groups (5120 byte limit)
output "inline_group_distribution" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "inline-group")
}

# For Service Control Policies (5120 byte limit)
output "scp_distribution" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "service-control-policy")
}

# For permissions boundary policies (6144 byte limit)
output "permissions_boundary_distribution" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "permissions-boundary")
}

# Legacy aliases still work
output "legacy_managed_distribution" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "managed")
}
```

### Handling Policies with IDs

```terraform
locals {
  policy_with_id = jsonencode({
    Version = "2012-10-17"
    Id      = "MyPolicyId"
    Statement = [
      # ... statements that exceed size limit
    ]
  })
}

output "distribution_with_ids" {
  value = provider::aws::iam_statements_distribute(local.policy_with_id, "inline-user")
}

# Result policies will have unique IDs: "MyPolicyId-split-1", "MyPolicyId-split-2", etc.
```

### Policy Size Optimization

Use the metadata to understand policy size characteristics and optimize policy structure:

```terraform
locals {
  large_policy = jsonencode({
    Version = "2012-10-17"
    Id      = "MyPolicy"
    Statement = [
      # ... statements that exceed size limit
    ]
  })    
  distribution_result = provider::aws::iam_statements_distribute(local.large_policy, "inline-user")
}

output "size_analysis" {
  value = {
    original_size_kb    = local.distribution_result.metadata.original_size / 1024
    average_policy_size = local.distribution_result.metadata.average_size
    size_efficiency     = local.distribution_result.metadata.average_size / local.distribution_result.metadata.original_size * length(local.distribution_result.policies)
  }
}
```

## Signature

```text
iam_statements_distribute(policy_json string, policy_type string) object
```

## Argument Reference

1. `policy_json` (Required) IAM policy JSON document to distribute statements from. Must be a valid IAM policy with required fields: `Version` and `Statement`.
2. `policy_type` (Required)  AWS policy type for size limits. Valid values:
   - `"customer-managed"` - 6144 bytes (for customer-managed policies)
   - `"inline-user"` - 2048 bytes (for inline policies attached to users)
   - `"inline-role"` - 10240 bytes (for inline policies attached to roles)
   - `"inline-group"` - 5120 bytes (for inline policies attached to groups)
   - `"service-control-policy"` - 5120 bytes (for AWS Organizations SCPs)

## Attribute Reference

The function returns an object with the following attributes:

- `policies` - List of complete IAM policy JSON documents. Each policy is standalone and can be used independently.
- `metadata` - Additional information about the distribution operation:
    - `original_size` - Size of the original policy in bytes
    - `average_size` - Average size of distributed policies in bytes
    - `largest_policy` - Size of the largest distributed policy in bytes
    - `smallest_policy` - Size of the smallest distributed policy in bytes
    - `total_size_reduction` - Total size difference in bytes between the original policy and all distributed policies combined. May be negative if distribution adds overhead.

## Behavior

### Size Limits and Validation

The function enforces AWS policy-specific size limits:

- **Customer-managed policies**: 6144 bytes maximum
- **Inline policies (User)**: 2048 bytes maximum
- **Inline policies (Role)**: 10240 bytes maximum  
- **Inline policies (Group)**: 5120 bytes maximum
- **Service Control Policies (SCP)**: 5120 bytes maximum

If the original policy is already within the specified size limit, it is returned as-is (but reformatted for consistency).

### Statement Integrity

- Individual policy statements are never split or modified
- Each statement remains intact and is moved as a complete unit
- If any single statement exceeds the size limit, the function returns an error

### Policy Structure Preservation

- All distributed policies maintain the same `Version` as the original
- If the original policy has an `Id` field, distributed policies get unique IDs: `"original-id-split-1"`, `"original-id-split-2"`, etc.
- Each distributed policy is a complete, valid IAM policy document

### Algorithm

The function uses an accurate size-based bin-packing algorithm:

1. Validates the input policy structure and policy type
2. Checks if any individual statement is too large to fit
3. Uses greedy bin-packing to optimally distribute statements across multiple policies
4. Validates that each resulting policy is within the size limit
5. Generates metadata about the distribution operation

## Error Conditions

The function will return an error in the following cases:

- **Invalid JSON**: The `policy_json` parameter is not valid JSON
- **Invalid policy structure**: Missing required fields (`Version`, `Statement`) or invalid field values
- **Unsupported version**: Policy version is not `"2012-10-17"` or `"2008-10-17"`
- **Invalid policy type**: Policy type is not one of the supported values
- **Missing policy type**: The `policy_type` parameter is required and cannot be empty
- **Statement too large**: Any individual statement exceeds the size limit for the specified policy type
- **Base policy too large**: The base policy structure (without statements) exceeds the size limit


