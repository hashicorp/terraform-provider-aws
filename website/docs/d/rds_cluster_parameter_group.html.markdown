---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster_parameter_group"
description: |-
  Information about an RDS cluster parameter group.
---

# Data Source: aws_rds_cluster_parameter_group

Information about an RDS cluster parameter group.

## Example Usage

```terraform
data "aws_rds_cluster_parameter_group" "test" {
  name = "default.postgres15"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) DB cluster parameter group name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the cluster parameter group.
* `family` - Family of the cluster parameter group.
* `description` - Description of the cluster parameter group.
