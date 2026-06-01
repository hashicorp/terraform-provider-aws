resource "aws_kinesis_stream" "test" {
{{- template "region" }}
  minimum_throughput_billing_commitment {
    status = "DISABLED"
  }
}
