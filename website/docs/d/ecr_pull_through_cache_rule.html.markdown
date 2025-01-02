---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_pull_through_cache_rule"
description: |-
  Provides details about an ECR Pull Through Cache Rule
---

# Data Source: aws_ecr_pull_through_cache_rule

The ECR Pull Through Cache Rule data source allows the upstream registry URL and registry ID to be retrieved for a Pull Through Cache Rule.

## Example Usage

```terraform
data "aws_ecr_pull_through_cache_rule" "ecr_public" {
  ecr_repository_prefix = "ecr-public"
}
```

## Argument Reference

This data source supports the following arguments:

- `ecr_repository_prefix` - (Required) The repository name prefix to use when caching images from the source registry.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

- `id` - The repository name prefix.
- `credential_arn` - ARN of the Secret which will be used to authenticate against the registry.
- `registry_id` - The registry ID where the repository was created.
- `upstream_registry_url` - The registry URL of the upstream public registry to use as the source.
