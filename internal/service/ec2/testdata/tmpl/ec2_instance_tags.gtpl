resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = "t4g.nano"

  metadata_options {
    http_tokens = "required"
  }

{{- template "tags" . }}
}

{{ template "acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI" }}
