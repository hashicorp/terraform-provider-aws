provider "aws" {
  version = "~> 2.0"
}

data "aws_ec2_instance_spot_price" "foo" {
  instance_type = "t3.medium"
  availability_zone = "us-west-2a"

  filter {
    name  = "product-description"
    values = ["Linux/UNIX"]
  }
}

output "foo" {
  value = data.aws_ec2_instance_spot_price.foo.spot_price
}

output "when" {
  value = data.aws_ec2_instance_spot_price.foo.spot_price_timestamp
}
