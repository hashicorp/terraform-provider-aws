---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_lifecycle_policy_document"
description: |-
    Generates an ECR lifecycle policy document in JSON format.
---

# Data Source: aws_ecr_lifecycle_policy_document

Generates an ECR lifecycle policy document in JSON format. Can be used with resources such as the [`aws_ecr_lifecycle_policy` resource](/docs/providers/aws/r/ecr_lifecycle_policy.html).

-> For more information about building AWS ECR lifecycle policy documents, see the [AWS ECR Lifecycle Policy Document Guide](https://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html).

## Example Usage

```hcl
data "aws_ecr_lifecycle_policy_document" "example" {
  rule {
    priority    = 1
    description = "This is a test."

    selection {
      tag_status      = "tagged"
      tag_prefix_list = ["prod"]
      count_type      = "imageCountMoreThan"
      count_number    = 100
    }
  }
}

resource "aws_ecr_lifecycle_policy" "example" {
  repository = aws_ecr_repository.example.name

  policy = data.aws_ecr_lifecycle_policy_document.example.json
}
```

## Argument Reference

This data source supports the following arguments:

Each document configuration may have one or more `rule` blocks, which each accept the following arguments:

* `action` (Optional) - Specifies the action type.
    * `type` (Required) - The supported value is `expire`.
* `description` (Optional) - Describes the purpose of a rule within a lifecycle policy.
* `priority` (Required) - Sets the order in which rules are evaluated, lowest to highest. When you add rules to a lifecycle policy, you must give them each a unique value for `priority`. Values do not need to be sequential across rules in a policy. A rule with a `tag_status` value of "any" must have the highest value for `priority` and be evaluated last.
* `selection` (Required) -  Collects parameters describing the selection criteria for the ECR lifecycle policy:
    * `tag_status` (Required) - Determines whether the lifecycle policy rule that you are adding specifies a tag for an image. Acceptable options are "tagged", "untagged", or "any". If you specify "any", then all images have the rule applied to them. If you specify "tagged", then you must also specify a `tag_prefix_list` value. If you specify "untagged", then you must omit `tag_prefix_list`.
    * `tag_pattern_list` (Required if `tag_status` is set to "tagged" and `tag_prefix_list` isn't specified) - You must specify a comma-separated list of image tag patterns that may contain wildcards (\*) on which to take action with your lifecycle policy. For example, if your images are tagged as `prod`, `prod1`, `prod2`, and so on, you would use the tag pattern list `["prod\*"]` to specify all of them. If you specify multiple tags, only the images with all specified tags are selected. There is a maximum limit of four wildcards (\*) per string. For example, `["*test*1*2*3", "test*1*2*3*"]` is valid but `["test*1*2*3*4*5*6"]` is invalid.
    * `tag_prefix_list` (Required if `tag_status` is set to "tagged" and `tag_pattern_list` isn't specified) - You must specify a comma-separated list of image tag prefixes on which to take action with your lifecycle policy. For example, if your images are tagged as `prod`, `prod1`, `prod2`, and so on, you would use the tag prefix "prod" to specify all of them. If you specify multiple tags, only images with all specified tags are selected.
    * `count_type` (Required) - Specify a count type to apply to the images. If `count_type` is set to "imageCountMoreThan", you also specify `count_number` to create a rule that sets a limit on the number of images that exist in your repository. If `count_type` is set to "sinceImagePushed", you also specify `count_unit` and `count_number` to specify a time limit on the images that exist in your repository.
    * `count_unit` (Required if `count_type` is set to "sinceImagePushed") - Specify a count unit of days to indicate that as the unit of time, in addition to `count_number`, which is the number of days.
    * `count_number` (Required) - Specify a count number. If the `count_type` used is "imageCountMoreThan", then the value is the maximum number of images that you want to retain in your repository. If the `count_type` used is "sinceImagePushed", then the value is the maximum age limit for your images.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `json` - The above arguments serialized as a standard JSON policy document.
