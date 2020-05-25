#!/bin/bash
make fmt
make build
terraform init
cp /Users/fcandela/Go/bin/terraform-provider-aws /Users/fcandela/Go/src/github.com/frc9/terraform-provider-aws/.terraform/plugins/darwin_amd64/terraform-provider-aws_v2.63.0_x4
terraform init
cp /Users/fcandela/Go/bin/terraform-provider-aws /Users/fcandela/Go/src/github.com/frc9/terraform-provider-aws/.terraform/plugins/darwin_amd64/terraform-provider-aws_v2.63.0_x4
TF_LOG=DEBUG terraform apply -auto-approve
