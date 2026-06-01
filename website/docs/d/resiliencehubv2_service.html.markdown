---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_service"
description: |-
  Terraform data source for reading an AWS Resilience Hub V2 Service.
---

# Data Source: aws_resiliencehubv2_service

Terraform data source for reading an AWS Resilience Hub V2 Service.

## Example Usage

### Basic Usage

```hcl
data "aws_resiliencehubv2_service" "example" {
  arn = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123"
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) ARN of the service.
* `region` - (Optional, **Deprecated**) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the service.
* `name` - Name of the service.
* `permission_model` - Permission model configuration. See [`permission_model` Block](#permission_model-block) below.
* `policy_arn` - ARN of the associated resilience policy.
* `regions` - List of AWS regions where the service operates.
* `tags` - Map of tags assigned to the resource.

### `permission_model` Block

* `invoker_role_name` - Name of the IAM role used for resource discovery.
