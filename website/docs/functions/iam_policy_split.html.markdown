---
subcategory: ""
layout: "aws"
page_title: "AWS: iam_policy_split"
description: |-
  Splits large IAM policy documents into multiple smaller policies that comply with AWS size limits.
---

# Function: iam_policy_split

Splits large IAM policy documents into multiple smaller policies that comply with AWS size limits.

This function helps manage policies that exceed AWS service-specific size constraints by intelligently distributing statements across multiple valid policy documents. Each output policy is a complete, standalone IAM policy document that can be used independently.

See the [AWS IAM documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements.html) for additional information on IAM policy structure and [AWS service quotas](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-quotas.html) for policy size limits.

## Example Usage

### Basic Policy Splitting

```terraform
# Split a large policy for inline policy usage (2048 byte limit)
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

output "split_policies" {
  value = provider::aws::iam_policy_split(local.large_policy, "inline")
}

# Result:
# {
#   "policies": [
#     "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"s3:GetObject\",\"s3:PutObject\",\"s3:DeleteObject\"],\"Resource\":\"arn:aws:s3:::my-bucket/*\"}]}",
#     "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"dynamodb:GetItem\",\"dynamodb:PutItem\",\"dynamodb:UpdateItem\",\"dynamodb:DeleteItem\"],\"Resource\":\"arn:aws:dynamodb:us-east-1:123456789012:table/MyTable\"}]}"
#   ],
#   "count": 2,
#   "total_size_reduction": 45,
#   "metadata": {
#     "original_size": 2150,
#     "average_size": 1052,
#     "largest_policy": 1100,
#     "smallest_policy": 1005
#   }
# }
```

### Complete Example: Automatic Policy Splitting for IAM Role

This example shows how a Terraform practitioner can define a large policy and automatically split it into multiple policies attached to an IAM role, without worrying about manual policy splitting:

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

  # Convert to JSON and automatically split for inline policies
  policy_json = jsonencode(local.comprehensive_policy)
  split_result = provider::aws::iam_policy_split(local.policy_json, "inline")
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

# Automatically create multiple inline policies from the split result
resource "aws_iam_role_policy" "example" {
  count = local.split_result.count

  name   = "app-policy-${count.index + 1}"
  role   = aws_iam_role.example.id
  policy = local.split_result.policies[count.index]
}

# Optional: Create an instance profile if needed for EC2
resource "aws_iam_instance_profile" "example" {
  name = "my-application-profile"
  role = aws_iam_role.example.name
}
```

This example demonstrates the key benefit: **the Terraform practitioner writes one comprehensive policy and the function automatically handles the splitting and attachment process**.

### Different Service Types

```terraform
# For managed policies (6144 byte limit)
output "managed_split" {
  value = provider::aws::iam_policy_split(local.large_policy, "managed")
}

# For resource-based policies (10240 byte limit)
output "resource_based_split" {
  value = provider::aws::iam_policy_split(local.large_policy, "resource-based")
}

