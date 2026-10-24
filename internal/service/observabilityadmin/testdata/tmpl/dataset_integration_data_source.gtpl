data "aws_observabilityadmin_dataset_integration" "test" {
{{- template "region" }}
  depends_on = [aws_observabilityadmin_dataset_integration.test]
}
