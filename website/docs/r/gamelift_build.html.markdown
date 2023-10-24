---
subcategory: "GameLift"
layout: "aws"
page_title: "AWS: aws_gamelift_build"
description: |-
  Provides a GameLift Build resource.
---

# Resource: aws_gamelift_build

Provides an GameLift Build resource.

## Example Usage

```terraform
resource "aws_gamelift_build" "test" {
  name             = "example-build"
  operating_system = "WINDOWS_2012"

  storage_location {
    bucket   = aws_s3_bucket.test.id
    key      = aws_s3_object.test.key
    role_arn = aws_iam_role.test.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name of the build
* `operating_system` - (Required) Operating system that the game server binaries are built to run onE.g., `WINDOWS_2012`, `AMAZON_LINUX` or `AMAZON_LINUX_2`.
* `storage_location` - (Required) Information indicating where your game build files are stored. See below.
* `version` - (Optional) Version that is associated with this build.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Nested Fields

#### `storage_location`

* `bucket` - (Required) Name of your S3 bucket.
* `key` - (Required) Name of the zip file containing your build files.
* `role_arn` - (Required) ARN of the access role that allows Amazon GameLift to access your S3 bucket.
* `object_version` - (Optional) A specific version of the file. If not set, the latest version of the file is retrieved.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - GameLift Build ID.
* `arn` - GameLift Build ARN.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import GameLift Builds using the ID. For example:

```terraform
import {
  to = aws_gamelift_build.example
  id = "<build-id>"
}
```

Using `terraform import`, import GameLift Builds using the ID. For example:

```console
% terraform import aws_gamelift_build.example <build-id>
```
