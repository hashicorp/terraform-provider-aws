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
