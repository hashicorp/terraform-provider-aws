---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_security_group_rules_exclusive"
description: |-
  Terraform resource for managing an exclusive set of AWS VPC (Virtual Private Cloud) Security Group Rules.
---

# Resource: aws_vpc_security_group_rules_exclusive

Terraform resource for managing an exclusive set of AWS VPC (Virtual Private Cloud) Security Group Rules.

This resource manages the complete set of ingress and egress rules assigned to a security group. It provides exclusive control by removing any rules not explicitly defined in the configuration.

!> This resource takes exclusive ownership over ingress and egress rules assigned to a security group. This includes removal of rules which are not explicitly configured. To prevent persistent drift, ensure any `aws_vpc_security_group_ingress_rule` and `aws_vpc_security_group_egress_rule` resources managed alongside this resource are included in the `ingress_rule_ids` and `egress_rule_ids` arguments.

~> Destruction of this resource means Terraform will no longer manage reconciliation of the configured security group rules. It **will not** revoke the configured rules from the security group.

~> When this resource detects a configured rule ID which must be created, a warning diagnostic is emitted. This is due to a limitation in the [`AuthorizeSecurityGroupEgress`](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupEgress.html) and [`AuthorizeSecurityGroupIngress`](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupIngress.html) APIs, which require the full rule definition to be provided rather than a reference to an existing rule ID.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "example" {
  name   = "example"
  vpc_id = aws_vpc.example.id
}

resource "aws_vpc_security_group_ingress_rule" "example" {
  security_group_id = aws_security_group.example.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  to_port     = 80
  ip_protocol = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "example" {
  security_group_id = aws_security_group.example.id

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"
}

resource "aws_vpc_security_group_rules_exclusive" "example" {
  security_group_id = aws_security_group.example.id
  ingress_rule_ids  = [aws_vpc_security_group_ingress_rule.example.id]
  egress_rule_ids   = [aws_vpc_security_group_egress_rule.example.id]
}
```

### Disallow All Rules

To automatically remove any configured security group rules, set both `ingress_rule_ids` and `egress_rule_ids` to empty lists.

~> This will not __prevent__ rules from being assigned to a security group via Terraform (or any other interface). This resource enables bringing security group rule assignments into a configured state, however, this reconciliation happens only when `apply` is proactively run.

```terraform
resource "aws_vpc_security_group_rules_exclusive" "example" {
  security_group_id = aws_security_group.example.id
  ingress_rule_ids  = []
  egress_rule_ids   = []
}
```

## Argument Reference

This resource supports the following arguments:

* `egress_rule_ids` - (Required) Egress rule IDs.
* `ingress_rule_ids` - (Required) Ingress rule IDs.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `security_group_id` - (Required, Forces new resource) ID of the security group.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import exclusive management of security group rules using the `security_group_id`. For example:

```terraform
import {
  to = aws_vpc_security_group_rules_exclusive.example
  id = "sg-1234567890abcdef0"
}
```

Using `terraform import`, import exclusive management of security group rules using the `security_group_id`. For example:

```console
% terraform import aws_vpc_security_group_rules_exclusive.example sg-1234567890abcdef0
```
