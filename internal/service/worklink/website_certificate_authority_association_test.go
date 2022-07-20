package worklink_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfworklink "github.com/hashicorp/terraform-provider-aws/internal/service/worklink"
)

func TestAccWorkLinkWebsiteCertificateAuthorityAssociation_basic(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_website_certificate_authority_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, worklink.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebsiteCertificateAuthorityAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebsiteCertificateAuthorityAssociationConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebsiteCertificateAuthorityAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(
						resourceName, "fleet_arn",
						"aws_worklink_fleet.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "certificate", regexp.MustCompile("^-----BEGIN CERTIFICATE-----")),
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

func TestAccWorkLinkWebsiteCertificateAuthorityAssociation_displayName(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_website_certificate_authority_association.test"
	displayName1 := fmt.Sprintf("tf-website-certificate-%s", sdkacctest.RandStringFromCharSet(5, sdkacctest.CharSetAlpha))
	displayName2 := fmt.Sprintf("tf-website-certificate-%s", sdkacctest.RandStringFromCharSet(5, sdkacctest.CharSetAlpha))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, worklink.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebsiteCertificateAuthorityAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebsiteCertificateAuthorityAssociationConfig_displayName(suffix, displayName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebsiteCertificateAuthorityAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayName1),
				),
			},
			{
				Config: testAccWebsiteCertificateAuthorityAssociationConfig_displayName(suffix, displayName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebsiteCertificateAuthorityAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayName2),
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

func TestAccWorkLinkWebsiteCertificateAuthorityAssociation_disappears(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_website_certificate_authority_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, worklink.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebsiteCertificateAuthorityAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebsiteCertificateAuthorityAssociationConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebsiteCertificateAuthorityAssociationExists(resourceName),
					testAccCheckWebsiteCertificateAuthorityAssociationDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWebsiteCertificateAuthorityAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_worklink_website_certificate_authority_association" {
			continue
		}

		_, err := conn.DescribeWebsiteCertificateAuthority(&worklink.DescribeWebsiteCertificateAuthorityInput{
			FleetArn:    aws.String(rs.Primary.Attributes["fleet_arn"]),
			WebsiteCaId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, worklink.ErrCodeResourceNotFoundException) {
				return nil
			}

			return err
		}
		return fmt.Errorf("Worklink Website Certificate Authority Association(%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckWebsiteCertificateAuthorityAssociationDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn
		fleetArn, websiteCaID, err := tfworklink.DecodeWebsiteCertificateAuthorityAssociationResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &worklink.DisassociateWebsiteCertificateAuthorityInput{
			FleetArn:    aws.String(fleetArn),
			WebsiteCaId: aws.String(websiteCaID),
		}

		if _, err := conn.DisassociateWebsiteCertificateAuthority(input); err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"DELETING"},
			Target:     []string{"DELETED"},
			Refresh:    tfworklink.WebsiteCertificateAuthorityAssociationStateRefresh(conn, websiteCaID, fleetArn),
			Timeout:    15 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()

		return err
	}

}

func testAccCheckWebsiteCertificateAuthorityAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Worklink Website Certificate Authority Association ID is set")
		}

		if _, ok := rs.Primary.Attributes["fleet_arn"]; !ok {
			return fmt.Errorf("WorkLink Fleet ARN is missing, should be set.")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn
		fleetArn, websiteCaID, err := tfworklink.DecodeWebsiteCertificateAuthorityAssociationResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeWebsiteCertificateAuthority(&worklink.DescribeWebsiteCertificateAuthorityInput{
			FleetArn:    aws.String(fleetArn),
			WebsiteCaId: aws.String(websiteCaID),
		})

		return err
	}
}

func testAccWebsiteCertificateAuthorityAssociationConfig_basic(r string) string {
	return acctest.ConfigCompose(
		testAccFleetConfig_basic(r), `
resource "aws_worklink_website_certificate_authority_association" "test" {
  fleet_arn   = aws_worklink_fleet.test.arn
  certificate = file("test-fixtures/worklink-website-certificate-authority-association.pem")
}
`)
}

func testAccWebsiteCertificateAuthorityAssociationConfig_displayName(r, displayName string) string {
	return acctest.ConfigCompose(
		testAccFleetConfig_basic(r),
		fmt.Sprintf(`
resource "aws_worklink_website_certificate_authority_association" "test" {
  fleet_arn    = aws_worklink_fleet.test.arn
  certificate  = file("test-fixtures/worklink-website-certificate-authority-association.pem")
  display_name = "%s"
}
`, displayName))
}
