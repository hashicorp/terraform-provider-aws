---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_organization_domain"
description: |-
  Manages an Amazon CloudWatch Omni organization domain.
---

# Resource: aws_cloudwatchomni_organization_domain

Manages an Amazon CloudWatch Omni organization domain.

An organization domain is a CloudWatch Omni domain shared by every account in an AWS Organizations organization. An organization can have one organization domain. While it exists, member accounts create their spaces in the organization domain instead of in an account domain.

~> **NOTE:** This resource must be managed from the organization's management account, in the domain's home Region. Trusted access for the `observabilityadmin.amazonaws.com` service principal must be enabled in AWS Organizations, and the management account must not have its own CloudWatch Omni domain.

~> **NOTE:** An organization domain cannot be deleted while any account in the organization has a space in it.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "cloudwatch.amazonaws.com"
      }
      Action = ["sts:AssumeRole", "sts:TagSession"]
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.current.account_id
        }
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/CloudWatchOmniDomainAccessPolicy"
}

resource "aws_cloudwatchomni_organization_domain" "example" {
  name                   = "example"
  identity_providers     = ["IAM"]
  domain_access_role_arn = aws_iam_role.example.arn

  depends_on = [aws_iam_role_policy_attachment.example]
}
```

### With IAM Identity Center

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_cloudwatchomni_organization_domain" "example" {
  name                   = "example"
  identity_providers     = ["IAM", "IDC"]
  domain_access_role_arn = aws_iam_role.example.arn

  identity_provider_configuration {
    identity_center_configuration {
      identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_access_role_arn` - (Required, Forces new resource) ARN of the IAM role that CloudWatch Omni assumes to access the organization. The role must be in the management account, trust `cloudwatch.amazonaws.com` for `sts:AssumeRole` and `sts:TagSession`, and have the `CloudWatchOmniDomainAccessPolicy` AWS managed policy attached.
* `identity_providers` - (Required) Set of identity providers that users can sign in with. Valid values: `IAM`, `IDC`.
* `name` - (Required) Name of the domain. Must be 3-63 characters, contain only lowercase letters, numbers, and hyphens, begin and end with a letter or number, and not contain consecutive hyphens. Changing the name changes the domain's custom endpoint URLs.

The following arguments are optional:

* `identity_provider_configuration` - (Optional) Identity provider configuration. Required when `identity_providers` includes `IDC`. [See below](#identity_provider_configuration-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `identity_provider_configuration` Block

The `identity_provider_configuration` configuration block supports the following arguments:

* `identity_center_configuration` - (Required) IAM Identity Center configuration. [See below](#identity_center_configuration-block).

### `identity_center_configuration` Block

The `identity_center_configuration` configuration block supports the following arguments:

* `identity_center_instance_arn` - (Required) ARN of the IAM Identity Center instance.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the organization domain.
* `custom_endpoint_urls` - Additional endpoint URLs derived from the domain name.
* `domain_endpoint_url` - HTTPS endpoint URL for accessing the domain.
* `domain_id` - Unique identifier of the organization domain.
* `identity_center_application_arn` - ARN of the IAM Identity Center application. Only set when `identity_providers` includes `IDC`.
* `organization_id` - ID of the organization that the domain belongs to.
* `owner_account_id` - ID of the management account that owns the domain.
* `status` - Status of the domain.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatchomni_organization_domain.example
  identity = {
    domain_id = "d-0123456789abcdefghijklmno"
  }
}

resource "aws_cloudwatchomni_organization_domain" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `domain_id` (String) Domain ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a CloudWatch Omni Organization Domain by domain ID. For example:

```terraform
import {
  to = aws_cloudwatchomni_organization_domain.example
  id = "d-0123456789abcdefghijklmno"
}
```

Using `terraform import`, import a CloudWatch Omni Organization Domain by domain ID. For example:

```console
% terraform import aws_cloudwatchomni_organization_domain.example d-0123456789abcdefghijklmno
```
