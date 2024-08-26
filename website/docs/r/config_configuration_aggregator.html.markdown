---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_configuration_aggregator"
description: |-
  Manages an AWS Config Configuration Aggregator.
---

# Resource: aws_config_configuration_aggregator

Manages an AWS Config Configuration Aggregator

## Example Usage

### Account Based Aggregation

```terraform
resource "aws_config_configuration_aggregator" "account" {
  name = "example"

  account_aggregation_source {
    account_ids = ["123456789012"]
    regions     = ["us-west-2"]
  }
}
```

### Organization Based Aggregation

```terraform
resource "aws_config_configuration_aggregator" "organization" {
  depends_on = [aws_iam_role_policy_attachment.organization]

  name = "example" # Required

  organization_aggregation_source {
    all_regions = true
    role_arn    = aws_iam_role.organization.arn
  }
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["config.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}
resource "aws_iam_role" "organization" {
  name               = "example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "organization" {
  role       = aws_iam_role.organization.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSConfigRoleForOrganizations"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the configuration aggregator.
* `account_aggregation_source` - (Optional) The account(s) to aggregate config data from as documented below.
* `organization_aggregation_source` - (Optional) The organization to aggregate config data from as documented below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

Either `account_aggregation_source` or `organization_aggregation_source` must be specified.

### `account_aggregation_source`

* `account_ids` - (Required) List of 12-digit account IDs of the account(s) being aggregated.
* `all_regions` - (Optional) If true, aggregate existing AWS Config regions and future regions.
* `regions` - (Optional) List of source regions being aggregated.

Either `regions` or `all_regions` (as true) must be specified.

### `organization_aggregation_source`

~> **Note:** If your source type is an organization, you must be signed in to the master account and all features must be enabled in your organization. AWS Config calls EnableAwsServiceAccess API to enable integration between AWS Config and AWS Organizations.

* `all_regions` - (Optional) If true, aggregate existing AWS Config regions and future regions.
* `regions` - (Optional) List of source regions being aggregated.
* `role_arn` - (Required) ARN of the IAM role used to retrieve AWS Organization details associated with the aggregator account.

Either `regions` or `all_regions` (as true) must be specified.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the aggregator
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Configuration Aggregators using the name. For example:

```terraform
import {
  to = aws_config_configuration_aggregator.example
  id = "foo"
}
```

Using `terraform import`, import Configuration Aggregators using the name. For example:

```console
% terraform import aws_config_configuration_aggregator.example foo
```
