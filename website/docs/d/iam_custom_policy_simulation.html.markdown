---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_custom_policy_simulation"
description: |-
  Simulates the effect of custom IAM policies.
---

# Data Source: aws_iam_custom_policy_simulation

Runs a simulation of the IAM policies provided as input against a set of API operations and optionally resource ARNs, using the [SimulateCustomPolicy](https://docs.aws.amazon.com/IAM/latest/APIReference/API_SimulateCustomPolicy.html) API.

Unlike [`aws_iam_principal_policy_simulation`](/docs/providers/aws/d/iam_principal_policy_simulation.html.markdown), this data source evaluates arbitrary inline policy documents without requiring an existing IAM principal.

## Example Usage

### Basic Usage

```terraform
data "aws_iam_custom_policy_simulation" "example" {
  action_names  = ["s3:GetObject"]
  resource_arns = ["arn:aws:s3:::my-bucket/*"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:GetObject"
      Resource = "arn:aws:s3:::my-bucket/*"
    }]
  })]
}
```

### Using as a Postcondition

```terraform
data "aws_iam_custom_policy_simulation" "check" {
  action_names  = ["s3:GetObject"]
  resource_arns = ["arn:aws:s3:::my-bucket/*"]

  policies_json = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "s3:GetObject"
      Resource = "arn:aws:s3:::my-bucket/*"
    }]
  })]

  lifecycle {
    postcondition {
      condition     = self.all_allowed
      error_message = "Policy does not allow the required access."
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `action_names` - (Required) Set of IAM action names to simulate, e.g. `s3:GetObject`.
* `policies_json` - (Required) Set of IAM policy documents (as JSON strings) to use as the policies for the simulation.

The following arguments are optional:

* `caller_arn` - (Optional) ARN of an IAM user to use as the caller of the simulated requests. Required if `resource_policy_json` is specified.
* `context` - (Optional) Each block specifies one context entry for condition evaluation. See [`context` Block](#context-block) below.
* `permissions_boundary_policies_json` - (Optional) Set of IAM permissions boundary policy documents (as JSON strings) to use in the simulation.
* `resource_arns` - (Optional) Set of ARNs of resources to include in the simulation. If not specified, the simulator assumes `*`.
* `resource_handling_option` - (Optional) Specifies the type of simulation to run for resource-level scenarios.
* `resource_owner_account_id` - (Optional) AWS account ID to use as the simulated owner for any resource whose ARN does not include a specific owner account ID.
* `resource_policy_json` - (Optional) Resource-based policy (as a JSON string) to associate with all target resources for simulation purposes.

### `context` Block

* `key` - (Required) Context key name, such as `aws:CurrentTime`.
* `type` - (Required) Type for the simulator to interpret the context values, such as `string`, `numeric`, `boolean`, `ip`, `date`, or their list variants.
* `values` - (Required) Set of values to assign to the context key.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `all_allowed` - `true` if every result in `results` has a decision of `allowed`.
* `results` - List of result objects, one per action-resource pair evaluated. See [`results` Attribute Reference](#results-attribute-reference) below.

### `results` Attribute Reference

* `action_name` - The action name that was tested.
* `allowed` - `true` if the `decision` is `allowed`.
* `decision` - The raw decision: `allowed`, `explicitDeny`, or `implicitDeny`.
* `decision_details` - Map of additional details about the decision.
* `matched_statements` - List of statements that contributed to the decision. See [`matched_statements` Attribute Reference](#matched_statements-attribute-reference) below.
* `missing_context_keys` - Set of context keys that were needed by the policies but not provided.
* `resource_arn` - ARN of the resource that the action was tested against.

### `matched_statements` Attribute Reference

* `source_policy_id` - Identifier of the policy that contained the matched statement.
* `source_policy_type` - Type of the policy that contained the matched statement.
