<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Adding `List Resource` Support

You are working on the [Terraform AWS Provider](https://github.com/hashicorp/terraform-provider-aws), specifically focused on adding `List Resource` support.

Follow the steps below to add a new List Resource.

- The git working branch name should begin with `f-list-resource-` and be suffixed with the name of the resource being added, e.g. `f-list-resource-ec2_instance`. If the current branch does not match this convention, create one. Ensure the branch is rebased with the `main` branch. Check if the branch already exists on the remote. If it does, do not proceed with this guide.
- Compile the latest `skaff` tool by running `make skaff` from the root of the repository.
- Navigate to the folder for the target resource. For example, to add a List Resource for `aws_instance`, navigate to `internal/service/ec2/`.
- Follow the steps on [this page](../add-a-new-list-resource.md) to add the List Resource for the target resource. Continue with the next step when the List Resource implementation is complete.
- Run acceptance tests for the new List Resource to ensure it functions correctly.
- Once acceptance tests pass, commit the changes with a message "add new list resource for `<service-name>` `<resource-name>`", replacing `<service-name>` and `<resource-name>` with the target service and resource. Be sure to include the COMPLETE output from acceptance testing in the commit body, wrapped in a `console` code block. e.g.

```console
% make testacc PKG=<service> TESTARGS='-run=TestAcc<service><resource-name>_List_'

<-- full results here -->
```

- After creating the commit, create a pull request with the changes. Use the Pull Request template provided by the repository. Fill the `Description` section with a summary of the changes. Also, add the test output from the commit message to `Output from Acceptance Testing` section.
- Create a new CHANGELOG entry in the `.changelog/` folder. The filename will be `<pr-number>.txt`. Replace `<pr-number>` with the pull request number. The content of the file should be as follows, replacing `aws_<service-name>_<resource-name>` with the name of the new List Resource added.

```release-note:new-list-resource
aws_<service-name>_<resource-name>
```
