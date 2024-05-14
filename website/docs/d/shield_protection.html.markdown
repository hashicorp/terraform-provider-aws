---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_protection"
description: |-
  Terraform data source for managing an AWS Shield Protection.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->

# Data Source: aws_shield_protection

Terraform data source for managing an AWS Shield Protection.

## Example Usage

### Basic Usage

```terraform
data "aws_shield_protection" "example" {
  resource_arn = "arn:${data.aws_partition.current.partition}:route53:::hostedzone/${aws_route53_zone.test.zone_id}"
}
```

## Argument Reference

The following arguments are optional:

* `resource_arn` - (Optional) The ARN (Amazon Resource Name) of the resource to be protected.
* `protection_id` - (Optional) The unique identifier (ID) for the Protection object.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `protection_id` - The unique identifier (ID) for the Protection object that is created.
* `arn` - The ARN of the Protection.
* `resource_arn` - The ARN (Amazon Resource Name) of the resource to be protected.
