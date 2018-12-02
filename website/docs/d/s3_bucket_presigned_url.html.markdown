---
layout: "aws"
page_title: "AWS: aws_s3_bucket_presigned_url"
sidebar_current: "docs-aws-datasource-s3-bucket-presigned-url"
description: |-
    Creates a presigned URL for getting or putting an object to a S3 bucket
---

# Data Source: aws_s3_bucket_presigned_url

Creates a presigned URL for getting or putting an object to a S3 bucket.

This datasource may be useful when waiting to access or upload content to a S3 bucket without wanting to create an IAM role.

## Example Usage

### Writing to a presigned URL in a AWS instances user data

```hcl
data "aws_bucket_presigned_url" "presigned_url" {
  bucket = "ourcorp-deploy-config"
  key = "some/file.txt"
  expiration_time = 300 // 5 minutes
  put = true
}

resource "aws_instance" "instance" {
  instance_type = "t2.micro"
  ami           = "ami-2757f631"
  user_data = <<-EOF
    #!/bin/bash
    echo test >> file.txt
    curl -T file.txt ${data.aws_bucket_presigned_url.presigned_url.url} 
  EOF
}
```

### Downloading a file from a presigned URL in a AWS instances user data

```hcl
data "aws_bucket_presigned_url" "presigned_url" {
  bucket = "ourcorp-deploy-config"
  key = "some/file.txt"
  expiration_time = 300 // 5 minutes
}

resource "aws_instance" "instance" {
  instance_type = "t2.micro"
  ami           = "ami-2757f631"
  user_data = <<-EOF
    #!/bin/bash
    curl ${data.aws_bucket_presigned_url.presigned_url.url} >> file.txt
  EOF
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to create the presigned URL for.
* `key` - (Required) The full path to the object inside the bucket.
* `expiration_time` - (Required) The time in seconds the presigned URL will be valid.
* `put` - (Optional) Flag that determines if a get or put presigned URL is created. Defaults to `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `url` - The presigned URL that was generated.