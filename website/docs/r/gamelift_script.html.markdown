---
subcategory: "GameLift"
layout: "aws"
page_title: "AWS: aws_gamelift_script"
description: |-
  Provides a GameLift Script resource.
---

# Resource: aws_gamelift_script

Provides an GameLift Script resource.

## Example Usage

```terraform
resource "aws_gamelift_script" "example" {
  name = "example-script"

  storage_location {
    bucket   = aws_s3_bucket.example.bucket
    key      = aws_s3_object.example.key
    role_arn = aws_iam_role.example.arn
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the script
* `storage_location` - (Optional) Information indicating where your game script files are stored. See below.
* `version` - (Optional) Version that is associated with this script.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `zip_file` - (Optional) A data object containing your Realtime scripts and dependencies as a zip  file. The zip file can have one or multiple files. Maximum size of a zip file is 5 MB.

### Nested Fields

#### `storage_location`

* `bucket` - (Required) Name of your S3 bucket.
* `key` - (Required) Name of the zip file containing your script files.
* `role_arn` - (Required) ARN of the access role that allows Amazon GameLift to access your S3 bucket.
* `object_version` - (Optional) A specific version of the file. If not set, the latest version of the file is retrieved.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - GameLift Script ID.
* `arn` - GameLift Script ARN.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

GameLift Scripts can be imported using the ID, e.g.,

```
$ terraform import aws_gamelift_script.example <script-id>
```
