package acmpca_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfacmpca "github.com/hashicorp/terraform-provider-aws/internal/service/acmpca"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccACMPCAPolicy_Basic(t *testing.T) {
	resourceName := "aws_acmpca_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acmpca.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acmpca_policy" {
			continue
		}

		_, err := tfacmpca.FindPolicyByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("ACM PCA Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ACM PCA Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAConn

		_, err := tfacmpca.FindPolicyByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccPolicyConfig_basic() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_acmpca_certificate_authority" "test" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "example.com"
    }
  }

  permanent_deletion_time_in_days = 7
}

resource "aws_acmpca_policy" "test" {
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
`
}
