---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_website_configuration"
description: |-
  Provides an S3 bucket website configuration resource.
---

# Resource: aws_s3_bucket_website_configuration

Provides an S3 bucket website configuration resource. For more information, see [Hosting Websites on S3](https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html).

## Example Usage

### With `routing_rule` configured

```terraform
resource "aws_s3_bucket_website_configuration" "example" {
  bucket = aws_s3_bucket.example.bucket

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rule {
    condition {
      key_prefix_equals = "docs/"
    }
    redirect {
      replace_key_prefix_with = "documents/"
    }
  }
}
```

### With `routing_rules` configured

```terraform
resource "aws_s3_bucket_website_configuration" "example" {
  bucket = aws_s3_bucket.example.bucket

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rules = <<EOF
[{
    "Condition": {
        "KeyPrefixEquals": "docs/"
    },
    "Redirect": {
        "ReplaceKeyPrefixWith": ""
    }
}]
EOF
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required, Forces new resource) The name of the bucket.
* `error_document` - (Optional, Conflicts with `redirect_all_requests_to`) The name of the error document for the website [detailed below](#error_document).
* `expected_bucket_owner` - (Optional, Forces new resource) The account ID of the expected bucket owner.
* `index_document` - (Optional, Required if `redirect_all_requests_to` is not specified) The name of the index document for the website [detailed below](#index_document).
* `redirect_all_requests_to` - (Optional, Required if `index_document` is not specified) The redirect behavior for every request to this bucket's website endpoint [detailed below](#redirect_all_requests_to). Conflicts with `error_document`, `index_document`, and `routing_rule`.
* `routing_rule` - (Optional, Conflicts with `redirect_all_requests_to` and `routing_rules`) List of rules that define when a redirect is applied and the redirect behavior [detailed below](#routing_rule).
* `routing_rules` - (Optional, Conflicts with `routing_rule` and `redirect_all_requests_to`) A json array containing [routing rules](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-websiteconfiguration-routingrules.html)
  describing redirect behavior and when redirects are applied. Use this parameter when your routing rules contain empty String values (`""`) as seen in the [example above](#with-routing_rules-configured).

### error_document

The `error_document` configuration block supports the following arguments:

* `key` - (Required) The object key name to use when a 4XX class error occurs.

### index_document

The `index_document` configuration block supports the following arguments:

* `suffix` - (Required) A suffix that is appended to a request that is for a directory on the website endpoint.
For example, if the suffix is `index.html` and you make a request to `samplebucket/images/`, the data that is returned will be for the object with the key name `images/index.html`.
The suffix must not be empty and must not include a slash character.

### redirect_all_requests_to

The `redirect_all_requests_to` configuration block supports the following arguments:

* `host_name` - (Required) Name of the host where requests are redirected.
* `protocol` - (Optional) Protocol to use when redirecting requests. The default is the protocol that is used in the original request. Valid values: `http`, `https`.

### routing_rule

The `routing_rule` configuration block supports the following arguments:

* `condition` - (Optional) A configuration block for describing a condition that must be met for the specified redirect to apply [detailed below](#condition).
* `redirect` - (Required) A configuration block for redirect information [detailed below](#redirect).

### condition

The `condition` configuration block supports the following arguments:

* `http_error_code_returned_equals` - (Optional, Required if `key_prefix_equals` is not specified) The HTTP error code when the redirect is applied. If specified with `key_prefix_equals`, then both must be true for the redirect to be applied.
* `key_prefix_equals` - (Optional, Required if `http_error_code_returned_equals` is not specified) The object key name prefix when the redirect is applied. If specified with `http_error_code_returned_equals`, then both must be true for the redirect to be applied.

### redirect

The `redirect` configuration block supports the following arguments:

* `host_name` - (Optional) The host name to use in the redirect request.
* `http_redirect_code` - (Optional) The HTTP redirect code to use on the response.
* `protocol` - (Optional) Protocol to use when redirecting requests. The default is the protocol that is used in the original request. Valid values: `http`, `https`.
* `replace_key_prefix_with` - (Optional, Conflicts with `replace_key_with`) The object key prefix to use in the redirect request. For example, to redirect requests for all pages with prefix `docs/` (objects in the `docs/` folder) to `documents/`, you can set a `condition` block with `key_prefix_equals` set to `docs/` and in the `redirect` set `replace_key_prefix_with` to `/documents`.
* `replace_key_with` - (Optional, Conflicts with `replace_key_prefix_with`) The specific object key to use in the redirect request. For example, redirect request to `error.html`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.
* `website_domain` - The domain of the website endpoint. This is used to create Route 53 alias records.
* `website_endpoint` - The website endpoint.

## Import

S3 bucket website configuration can be imported in one of two ways.

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider,
the S3 bucket website configuration resource should be imported using the `bucket` e.g.,

```
$ terraform import aws_s3_bucket_website_configuration.example bucket-name
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider,
the S3 bucket website configuration resource should be imported using the `bucket` and `expected_bucket_owner` separated by a comma (`,`) e.g.,

```
$ terraform import aws_s3_bucket_website_configuration.example bucket-name,123456789012
```
