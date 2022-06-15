package amplify_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/amplify"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccDomainAssociation_basic(t *testing.T) {
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain amplify.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_basic(rName, domainName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainAssociationExists(resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+/domains/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name": rName,
						"prefix":      "",
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", "false"),
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

func testAccDomainAssociation_disappears(t *testing.T) {
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain amplify.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_basic(rName, domainName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainAssociationExists(resourceName, &domain),
					acctest.CheckResourceDisappears(acctest.Provider, tfamplify.ResourceDomainAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomainAssociation_update(t *testing.T) {
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain amplify.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_basic(rName, domainName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainAssociationExists(resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+/domains/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name": rName,
						"prefix":      "",
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
			{
				Config: testAccDomainAssociationConfig_updated(rName, domainName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainAssociationExists(resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+/domains/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name": rName,
						"prefix":      "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name": fmt.Sprintf("%s-2", rName),
						"prefix":      "www",
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", "true"),
				),
			},
		},
	})
}

func testAccCheckDomainAssociationExists(resourceName string, v *amplify.DomainAssociation) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

		domainAssociation, err := tfamplify.FindDomainAssociationByAppIDAndDomainName(conn, appID, domainName)

		if err != nil {
			return err
		}

		*v = *domainAssociation

		return nil
	}
}

func testAccCheckDomainAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_domain_association" {
			continue
		}

		appID, domainName, err := tfamplify.DomainAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfamplify.FindDomainAssociationByAppIDAndDomainName(conn, appID, domainName)

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

func testAccDomainAssociationConfig_basic(rName, domainName string, waitForVerification bool) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  wait_for_verification = %[3]t
}
`, rName, domainName, waitForVerification)
}

func testAccDomainAssociationConfig_updated(rName, domainName string, waitForVerification bool) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_branch" "test2" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-2"
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  sub_domain {
    branch_name = aws_amplify_branch.test2.branch_name
    prefix      = "www"
  }

  wait_for_verification = %[3]t
}
`, rName, domainName, waitForVerification)
}
