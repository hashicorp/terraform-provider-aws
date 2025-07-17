---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_web_acl"
description: |-
  Retrieves a WAF Regional Web ACL id.
---

# Data Source: aws_wafregional_web_acl

`aws_wafregional_web_acl` Retrieves a WAF Regional Web ACL Resource Id.

## Example Usage

```terraform
data "aws_wafregional_web_acl" "example" {
  name = "tfWAFRegionalWebACL"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the WAF Regional Web ACL.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF Regional Web ACL.
