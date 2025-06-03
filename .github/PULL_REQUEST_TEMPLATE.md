<!---
See what makes a good Pull Request at: https://hashicorp.github.io/terraform-provider-aws/raising-a-pull-request/
--->

<!-- heimdall_github_prtemplate:grc-pci_dss-2024-01-05 -->

## Rollback Plan

If a change needs to be reverted, we will publish an updated version of the library.

## Changes to Security Controls

Are there any changes to security controls (access controls, encryption, logging) in this pull request? If so, explain.

### Description
<!---
Please provide a helpful description of what change this pull request will introduce.
--->


### Relations
<!---
If your pull request fully resolves and should automatically close the linked issue, use Closes. Otherwise, use Relates.

For Example:

Relates #0000
or 
Closes #0000
--->

Closes #0000

### References
<!---
Optionally, provide any helpful references that may help the reviewer(s).
--->


### Output from Acceptance Testing
<!--
Replace TestAccXXX with a pattern that matches the tests affected by this PR.

Replace ec2 with the service package corresponding to your tests.

For more information on the `-run` flag, see the `go test` documentation at https://tip.golang.org/cmd/go/#hdr-Testing_flags.
-->

```console
% make testacc TESTS=TestAccXXX PKG=ec2

...
```
