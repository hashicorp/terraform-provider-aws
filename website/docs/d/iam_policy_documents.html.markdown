---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_policy_documents"
description: Provides details about one or more generated IAM policy documents in JSON format, split across documents to stay within a size limit
---

# Data Source: aws_iam_policy_documents

Provides details about IAM policy documents in JSON format. This data source generates one or more policy documents for use with resources that expect policy documents such as [`aws_iam_policy`](/docs/providers/aws/r/iam_policy.html). It accepts the same arguments as [`aws_iam_policy_document`](/docs/providers/aws/d/iam_policy_document.html), then packs the merged statements into as many documents as needed so that no document exceeds `max_policy_size`.

Statements keep their order. Each document is filled until the next statement would make it too large, and that statement starts a new document. Adding a statement to the end of `statement` therefore only changes the last document, unless its `sid` matches an earlier statement (which it replaces in place) or an override document adds new statements after it.

-> **Size limits:** Size is measured on `minified_json`. AWS does not count whitespace towards policy size, so either `json` or `minified_json` can be passed to a resource, but `minified_json` matches how this data source measures size. Size is measured in bytes, which is never less than the number of characters AWS counts. It can be larger for non-ASCII characters and for characters such as `&`, `<` and `>`, which are escaped in the JSON output. If a single statement is larger than `max_policy_size` on its own, this data source returns an error.

