package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var certificate lightsail.Certificate
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &certificate),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "lightsail", regexp.MustCompile(`Certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domainName),
					// When using a .test domain, Domain Validation Records are not returned
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccLightsailCertificate_subjectAlternativeNames(t *testing.T) {
	ctx := acctest.Context(t)
	var certificate lightsail.Certificate
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())
	subjectAlternativeName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(rName, domainName, subjectAlternativeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &certificate),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", subjectAlternativeName),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domainName),
				),
			},
		},
	})
}

func TestAccLightsailCertificate_DomainValidationOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var certificate lightsail.Certificate
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// Lightsail will only return Domain Validation Options when using a non-test domain.
	// We need to provide a non-test domain in order to test these values.
	domainName := fmt.Sprintf("%s.com", acctest.ResourcePrefix)
	subjectAlternativeName := fmt.Sprintf("%s.com", acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(rName, domainName, subjectAlternativeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &certificate),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domainName,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          subjectAlternativeName,
						"resource_record_type": "CNAME",
					}),
				),
			},
		},
	})
}

func TestAccLightsailCertificate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var certificate lightsail.Certificate
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_tags1(rName, domainName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &certificate),
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
				Config: testAccCertificateConfig_tags2(rName, domainName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &certificate),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCertificateConfig_tags1(rName, domainName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &certificate),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccLightsailCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var certificate lightsail.Certificate
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Certificate
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()
		_, err := conn.DeleteCertificateWithContext(ctx, &lightsail.DeleteCertificateInput{
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
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &certificate),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_certificate" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

			_, err := tflightsail.FindCertificateByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResCertificate, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccCheckCertificateExists(ctx context.Context, n string, certificate *lightsail.Certificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		respCertificate, err := tflightsail.FindCertificateByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if respCertificate == nil {
			return fmt.Errorf("Certificate %q does not exist", rs.Primary.ID)
		}

		*certificate = *respCertificate

		return nil
	}
}

func testAccCertificateConfig_basic(rName string, domainName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[2]q
}
`, rName, domainName)
}

func testAccCertificateConfig_subjectAlternativeNames(rName string, domainName string, san string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name                      = %[1]q
  domain_name               = %[2]q
  subject_alternative_names = [%[3]q]
}
`, rName, domainName, san)
}

func testAccCertificateConfig_tags1(resourceName string, domainName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[2]q
  tags = {
    %[3]q = %[4]q
  }
}
`, resourceName, domainName, tagKey1, tagValue1)
}

func testAccCertificateConfig_tags2(resourceName, domainName string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[2]q
  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, resourceName, domainName, tagKey1, tagValue1, tagKey2, tagValue2)
}
