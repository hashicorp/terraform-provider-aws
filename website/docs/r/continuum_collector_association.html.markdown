---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_collector_association"
description: |-
  Associates an AWS Continuum collector with an environment.
---

# Resource: aws_continuum_collector_association

Associates an AWS Continuum collector with an environment and sets the role that Continuum assumes to read collector data.

~> **NOTE:** The collector data access role must trust `continuum.amazonaws.com` (scoped with `aws:SourceAccount` and `aws:SourceArn` for the environment) and allow `continuum:GetCollectorDataAccessAuthorization` on the collector. Continuum assumes the role while creating the association, so create the environment and collector first.

## Example Usage

```terraform
resource "aws_continuum_collector_association" "example" {
  environment_id                 = aws_continuum_environment.example.id
  collector_arn                  = aws_continuum_collector.example.arn
  collector_data_access_role_arn = aws_iam_role.collector_data_access.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `collector_arn` - (Required, Forces new resource) ARN of the collector.
* `collector_data_access_role_arn` - (Required) ARN of the IAM role that Continuum assumes to read collector data. Can be updated in place.
* `environment_id` - (Required, Forces new resource) ID of the environment.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `collector_id` - ID of the collector.
* `id` - (**Deprecated**) Import ID in the format `environment_id,collector_arn`. Use `environment_id` and `collector_arn` instead.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_continuum_collector_association.example
  identity = {
    environment_id = "01a17e22728b7d25b3a2294e0ca5c35e"
    collector_arn  = "arn:aws:continuum:us-east-1:123456789012:collectors/01a17e22728b7d25b3a2294e0ca5c35f"
  }
}

resource "aws_continuum_collector_association" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `environment_id` (String) ID of the environment.
* `collector_arn` (String) ARN of the collector.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_continuum_collector_association` using the ID. For example:

```terraform
import {
  to = aws_continuum_collector_association.example
  id = "01a17e22728b7d25b3a2294e0ca5c35e,arn:aws:continuum:us-east-1:123456789012:collectors/01a17e22728b7d25b3a2294e0ca5c35f"
}
```

Using `terraform import`, import `aws_continuum_collector_association` using the ID. For example:

```console
% terraform import aws_continuum_collector_association.example 01a17e22728b7d25b3a2294e0ca5c35e,arn:aws:continuum:us-east-1:123456789012:collectors/01a17e22728b7d25b3a2294e0ca5c35f
```
