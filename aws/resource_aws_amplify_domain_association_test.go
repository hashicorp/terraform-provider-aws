package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfamplify "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSAmplifyDomainAssociation_basic(t *testing.T) {
	var domain amplify.DomainAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_domain_association.test"

	domainName := "example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyDomainAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyDomainAssociationConfig_Required(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyDomainAssociationExists(resourceName, &domain),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile("^arn:[^:]+:amplify:[^:]+:[^:]+:apps/[^/]+/domains/[^/]+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "sub_domain_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sub_domain_setting.0.branch_name", "master"),
					resource.TestCheckResourceAttr(resourceName, "sub_domain_setting.0.prefix", ""),
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

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify Domain Association ID is set")
		}

		appID, domainName, err := tfamplify.DomainAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		domainAssociation, err := finder.DomainAssociationByAppIDAndDomainName(conn, appID, domainName)

		if err != nil {
			return err
		}

		*v = *domainAssociation

		return nil
	}
}

func testAccCheckAWSAmplifyDomainAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).amplifyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_domain_association" {
			continue
		}

		appID, domainName, err := tfamplify.DomainAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.DomainAssociationByAppIDAndDomainName(conn, appID, domainName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Amplify Domain Association %s still exists", rs.Primary.ID)
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

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  wait_for_verification = false
}
`, rName, domainName)
}
