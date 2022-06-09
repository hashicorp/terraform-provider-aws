package lightsail_test

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/aws/smithy-go"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccCertificate_basic(t *testing.T) {
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rName,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccCertificate_SubjectAlternativeNames(t *testing.T) {
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_SubjectAlternativeNames(rName, "www.test.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", "www.test.com"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rName,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccCertificate_Tags(t *testing.T) {
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck: acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(resourceName),
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
				Config: testAccCertificateConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCertificateConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCertificate_disappears(t *testing.T) {
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Certificate
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.DeleteCertificate(&lightsail.DeleteCertificateInput{
			CertificateName: aws.String(rName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Certificate in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(resourceName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCertificateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_certificate" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		respCertificate, err := conn.GetCertificates(&lightsail.GetCertificatesInput{
			CertificateName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if len(respCertificate.Certificates) > 0 {
				return fmt.Errorf("Certificate %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if err != nil {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
			}
			return nil
		}
		return err
	}

	return nil
}

func testAccCheckCertificateExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		respCertificate, err := conn.GetCertificates(&lightsail.GetCertificatesInput{
			CertificateName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if respCertificate == nil || respCertificate.Certificates == nil {
			return fmt.Errorf("Certificate (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCertificateConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[1]q
}
`, rName)
}

func testAccCertificateConfig_SubjectAlternativeNames(rName string, san string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name                      = %[1]q
  domain_name               = %[1]q
  subject_alternative_names = [%[2]q]
}
`, rName, san)
}

func testAccCertificateConfigTags1(resourceName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, resourceName, tagKey1, tagValue1)
}

func testAccCertificateConfigTags2(resourceName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, resourceName, tagKey1, tagValue1, tagKey2, tagValue2)
}
