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
* `ecr_repository_prefix` - (Required, Forces new resource) The repository name prefix to use when caching images from the source registry.
* `upstream_registry_url` - (Required, Forces new resource) The registry URL of the upstream public registry to use as the source.

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
