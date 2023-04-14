package ec2_test

import (
	"context"
	"fmt"
	"regexp"
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
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^vpce-svc-perm-\w{17}$`)),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_service_id", "aws_vpc_endpoint_service.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_arn", "data.aws_iam_session_context.current", "issuer_arn"),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceAllowedPrincipal_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tfacctest")

	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"
	tagResourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceAllowedPrincipalConfig_tag(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(tagResourceName, "resource_id", resourceName, "id"),
					resource.TestCheckResourceAttr(tagResourceName, "key", "Name"),
					resource.TestCheckResourceAttr(tagResourceName, "value", rName),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceAllowedPrincipal_migrateID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		CheckDestroy: testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.63.0",
					},
				},
				Config: testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_endpoint_service_allowed_principal" {
				continue
			}

			err := tfec2.FindVPCEndpointServicePermissionExists(ctx, conn, rs.Primary.Attributes["vpc_endpoint_service_id"], rs.Primary.Attributes["principal_arn"])

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
}

func testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC Endpoint Service Allowed Principal ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

		return tfec2.FindVPCEndpointServicePermissionExists(ctx, conn, rs.Primary.Attributes["vpc_endpoint_service_id"], rs.Primary.Attributes["principal_arn"])
	}
}

func testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, 1), `
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
}

resource "aws_vpc_endpoint_service_allowed_principal" "test" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.test.id

  principal_arn = data.aws_iam_session_context.current.issuer_arn
}
`)
}

func testAccVPCEndpointServiceAllowedPrincipalConfig_tag(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_networkLoadBalancerBase(rName, 1), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
}

resource "aws_vpc_endpoint_service_allowed_principal" "test" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.test.id

  principal_arn = data.aws_iam_session_context.current.issuer_arn
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_vpc_endpoint_service_allowed_principal.test.id

  key   = "Name"
  value = %[1]q
}
`, rName))
}
