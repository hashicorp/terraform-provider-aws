resource "aws_sagemaker_servicecatalog_portfolio_status" "test" {
{{- template "region" }}
  status = "Enabled"
}
