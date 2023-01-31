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

func TestAccLightsailLoadBalancerHTTPSRedirectionPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	enabled := "true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccLoadBalancerHTTPSRedirectionPolicyConfig_basic(rName, enabled),
				ExpectError: regexp.MustCompile(`cannot enable https redirection while https is disabled.`),
			},
		},
	})
}

func testAccLoadBalancerHTTPSRedirectionPolicyConfig_basic(rName string, enabled string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_lb" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
}
resource "aws_lightsail_lb_https_redirection_policy" "test" {
  enabled = %[2]s
  lb_name = aws_lightsail_lb.test.name
}
`, rName, enabled)
}
