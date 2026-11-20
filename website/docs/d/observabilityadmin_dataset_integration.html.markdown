---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_dataset_integration"
description: |-
  Provides details about the CloudWatch Observability Admin Dataset Integration for the current account and Region.
---

# Data Source: aws_observabilityadmin_dataset_integration

Provides details about the CloudWatch Observability Admin Dataset Integration for the current account and Region. Reading fails if no dataset integration exists in the Region.

## Example Usage

### Basic Usage

```terraform
data "aws_observabilityadmin_dataset_integration" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the dataset integration.
* `created_at` - Timestamp when the dataset integration was created, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).
* `role_arn` - ARN of the IAM role associated with the dataset integration.
* `tags` - Map of tags assigned to the dataset integration.
* `updated_at` - Timestamp when the dataset integration was last updated, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).