# Default to inline if service_type is omitted
output "default_split" {
  value = provider::aws::iam_policy_split(local.large_policy, "")
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

output "split_with_ids" {
  value = provider::aws::iam_policy_split(local.policy_with_id, "inline")
}

# Result policies will have unique IDs: "MyPolicyId-split-1", "MyPolicyId-split-2", etc.
```

## Signature

```text
iam_policy_split(policy_json string, service_type string) object
```

## Arguments

1. `policy_json` (String) IAM policy JSON document to split. Must be a valid IAM policy with required fields: `Version` and `Statement`.
1. `service_type` (String) AWS service type for size limits. Valid values:
   - `"inline"` - 2048 bytes (for inline policies attached to users, groups, or roles)
   - `"managed"` - 6144 bytes (for customer-managed policies)
   - `"resource-based"` - 10240 bytes (for resource-based policies like S3 bucket policies)
   - Empty string defaults to `"inline"`

## Return Value

The function returns an object with the following attributes:

- `policies` (List of Strings) - Array of complete IAM policy JSON documents. Each policy is standalone and can be used independently.
- `count` (Number) - Number of policies generated from the split operation.
- `total_size_reduction` (Number) - Total size difference in bytes between the original policy and all split policies combined. May be negative if splitting adds overhead.
- `metadata` (Object) - Additional information about the splitting operation:
    - `original_size` (Number) - Size of the original policy in bytes
    - `average_size` (Number) - Average size of split policies in bytes
    - `largest_policy` (Number) - Size of the largest split policy in bytes
    - `smallest_policy` (Number) - Size of the smallest split policy in bytes

## Behavior

### Size Limits and Validation

The function enforces AWS service-specific size limits:

- **Inline policies**: 2048 bytes maximum
- **Managed policies**: 6144 bytes maximum  
- **Resource-based policies**: 10240 bytes maximum

If the original policy is already within the specified size limit, it is returned as-is (but reformatted for consistency).

### Statement Integrity

- Individual policy statements are never split or modified
- Each statement remains intact and is moved as a complete unit
- If any single statement exceeds the size limit, the function returns an error

### Policy Structure Preservation

- All split policies maintain the same `Version` as the original
- If the original policy has an `Id` field, split policies get unique IDs: `"original-id-split-1"`, `"original-id-split-2"`, etc.
- Each split policy is a complete, valid IAM policy document

### Algorithm

The function uses an accurate size-based bin-packing algorithm:

1. Validates the input policy structure and service type
2. Checks if any individual statement is too large to fit
3. Uses greedy bin-packing to optimally distribute statements across multiple policies
4. Validates that each resulting policy is within the size limit
5. Generates metadata about the splitting operation

## Error Conditions

The function will return an error in the following cases:

- **Invalid JSON**: The `policy_json` parameter is not valid JSON
- **Invalid policy structure**: Missing required fields (`Version`, `Statement`) or invalid field values
- **Unsupported version**: Policy version is not `"2012-10-17"` or `"2008-10-17"`
- **Invalid service type**: Service type is not one of the supported values
- **Statement too large**: Any individual statement exceeds the size limit for the specified service type
- **Base policy too large**: The base policy structure (without statements) exceeds the size limit

## Use Cases

### Automated Policy Management for IAM Roles

The primary use case is enabling Terraform practitioners to define comprehensive policies without worrying about AWS size limits. The function automatically splits large policies and attaches them to IAM roles:

```terraform
# Practitioner defines one large policy - no manual splitting needed
locals {
  my_app_permissions = {
    Version = "2012-10-17"
    Statement = [
      # ... many statements that exceed 2048 bytes
    ]
  }
  
  # Function automatically handles splitting and attachment
  split_policies = provider::aws::iam_policy_split(jsonencode(local.my_app_permissions), "inline")
}

resource "aws_iam_role_policy" "auto_split" {
  count  = local.split_policies.count
  name   = "app-permissions-${count.index + 1}"
  role   = aws_iam_role.app.id
  policy = local.split_policies.policies[count.index]
}
```

This eliminates the need for manual policy splitting and ensures all permissions are properly attached to the role.

### Managing Large Inline Policies

When attaching policies directly to IAM users, groups, or roles, AWS limits inline policies to 2048 bytes. This function helps split large policies into multiple smaller inline policies.

### Converting Between Policy Types

Use different service types to understand how a policy would need to be split for different AWS services:

```terraform
# Check if a policy needs splitting for different service types
locals {
  policy_json = jsonencode(local.my_policy)
  
  inline_split    = provider::aws::iam_policy_split(local.policy_json, "inline")
  managed_split   = provider::aws::iam_policy_split(local.policy_json, "managed")
  resource_split  = provider::aws::iam_policy_split(local.policy_json, "resource-based")
}

output "policy_analysis" {
  value = {
    needs_splitting_for_inline    = local.inline_split.count > 1
    needs_splitting_for_managed   = local.managed_split.count > 1
    needs_splitting_for_resource  = local.resource_split.count > 1
  }
}
```

### Policy Size Optimization

Use the metadata to understand policy size characteristics and optimize policy structure:

```terraform
locals {
  split_result = provider::aws::iam_policy_split(local.large_policy, "inline")
}

output "size_analysis" {
  value = {
    original_size_kb    = local.split_result.metadata.original_size / 1024
    average_policy_size = local.split_result.metadata.average_size
    size_efficiency     = local.split_result.metadata.average_size / local.split_result.metadata.original_size * local.split_result.count
  }
}
```
