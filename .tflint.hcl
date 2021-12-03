plugin "aws" {
  enabled = true
  version = "0.7.1"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

# https://github.com/terraform-linters/tflint-ruleset-aws/issues/206
rule "aws_transfer_server_invalid_identity_provider_type" {
  enabled = false
}
