---
subcategory: "ACM PCA (Certificate Manager Private Certificate Authority)"
layout: "aws"
page_title: "AWS: aws_acmpca_policy"
description: |-
  Attaches a resource based policy to an AWS Certificate Manager Private Certificate Authority (ACM PCA)
---

# Resource: aws_acmpca_policy

Attaches a resource based policy to a private CA.

## Example Usage

### Basic

```terraform
resource "aws_acmpca_policy" "example" {
  resource_arn = aws_acmpca_certificate_authority.example.arn
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
         "Resource":"${aws_acmpca_certificate_authority.example.arn}"
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
         "Resource":"${aws_acmpca_certificate_authority.example.arn}",
         "Condition":{
            "StringEquals":{
               "acm-pca:TemplateArn":"arn:aws:acm-pca:::template/EndEntityCertificate/V1"
            }
         }
      }
   ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `resource_arn` - (Required) Amazon Resource Name (ARN) of the private CA to associate with the policy.
* `policy` - (Required) JSON-formatted IAM policy to attach to the specified private CA resource.

## Attributes Reference

No additional attributes are exported.

## Import

`aws_acmpca_policy` can be imported using the `resource_arn` value.

$ terraform import aws_acmpca_policy.example arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012
