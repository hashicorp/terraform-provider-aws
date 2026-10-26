resource "aws_ec2_transit_gateway_route" "test" {
{{- template "region" }}
  blackhole                      = true
  destination_cidr_block         = "10.1.0.0/16"
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}

resource "aws_ec2_transit_gateway" "test" {
{{- template "region" }}
}
