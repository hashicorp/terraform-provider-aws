package lightsail_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLightsailLoadBalancerCertificateAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	lbName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cName := sdkacctest.RandomWithPrefix("tf-acc-test")
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
				Config:      testAccLoadBalancerCertificateAttachmentConfig_basic(lbName, cName, domainName),
				ExpectError: regexp.MustCompile(`Sorry, you can only attach a validated certificate to a load balancer.`),
			},
		},
	})
}

func testAccLoadBalancerCertificateAttachmentConfig_basic(lbName string, cName string, domainName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_lb" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
}
resource "aws_lightsail_lb_certificate" "test" {
  name        = %[2]q
  lb_name     = aws_lightsail_lb.test.id
  domain_name = %[3]q
}
resource "aws_lightsail_lb_certificate_attachment" "test" {
  lb_name          = aws_lightsail_lb.test.name
  certificate_name = aws_lightsail_lb_certificate.test.name
}
`, lbName, cName, domainName)
}
