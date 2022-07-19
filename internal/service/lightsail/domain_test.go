package lightsail_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
)

func TestAccLightsailDomain_basic(t *testing.T) {
	var domain lightsail.Domain
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))
	resourceName := "aws_lightsail_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckDomain(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(lightsailDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
		},
	})
}

func TestAccLightsailDomain_disappears(t *testing.T) {
	var domain lightsail.Domain
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))
	resourceName := "aws_lightsail_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckDomain(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(lightsailDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					acctest.CheckResourceDisappears(testAccProviderLightsailDomain, tflightsail.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainExists(n string, domain *lightsail.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain ID is set")
		}

		conn := testAccProviderLightsailDomain.Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetDomain(&lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.Domain == nil {
			return fmt.Errorf("Domain (%s) not found", rs.Primary.ID)
		}
		*domain = *resp.Domain
		return nil
	}
}

func testAccCheckDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_domain" {
			continue
		}

		conn := testAccProviderLightsailDomain.Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetDomain(&lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
			continue
		}

		if err == nil {
			if resp.Domain != nil {
				return fmt.Errorf("Lightsail Domain %q still exists", rs.Primary.ID)
			}
		}

		return err
	}

	return nil
}

func testAccDomainConfig_basic(lightsailDomainName string) string {
	return acctest.ConfigCompose(
		testAccDomainRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = "%s"
}
`, lightsailDomainName))
}
