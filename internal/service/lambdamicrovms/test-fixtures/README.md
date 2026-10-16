  <!-- Copyright IBM Corp. 2014, 2026 -->
  <!-- SPDX-License-Identifier: MPL-2.0 -->

# Test Fixtures

## code.zip

`code.zip` is the code artifact uploaded to S3 by the acceptance tests and referenced
by the `code_artifact` block of `aws_lambdamicrovms_image`. It contains the files in
`code/` (a minimal Go program and the `Dockerfile` that builds it) at the root of the
archive — the service requires the `Dockerfile` at the archive root, not inside a
subdirectory.

To regenerate after changing anything in `code/`:

```console
% cd code && zip ../code.zip main.go Dockerfile
```

## code-hooks.zip

`code-hooks.zip` is the code artifact used by tests that enable lifecycle hooks
(`aws_lambdamicrovms_image` with a `hooks` block). It contains the files in
`code-hooks/`: like `code/`, plus an HTTP server on port 9000 that answers the
lifecycle endpoints (`POST /aws/lambda-microvms/runtime/v1/<hook>`). Images built
with hooks enabled fail (`CREATE_FAILED`) unless the application answers the
`/ready` build hook, so the plain `code.zip` fixture cannot be used.

To regenerate after changing anything in `code-hooks/`:

```console
% cd code-hooks && zip ../code-hooks.zip main.go Dockerfile
```
