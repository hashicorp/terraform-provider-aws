package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAmplifyDomainAssociation_basic(t *testing.T) {
	var domain amplify.DomainAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_domain_association.test"

	domainName := "example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyDomainAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyDomainAssociationConfig_Required(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyDomainAssociationExists(resourceName, &domain),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile("^arn:[^:]+:amplify:[^:]+:[^:]+:apps/[^/]+/domains/[^/]+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "sub_domain_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sub_domain_settings.0.branch_name", "master"),
					resource.TestCheckResourceAttr(resourceName, "sub_domain_settings.0.prefix", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
		},
	})
}

func testAccCheckAWSAmplifyDomainAssociationExists(resourceName string, v *amplify.DomainAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		id := strings.Split(rs.Primary.ID, "/")
		app_id := id[0]
		domain_name := id[2]

		output, err := conn.GetDomainAssociation(&amplify.GetDomainAssociationInput{
			AppId:      aws.String(app_id),
			DomainName: aws.String(domain_name),
		})
		if err != nil {
			return err
		}

		if output == nil || output.DomainAssociation == nil {
			return fmt.Errorf("Amplify DomainAssociation (%s) not found", rs.Primary.ID)
		}

		*v = *output.DomainAssociation

		return nil
	}
}

func testAccCheckAWSAmplifyDomainAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_domain_association" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		s := strings.Split(rs.Primary.ID, "/")
		app_id := s[0]
		domain_name := s[2]

		_, err := conn.GetDomainAssociation(&amplify.GetDomainAssociationInput{
			AppId:      aws.String(app_id),
			DomainName: aws.String(domain_name),
		})

		if isAWSErr(err, amplify.ErrCodeNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccAWSAmplifyDomainAssociationConfig_Required(rName string, domainName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = "%s"

  sub_domain_settings {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  wait_for_verification = false
}
`, rName, domainName)
}
