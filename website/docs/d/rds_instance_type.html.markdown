---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_instance_type"
description: |-
  Information about hardware specifications for an RDS DB instance class.
---

# Data Source: aws_rds_instance_type

Information about hardware specifications (memory, vCPUs) for an RDS DB instance class.

The RDS API does not expose memory or vCPU information for DB instance classes. This data source derives the underlying EC2 instance type from the DB instance class (for example, `db.t3.medium` corresponds to EC2 instance type `t3.medium`) and looks up its hardware specifications. As a result, DB instance classes with no directly corresponding EC2 instance type (for example, `db.serverless`) are not supported.

## Example Usage

```terraform
data "aws_rds_instance_type" "example" {
  instance_class = "db.t3.medium"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `instance_class` - (Required) RDS DB instance class, for example `db.t3.medium`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `burstable_performance_supported` - Whether the instance class is a burstable performance instance class.
* `current_generation` - Whether the instance class is a current generation instance class.
* `default_cores` - Default number of cores.
* `default_threads_per_core` - Default number of threads per core.
* `default_vcpus` - Default number of vCPUs.
* `ec2_instance_type` - EC2 instance type derived from `instance_class` and used to look up hardware specifications.
* `free_tier_eligible` - Whether the instance class is eligible for the free tier.
* `id` - RDS DB instance class (same value as `instance_class`).
* `memory_size` - Size of memory for the instance class, in MiB.
* `supported_architectures` - Supported processor architectures.
