---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_delete_default_vpc"
description: |-
  Deletes the default VPC in the current region.
---

# Action: aws_vpc_delete_default_vpc

!> **Warning:** This action is destructive and cannot be undone. When triggered, the `aws_vpc_delete_default_vpc` action permanently deletes the default VPC and all its dependencies including internet gateways, subnets, route tables, network ACLs, and security groups. Any resources running in the default VPC (EC2 instances, RDS databases, etc.) will be affected. Use extreme caution—this action should be limited to development environments or accounts where the default VPC is confirmed to be unused.

Deletes the default VPC in the current region. This action removes the default VPC and all its associated resources, providing a clean slate for VPC configuration.

For information about Amazon VPC, see the [Amazon VPC User Guide](https://docs.aws.amazon.com/vpc/latest/userguide/). For specific information about default VPCs, see the [Default VPC and default subnets](https://docs.aws.amazon.com/vpc/latest/userguide/default-vpc.html) page in the Amazon VPC User Guide.

~> **Note:** If no default VPC exists in the region, this action will complete successfully without error. The action automatically handles cleanup of VPC dependencies in the correct order.

## Example Usage

### Basic Usage

```terraform
action "aws_vpc_delete_default_vpc" "example" {
  config {}
}
```

### Multi-Region Cleanup

```terraform
# Get all available AWS regions
data "aws_regions" "all" {
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required", "opted-in"]
  }
}

# Create an action for each region to delete default VPC
action "aws_vpc_delete_default_vpc" "all_regions" {
  for_each = toset(data.aws_regions.all.names)

  config {
    region  = each.value
    timeout = 900
  }
}

resource "terraform_data" "multi_region_cleanup" {
  input = "cleanup-all"

  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = values(action.aws_vpc_delete_default_vpc.all_regions)
    }
  }
}
```

## Argument Reference

This action supports the following arguments:

* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `timeout` - (Optional) Timeout in seconds to wait for VPC deletion to complete. Must be between 60 and 3600 seconds. Default: `600`.

## Behavior

When executed, this action performs the following steps:

1. Locates the default VPC in the specified region
2. If no default VPC exists, the action completes successfully
3. If a default VPC is found, it deletes the following dependencies in order:
   - Internet gateways (detached then deleted)
   - Subnets
   - Route tables (non-main only)
   - Security groups (non-default only)
   - Network ACLs (non-default only)
4. Deletes the default VPC itself
5. Waits for VPC deletion to complete

Default resources (default security group, default network ACL, main route table) are automatically deleted when the VPC is removed.
