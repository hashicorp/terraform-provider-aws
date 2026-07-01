# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# OPA-only configuration for embedded test linting.
# The main .tflint.hcl uses --only for standard rules, which is
# incompatible with OPA rules. This config runs OPA rules separately.

plugin "terraform" {
  enabled = false
}

plugin "opa" {
  enabled = true
  version = "0.10.0"
  source  = "github.com/terraform-linters/tflint-ruleset-opa"
}
