---
subcategory: "ECR"
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
}
```

## Argument Reference

The following arguments are supported:

* `ecr_repository_prefix` - (Required, Forces new resource) The repository name prefix to use when caching images from the source registry.
* `upstream_registry_url` - (Required, Forces new resource) The registry URL of the upstream public registry to use as the source.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `registry_id` - The registry ID where the repository was created.

## Import

Use the `ecr_repository_prefix` to import a Pull Through Cache Rule. For example:

```
$ terraform import aws_ecr_pull_through_cache_rule.example ecr-public
```
