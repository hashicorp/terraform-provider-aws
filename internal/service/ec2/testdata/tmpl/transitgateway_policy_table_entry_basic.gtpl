resource "aws_ec2_transit_gateway_policy_table_entry" "test" {
{{- template "region" }}
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
  policy_rule_number              = 100
  target_route_table_id           = aws_ec2_transit_gateway_route_table.test.id
}

resource "aws_ec2_transit_gateway" "test" {
{{- template "region" }}
  tags = {
    Name = var.rName
  }
}

resource "aws_ec2_transit_gateway_policy_table" "test" {
{{- template "region" }}
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_ec2_transit_gateway_route_table" "test" {
{{- template "region" }}
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = var.rName
  }
}
