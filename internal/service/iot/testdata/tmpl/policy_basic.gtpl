resource "aws_iot_policy" "test" {
{{- template "region" }}
  name = var.rName

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["iot:*"]
      Resource = ["*"]
    }]
  })

{{- template "tags" . }}
}
