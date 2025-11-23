resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = var.rName
  }
}

resource "aws_vpn_concentrator" "test" {
  type               = "ipsec.1"
  transit_gateway_id = aws_ec2_transit_gateway.test.id

{{- template "tags" . }}
}
