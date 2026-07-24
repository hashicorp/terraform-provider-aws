resource "aws_securityhub_product_subscription" "test" {
{{- template "region" }}
  depends_on  = [aws_securityhub_account.test]
  product_arn = "arn:${data.aws_partition.current.partition}:securityhub:${data.aws_region.current.region}::product/aws/guardduty"
}

data "aws_region" "current" {
{{- template "region" }}
}
data "aws_partition" "current" {}

resource "aws_securityhub_account" "test" {
{{- template "region" }}
}