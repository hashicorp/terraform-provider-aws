resource "aws_network_interface" "test" {
{{- template "region" }}
  subnet_id = aws_subnet.test[0].id
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
