---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_service"
description: |-
  Terraform resource for managing an AWS Resilience Hub V2 Service.
---

# Resource: aws_resiliencehubv2_service

Terraform resource for managing an AWS Resilience Hub V2 Service.

A service is the primary building block in Resilience Hub. It comprises AWS resources, code, and observability that together deliver a specific capability. Services can be associated with a resilience policy and a permission model for resource discovery.

## Example Usage

### Basic Usage

```hcl
resource "aws_resiliencehubv2_service" "example" {
  name    = "example-service"
  regions = ["us-west-2"]

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}
```

### With Policy

```hcl
resource "aws_resiliencehubv2_policy" "example" {
  name = "example-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "example" {
  name        = "example-service"
  description = "Production API service"
  policy_arn  = aws_resiliencehubv2_policy.example.arn
  regions     = ["us-west-2", "us-east-1"]

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the service. Changing this value requires creating a new resource.
* `regions` - (Required) List of AWS regions where the service operates.

The following arguments are optional:

* `description` - (Optional) Description of the service.
* `permission_model` - (Optional) Permission model for resource discovery. See [`permission_model` Block](#permission_model-block) below.
* `policy_arn` - (Optional) ARN of the resilience policy to associate with this service.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `permission_model` Block

The `permission_model` block supports:

* `invoker_role_name` - (Required) Name of the IAM role that Resilience Hub assumes for resource discovery.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the service.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_resiliencehubv2_service.example
  identity = {
    "arn" = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123"
  }
}

resource "aws_resiliencehubv2_service" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the Resilience Hub V2 Service.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub V2 Service using the `arn`. For example:

```terraform
import {
  to = aws_resiliencehubv2_service.example
  id = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123"
}
```

Using `terraform import`, import Resilience Hub V2 Service using the `arn`. For example:

```console
% terraform import aws_resiliencehubv2_service.example arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123
```
