// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccLoadBalancerCertificateAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	lbName := acctest.RandomWithPrefix(t, "tf-acc-test")
	cName := acctest.RandomWithPrefix(t, "tf-acc-test")
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccLoadBalancerCertificateAttachmentConfig_basic(lbName, cName, domainName),
				ExpectError: regexache.MustCompile(`Sorry, you can only attach a validated certificate to a load balancer.`),
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
