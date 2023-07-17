plugin "aws" {
  enabled = true
  version = "0.23.1"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

rule "aws_acm_certificate_lifecycle" {
  enabled = false
}
