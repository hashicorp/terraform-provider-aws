---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_signing_configuration"
description: |-
  Manages the Amazon ECR managed signing configuration for a registry.
---

# Resource: aws_ecr_signing_configuration

Manages the Amazon ECR [managed signing](https://docs.aws.amazon.com/AmazonECR/latest/userguide/managed-signing.html) configuration for the registry in a Region. When an image is pushed, ECR signs it with the AWS Signer signing profile of every rule whose repository filters match the repository.

~> **NOTE:** A registry has one signing configuration per Region. This resource manages the whole configuration, so declare only one `aws_ecr_signing_configuration` per Region. Destroying the resource deletes the signing configuration.

~> **NOTE:** A `rule` with no `repository_filter` blocks signs every image pushed to the registry.

~> **NOTE:** Creating this resource replaces any signing configuration that already exists in the Region, including one created in the console or with the AWS CLI. To manage an existing configuration, [import](#import) it first.

-> **NOTE:** Signing happens when an image is pushed, using the identity of the principal that pushes it. That principal needs `signer:SignPayload` on the signing profile, or signing fails even though this resource applies successfully. Use `aws ecr describe-image-signing-status` to check whether an image was signed. See the [managed signing prerequisites](https://docs.aws.amazon.com/AmazonECR/latest/userguide/managed-signing.html#managed-signing-prerequisites).

## Example Usage

### Basic Usage

```terraform
resource "aws_signer_signing_profile" "example" {
  name        = "example"
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_ecr_signing_configuration" "example" {
  rule {
    signing_profile_arn = aws_signer_signing_profile.example.arn

    repository_filter {
      filter      = "team-a/*"
      filter_type = "WILDCARD_MATCH"
    }
  }
}
```

### Multiple Rules

```terraform
resource "aws_signer_signing_profile" "team_a" {
  name        = "team_a"
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "production" {
  name        = "production"
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_ecr_signing_configuration" "example" {
  rule {
    signing_profile_arn = aws_signer_signing_profile.team_a.arn

    repository_filter {
      filter      = "team-a/*"
      filter_type = "WILDCARD_MATCH"
    }
  }

  rule {
    signing_profile_arn = aws_signer_signing_profile.production.arn

    repository_filter {
      filter      = "*/production"
      filter_type = "WILDCARD_MATCH"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rule` - (Required) One to 10 signing rules. See [`rule`](#rule-block) below.

### `rule` Block

* `repository_filter` - (Optional) Up to 100 filters that select which repositories this rule signs images for. If omitted, the rule signs every image pushed to the registry. See [`repository_filter`](#repository_filter-block) below.
* `signing_profile_arn` - (Required) ARN of the AWS Signer signing profile used to sign matching images. The profile must use the `Notation-OCI-SHA384-ECDSA` platform and be in the same Region as the registry. It can be in a different account.

### `repository_filter` Block

* `filter` - (Required) Repository name pattern. With `WILDCARD_MATCH`, `*` matches any sequence of characters, for example `myapp/*`, `*/production`, or `*prod*`.
* `filter_type` - (Required) Type of filter. Valid values: `WILDCARD_MATCH`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `registry_id` - AWS account ID of the registry.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ecr_signing_configuration.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_ecr_signing_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the ECR signing configuration using the Region. For example:

```terraform
import {
  to = aws_ecr_signing_configuration.example
  id = "us-west-2"
}
```

Using `terraform import`, import the ECR signing configuration using the Region. For example:

```console
% terraform import aws_ecr_signing_configuration.example us-west-2
```
