---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_centralization_rule_for_organization"
description: |-
  Manages an AWS CloudWatch Observability Admin Centralization Rule For Organization.
---

# Resource: aws_observabilityadmin_centralization_rule_for_organization

Manages an AWS CloudWatch Observability Admin Centralization Rule For Organization.

Centralization rules enable you to centralize log data from multiple AWS accounts and regions within your organization to a single destination account and region. This helps with log management, compliance, and cost optimization by consolidating logs in a central location.

This requires an AWS account within an organization with at least [delegated administrator permissions](https://docs.aws.amazon.com/organizations/latest/APIReference/API_RegisterDelegatedAdministrator.html).

## Example Usage

### Basic Centralization Rule

```terraform
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "example" {
  rule_name = "example-centralization-rule"

  rule {
    destination {
      region  = "eu-west-1"
      account = data.aws_caller_identity.current.account_id
    }

    source {
      regions = ["ap-southeast-1"]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "SKIP"
        log_group_selection_criteria = "*"
      }
    }
  }

  tags = {
    Name        = "example-centralization-rule"
    Environment = "production"
  }
}
```

### Advanced Configuration with Encryption and Backup

```terraform
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "advanced" {
  rule_name = "advanced-centralization-rule"

  rule {
    destination {
      region  = "eu-west-1"
      account = data.aws_caller_identity.current.account_id

      destination_logs_configuration {
        logs_encryption_configuration {
          encryption_strategy = "AWS_OWNED"
        }

        backup_configuration {
          region = "us-west-1"
        }
      }
    }

    source {
      regions = ["ap-southeast-1", "us-east-1"]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "ALLOW"
        log_group_selection_criteria = "*"
      }
    }
  }

  tags = {
    Name        = "advanced-centralization-rule"
    Environment = "production"
    Team        = "observability"
  }
}
```

### Selective Log Group Filtering

```terraform
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "filtered" {
  rule_name = "filtered-centralization-rule"

  rule {
    destination {
      region  = "eu-west-1"
      account = data.aws_caller_identity.current.account_id
    }

    source {
      regions = ["ap-southeast-1", "us-east-1"]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "ALLOW"
        log_group_selection_criteria = "LogGroupName LIKE '/aws/lambda%'"
      }
    }
  }

  tags = {
    Name   = "filtered-centralization-rule"
    Filter = "lambda-logs"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `rule_name` - (Required) Name of the centralization rule. Must be unique within the organization.
* `rule` - (Required) Configuration block for the centralization rule. See [`rule`](#rule) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### rule

* `destination` - (Required) Configuration block for the destination where logs will be centralized. See [`destination`](#destination) below.
* `source` - (Required) Configuration block for the source of logs to be centralized. See [`source`](#source) below.

### destination

* `account` - (Required) AWS account ID where logs will be centralized.
* `region` - (Required) AWS region where logs will be centralized.
* `destination_logs_configuration` - (Optional) Configuration block for destination logs settings. See [`destination_logs_configuration`](#destination_logs_configuration) below.

#### destination_logs_configuration

* `backup_configuration` - (Optional) Configuration block for backup settings. See [`backup_configuration`](#backup_configuration) below.
* `logs_encryption_configuration` - (Optional) Configuration block for logs encryption settings. See [`logs_encryption_configuration`](#logs_encryption_configuration) below.

##### backup_configuration

* `region` - (Required) AWS region for backup storage.
* `kms_key_arn` - (Optional) ARN of the KMS key to use for backup encryption.

##### logs_encryption_configuration

* `encryption_strategy` - (Required) Encryption strategy for logs. Valid values: `AWS_OWNED`, `CUSTOMER_MANAGED`.
* `encryption_conflict_resolution_strategy` - (Optional) Strategy for resolving encryption conflicts. Valid values: `ALLOW`, `SKIP`.
* `kms_key_arn` - (Optional) ARN of the KMS key to use for encryption when `encryption_strategy` is `CUSTOMER_MANAGED`.

### source

* `regions` - (Required) Set of AWS regions from which to centralize logs. Must contain at least one region.
* `scope` - (Required) Scope defining which resources to include. Use organization ID format: `OrganizationId = 'o-example123456'`.
* `source_logs_configuration` - (Optional) Configuration block for source logs settings. See [`source_logs_configuration`](#source_logs_configuration) below.

#### source_logs_configuration

* `encrypted_log_group_strategy` - (Required) Strategy for handling encrypted log groups. Valid values: `ALLOW`, `SKIP`.
* `log_group_selection_criteria` - (Required) Criteria for selecting log groups. Use `*` for all log groups or OAM filter syntax like `LogGroupName LIKE '/aws/lambda%'`. Must be between 1 and 2000 characters.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `rule_arn` - ARN of the centralization rule.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin Centralization Rule For Organization using the `rule_name`. For example:

```terraform
import {
  to = aws_observabilityadmin_centralization_rule_for_organization.example
  id = "example-centralization-rule"
}
```

Using `terraform import`, import CloudWatch Observability Admin Centralization Rule For Organization using the `rule_name`. For example:

```console
% terraform import aws_observabilityadmin_centralization_rule_for_organization.example example-centralization-rule
```
