resource "aws_cloudwatch_query_definition" "test" {
{{- template "region" }}
  name = var.rName

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}
