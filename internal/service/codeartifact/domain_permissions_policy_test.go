package codeartifact_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodeartifact "github.com/hashicorp/terraform-provider-aws/internal/service/codeartifact"
)

func TestAccAWSCodeArtifactDomainPermissionsPolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain_permissions_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCodeArtifactDomainPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactDomainPermissionsPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactDomainPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codeartifact_domain.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:CreateRepository")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeArtifactDomainPermissionsPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactDomainPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codeartifact_domain.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:CreateRepository")),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:ListRepositoriesInDomain")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
				),
			},
		},
	})
}

func TestAccAWSCodeArtifactDomainPermissionsPolicy_owner(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain_permissions_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCodeArtifactDomainPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactDomainPermissionsPolicyOwnerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactDomainPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codeartifact_domain.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:CreateRepository")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
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

func TestAccAWSCodeArtifactDomainPermissionsPolicy_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain_permissions_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCodeArtifactDomainPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactDomainPermissionsPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactDomainPermissionsExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodeartifact.ResourceDomainPermissionsPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCodeArtifactDomainPermissionsPolicy_disappears_domain(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain_permissions_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCodeArtifactDomainPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactDomainPermissionsPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactDomainPermissionsExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodeartifact.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCodeArtifactDomainPermissionsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no CodeArtifact domain set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactConn

		domainOwner, domainName, err := tfcodeartifact.DecodeDomainID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetDomainPermissionsPolicy(&codeartifact.GetDomainPermissionsPolicyInput{
			Domain:      aws.String(domainName),
			DomainOwner: aws.String(domainOwner),
		})

		return err
	}
}

func testAccCheckAWSCodeArtifactDomainPermissionsDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codeartifact_domain_permissions_policy" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactConn

		domainOwner, domainName, err := tfcodeartifact.DecodeDomainID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.GetDomainPermissionsPolicy(&codeartifact.GetDomainPermissionsPolicyInput{
			Domain:      aws.String(domainName),
			DomainOwner: aws.String(domainOwner),
		})

		if err == nil {
			if aws.StringValue(resp.Policy.ResourceArn) == rs.Primary.ID {
				return fmt.Errorf("CodeArtifact Domain %s still exists", rs.Primary.ID)
			}
		}

		if tfawserr.ErrMessageContains(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSCodeArtifactDomainPermissionsPolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_domain_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}

func testAccAWSCodeArtifactDomainPermissionsPolicyOwnerConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_domain_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  domain_owner    = aws_codeartifact_domain.test.owner
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}

func testAccAWSCodeArtifactDomainPermissionsPolicyUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_domain_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
 				"codeartifact:CreateRepository",
				"codeartifact:ListRepositoriesInDomain"
			],
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}
