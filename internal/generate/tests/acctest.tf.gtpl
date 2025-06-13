{{ define "acctest.ConfigVPCWithSubnets" -}}
# acctest.ConfigVPCWithSubnets(rName, {{ . }})

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
{{- template "region" }}
  count = {{ . }}

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
{{- end }}

{{ define "acctest.ConfigAvailableAZsNoOptInDefaultExclude" -}}
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
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}
{{- end }}

{{ define "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" -}}
# acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI

{{ template "acctest.configLatestAmazonLinux2HVMEBSAMI" "x86_64" }}
{{- end }}

{{ define "acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI" -}}
# acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI

{{ template "acctest.configLatestAmazonLinux2HVMEBSAMI" "arm64" }}
{{- end }}

{{ define "acctest.configLatestAmazonLinux2HVMEBSAMI" -}}
# acctest.configLatestAmazonLinux2HVMEBSAMI("{{ . }}")

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-{{ . }}" {
{{- template "region" }}
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "architecture"
    values = ["{{ . }}"]
  }
}
{{- end }}

{{ define "acctest.ConfigAlternateAccountProvider" -}}
provider "awsalternate" {
  access_key = var.AWS_ALTERNATE_ACCESS_KEY_ID
  profile    = var.AWS_ALTERNATE_PROFILE
  secret_key = var.AWS_ALTERNATE_SECRET_ACCESS_KEY
}

variable "AWS_ALTERNATE_ACCESS_KEY_ID" {
  type     = string
  nullable = true
  default  = null
}

variable "AWS_ALTERNATE_PROFILE" {
  type     = string
  nullable = true
  default  = null
}

variable "AWS_ALTERNATE_SECRET_ACCESS_KEY" {
  type     = string
  nullable = true
  default  = null
}
{{- end }}
