---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_principal_policy_simulation"
description: |-
  Runs a simulation of the IAM policies of a particular principal against a given hypothetical request.
---

# Data Source: aws_iam_principal_policy_simulation

Runs a simulation of the IAM policies of a particular principal against a given hypothetical request.

You can use this data source in conjunction with
[Preconditions and Postconditions](https://www.terraform.io/language/expressions/custom-conditions#preconditions-and-postconditions) so that your configuration can test either whether it should have sufficient access to do its own work, or whether policies your configuration declares itself are sufficient for their intended use elsewhere.

-> **Note:** Correctly using this data source requires familiarity with various details of AWS Identity and Access Management, and how various AWS services integrate with it. For general information on the AWS IAM policy simulator, see [Testing IAM policies with the IAM policy simulator](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_testing-policies.html). This data source wraps the `iam:SimulatePrincipalPolicy` API action described on that page.

## Example Usage

### Self Access-checking Example

The following example raises an error if the credentials passed to the AWS provider do not have access to perform the three actions `s3:GetObject`, `s3:PutObject`, and `s3:DeleteObject` on the S3 bucket with the given ARN. It combines `aws_iam_principal_policy_simulation` with the core Terraform postconditions feature.

```terraform
data "aws_caller_identity" "current" {}

data "aws_iam_principal_policy_simulation" "s3_object_access" {
  action_names = [
    "s3:GetObject",
    "s3:PutObject",
    "s3:DeleteObject",
  ]
  policy_source_arn = data.aws_caller_identity.current.arn
  resource_arns     = ["arn:aws:s3:::my-test-bucket"]

  # The "lifecycle" and "postcondition" block types are part of
  # the main Terraform language, not part of this data source.
  lifecycle {
    postcondition {
      condition     = self.all_allowed
      error_message = <<EOT
        Given AWS credentials do not have sufficient access to manage ${join(", ", self.resource_arns)}.
      EOT
    }
  }
}
```

If you intend to use this data source to quickly raise an error when the given credentials are insufficient then you must use [`depends_on`](https://www.terraform.io/language/meta-arguments/depends_on) inside any resource which would require those credentials, to ensure that the policy check will run first:

```terraform
resource "aws_s3_bucket_object" "example" {
  bucket = "my-test-bucket"
  # ...

  depends_on = [data.aws_iam_principal_policy_simulation.s3_object_access]
}
```

### Testing the Effect of a Declared Policy

The following example declares an S3 bucket and a user that should have access to the bucket, and then uses `aws_iam_principal_policy_simulation` to verify that the user does indeed have access to perform needed operations against the bucket.

```terraform
data "aws_caller_identity" "current" {}

resource "aws_iam_user" "example" {
  name = "example"
}

resource "aws_s3_bucket" "example" {
  bucket = "my-test-bucket"
}

resource "aws_iam_user_policy" "s3_access" {
  name = "example_s3_access"
  user = aws_iam_user.example.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = "s3:GetObject"
        Effect   = "Allow"
        Resource = aws_s3_bucket.example.arn
      },
    ]
  })
}

resource "aws_s3_bucket_policy" "account_access" {
  bucket = aws_s3_bucket.example.bucket
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "s3:*"
        Effect = "Allow"
        Principal = {
          # Any caller belonging to the current AWS account
          # is allowed full access to this S3 bucket.
          AWS = data.aws_caller_identity.current.account_id
        }
        Resource = [
          aws_s3_bucket.example.arn,
          "${aws_s3_bucket.example.arn}/*",
        ]
      },
    ]
  })
}

data "aws_iam_principal_policy_simulation" "s3_object_access" {
  action_names = [
    "s3:GetObject",
  ]
  policy_source_arn = aws_iam_user.example.arn
  resource_arns     = [aws_s3_bucket.example.arn]

  # The IAM policy simulator automatically imports the policies associated
  # with the user (policy_source_arn) but cannot automatically import the
  # policies from the S3 bucket, because in a real request those would be
  # provided by the S3 service itself. Therefore we need to provide the
  # same policy that we associated with the S3 bucket above.
  resource_policy_json = aws_s3_bucket_policy.account_access.policy

  # The policy simulation won't be valid until the user's policy
  # has been created or updated.
  depends_on = [aws_iam_user_policy.s3_access]

  # The "lifecycle" and "postcondition" block types are part of
  # the main Terraform language, not part of this data source.
  lifecycle {
    postcondition {
      condition     = self.all_allowed
      error_message = <<EOT
        ${self.policy_source_arn} does not have the expected access to ${join(", ", self.resource_arns)}.
      EOT
    }
  }
}
```

When using `aws_iam_principal_policy_simulation` to test the effect of a policy declared elsewhere in the same configuration, it's important to use [`depends_on`](https://www.terraform.io/language/meta-arguments/depends_on) to make sure that the needed policy has been fully created or updated before running the simulation.

## Argument Reference

The following arguments are required for any principal policy simulation:

* `action_names` (Required) - A set of IAM action names to run simulations for. Each entry in this set adds an additional hypothetical request to the simulation.

    Action names consist of a service prefix and an action verb separated by a colon, such as `s3:GetObject`. Refer to [Actions, resources, and condition keys for AWS services](https://docs.aws.amazon.com/service-authorization/latest/reference/reference_policies_actions-resources-contextkeys.html) to see the full set of possible IAM action names across all AWS services.

* `policy_source_arn` (Required) - The [ARN](https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html) of the IAM user, group, or role whose policies will be included in the simulation.

You must closely match the form of the real service request you are simulating in order to achieve a realistic result. You can use the following additional arguments to specify other characteristics of the simulated requests:

* `caller_arn` (Optional) - The ARN of an user that will appear as the "caller" of the simulated requests. If you do not specify `caller_arn` then the simulation will use the `policy_source_arn` instead, if it contains a user ARN.

* `context` (Optional) - Each [`context` block](#context-block-arguments) defines an entry in the table of additional context keys in the simulated request.

    IAM uses context keys for both custom conditions and for interpolating dynamic request-specific values into policy values. If you use policies that include those features then you will need to provide suitable example values for those keys to achieve a realistic simulation.

* `additional_policies_json` (Optional) - A set of additional principal policy documents to include in the simulation. The simulator will behave as if each of these policies were associated with the object specified in `policy_source_arn`, allowing you to test the effect of hypothetical policies not yet created.

* `permissions_boundary_policies_json` (Optional) - A set of [permissions boundary policy documents](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_boundaries.html) to include in the simulation.

* `resource_arns` (Optional) - A set of ARNs of resources to include in the simulation.

    This argument is important for actions that have either required or optional resource types listed in [Actions, resources, and condition keys for AWS services](https://docs.aws.amazon.com/service-authorization/latest/reference/reference_policies_actions-resources-contextkeys.html), and you must provide ARNs that identify AWS objects of the appropriate types for the chosen actions.

    The policy simulator only automatically loads policies associated with the `policy_source_arn`, so if your given resources have their own resource-level policy then you'll also need to provide that explicitly using the `resource_policy_json` argument to achieve a realistic simulation.

* `resource_handling_option` (Optional) - Specifies a special simulation type to run. Some EC2 actions require special simulation behaviors and a particular set of resource ARNs to achieve a realistic result.

    For more details, see the `ResourceHandlingOption` request parameter for [the underlying `iam:SimulatePrincipalPolicy` action](https://docs.aws.amazon.com/IAM/latest/APIReference/API_SimulatePrincipalPolicy.html).

* `resource_owner_account_id` (Optional) - An AWS account ID to use for any resource ARN in `resource_arns` that doesn't include its own AWS account ID. If unspecified, the simulator will use the account ID from the `caller_arn` argument as a placeholder.

* `resource_policy_json` (Optional) - An IAM policy document representing the resource-level policy of all of the resources specified in `resource_arns`.

    The policy simulator cannot automatically load policies that are associated with individual resources, as described in the documentation for `resource_arns` above.

### `context` block arguments

The following arguments are all required in each `context` block:

* `key` (Required) - The context _condition key_ to set.

    If you have policies containing `Condition` elements or using dynamic interpolations then you will need to provide suitable values for each condition key your policies use. See [Actions, resources, and condition keys for AWS services](https://docs.aws.amazon.com/service-authorization/latest/reference/reference_policies_actions-resources-contextkeys.html) to find the various condition keys that are normally provided for real requests to each action of each AWS service.

* `type` (Required) - An IAM value type that determines how the policy simulator will interpret the strings given in `values`.

    For more information, see the `ContextKeyType` field of [`iam.ContextEntry`](https://docs.aws.amazon.com/IAM/latest/APIReference/API_ContextEntry.html) in the underlying API.

* `values` (Required) - A set of one or more values for this context entry.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `all_allowed` - `true` if all of the simulation results have decision "allowed", or `false` otherwise.

    This is a convenient shorthand for the common case of requiring that all of the simulated requests passed in a postcondition associated with the data source. If you need to describe a more granular condition, use the `results` attribute instead.

* `results` - A set of result objects, one for each of the simulated requests, with the following nested attributes:

    * `action_name` - The name of the single IAM action used for this particular request.

    * `decision` - The raw decision determined from all of the policies in scope; either "allowed", "explicitDeny", or "implicitDeny".

    * `allowed` - `true` if `decision` is "allowed", and `false` otherwise.

    * `decision_details` - A map of arbitrary metadata entries returned by the policy simulator for this request.

    * `resource_arn` - ARN of the resource that was used for this particular request. When you specify multiple actions and multiple resource ARNs, that causes a separate policy request for each combination of unique action and resource.

    * `matched_statements` - A nested set of objects describing which policies contained statements that were relevant to this simulation request. Each object has attributes `source_policy_id` and `source_policy_type` to identify one of the policies.

    * `missing_context_keys` - A set of context keys (or condition keys) that were needed by some of the policies contributing to this result but not specified using a `context` block in the configuration. Missing or incorrect context keys will typically cause a simulated request to be disallowed.
