# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_observabilityadmin_telemetry_enrichment" "test" {
}

resource "aws_cloudwatch_otel_enrichment" "test" {

  depends_on = [aws_observabilityadmin_telemetry_enrichment.test]
}
