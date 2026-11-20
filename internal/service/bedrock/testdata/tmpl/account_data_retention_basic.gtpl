resource "aws_bedrock_account_data_retention" "test" {
{{- template "region" }}
  mode = "default"
}
