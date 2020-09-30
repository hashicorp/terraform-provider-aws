---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_custom_data_identifier"
description: |-
  Provides a Macie Custom Data Identifier resource.
---

# Resource: aws_macie2_custom_data_identifier

Provides a Macie Custom Data Identifier resource.

~> **NOTE:** Before using Amazon Macie for the first time it must be enabled manually. Instructions are [here](https://docs.aws.amazon.com/macie/latest/user/getting-started.html#enable-macie).

~> **NOTE:** It is highly recommended that you test and refine the custom data identifier in the AWS Console before managing it in Terraform. Because custom data identifiers are used by sensitive data discovery jobs, you can't edit a custom data identifier after you save it. This helps ensure that you have an immutable history of sensitive data findings and discovery results for data privacy and protection audits or investigations that you perform. As such, if you change any of the attributes of an `aws_macie2_custom_data_identifier` resource, Terraform will create a new resource.

## Example Usage

```hcl
resource "aws_macie2_custom_data_identifier" "example" {
  name                   = "tf-macie-example"
  regex                  = "\\d"
  description            = "An example Macie custom data identifier"
  ignore_words           = ["ignorethis", "andthis"]
  keywords               = ["keyword1", "keyword2"]
  maximum_match_distance = 100
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the custom data identifier.
* `regex` - (Required) The regular expression that defines a text pattern to match in data. See [the Macie documentation](https://docs.aws.amazon.com/macie/latest/user/custom-data-identifiers.html#custom-data-identifiers-regex-support) for specific details on the supported pattern syntax.
* `description` - (Optional) A description of the custom data identifier, up to 512 characters in length.
* `ignore_words` - (Optional) Up to 10 expressions that define specific text to exclude from the results. Each ignore word can contain 4–90 characters. Ignore words are case sensitive.
* `keywords` - (Optional) Up to 50 keywords that define specific text to match. Each keyword can contain 4–90 characters. Keywords aren't case sensitive.
* `maximum_match_distance` - (Optional) The maximum allowable distance between text that matches the regex pattern and any of the keywords. Must be between 1 and 300. The default distance is 50 characters.
* `tags` - (Optional) Key-value mapping of resource tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the custom data identifier
* `id` - The ID of the custom data identifier.

## Import

Custom Data Identifiers can be imported using the `id`, e.g.

```
$ terraform import aws_macie2_custom_data_identifier 9afbae86-c154-4ae9-ac5f-667a0a83891e
```
