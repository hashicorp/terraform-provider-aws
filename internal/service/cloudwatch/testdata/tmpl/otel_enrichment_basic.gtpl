resource "aws_observabilityadmin_telemetry_enrichment" "test" {
}

resource "aws_cloudwatch_otel_enrichment" "test" {
{{- template "region" }}

  depends_on = [aws_observabilityadmin_telemetry_enrichment.test]
}