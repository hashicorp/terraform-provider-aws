# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

plugin "terraform" {
  enabled = false
}

plugin "opa" {
  enabled = true
  version = "0.10.0"
  source  = "github.com/terraform-linters/tflint-ruleset-opa"
}
