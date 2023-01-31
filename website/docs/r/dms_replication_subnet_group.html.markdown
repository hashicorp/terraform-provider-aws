---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_replication_subnet_group"
description: |-
  Provides a DMS (Data Migration Service) subnet group resource.
---

# Resource: aws_dms_replication_subnet_group

Provides a DMS (Data Migration Service) replication subnet group resource. DMS replication subnet groups can be created, updated, deleted, and imported.

~> **Note:** AWS requires a special IAM role called `dms-vpc-role` when using this resource. See the example below to create it as part of your configuration.

## Example Usage

### Basic

```terraform
# Create a new replication subnet group
resource "aws_dms_replication_subnet_group" "example" {
  replication_subnet_group_description = "Example replication subnet group"
  replication_subnet_group_id          = "example-dms-replication-subnet-group-tf"

  subnet_ids = [
    "subnet-12345678",
    "subnet-12345679",
  ]

  tags = {
    Name = "example"
  }
}
```

### Creating special IAM role

If your account does not already include the `dms-vpc-role` IAM role, you will need to create it to allow DMS to manage subnets in the VPC.

```terraform
resource "aws_iam_role" "dms-vpc-role" {
  name        = "dms-vpc-role"
  description = "Allows DMS to manage VPC"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "dms.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.dms-vpc-role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonDMSVPCManagementRole"
}

resource "aws_dms_replication_subnet_group" "example" {
  replication_subnet_group_description = "Example"
  replication_subnet_group_id          = "example-id"

  subnet_ids = [
    "subnet-12345678",
    "subnet-12345679",
  ]

  tags = {
    Name = "example-id"
  }

  # explicit depends_on is needed since this resource doesn't reference the role or policy attachment
  depends_on = [aws_iam_role_policy_attachment.example]
}
```

## Argument Reference

The following arguments are supported:

* `replication_subnet_group_description` - (Required) Description for the subnet group.
* `replication_subnet_group_id` - (Required) Name for the replication subnet group. This value is stored as a lowercase string. It must contain no more than 255 alphanumeric characters, periods, spaces, underscores, or hyphens and cannot be `default`.
* `subnet_ids` - (Required) List of at least 2 EC2 subnet IDs for the subnet group. The subnets must cover at least 2 availability zones.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_id` - The ID of the VPC the subnet group is in.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `15m`)
- `update` - (Default `15m`)
- `delete` - (Default `15m`)

## Import

Replication subnet groups can be imported using the `replication_subnet_group_id`, e.g.,

```
$ terraform import aws_dms_replication_subnet_group.test test-dms-replication-subnet-group-tf
```
