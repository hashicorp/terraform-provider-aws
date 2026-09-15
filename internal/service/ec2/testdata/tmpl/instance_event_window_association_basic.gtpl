resource "aws_ec2_instance_event_window" "test" {
{{- template "region" }}
  name = var.rName

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
  }
}

resource "aws_ec2_instance_event_window_association" "test" {
{{- template "region" }}
  instance_event_window_id = aws_ec2_instance_event_window.test.id

  association_target {
    instance_ids = [aws_instance.test.id]
  }
}

{{ template "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" }}

data "aws_ec2_instance_type_offering" "available" {
{{- template "region" }}
  filter {
    name   = "instance-type"
    values = ["t3.micro", "t2.micro", "t1.micro", "m1.small"]
  }

  preferred_instance_types = ["t3.micro", "t2.micro", "t1.micro", "m1.small"]
}

resource "aws_instance" "test" {
{{- template "region" }}
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}
