# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_acmpca_policy" "test" {
  region = var.region

  resource_arn = aws_acmpca_certificate_authority.test.arn
  policy       = <<EOF
{                        
   "Version":"2012-10-17",
   "Statement":[
      {    
         "Sid":"1",
         "Effect":"Allow",         
         "Principal":{                                                                                                                                               
            "AWS":"${data.aws_caller_identity.current.account_id}"                                                                                
         },
         "Action":[
            "acm-pca:DescribeCertificateAuthority",
            "acm-pca:GetCertificate",
            "acm-pca:GetCertificateAuthorityCertificate",
            "acm-pca:ListPermissions",
            "acm-pca:ListTags"                                                                                   
         ],                                                                                              
         "Resource":"${aws_acmpca_certificate_authority.test.arn}"
      },
      {
         "Sid":"1",  
         "Effect":"Allow",
         "Principal":{
            "AWS":"${data.aws_caller_identity.current.account_id}"
         },
         "Action":[
            "acm-pca:IssueCertificate"
         ],
         "Resource":"${aws_acmpca_certificate_authority.test.arn}",
         "Condition":{
            "StringEquals":{
               "acm-pca:TemplateArn":"arn:${data.aws_partition.current.partition}:acm-pca:::template/EndEntityCertificate/V1"
            }
         }
      }
   ]
}
EOF
}

resource "aws_acmpca_certificate_authority" "test" {
  region = var.region

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "example.com"
    }
  }

  permanent_deletion_time_in_days = 7
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
