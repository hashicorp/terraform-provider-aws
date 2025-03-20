---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_pull_through_cache_rule"
description: |-
  Provides an Elastic Container Registry Pull Through Cache Rule.
---

# Resource: aws_ecr_pull_through_cache_rule

Provides an Elastic Container Registry Pull Through Cache Rule.

More information about pull through cache rules, including the set of supported
upstream repositories, see [Using pull through cache rules](https://docs.aws.amazon.com/AmazonECR/latest/userguide/pull-through-cache.html).

## Example Usage

```terraform
resource "aws_ecr_pull_through_cache_rule" "example" {
  ecr_repository_prefix = "ecr-public"
  upstream_registry_url = "public.ecr.aws"
  credential_arn        = "arn:aws:secretsmanager:us-east-1:123456789:secret:ecr-pullthroughcache/ecrpublic"
}
```

## Argument Reference

This resource supports the following arguments:

* `credential_arn` - (Optional) ARN of the Secret which will be used to authenticate against the registry.
* `custom_role_arn` - (Optional) The ARN of the IAM role associated with the pull through cache rule. Must be specified if the upstream registry is a cross-account ECR private registry. See [AWS Document - Setting up permissions for cross-account ECR to ECR PTC](https://docs.aws.amazon.com/AmazonECR/latest/userguide/pull-through-cache-private.html).
* `ecr_repository_prefix` - (Required, Forces new resource) The repository name prefix to use when caching images from the source registry. Use `ROOT` as the prefix to apply a template to all repositories in your registry that don't have an associated pull through cache rule.
* `upstream_registry_url` - (Required, Forces new resource) The registry URL of the upstream registry to use as the source.
* `upstream_repository_prefix` - (Optional, Forces new resource) The upstream repository prefix associated with the pull through cache rule. Used if the upstream registry is an ECR private registry. If not specified, it's set to `ROOT`, which allows matching with any upstream repository. See [AWS Document - Customizing repository prefixes for ECR to ECR pull through cache](https://docs.aws.amazon.com/AmazonECR/latest/userguide/pull-through-cache-private-wildcards.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `registry_id` - The registry ID where the repository was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a pull-through cache rule using the `ecr_repository_prefix`. For example:

```terraform
import {
  to = aws_ecr_pull_through_cache_rule.example
  id = "ecr-public"
}
```

Using `terraform import`, import a pull-through cache rule using the `ecr_repository_prefix`. For example:

```console
% terraform import aws_ecr_pull_through_cache_rule.example ecr-public
```
