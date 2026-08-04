resource "aws_ec2_secondary_subnet" "test" {
{{- template "region" }}
  secondary_network_id = aws_ec2_secondary_network.test.id
  ipv4_cidr_block      = "10.0.0.0/24"
  availability_zone    = data.aws_availability_zones.available.names[0]
}

resource "aws_ec2_secondary_network" "test" {
{{- template "region" }}
  ipv4_cidr_block = "10.0.0.0/16"
  network_type    = "rdma"
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
