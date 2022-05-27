plugin "aws" {
  enabled = true
  version = "0.13.4"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

rule "aws_acm_certificate_lifecycle" {
  enabled = false
}

rule "aws_route_not_specified_target" {
  enabled = false
}
