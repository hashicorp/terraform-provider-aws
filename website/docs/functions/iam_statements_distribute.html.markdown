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
# Distribute statements from a large policy for customer-managed policy usage (6144 byte limit)
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
      # ... many more statements that exceed 6144 bytes
    ]
  })
}

output "distributed_policies" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "customer-managed")
}

# Result:
# {
#   "policies": [
#     "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"s3:GetObject\",\"s3:PutObject\",\"s3:DeleteObject\"],\"Resource\":\"arn:aws:s3:::my-bucket/*\"}]}",
#     "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"dynamodb:GetItem\",\"dynamodb:PutItem\",\"dynamodb:UpdateItem\",\"dynamodb:DeleteItem\"],\"Resource\":\"arn:aws:dynamodb:us-east-1:123456789012:table/MyTable\"}]}"
#   ],
#   "metadata": {
#     "original_size": 6200,
#     "average_size": 3052,
#     "largest_policy": 3100,
#     "smallest_policy": 3005,
#     "total_size_reduction": 45
#   }
# }
```

### Complete Example: Automatic Statement Distribution for Customer-Managed Policies

This example shows a large policy and automatically distributes its statements across multiple customer-managed policies:

```terraform
# Define a comprehensive policy that exceeds customer-managed policy size limits
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
      # Additional permissions that would make this policy exceed 6144 bytes
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

  # Convert to JSON and automatically distribute statements for customer-managed policies
  policy_json         = jsonencode(local.comprehensive_policy)
  distribution_result = provider::aws::iam_statements_distribute(local.policy_json, "customer-managed")
}

# Create multiple customer-managed policies from the distribution result
resource "aws_iam_policy" "distributed_policies" {
  count = length(local.distribution_result.policies)

  name        = "my-application-policy-${count.index + 1}"
  description = "Distributed policy ${count.index + 1} for my application"
  policy      = local.distribution_result.policies[count.index]

  tags = {
    Environment = "production"
    Application = "my-app"
    PolicySet   = "distributed-permissions"
  }
}

# Create the IAM role and attach all distributed policies
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

# Attach all distributed policies to the role
resource "aws_iam_role_policy_attachment" "example" {
  count = length(aws_iam_policy.distributed_policies)

  role       = aws_iam_role.example.name
  policy_arn = aws_iam_policy.distributed_policies[count.index].arn
}

# Optional: Create an instance profile if needed for EC2
resource "aws_iam_instance_profile" "example" {
  name = "my-application-profile"
  role = aws_iam_role.example.name
}
```

This example demonstrates the key benefit: **a comprehensive policy can be written and the function automatically handles the statement distribution and policy creation process**.

### Different Policy Types

```terraform
# For customer-managed policies (6144 byte limit)
output "customer_managed_distribution" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "customer-managed")
}

# For Service Control Policies (5120 byte limit)
output "scp_distribution" {
  value = provider::aws::iam_statements_distribute(local.large_policy, "service-control-policy")
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
  distribution_result = provider::aws::iam_statements_distribute(local.large_policy, "customer-managed")
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

## Arguments

1. `policy_json` (String) IAM policy JSON document to distribute statements from. Must be a valid IAM policy with required fields: `Version` and `Statement`.
2. `policy_type` (String)  AWS policy type for size limits. Valid values:
   - `"customer-managed"` - 6144 bytes (for customer-managed policies)
   - `"service-control-policy"` - 5120 bytes (for AWS Organizations SCPs)

## Attribute Reference

The function returns an object with the following attributes:

- `policies` (List of Strings) List of complete IAM policy JSON documents. Each policy is standalone and can be used independently.
- `metadata` (Object) Additional information about the distribution operation:
    - `original_size` (Number) Size of the original policy in bytes
    - `average_size` (Number) Average size of distributed policies in bytes
    - `largest_policy` (Number) Size of the largest distributed policy in bytes
    - `smallest_policy` (Number) Size of the smallest distributed policy in bytes
    - `total_size_reduction` (Number) Total size difference in bytes between the original policy and all distributed policies combined. May be negative if distribution adds overhead.

## Behavior

### Size Limits and Validation

The function enforces AWS policy-specific size limits:

- **Customer-managed policies**: 6144 bytes maximum
- **Service Control Policies (SCP)**: 5120 bytes maximum

If the original policy is already within the specified size limit, it is returned as-is (but reformatted for consistency).

### Inline Policy Considerations

**Note**: This function does not support inline policy types (`inline-user`, `inline-role`, `inline-group`) because AWS enforces aggregate size limits for inline policies per entity:

- User inline policies: 2,048 characters total across all inline policies
- Role inline policies: 10,240 characters total across all inline policies  
- Group inline policies: 5,120 characters total across all inline policies

Since these are aggregate limits, distributing statements across multiple inline policies attached to the same entity does not solve size constraint issues. For large policies, use customer-managed policies instead, which have individual policy limits and no aggregate restrictions.

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
