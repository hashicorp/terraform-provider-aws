package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCEndpointServiceAllowedPrincipal_basic(t *testing.T) {
	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointServiceAllowedPrincipalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_service_allowed_principal" {
			continue
		}

		err := tfec2.FindVPCEndpointServicePermissionExists(conn, rs.Primary.Attributes["vpc_endpoint_service_id"], rs.Primary.Attributes["principal_arn"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 VPC Endpoint Service Allowed Principal %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVPCEndpointServiceAllowedPrincipalExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC Endpoint Service Allowed Principal ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		return tfec2.FindVPCEndpointServicePermissionExists(conn, rs.Primary.Attributes["vpc_endpoint_service_id"], rs.Primary.Attributes["principal_arn"])
	}
}

func testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, 1), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service_allowed_principal" "test" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.test.id

  principal_arn = data.aws_iam_session_context.current.issuer_arn
}
`, rName))
}
