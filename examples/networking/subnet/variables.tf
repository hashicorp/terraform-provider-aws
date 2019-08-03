variable "vpc_id" {}

variable "availability_zone" {}

data "aws_availability_zone" "target" {
  name = "${var.availability_zone}"
}

data "aws_vpc" "target" {
  id = "${var.vpc_id}"
}

variable "az_numbers" {
  default = {
    a = 1
    b = 2
    c = 3
    d = 4
    e = 5
    f = 6
    g = 7
    h = 8
    i = 9
    j = 10
    k = 11
    l = 12
    m = 13
    n = 14
  }
}
