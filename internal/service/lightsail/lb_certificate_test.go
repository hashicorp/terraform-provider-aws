package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

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

func TestAccLightsailLoadBalancerCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var certificate lightsail.LoadBalancerTlsCertificate
	resourceName := "aws_lightsail_lb_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerCertificateConfig_basic(rName, lbName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerCertificateExists(ctx, resourceName, &certificate),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "lightsail", regexp.MustCompile(`LoadBalancerTlsCertificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domainName),
					// When using a .test domain, Domain Validation Records return a single FAILED entry
					resource.TestCheckResourceAttr(resourceName, "domain_validation_records.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccLightsailLoadBalancerCertificate_subjectAlternativeNames(t *testing.T) {
	ctx := acctest.Context(t)
	var certificate lightsail.LoadBalancerTlsCertificate
	resourceName := "aws_lightsail_lb_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
		CheckDestroy:             testAccCheckLoadBalancerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerCertificateConfig_subjectAlternativeNames(rName, lbName, domainName, subjectAlternativeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerCertificateExists(ctx, resourceName, &certificate),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", subjectAlternativeName),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domainName),
				),
			},
		},
	})
}

func TestAccLightsailLoadBalancerCertificate_domainValidationRecords(t *testing.T) {
	ctx := acctest.Context(t)
	var certificate lightsail.LoadBalancerTlsCertificate
	resourceName := "aws_lightsail_lb_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
		CheckDestroy:             testAccCheckLoadBalancerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerCertificateConfig_subjectAlternativeNames(rName, lbName, domainName, subjectAlternativeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerCertificateExists(ctx, resourceName, &certificate),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_records.*", map[string]string{
						"domain_name":          domainName,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_records.*", map[string]string{
						"domain_name":          subjectAlternativeName,
						"resource_record_type": "CNAME",
					}),
				),
			},
		},
	})
}

func TestAccLightsailLoadBalancerCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var certificate lightsail.LoadBalancerTlsCertificate
	resourceName := "aws_lightsail_lb_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerCertificateConfig_basic(rName, lbName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerCertificateExists(ctx, resourceName, &certificate),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceLoadBalancerCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLoadBalancerCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_lb_certificate" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

			_, err := tflightsail.FindLoadBalancerCertificateById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResLoadBalancerCertificate, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccCheckLoadBalancerCertificateExists(ctx context.Context, n string, certificate *lightsail.LoadBalancerTlsCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

		respCertificate, err := tflightsail.FindLoadBalancerCertificateById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if respCertificate == nil {
			return fmt.Errorf("Load Balancer Certificate %q does not exist", rs.Primary.ID)
		}

		*certificate = *respCertificate

		return nil
	}
}

func testAccLoadBalancerCertificateConfigBase(lbName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_lb" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
}
`, lbName)
}

func testAccLoadBalancerCertificateConfig_basic(rName string, lbName string, domainName string) string {
	return acctest.ConfigCompose(
		testAccLoadBalancerCertificateConfigBase(lbName),
		fmt.Sprintf(`
resource "aws_lightsail_lb_certificate" "test" {
  name        = %[1]q
  lb_name     = aws_lightsail_lb.test.id
  domain_name = %[2]q
}
`, rName, domainName))
}

func testAccLoadBalancerCertificateConfig_subjectAlternativeNames(rName string, lbName string, domainName string, san string) string {
	return acctest.ConfigCompose(
		testAccLoadBalancerCertificateConfigBase(lbName),
		fmt.Sprintf(`
resource "aws_lightsail_lb_certificate" "test" {
  name                      = %[1]q
  lb_name                   = aws_lightsail_lb.test.id
  domain_name               = %[2]q
  subject_alternative_names = [%[3]q]
}
`, rName, domainName, san))
}
