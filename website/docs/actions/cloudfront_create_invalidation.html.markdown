---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_create_invalidation"
description: |-
  Invalidates CloudFront distribution cache for specified paths.
---

# Action: aws_cloudfront_create_invalidation

~> **Note:** `aws_cloudfront_create_invalidation` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Invalidates CloudFront distribution cache for specified paths. This action creates an invalidation request and waits for it to complete.

For information about CloudFront cache invalidation, see the [Amazon CloudFront Developer Guide](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Invalidation.html). For specific information about creating invalidation requests, see the [CreateInvalidation](https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_CreateInvalidation.html) page in the Amazon CloudFront API Reference.

~> **Note:** CloudFront invalidation requests can take several minutes to complete. This action will wait for the invalidation to finish before continuing. You can only have a limited number of invalidation requests in progress at any given time.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_distribution" "example" {
  # ... distribution configuration
}

action "aws_cloudfront_create_invalidation" "example" {
  config {
    distribution_id = aws_cloudfront_distribution.example.id
    paths           = ["/*"]
  }
}

resource "terraform_data" "example" {
  input = "trigger-invalidation"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_cloudfront_create_invalidation.example]
    }
  }
}
```

### Invalidate Specific Paths

```terraform
action "aws_cloudfront_create_invalidation" "assets" {
  config {
    distribution_id = aws_cloudfront_distribution.example.id
    paths = [
      "/images/*",
      "/css/*",
      "/js/app.js",
      "/index.html"
    ]
    timeout = 1200 # 20 minutes
  }
}
```

### With Custom Caller Reference

```terraform
action "aws_cloudfront_create_invalidation" "deployment" {
  config {
    distribution_id  = aws_cloudfront_distribution.example.id
    paths            = ["/*"]
    caller_reference = "deployment-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
    timeout          = 900
  }
}
```

### CI/CD Pipeline Integration

Use this action in your deployment pipeline to invalidate cache after updating static assets:

```terraform
# Trigger invalidation after S3 sync
resource "terraform_data" "deploy_complete" {
  input = local.deployment_id

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_cloudfront_create_invalidation.post_deploy]
    }
  }

  depends_on = [aws_s3_object.assets]
}

action "aws_cloudfront_create_invalidation" "post_deploy" {
  config {
    distribution_id = aws_cloudfront_distribution.main.id
    paths = [
      "/index.html",
      "/manifest.json",
      "/static/js/*",
      "/static/css/*"
    ]
  }
}
```

### Environment-Specific Invalidation

```terraform
locals {
  cache_paths = var.environment == "production" ? [
    "/api/*",
    "/assets/*"
  ] : ["/*"]
}

action "aws_cloudfront_create_invalidation" "env_specific" {
  config {
    distribution_id = aws_cloudfront_distribution.app.id
    paths           = local.cache_paths
    timeout         = var.environment == "production" ? 1800 : 900
  }
}
```

## Argument Reference

This action supports the following arguments:

* `distribution_id` - (Required) ID of the CloudFront distribution to invalidate cache for. Must be a valid CloudFront distribution ID (e.g., E1GHKQ2EXAMPLE).
* `paths` - (Required) List of file paths or patterns to invalidate. Use `/*` to invalidate all files. Supports specific files (`/index.html`), directory wildcards (`/images/*`), or all files (`/*`). Maximum of 3000 paths per invalidation request. Note: The first 1,000 invalidation paths per month are free, additional paths are charged per path.
* `caller_reference` - (Optional) Unique identifier for the invalidation request. If not provided, one will be generated automatically. Maximum length of 128 characters.
* `timeout` - (Optional) Timeout in seconds to wait for the invalidation to complete. Defaults to 900 seconds (15 minutes). Must be between 60 and 3600 seconds. Invalidation requests typically take 5-15 minutes to process.
