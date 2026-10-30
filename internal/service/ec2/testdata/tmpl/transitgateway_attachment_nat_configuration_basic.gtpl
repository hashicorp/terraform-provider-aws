resource "aws_ec2_transit_gateway_attachment_nat_configuration" "test" {
{{- template "region" }}
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
  nat_gateway_support           = "enable"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}

resource "aws_ec2_transit_gateway" "test" {
{{- template "region" }}
  tags = {
    Name = var.rName
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
{{- template "region" }}
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
