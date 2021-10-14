package macie2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
)

func testAccAccount_basic(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMacieAccountBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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

func testAccAccount_FindingPublishingFrequency(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMacieAccountWithFindingConfig(macie2.FindingPublishingFrequencyFifteenMinutes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccMacieAccountWithFindingConfig(macie2.FindingPublishingFrequencyOneHour),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyOneHour),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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

func testAccAccount_WithStatus(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMacieAccountWithstatusConfig(macie2.MacieStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccMacieAccountWithstatusConfig(macie2.MacieStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusPaused),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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

func testAccAccount_WithFindingAndStatus(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMacieAccountWithfindingandstatusConfig(macie2.FindingPublishingFrequencyFifteenMinutes, macie2.MacieStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccMacieAccountWithfindingandstatusConfig(macie2.FindingPublishingFrequencyOneHour, macie2.MacieStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyOneHour),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusPaused),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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

func testAccAccount_disappears(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMacieAccountBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &macie2Output),
					acctest.CheckResourceDisappears(acctest.Provider, tfmacie2.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_account" {
			continue
		}

		input := &macie2.GetMacieSessionInput{}
		resp, err := conn.GetMacieSession(input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("macie account %q still enabled", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAccountExists(resourceName string, macie2Session *macie2.GetMacieSessionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn
		input := &macie2.GetMacieSessionInput{}

		resp, err := conn.GetMacieSession(input)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("macie account %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testAccMacieAccountBasicConfig() string {
	return `
resource "aws_macie2_account" "test" {}
`
}

func testAccMacieAccountWithFindingConfig(finding string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "%s"
}
`, finding)
}

func testAccMacieAccountWithstatusConfig(status string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  status = "%s"
}
`, status)
}

func testAccMacieAccountWithfindingandstatusConfig(finding, status string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "%s"
  status                       = "%s"
}
`, finding, status)
}