AWS enforces the following [IAM character limits](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-quotas.html#reference_iam-quotas-entity-length):

* Customer managed policies: 6,144 characters per policy.
* Inline policies: 2,048 characters for users, 5,120 for groups and 10,240 for roles.

~> **NOTE:** Inline policy limits apply to the total size of all inline policies attached to a user, group or role, not to each policy. Setting `max_policy_size` to the inline limit does not guarantee that several generated documents fit when attached as inline policies to the same principal.

~> **NOTE:** AWS's IAM policy document syntax allows for replacement of [policy variables](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_variables.html) within a statement using `${...}`-style notation, which conflicts with Terraform's interpolation syntax. In order to use AWS policy variables with this data source, use `&{...}` notation for interpolations that should be processed by AWS rather than by Terraform.

## Example Usage

### Customer Managed Policies

```terraform
data "aws_iam_policy_documents" "example" {
  dynamic "statement" {
    for_each = var.bucket_names

    content {
      sid       = "Read${replace(statement.value, "-", "")}"
      actions   = ["s3:GetObject", "s3:ListBucket"]
      resources = ["arn:aws:s3:::${statement.value}", "arn:aws:s3:::${statement.value}/*"]
    }
  }
}

resource "aws_iam_policy" "example" {
  count = length(data.aws_iam_policy_documents.example.minified_json)

  name   = "example-${count.index}"
  policy = data.aws_iam_policy_documents.example.minified_json[count.index]
}

resource "aws_iam_role_policy_attachment" "example" {
  count = length(aws_iam_policy.example)

  role       = aws_iam_role.example.name
  policy_arn = aws_iam_policy.example[count.index].arn
}
```

### Splitting Existing Policy Documents

```terraform
data "aws_iam_policy_documents" "example" {
  source_policy_documents = [
    data.aws_iam_policy_document.s3.json,
    data.aws_iam_policy_document.dynamodb.json,
  ]
}
```

## Argument Reference

This data source supports the following arguments:

~> **NOTE:** Statements without a `sid` cannot be overridden. In other words, a statement without a `sid` from `source_policy_documents` cannot be overridden by statements from `override_policy_documents`.

* `max_policy_size` - (Optional) Maximum size, in characters, of each generated policy document. Defaults to `6144`, the limit for customer managed policies. Size is measured on the minified JSON as exported in `minified_json`. See the size limits note above.
* `override_policy_documents` - (Optional) List of IAM policy documents that are merged together into the exported documents. In merging, statements with non-blank `sid`s will override statements with the same `sid` from earlier documents in the list. Statements with non-blank `sid`s will also override statements with the same `sid` from `source_policy_documents`. Non-overriding statements will be added to the exported documents.
* `policy_id` - (Optional) ID for the policy documents. Every generated document uses the same ID.
* `source_policy_documents` - (Optional) List of IAM policy documents that are merged together into the exported documents. Statements defined in `source_policy_documents` must have unique `sid`s. Statements with the same `sid` from `override_policy_documents` will override source statements.
* `statement` - (Optional) Configuration block for a policy statement. Detailed below.
* `version` - (Optional) IAM policy document version. Valid values are `2008-10-17` and `2012-10-17`. Defaults to `2012-10-17`. For more information, see the [AWS IAM User Guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_version.html).

### `statement` Block

The following arguments are optional:

* `actions` - (Optional) List of actions that this statement either allows or denies. For example, `["ec2:RunInstances", "s3:*"]`.
* `condition` - (Optional) Configuration block for a condition. Detailed below.
* `effect` - (Optional) Whether this statement allows or denies the given actions. Valid values are `Allow` and `Deny`. Defaults to `Allow`.
* `not_actions` - (Optional) List of actions that this statement does *not* apply to. Use to apply a policy statement to all actions *except* those listed.
* `not_principals` - (Optional) Like `principals` except these are principals that the statement does *not* apply to.
* `not_resources` - (Optional) List of resource ARNs that this statement does *not* apply to. Use to apply a policy statement to all resources *except* those listed. Conflicts with `resources`.
* `principals` - (Optional) Configuration block for principals. Detailed below.
* `resources` - (Optional) List of resource ARNs that this statement applies to. This is required by AWS if used for an IAM policy. Conflicts with `not_resources`.
* `sid` - (Optional) Sid (statement ID) is an identifier for a policy statement.

### `condition` Block

A `condition` constrains whether a statement applies in a particular situation. Conditions can be specific to an AWS service. When using multiple `condition` blocks, they must *all* evaluate to true for the policy statement to apply. In other words, AWS evaluates the conditions as though with an "AND" boolean operation.

The following arguments are required:

* `test` - (Required) Name of the [IAM condition operator](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition_operators.html) to evaluate.
* `values` - (Required) Values to evaluate the condition against. If multiple values are provided, the condition matches if at least one of them applies. That is, AWS evaluates multiple values as though using an "OR" boolean operation.
* `variable` - (Required) Name of a [Context Variable](http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements.html#AvailableKeys) to apply the condition to. Context variables may either be standard AWS variables starting with `aws:` or service-specific variables prefixed with the service name.

### `principals` Block

The `principals` block defines to whom a statement applies.

~> **NOTE:** Even though the [IAM Documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html) states that `"Principal": "*"` and `"Principal": {"AWS": "*"}` are equivalent, those principal elements have different behavior in some situations, e.g., IAM Role Trust Policy. To have Terraform render JSON containing `"Principal": "*"`, use `type = "*"` and `identifiers = ["*"]`. To have Terraform render JSON containing `"Principal": {"AWS": "*"}`, use `type = "AWS"` and `identifiers = ["*"]`.

The following arguments are required:

* `identifiers` - (Required) List of identifiers for principals. When `type` is `AWS`, these are IAM principal ARNs, e.g., `arn:aws:iam::12345678901:role/yak-role`. When `type` is `Service`, these are AWS Service roles, e.g., `lambda.amazonaws.com`. When `type` is `Federated`, these are web identity users or SAML provider ARNs, e.g., `accounts.google.com` or `arn:aws:iam::12345678901:saml-provider/yak-saml-provider`. When `type` is `CanonicalUser`, these are [canonical user IDs](https://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html#FindingCanonicalId), e.g., `79a59df900b949e55d96a1e698fbacedfd6e09d98eacf8f8d5218e7cd47ef2be`.
* `type` - (Required) Type of principal. Valid values include `AWS`, `Service`, `Federated`, `CanonicalUser` and `*`.

### `not_principals` Block

The `not_principals` block defines to whom a statement does not apply.

The following arguments are required:

* `identifiers` - (Required) List of identifiers for principals. Accepts the same values as `identifiers` in the `principals` block.
* `type` - (Required) Type of principal. Valid values include `AWS`, `Service`, `Federated`, `CanonicalUser` and `*`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `json` - List of standard JSON policy documents rendered based on the arguments above. Elements correspond by index to `minified_json`.
* `minified_json` - List of minified JSON policy documents rendered based on the arguments above. Each element is at most `max_policy_size` characters.
