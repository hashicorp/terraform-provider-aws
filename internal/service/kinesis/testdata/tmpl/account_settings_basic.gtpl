resource "aws_kinesis_account_settings" "test" {
{{- template "region" }}
  minimum_throughput_billing_commitment {
    status = "DISABLED"
  }
}
