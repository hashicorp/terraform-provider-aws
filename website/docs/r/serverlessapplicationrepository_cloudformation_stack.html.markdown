---
subcategory: "Serverless Application Repository"
layout: "aws"
page_title: "AWS: aws_serverlessapplicationrepository_cloudformation_stack"
description: |-
  Deploys an Application CloudFormation Stack from the Serverless Application Repository.
---

# Resource: aws_serverlessapplicationrepository_cloudformation_stack

Deploys an Application CloudFormation Stack from the Serverless Application Repository.

## Example Usage

### Basic Usage

```terraform
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = "postgres-rotator"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "func-postgres-rotator"
    endpoint     = "secretsmanager.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
```

### `NoEcho` parameters and state

CloudFormation masks the values of parameters declared with `NoEcho: true` in API responses, returning `****` instead of the configured value. Because the provider cannot read the actual value back from AWS, it persists the value from the Terraform configuration in state instead. This means `NoEcho` parameter values are stored in Terraform state in plaintext. Ensure your state backend is appropriately secured.

### Recovering from versions affected by issue #16485

In provider versions prior to the fix for [#16485](https://github.com/hashicorp/terraform-provider-aws/issues/16485), the resource produced false drift in two cases:

1. `parameters` whose value matched the application's default were pruned from state on Read.
2. `parameters` declared as `NoEcho` were stored in state as `****`.

After upgrading, **new resources are correct from the first apply.** Resources whose state was produced by an affected version retain the broken state, and the first apply against them will still fail with `change set: unexpected state 'FAILED' … No updates are to be performed`. A state upgrader cannot repair the state automatically because the SDKv2 state upgrade interface has no access to either the HCL configuration or the original `NoEcho` values.

Choose one of the following procedures to recover an affected resource.

#### Recovery option 1: Parameter perturbation (non-sensitive parameters)

Use this when the affected parameter does not control sensitive state (for example, a `passwordLength` left at its default). The resource is briefly updated in AWS but the end state matches your original configuration.

1. Temporarily change the affected parameter to a non-default value (for example, `passwordLength = "33"`).
2. Run `terraform apply`. The CloudFormation stack updates and state is repopulated for that parameter.
3. Revert the parameter to its intended value.
4. Run `terraform apply` again. The stack updates back to the intended value and state is now correct.
5. Run `terraform plan` to confirm no remaining diff.

For `NoEcho` parameters this is generally not appropriate because it would cause an actual change to the underlying secret. Use option 2 instead.

#### Recovery option 2: State surgery (no AWS-side change)

Use this when you cannot or do not want to perturb the value in AWS. This is the recommended path for `NoEcho` parameters.

1. Add `ignore_changes = [parameters]` to the resource's `lifecycle` block so subsequent plans do not fail on the existing drift.
2. Pull state to a local file:

    ```console
    % terraform state pull > stack.tfstate
    ```

3. Edit `stack.tfstate`. Locate the affected resource's `instances[].attributes.parameters` map and:
    - Add any missing default-matching keys with the value you have configured.
    - Replace any `****` placeholder values with the real values from your configuration.
    - Increment the top-level `serial` field by one.
4. Push the corrected state back:

    ```console
    % terraform state push stack.tfstate
    ```

5. Remove the `ignore_changes = [parameters]` lifecycle block.
6. Run `terraform plan` and confirm there is no remaining diff.

#### Recovery option 3: Recreate the resource

If neither option above is workable, recreate the stack. This is destructive and will incur whatever downtime or side effects the underlying CloudFormation stack delete and recreate cycle implies for your application.

```console
% terraform apply -replace="aws_serverlessapplicationrepository_cloudformation_stack.example"
```

#### If you cannot upgrade right away

If you must remain on an affected provider version, the long-standing workaround documented in the issue is to add `lifecycle { ignore_changes = [parameters] }` to the resource. The trade-off is that subsequent intentional changes to `parameters` will not be applied by Terraform until the block is removed.

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) ARN of the application from the Serverless Application Repository.
* `capabilities` - (Optional) List of capabilities. Valid values are `CAPABILITY_IAM`, `CAPABILITY_NAMED_IAM`, `CAPABILITY_RESOURCE_POLICY`, or `CAPABILITY_AUTO_EXPAND`. If the application contains IAM resources, IAM resources with custom names, resource-based policies, or nested applications, the corresponding capability must be specified. If omitted, the value applied by AWS is tracked in state.
* `name` - (Required) Name of the stack to create. The resource deployed in AWS will be prefixed with `serverlessrepo-`
* `parameters` - (Optional) Map of Parameter structures that specify input parameters for the stack. `NoEcho` parameter values are persisted in state in plaintext; see the [`NoEcho` parameters and state](#noecho-parameters-and-state) section above.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `semantic_version` - (Optional) Version of the application to deploy. If not supplied, deploys the latest version.
* `tags` - (Optional) Map of tags to associate with this stack. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier of the stack.
* `outputs` - Map of outputs from the stack.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Serverless Application Repository Stack using the CloudFormation Stack name (with or without the `serverlessrepo-` prefix) or the CloudFormation Stack ID. For example:

```terraform
import {
  to = aws_serverlessapplicationrepository_cloudformation_stack.example
  id = "serverlessrepo-postgres-rotator"
}
```

Using `terraform import`, import Serverless Application Repository Stack using the CloudFormation Stack name (with or without the `serverlessrepo-` prefix) or the CloudFormation Stack ID. For example:

```console
% terraform import aws_serverlessapplicationrepository_cloudformation_stack.example serverlessrepo-postgres-rotator
```
