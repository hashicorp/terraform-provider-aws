# Local AWS provider: ECS sigint_rollback fix

## Bug

`sigint_rollback = true` started `rollbackRoutine` on the **per-refresh**
context created by `retry.refreshWithTimeout` (`context.WithTimeout` +
`defer cancel()`). When each poll returned, that child context was cancelled,
the routine treated it as SIGINT, and called `StopServiceDeployment(Rollback)`.
AWS then reported `Service deployment rolled back by user` within ~20s —
even for healthy rolling deploys (seen on Braintrust prod-eu gateway).

## Fix

- Attach `rollbackRoutine` to the parent `waitServiceStable` context.
- When checking whether to skip rollback, use raw AWS deployment status
  (not the `tfSTABLE` sentinel from `findDeploymentStatus`).

## Build + Terraform override

```bash
cd ~/dev/src/hashicorp/terraform-provider-aws
go build -o "$HOME/.terraform.d/plugins/registry.terraform.io/hashicorp/aws/99.0.0/$(go env GOOS)_$(go env GOARCH)/terraform-provider-aws" .

mkdir -p ~/.terraform.d
cat > ~/.terraform.d/dev-aws.tfrc <<EOF
provider_installation {
  dev_overrides {
    "hashicorp/aws" = "$HOME/.terraform.d/plugins/registry.terraform.io/hashicorp/aws/99.0.0/$(go env GOOS)_$(go env GOARCH)"
  }
  direct {}
}
EOF
export TF_CLI_CONFIG_FILE="$HOME/.terraform.d/dev-aws.tfrc"
```

`terraform init` will warn about overrides; that is expected.

## Sandbox repro / verify

1. On data-plane module, temporarily set gateway `sigint_rollback = true`
   (today it is `false` after the mitigation).
2. Force a gateway task-definition change (e.g. bump `ai_gateway_version_override`
   or an env var) on a sandbox with `create_ai_gateway = true`.
3. **Broken provider (main):** apply fails quickly with
   `Service deployment rolled back by user`.
4. **This branch:** apply waits for steady state and succeeds.
5. Optional: with override still on, `kill -INT <terraform apply pid>` mid-wait
   should still roll back (real SIGINT path).

Example sandbox dir:
`terraform-aws-braintrust-data-plane/examples/braintrust-data-plane-sandbox-private-gateway.deployed`
