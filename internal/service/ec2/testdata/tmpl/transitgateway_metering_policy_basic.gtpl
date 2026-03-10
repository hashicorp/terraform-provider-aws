resource "aws_ec2_transit_gateway_metering_policy" "test" {
{{- template "region" }}
  transit_gateway_id = aws_ec2_transit_gateway.test.id

{{- template "tags" . }}
}

resource "aws_ec2_transit_gateway" "test" {}