package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccAWSCodeArtifactDomain_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactDomainBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactDomainExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "codeartifact", fmt.Sprintf("domain/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestCheckResourceAttr(resourceName, "asset_size_bytes", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository_count", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					testAccCheckResourceAttrAccountID(resourceName, "owner"),
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

func TestAccAWSCodeArtifactDomain_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactDomainBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactDomainExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCodeArtifactDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCodeArtifactDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no CodeArtifact domain set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codeartifactconn

		_, err := conn.DescribeDomain(&codeartifact.DescribeDomainInput{
			Domain: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckAWSCodeArtifactDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codeartifact_domain" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).codeartifactconn
		resp, err := conn.DescribeDomain(&codeartifact.DescribeDomainInput{
			Domain: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if aws.StringValue(resp.Domain.Name) == rs.Primary.ID {
				return fmt.Errorf("CodeArtifact Domain %s still exists", rs.Primary.ID)
			}
		}

		if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSCodeArtifactDomainBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = "${aws_kms_key.test.arn}"
}
`, rName)
}
