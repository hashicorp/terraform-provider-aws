resource "aws_ec2_secondary_network" "test" {
{{- template "region" }}
  ipv4_cidr_block = "10.0.0.0/16"
  network_type    = "rdma"
}
