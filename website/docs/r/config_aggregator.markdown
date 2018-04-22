---
layout: "aws"
page_title: "AWS: aws_config_aggregator"
sidebar_current: "docs-aws-resource-config-aws_config_aggregator"
description: |-
  Provides an AWS Config Aggregator.
---

# aws_config_aggregator

Provides an AWS Config Aggregator

## Example Usage

```hcl
resource "aws_config_aggregator" "account" {
  name = "example" # Required

  account_aggregation_source {
    account_ids = ["123456789012"] # Required
    regions     = ["us-west-2"]    # Optional
  }
}
```

```hcl
resource "aws_config_aggregator" "organization" {
  depends_on = ["aws_iam_role_policy_attachment.organization"]

  name = "example" # Required

  organization_aggregation_source {
    all_regions = true                               # Optional
    role_arn    = "${aws_iam_role.organization.arn}" # Required
  }
}

resource "aws_iam_role" "organization" {
  name = "example"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "organization" {
  role       = "${aws_iam_role.organization.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSConfigRoleForOrganizations"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the configuration aggregator.
* `account_aggregation_source` - (Optional) The account(s) to aggregate config data from as documented below.
* `organization_aggregation_source` - (Optional) The organization to aggregate config data from as documented below.

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
* `role_arn` - (Required) ARN of the IAM role used to retreive AWS Organization details associated with the aggregator account.

Either `regions` or `all_regions` (as true) must be specified.

## Attributes Reference

The following attributes are exported:

* `arn` - The ARN of the aggregator

## Import

Aggregators can be imported using the name, e.g.

```
$ terraform import aws_config_aggregator.example foo
```
