package redshift_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftHsmClientCertificate_basic(t *testing.T) {
	resourceName := "aws_redshift_hsm_client_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHsmClientCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHsmClientCertificate_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHsmClientCertificateExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift", regexp.MustCompile(`hsmclientcertificate:.+`)),
					resource.TestCheckResourceAttr(resourceName, "hsm_client_certificate_identifier", rName),
					resource.TestCheckResourceAttrSet(resourceName, "hsm_client_certificate_public_key"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccRedshiftHsmClientCertificate_tags(t *testing.T) {
	resourceName := "aws_redshift_hsm_client_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHsmClientCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHsmClientCertificateConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHsmClientCertificateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHsmClientCertificateConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHsmClientCertificateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			}, {
				Config: testAccHsmClientCertificateConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHsmClientCertificateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRedshiftHsmClientCertificate_disappears(t *testing.T) {
	resourceName := "aws_redshift_hsm_client_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHsmClientCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHsmClientCertificate_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHsmClientCertificateExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceHsmClientCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHsmClientCertificateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_hsm_client_certificate" {
			continue
		}

		_, err := tfredshift.FindHsmClientCertificateByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Hsm Client Certificate %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckHsmClientCertificateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot Copy Grant ID (HsmClientCertificateName) is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		_, err := tfredshift.FindHsmClientCertificateByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccHsmClientCertificate_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_hsm_client_certificate" "test" {
  hsm_client_certificate_identifier = %[1]q
}
`, rName)
}

func testAccHsmClientCertificateConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_hsm_client_certificate" "test" {
  hsm_client_certificate_identifier = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccHsmClientCertificateConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_hsm_client_certificate" "test" {
  hsm_client_certificate_identifier = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
