<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Adding `go-vcr` Support

You are working on the [Terraform AWS Provider](https://github.com/hashicorp/terraform-provider-aws), specifically focused on enabling support for `go-vcr`.

Follow the steps below to enable support for a single service.

- The working branch name should begin with `f-go-vcr-` and be suffixed with the name of the service being updated, e.g. `f-go-vcr-s3`. If the current branch does not match this convention, create one. Ensure the branch is rebased with the `main` branch.
- Follow the steps on [this page](../go-vcr.md) to enable `go-vcr` for the target service.
- Once all acceptance tests are passing, commit the changes with a message like "service-name: enable `go-vcr` support", replacing `service-name` with the target service. Be sure to include the COMPLETE output from acceptance testing in the commit body, wrapped in a `console` code block. e.g.

```console
% make testacc PKG=polly VCR_MODE=REPLAY_ONLY VCR_PATH=/tmp/polly-vcr-testdata/

Enables `go-vcr` for the `<service-name>` service.

<-- full results here -->
```
