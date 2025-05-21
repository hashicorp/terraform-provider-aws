{{ define "acctest.ConfigVPCWithSubnets" -}}
# acctest.ConfigVPCWithSubnets(rName, {{ . }})

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
{{- template "region" }}
  count = local.subnet_count

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
{{- template "region" }}
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  subnet_count             = {{ . }}
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}
{{- end }}
