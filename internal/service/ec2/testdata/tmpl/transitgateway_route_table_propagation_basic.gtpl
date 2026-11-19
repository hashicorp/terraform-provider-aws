resource "aws_ec2_transit_gateway_route_table_propagation" "test" {
{{- template "region" }}
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}

resource "aws_ec2_transit_gateway" "test" {
{{- template "region" }}
  tags = {
    Name = var.rName
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
{{- template "region" }}
  subnet_ids         = [aws_subnet.test[0].id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

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

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
