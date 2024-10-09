# Exclusive Relationship Management Resources

**Summary:** A proposal describing the use case for "exclusive relationship management" resources and their function within the Terraform AWS provider.  
**Created**: 2024-09-06  
**Author**: [@jar-b](https://github.com/jar-b)  

---

Within AWS there are several resource types which have direct relationships or dependencies on one another. These can be either "one-to-one", where a single entity is linked to another, or "one-to-many", where a single parent entity can be linked to multiple children. As the Terraform AWS provider has matured, the patterns for modeling these relationships have evolved, most notably in the "one-to-many" style relationships.

This RFC will cover a brief history of "one-to-many" resource relationships and their representation in the Terraform AWS provider, followed by a proposal for enabling exclusive management of these relationships via a standalone resource. For an in-depth review of all relationship resources in the Terraform AWS provider, refer to the [relationship resource design standards](./relationship-resource-design-standards.md) design decision.

## Background

During early development of the Terraform AWS provider, resources sometimes represented "one-to-many" relationships as arguments on the parent resource. For example, the [`inline_policy`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_role#inline_policy) argument on the [`aws_iam_role`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_role) resource allows inline policies to be created as part of the role resource lifecycle.

```terraform
resource "aws_iam_role" "example" {
  name               = "exampleRole"
  assume_role_policy = data.aws_iam_policy_document.instance_assume_role_policy.json

  inline_policy {
    name   = "exampleInlinePolicy"
    policy = data.aws_iam_policy_document.inline_policy.json
  }
}
```

While simplifying the syntax for practitioners, this approach introduces complexity within the provider implementation; a single resource now manages the lifecycle of several different remote resources. This can leave the resource in a partially provisioned state if creation of one of the child resources fails. One benefit to this design is that it enables the parent resource to retain "exclusive" control of the relationships. That is, if the parent resource includes a relationship with a child entity that is not explicitly configured, the provider can remove it.

Due to the complexity of the design described above, the Terraform AWS provider has moved toward representing "one-to-many" relationships via standalone resources. These resources are typically added alongside the argument-based method to preserve backward compatibility. Continuing with the example above, inline policies can now be managed distinctly from the assigned role via the [`aws_iam_role_policy`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_role_policy) resource.

```terraform
resource "aws_iam_role" "example" {
  name               = "exampleRole"
  assume_role_policy = data.aws_iam_policy_document.instance_assume_role_policy.json
}

resource "aws_iam_role_policy" "example" {
  name = "exampleInlinePolicy"
  role = aws_iam_role.test_role.id

  policy = data.aws_iam_policy_document.inline_policy.json
}
```

While this separates the lifecycle of distinct resource types to more closely align with the underlying AWS APIs, this approach alone does not provide a mechanism for exclusive management of all relationships to the parent. The absence of this benefit is often [cited by the community](https://github.com/hashicorp/terraform-provider-aws/issues/22336#issuecomment-1586581208) as a reason not to move away from the argument-based definitions to standalone resources. The lack of parity has also prevented maintainers from formally deprecating and removing arguments on the parent resource for configuring relationships.

### Existing Resources

There is some precedent in the Terraform AWS provider for a resource which handles only exclusive relationship management in [`aws_iam_policy_attachment`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_policy_attachment). However, the scope of responsibility in this resource (managing attachments to roles, users, and groups simultaneously), is broader than what this RFC proposes, and cannot be replicated widely due to the unique nature of customer managed IAM policies which can be associated with multiple distinct parent resource types.

## Proposal

To provide practitioners with a consistent and maintainable mechanism for exclusive management of "one-to-many" relationships, the Terraform AWS provider will introduce a new pattern for developing "exclusive relationship management" resources. Resources with this function will end in the suffix `_exclusive`, with the purpose of reconciling the relationships present in AWS against what is configured in Terraform, adding and removing relationships as necessary.

In general, exclusive relationship management resources should have the following characteristics:

1. A required argument (typically `TypeString`) storing the parent resource identifier.
1. A required argument (typically `TypeSet` with `TypeString` elements) storing identifiers of all child resources. An empty set should remove all relationships.
1. The ability to read the current state of relationships in AWS and add or remove them as necessary.
1. The ability to "inherit" exclusive ownership of existing relationship definitions without destructive action. Adding this resource to a configuration should __not__ result in a destroy/re-create relationship workflow.

As an example, the implementation for inline IAM policies would be named `aws_iam_role_policies_exclusive`, and used as follows:

```terraform
resource "aws_iam_role" "example" {
  name               = "exampleRole"
  assume_role_policy = data.aws_iam_policy_document.instance_assume_role_policy.json
}

resource "aws_iam_role_policy" "example" {
  name = "exampleInlinePolicy"
  role = aws_iam_role.example.id

  policy = data.aws_iam_policy_document.inline_policy.json
}

# This resource ensures that only `exampleInlinePolicy` is assigned
# to this role. Any other inline policies will be removed.
resource "aws_iam_role_policies_exclusive" "example" {
  role_name    = aws_iam_role.example.name
  policy_names = [aws_iam_role_policy.example.name]
}
```

A working implementation can be found on [this branch](https://github.com/hashicorp/terraform-provider-aws/tree/f-iam_role_policies_lock). The `_exclusive` resource will detect any additional inline policies assigned to the role during `plan` operations and remove them during `apply`. This behavior provides parity with the `inline_policy` argument on the `aws_iam_role` resource, allowing maintainers to formally deprecate this argument and suggest practitioners migrate to the standalone inline policy resource, optionally including an `_exclusive` resource when exclusive management of assignments is desired.

### Potential Resources

In addition to the IAM inline policy example used throughout this document, there are several other resources with "one-to-many" relationships that could benefit from exclusive management of relationships.

!!! note
    Due to the popularity of the resources in this section, argument deprecations are likely to be "soft" deprecations where removal will not happen for several major releases, or until tooling is available to limit the amount of manual changes required to migrate to the preferred pattern. Despite this long removal window, a soft deprecation is still helpful for maintainers to reference when making best practice recommendations to the community.

#### Inline IAM policies to role

Manage inline IAM policies assigned to a role. Related resources:

- [`aws_iam_role`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_role)
- [`aws_iam_role_policy`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_role_policy)

Deprecate `aws_iam_role.inline_policy`.

#### Inline IAM policies to user

Manage inline IAM policies assigned to a user. Related resources:

- [`aws_iam_user`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_user)
- [`aws_iam_user_policy`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_user_policy)

There are no arguments to deprecate on `aws_iam_user`. This may lower the relative priority.

#### Inline IAM policies to group

Manage inline IAM policies assigned to a group. Related resources:

- [`aws_iam_group`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_group)
- [`aws_iam_group_policy`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_group_policy)

There are no arguments to deprecate on `aws_iam_group`. This may lower the relative priority.

#### Customer managed IAM policies to role

Manage customer managed IAM policies attached to a role. Related resources:

- [`aws_iam_role`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_role)
- [`aws_iam_policy`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_policy)
- [`aws_iam_role_policy_attachment`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_role_policy_attachment)

Deprecate `aws_iam_role.managed_policy_arns` .

#### Customer managed IAM policies to user

Manage customer managed IAM policies attached to a user. Related resources:

- [`aws_iam_user`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_user)
- [`aws_iam_policy`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_policy)
- [`aws_iam_user_policy_attachment`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_user_policy_attachment)

There are no arguments to deprecate on `aws_iam_user`. This may lower the relative priority.

#### Customer managed IAM policies to group

Manage customer managed IAM policies attached to a group. Related resources:

- [`aws_iam_group`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_group)
- [`aws_iam_policy`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_policy)
- [`aws_iam_group_policy_attachment`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/iam_group_policy_attachment)

There are no arguments to deprecate on `aws_iam_group`. This may lower the relative priority.

#### EC2 security groups to network interface

Manage EC2 security group attachments to network interfaces. Related resources:

- [`aws_instance`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/instance)
- [`aws_network_interface`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/network_interface)
- [`aws_security_group`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/security_group)
- [`aws_ec2_network_interface_sg_attachment`](https://registry.terraform.io/providers/-/aws/latest/docs/resources/network_interface_sg_attachment)

Deprecate `aws_instance.security_groups`, `aws_instance.vpc_security_group_ids`, and `aws_network_interface.security_groups`.

#### VPC security group rules to security group

Manage ingress and egress rules assigned to a VPC security group. Related resources:

- [`aws_security_group`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group)
- [`aws_security_group_rule`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule)
- [`aws_vpc_security_group_egress_rule`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_security_group_egress_rule)
- [`aws_vpc_security_group_ingress_rule`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_security_group_ingress_rule)

Deprecate `aws_security_group.ingress` and `aws_security_group.egress`. Consider deprecating `aws_security_group_rule` entirely.

### Alternate Naming Conventions

Ideally all resources providing the "exclusive relationship management" function should utilize the same suffix. While there are several options to describe this behavior, `_exclusive` seemed the most concise. The options considered were:

| Suffix | Example |
| --- | --- |
| `_exclusive` _(selected)_ | `aws_iam_role_policies_exclusive` |
| `_management` | `aws_iam_role_policies_management` |
| `_lock` | `aws_iam_role_policies_lock` |
| `_exclusive_lock` | `aws_iam_role_policies_exclusive_lock` |
| `_exclusive_management` | `aws_iam_role_policies_exclusive_management` |

## Next Steps

If this design decision is approved, a meta-issue will be opened to track all resources listed in the [Potential Resources](#potential-resources) section, with linked issues for each individual implementation.
