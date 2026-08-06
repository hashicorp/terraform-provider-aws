resource "aws_vpn_gateway" "test" {
  {{- template "region" }}
  tags = {
    Name = var.rName
  }
}

resource "aws_customer_gateway" "test" {
  {{- template "region" }}
  bgp_asn    = 65000
  ip_address = "182.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = var.rName
  }
}

resource "aws_vpn_connection" "test" {
  {{- template "region" }}
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    Name = var.rName
  }
}

resource "aws_vpn_connection_route" "test" {
  {{- template "region" }}
  destination_cidr_block = "172.168.10.0/24"
  vpn_connection_id      = aws_vpn_connection.test.id
}
