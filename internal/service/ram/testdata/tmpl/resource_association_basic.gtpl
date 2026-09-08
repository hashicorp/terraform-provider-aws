resource "aws_ram_resource_association" "test" {
{{- template "region" }}
  resource_arn       = aws_ec2_managed_prefix_list.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_share" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_ec2_managed_prefix_list" "test" {
{{- template "region" }}
  name           = var.rName
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr        = "10.0.0.0/8"
    description = "Test entry"
  }
}