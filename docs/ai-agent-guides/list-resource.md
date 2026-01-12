<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Adding `List Resource` Support

You are working on the [Terraform AWS Provider](https://github.com/hashicorp/terraform-provider-aws), specifically focused on adding `List Resource` support.

Follow the steps below to add a new List Resource.

- The working branch name should begin with `f-list-resource-` and be suffixed with the name of the resource being added, e.g. `f-list-resource-ec2_instance`. If the current branch does not match this convention, create one. Ensure the branch is rebased with the `main` branch.
- Navigate to the folder for the target resource. For example, to add a List Resource for `aws_instance`, navigate to `internal/service/ec2/`.
- Follow the steps on [this page](../add-a-new-list-resource.md) to add the List Resource for the target resource.
