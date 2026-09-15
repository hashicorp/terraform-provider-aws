resource "aws_volume_attachment" "test" {
{{- template "region" }}
  device_name = "/dev/sdh"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}

{{ template "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" }}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}

data "aws_ec2_instance_type_offering" "available" {
{{- template "region" }}
  filter {
    name   = "instance-type"
    values = ["t3.micro", "t2.micro"]
  }

  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type            = "availability-zone"
  preferred_instance_types = ["t3.micro", "t2.micro"]
}

resource "aws_instance" "test" {
{{- template "region" }}
  ami               = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_ebs_volume" "test" {
{{- template "region" }}
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
}
