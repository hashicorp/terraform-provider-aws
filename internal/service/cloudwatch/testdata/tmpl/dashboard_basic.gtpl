resource "aws_cloudwatch_dashboard" "test" {
{{- template "region" }}
  dashboard_name = var.rName

  dashboard_body = <<EOF
{
  "widgets": [
    {
      "type": "text",
      "x": 0,
      "y": 0,
      "width": 6,
      "height": 6,
      "properties": {
        "markdown": "Hi there from Terraform: CloudWatch"
      }
    }
  ]
}
EOF
}