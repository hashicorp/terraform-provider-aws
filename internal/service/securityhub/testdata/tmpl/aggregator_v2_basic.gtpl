resource "aws_securityhub_aggregator_v2" "test" {
{{- template "region" }}
  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = ["ap-southeast-1"]

{{- template "tags" . }}

  depends_on = [aws_securityhub_account_v2.test]
}

resource "aws_securityhub_account_v2" "test" {
{{- template "region" }}
}